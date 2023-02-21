package model

import (
	"context"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/monc"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var _ MessageModel = (*customMessageModel)(nil)

type (
	// MessageModel is an interface to be customized, add more methods here,
	// and implement the added methods in customMessageModel.
	MessageModel interface {
		messageModel
		// 添加自定义的业务逻辑方法接口
		businessMessageModelInterface
	}

	customMessageModel struct {
		*defaultMessageModel
		// 添加自定义的业务逻辑方法实例类型
		*businessMessageModel
	}
)

type (
	businessMessageModel struct {
		conn *monc.Model
	}
)

const (
	cachePrefix      = "messageListUserId:"
	inboxCachePrefix = "inboxList:"
)

// 自定义业务逻辑方法接口
type businessMessageModelInterface interface {
	GetMessageListByUserId(ctx context.Context, userId string, toUserId string) (*Message, error)
	PutMessageIntoUserMessageList(ctx context.Context, fromUserId string, toUserId string, content string) error
}

// 拿到 mongo 的连接返回一个自定义业务逻辑的 model 实例
func newBusinessMessageModel(conn *monc.Model) *businessMessageModel {
	return &businessMessageModel{conn: conn}
}

// 自定义业务逻辑方法实现

// 通过用户 id 查找消息列表
// @param userId 用户 id
// @param toUserId 对方（发消息的用户）的 id
func (m *businessMessageModel) GetMessageListByUserId(ctx context.Context, userId string, toUserId string) (*Message, error) {

	var err error = nil

	var data Message

	// find the message from cache
	fromCacheKey := prefixMessageCacheKey + inboxCachePrefix + userId + ":" + toUserId

	// 从缓存中查找
	err = m.conn.GetCache(fromCacheKey, &data)
	if err != nil && err != monc.ErrNotFound {
		return nil, err
	}

	if data.ID != primitive.NilObjectID {
		return &data, nil
	}

	var aggregatedData []Message

	// 缓存中找不到，从数据库中查找
	err = m.conn.Aggregate(ctx, &aggregatedData, bson.A{
		bson.D{{"$match", bson.D{{"inboxUserId", userId}}}},
		bson.D{
			{"$project",
				bson.D{
					{"_id", true},
					{"inboxUserId", true},
					{"createAt", true},
					{"updateAt", true},
					{"messageList",
						bson.D{
							{"$filter",
								bson.D{
									{"input", "$messageList"},
									{"as", "item"},
									{"cond",
										bson.D{
											{"$eq",
												bson.A{
													"$$item.fromUserId",
													toUserId,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})

	switch err {
	case nil:
		if len(aggregatedData) == 0 {
			return &Message{}, nil
		}
		// 异步更新缓存
		go func(m *businessMessageModel, fromCacheKey string, aggregatedData []Message) {
			updateCacheErr := m.conn.SetCache(fromCacheKey, aggregatedData[0])
			if updateCacheErr != nil {
				// log
				_ = errors.New("update cache error")
			}
		}(m, fromCacheKey, aggregatedData)

		return &aggregatedData[0], nil
	case monc.ErrNotFound:
		// 查不到当前用户名，返回空
		return &Message{}, nil
	default:
		return nil, err
	}
}

// 将消息放入用户的消息列表中
func (m *businessMessageModel) PutMessageIntoUserMessageList(ctx context.Context, fromUserId string, toUserId string, content string) error {
	now := time.Now()

	cacheKey := prefixMessageCacheKey + cachePrefix + toUserId

	fromUserCacheKey := prefixMessageCacheKey + inboxCachePrefix + toUserId + ":" + fromUserId

	var data Message
	// find the message
	// filter := bson.M{"InboxUserId": userId}
	findOneErr := m.conn.FindOne(ctx, cacheKey, &data, bson.M{"inboxUserId": toUserId})

	if findOneErr != nil && findOneErr != monc.ErrNotFound {
		return findOneErr
	}

	newData := Message{
		ID:          primitive.NewObjectID(),
		CreateAt:    now,
		UpdateAt:    now,
		InboxUserId: toUserId,
		MessageList: []MessageItem{
			{
				Content:    content,
				CreateAt:   now,
				UpdateAt:   now,
				FromUserId: fromUserId,
				ToUserId:   toUserId,
			},
		},
	}

	if findOneErr == monc.ErrNotFound {
		// insert
		_, insertErr := m.conn.InsertOne(ctx, cacheKey, newData)
		if insertErr != nil {
			return insertErr
		}

		// 异步更新缓存
		go func(m *businessMessageModel, fromUserCacheKey string, newData Message) {
			cacheData := Message{}

			formUserCacheErr := m.conn.GetCache(fromUserCacheKey, &cacheData)

			if formUserCacheErr != nil && formUserCacheErr != monc.ErrNotFound {
				// log
				_ = errors.New("get cache error")
			}

			if formUserCacheErr == monc.ErrNotFound {
				updateCacheErr := m.conn.SetCache(fromUserCacheKey, newData)
				if updateCacheErr != nil {
					// log
					_ = errors.New("update cache error")
				}
			}

			if cacheData.ID != primitive.NilObjectID {
				cacheData.MessageList = append(cacheData.MessageList, newData.MessageList...)
				m.conn.SetCache(fromUserCacheKey, cacheData)
			}
		}(m, fromUserCacheKey, newData)

		return nil
	}

	// update
	data.UpdateAt = now
	data.MessageList = append(data.MessageList, MessageItem{
		ID:         primitive.NewObjectID(),
		Content:    content,
		CreateAt:   now,
		UpdateAt:   now,
		FromUserId: fromUserId,
		ToUserId:   toUserId,
	})

	updateRes, updateErr := m.conn.UpdateOne(ctx, cacheKey, bson.M{"inboxUserId": toUserId}, bson.M{"$set": data})

	if updateErr != nil {
		return updateErr
	}

	if updateRes.ModifiedCount == 0 {
		return errors.New("update failed")
	}

	cacheData := Message{}
	formUserCacheErr := m.conn.GetCache(fromUserCacheKey, &cacheData)
	if formUserCacheErr != nil && formUserCacheErr != monc.ErrNotFound {
		return formUserCacheErr
	}

	// 异步更新缓存
	go func(m *businessMessageModel, data Message, fromUserCacheKey string, cacheData Message) {

		filteredMessageList := []MessageItem{}

		for _, item := range data.MessageList {
			if item.FromUserId == fromUserId {
				filteredMessageList = append(filteredMessageList, item)
			}
		}

		if formUserCacheErr == monc.ErrNotFound {
			updateCacheErr := m.conn.SetCache(fromUserCacheKey, Message{
				ID:          data.ID,
				CreateAt:    data.CreateAt,
				UpdateAt:    data.UpdateAt,
				InboxUserId: data.InboxUserId,
				MessageList: filteredMessageList,
			})

			if updateCacheErr != nil {
				// log
				_ = errors.New("update cache error")
			}
		}

		if cacheData.ID != primitive.NilObjectID {
			m.conn.SetCache(fromUserCacheKey, Message{
				ID:          data.ID,
				CreateAt:    data.CreateAt,
				UpdateAt:    data.UpdateAt,
				InboxUserId: data.InboxUserId,
				MessageList: filteredMessageList,
			})
		}
	}(m, data, fromUserCacheKey, cacheData)

	return nil
}

// NewMessageModel returns a model for the mongo.
func NewMessageModel(url, db, collection string, c cache.CacheConf) MessageModel {
	conn := monc.MustNewModel(url, db, collection, c)
	return &customMessageModel{
		defaultMessageModel: newDefaultMessageModel(conn),
		// 将自定义的业务逻辑方法实例注入到自定义的 model 实例中
		businessMessageModel: newBusinessMessageModel(conn),
	}
}

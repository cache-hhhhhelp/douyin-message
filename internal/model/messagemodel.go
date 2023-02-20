package model

import (
	"context"
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
	cachePrefix = "messageListUserId:"
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
	cacheKey := prefixMessageCacheKey + cachePrefix + userId

	err = m.conn.FindOne(ctx, cacheKey, &data, bson.M{
		"inboxUserId": userId,
		// FIXME: 无法使用 $elemMatch 过滤数组中的元素，原因未知
		// "messageList": bson.M{
		// 	"$elemMatch": bson.M{
		// 		"fromUserId": toUserId,
		// 	},
		// },
	})

	switch err {
	case nil:
		// filter formUserId
		var messageList []MessageItem
		// FIXME: 无法使用 $elemMatch 过滤数组中的元素，原因未知，所以这里使用循环过滤
		for _, item := range data.MessageList {
			if item.FromUserId == toUserId {
				messageList = append(messageList, item)
			}
		}

		data.MessageList = messageList
		return &data, nil
	case monc.ErrNotFound:
		// 查不到当前用户名，返回空
		return nil, nil
	default:
		return nil, err
	}
}

// 将消息放入用户的消息列表中
func (m *businessMessageModel) PutMessageIntoUserMessageList(ctx context.Context, fromUserId string, toUserId string, content string) error {
	now := time.Now()
	cacheKey := prefixMessageCacheKey + cachePrefix + toUserId

	var data Message
	// find the message
	// filter := bson.M{"InboxUserId": userId}
	findOneErr := m.conn.FindOne(ctx, cacheKey, &data, bson.M{"inboxUserId": toUserId})

	if findOneErr != nil && findOneErr != monc.ErrNotFound {
		return findOneErr
	}

	if findOneErr == monc.ErrNotFound {
		// insert
		data = Message{
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

		_, updateErr := m.conn.InsertOne(ctx, cacheKey, data)
		if updateErr != nil {
			return updateErr
		}

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

	if updateRes.MatchedCount == 0 {
		return monc.ErrNotFound
	}

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

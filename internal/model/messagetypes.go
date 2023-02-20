package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserId string

type MessageList []MessageItem

type MessageItem struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UpdateAt time.Time          `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
	CreateAt time.Time          `bson:"createAt,omitempty" json:"createAt,omitempty"`
	// 消息发送方的ID
	FromUserId string `bson:"fromUserId,omitempty" json:"fromUserId,omitempty"`
	// 消息接收方的ID
	ToUserId string `bson:"toUserId,omitempty" json:"toUserId,omitempty"`
	// 消息内容
	Content string `bson:"content,omitempty" json:"content,omitempty"`
}

type Message struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UpdateAt time.Time          `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
	CreateAt time.Time          `bson:"createAt,omitempty" json:"createAt,omitempty"`
	// 消息盒子的ID
	InboxUserId string `bson:"inboxUserId,omitempty" json:"inboxUserId,omitempty"`
	// 消息盒子列表
	MessageList MessageList `bson:"messageList,omitempty" json:"messageList,omitempty"`
}

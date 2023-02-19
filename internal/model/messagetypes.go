package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UpdateAt time.Time `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
	CreateAt time.Time `bson:"createAt,omitempty" json:"createAt,omitempty"`

	/**
	* 消息接收者id
	 */
	ToUserId string `bson:"toUserId,omitempty" json:"toUserId,omitempty"`
	/**
	* 消息发送者 id
	 */
	FromUserId string `bson:"fromUserId,omitempty" json:"fromUserId,omitempty"`
	Content    string `bson:"content,omitempty" json:"content,omitempty"`
}

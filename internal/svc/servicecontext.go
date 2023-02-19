package svc

import (
	"douyin-message/internal/config"
	"douyin-message/internal/model"
)

type ServiceContext struct {
	Config       config.Config
	MessageModel model.MessageModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:       c,
		MessageModel: model.NewMessageModel(c.Mongodb.URL, c.Mongodb.Database, c.Mongodb.Collection, c.CacheRedis),
	}
}

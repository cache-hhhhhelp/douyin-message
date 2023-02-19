package logic

import (
	"context"

	"douyin-message/internal/svc"
	"douyin-message/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChatLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewChatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChatLogic {
	return &ChatLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ChatLogic) Chat(in *__.ChatRequest) (*__.ChatResponse, error) {
	// todo: add your logic here and delete this line

	return &__.ChatResponse{}, nil
}

package logic

import (
	"context"

	"douyin-message/internal/svc"
	__ "douyin-message/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ActionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewActionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ActionLogic {
	return &ActionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ActionLogic) Action(in *__.ActionRequest) (*__.ActionResponse, error) {
	// todo: add your logic here and delete this line

	return &__.ActionResponse{}, nil
}

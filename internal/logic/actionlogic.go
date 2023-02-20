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
	// 参数校验
	if in.ToUserId == "" {
		return &__.ActionResponse{
				Data: nil,
				BaseResp: &__.BaseResp{
					StatusCode:    __.StatusCode_UNKNOWN_ERROR,
					StatusMessage: "invalid param: ToUserId is empty",
				},
			}, &__.ErrorResp{
				StatusCode:    __.StatusCode_UNKNOWN_ERROR,
				StatusMessage: "invalid param: ToUserId is empty",
			}
	}
	// 将消息放入用户消息盒子
	err := l.svcCtx.MessageModel.PutMessageIntoUserMessageList(l.ctx, in.FromUserId, in.ToUserId, in.Content)

	if err != nil {
		return &__.ActionResponse{
				Data: nil,
				BaseResp: &__.BaseResp{
					StatusCode:    __.StatusCode_UNKNOWN_ERROR,
					StatusMessage: err.Error(),
				},
			}, &__.ErrorResp{
				StatusCode:    __.StatusCode_UNKNOWN_ERROR,
				StatusMessage: err.Error(),
			}
	}

	return &__.ActionResponse{
		BaseResp: &__.BaseResp{
			StatusCode:    __.StatusCode_OK,
			StatusMessage: "success",
		},
		Data: &__.ActionResponseData{},
	}, nil
}

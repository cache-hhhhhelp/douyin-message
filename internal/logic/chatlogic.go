package logic

import (
	"context"

	"douyin-message/internal/svc"
	__ "douyin-message/types"

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
	// 参数校验
	if in.ToUserId == "" {
		return &__.ChatResponse{
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
	// 查询用户所有消息盒子
	message, err := l.svcCtx.MessageModel.GetMessageListByUserId(l.ctx, in.UserId, in.ToUserId)

	if err != nil {
		return &__.ChatResponse{
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

	// 组合 messageList 为一个数组

	var messageListRes []*__.ChatMessageItem

	if message == nil {
		return &__.ChatResponse{
			BaseResp: &__.BaseResp{
				StatusCode:    __.StatusCode_OK,
				StatusMessage: "success",
			},
			Data: &__.ChatResponseData{
				MessageList: messageListRes,
			},
		}, nil
	}

	for _, messageItem := range message.MessageList {
		messageListRes = append(messageListRes, &__.ChatMessageItem{
			Id:         messageItem.ID.Hex(),
			ToUserId:   messageItem.ToUserId,
			FromUserId: messageItem.FromUserId,
			Content:    messageItem.Content,
			CreateTime: messageItem.CreateAt.UnixMilli(),
		})
	}

	return &__.ChatResponse{
		BaseResp: &__.BaseResp{
			StatusCode:    __.StatusCode_OK,
			StatusMessage: "success",
		},
		Data: &__.ChatResponseData{
			MessageList: messageListRes,
		},
	}, nil
}

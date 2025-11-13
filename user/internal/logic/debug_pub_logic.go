package logic

import (
	"context"

	"github.com/uwu-octane/antBackend/common/eventbus/publisher"
	"github.com/uwu-octane/antBackend/user/internal/event"
	"github.com/uwu-octane/antBackend/user/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type DebugPubLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDebugPubLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DebugPubLogic {
	return &DebugPubLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DebugPubLogic) Test() error {
	if l.svcCtx.UserEventsPusher == nil {
		return nil
	}

	userID := "user.prc"
	key := event.KeyForUser(userID)
	evt := event.NewUserRegisteredEvent(
		userID,
		"", // traceID，可选
		"user-1",
		"user@example.com",
		"DemoUser",
		"",
	)
	logx.Infow("send user registered event", logx.Field("user_id", userID), logx.Field("key", key))
	return publisher.Send(l.ctx, l.svcCtx.UserEventsPusher, evt, key, nil)
}

package logic

import (
	"context"
	"encoding/json"

	"github.com/uwu-octane/antBackend/common/eventbus"
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

	evt := eventbus.NewUserRegisteredEvent(
		"user.rpc",
		"", // traceID，可选
		"user-1",
		"user@example.com",
		"DemoUser",
		"",
	)

	body, _ := json.Marshal(evt)
	return l.svcCtx.UserEventsPusher.Push(l.ctx, string(body))
}

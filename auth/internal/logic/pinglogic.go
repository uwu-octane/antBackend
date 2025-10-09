package logic

import (
	"context"

	"api/v1/auth"
	"auth/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PingLogic {
	return &PingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *PingLogic) Ping(in *auth.PingReq) (*auth.PingResp, error) {
	// todo: add your logic here and delete this line

	return &auth.PingResp{}, nil
}

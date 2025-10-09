// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package logic

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GatewayLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGatewayLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GatewayLogic {
	return &GatewayLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GatewayLogic) Gateway(req *types.Request) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line

	return
}

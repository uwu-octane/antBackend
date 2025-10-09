// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package logic

import (
	"context"

	"github.com/uwu-octane/antBackend/gateway/internal/svc"
	"github.com/uwu-octane/antBackend/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginRequestLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginRequestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginRequestLogic {
	return &LoginRequestLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginRequestLogic) LoginRequest(req *types.Request) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line

	return
}

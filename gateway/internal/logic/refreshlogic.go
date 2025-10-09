// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package logic

import (
	"context"

	"github.com/uwu-octane/antBackend/gateway/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRefreshLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshLogic {
	return &RefreshLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RefreshLogic) Refresh() error {
	// todo: add your logic here and delete this line

	return nil
}

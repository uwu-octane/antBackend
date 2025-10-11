package logic

import (
	"context"
	"errors"

	"github.com/uwu-octane/antBackend/api/v1/auth"
	"github.com/uwu-octane/antBackend/auth/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type LogoutAllLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLogoutAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutAllLogic {
	return &LogoutAllLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LogoutAllLogic) LogoutAll(in *auth.LogoutAllReq) (*auth.LogoutResp, error) {
	if in.GetRefreshToken() == "" {
		return nil, errors.New("refresh token is required")
	}

	return &auth.LogoutResp{}, nil
}

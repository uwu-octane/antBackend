// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package auth

import (
	"context"

	"github.com/uwu-octane/antBackend/api/v1/auth"
	"github.com/uwu-octane/antBackend/gateway/internal/svc"
	"github.com/uwu-octane/antBackend/gateway/internal/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/zeromicro/go-zero/core/logx"
)

type LogoutAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLogoutAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutAllLogic {
	return &LogoutAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LogoutAllLogic) LogoutAll(sid string) (resp *types.LogoutResp, err error) {
	if sid == "" {
		return nil, status.Error(codes.InvalidArgument, "session id is required")
	}
	_, err = l.svcCtx.AuthRpc.Logout(l.ctx, &auth.LogoutReq{
		SessionId: sid,
		All:       true,
	})
	if err != nil {
		return nil, err
	}

	return &types.LogoutResp{
		Ok:      true,
		Message: "logout all success",
	}, nil
}

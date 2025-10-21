// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package auth

import (
	"context"

	"github.com/uwu-octane/antBackend/auth/authservice"
	"github.com/uwu-octane/antBackend/gateway/internal/svc"
	"github.com/uwu-octane/antBackend/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func (l *RefreshLogic) Refresh(sessionId string) (resp *types.LoginResp, err error) {
	if sessionId == "" {
		return nil, status.Error(codes.InvalidArgument, "session id is required")
	}
	r, err := l.svcCtx.AuthRpc.Refresh(l.ctx, &authservice.RefreshReq{
		SessionId: sessionId,
	})
	if err != nil {
		return nil, err
	}
	return &types.LoginResp{
		AccessToken: r.AccessToken,
		ExpiresIn:   r.ExpiresIn,
		TokenType:   r.TokenType,
	}, nil
}

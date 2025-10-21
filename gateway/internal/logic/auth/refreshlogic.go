// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package auth

import (
	"context"

	"github.com/uwu-octane/antBackend/auth/authservice"
	"github.com/uwu-octane/antBackend/gateway/internal/handler/constvar"
	"github.com/uwu-octane/antBackend/gateway/internal/svc"
	"github.com/uwu-octane/antBackend/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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

func (l *RefreshLogic) Refresh(sessionId string, refreshToken string) (resp *types.LoginResp, header metadata.MD, err error) {
	md := metadata.Pairs(constvar.HeaderRefreshToken, refreshToken)
	ctx := metadata.NewOutgoingContext(l.ctx, md)

	var hdr metadata.MD
	r, err := l.svcCtx.AuthRpc.Refresh(ctx, &authservice.RefreshReq{
		SessionId: sessionId,
	}, grpc.Header(&hdr))
	if err != nil {
		return nil, nil, err
	}
	logx.Infof("refresh response: %+v", r)
	logx.Infof("refresh header: %+v", hdr)

	return &types.LoginResp{
		AccessToken: r.AccessToken,
		ExpiresIn:   r.ExpiresIn,
		TokenType:   r.TokenType,
	}, hdr, nil
}

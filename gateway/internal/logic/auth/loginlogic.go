// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package auth

import (
	"context"

	"github.com/uwu-octane/antBackend/auth/authservice"
	"github.com/uwu-octane/antBackend/gateway/internal/svc"
	"github.com/uwu-octane/antBackend/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.LoginResp, header metadata.MD, err error) {
	var md metadata.MD
	r, err := l.svcCtx.AuthRpc.Login(l.ctx, &authservice.LoginReq{
		Username: req.Username,
		Password: req.Password,
	},
		grpc.Header(&md), // get grpc Header
	)
	if err != nil {
		return nil, nil, err
	}
	logx.Infof("login response: %+v", r)
	logx.Infof("login metadata: %+v", md)
	return &types.LoginResp{
		AccessToken: r.GetAccessToken(),
		SessionId:   r.GetSessionId(),
		ExpiresIn:   r.GetExpiresIn(),
		TokenType:   r.GetTokenType(),
	}, md, nil
}

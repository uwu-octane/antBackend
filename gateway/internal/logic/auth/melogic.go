// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package auth

import (
	"context"

	"github.com/uwu-octane/antBackend/gateway/internal/handler/constvar"
	"github.com/uwu-octane/antBackend/gateway/internal/svc"
	"github.com/uwu-octane/antBackend/gateway/internal/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/zeromicro/go-zero/core/logx"
)

type MeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MeLogic {
	return &MeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MeLogic) Me() (resp *types.MeResp, err error) {
	uid, _ := l.ctx.Value(constvar.CtxUID).(string)
	jti, _ := l.ctx.Value(constvar.CtxJTI).(string)
	iat, _ := l.ctx.Value(constvar.CtxIAT).(int64)

	if uid == "" || jti == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	return &types.MeResp{
		Uid: uid,
		Jti: jti,
		Iat: iat,
	}, nil
}

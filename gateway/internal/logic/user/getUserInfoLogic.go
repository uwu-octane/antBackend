// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package user

import (
	"context"

	"github.com/uwu-octane/antBackend/api/v1/user"
	"github.com/uwu-octane/antBackend/gateway/internal/middleware"
	"github.com/uwu-octane/antBackend/gateway/internal/svc"
	"github.com/uwu-octane/antBackend/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetUserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserInfoLogic) GetUserInfo() (resp *types.UserInfoResp, err error) {
	uid, ok := middleware.UIDFromContext(l.ctx)
	if !ok || uid == "" {
		logx.Info("l.ctx.Value(sub) is empty")
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}
	userResp, err := l.svcCtx.UserRpc.GetUserInfo(l.ctx, &user.GetUserInfoReq{
		UserId: uid,
	})
	if err != nil {
		return nil, err
	}
	return &types.UserInfoResp{
		UserId:      userResp.UserId,
		Username:    userResp.Username,
		Email:       userResp.Email,
		DisplayName: userResp.DisplayName,
		AvatarUrl:   userResp.AvatarUrl,
	}, nil
}

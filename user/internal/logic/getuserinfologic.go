package logic

import (
	"context"
	"database/sql"

	"github.com/uwu-octane/antBackend/api/v1/user"
	"github.com/uwu-octane/antBackend/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserInfoLogic) GetUserInfo(in *user.GetUserInfoReq) (*user.GetUserInfoResp, error) {
	u, err := l.svcCtx.Users.FindOne(l.ctx, in.GetUserId())
	if err != nil {
		l.Errorf("failed to find user: %v", err)
		return nil, err
	}

	return &user.GetUserInfoResp{
		UserId:      u.Id,
		Username:    u.Username,
		Email:       nullable(u.Email),
		DisplayName: nullable(u.DisplayName),
		AvatarUrl:   nullable(u.AvatarUrl),
	}, nil
}

func nullable(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

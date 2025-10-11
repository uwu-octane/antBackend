package logic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/uwu-octane/antBackend/api/v1/auth"
	"github.com/uwu-octane/antBackend/auth/internal/svc"
	"github.com/uwu-octane/antBackend/auth/internal/util"

	"github.com/zeromicro/go-zero/core/logx"
)

type LogoutLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LogoutLogic) Logout(in *auth.LogoutReq) (*auth.LogoutResp, error) {
	if in.GetRefreshToken() == "" {
		return nil, errors.New("refresh token is required")
	}

	cfg := l.svcCtx.Config.JwtAuth
	tokenHelper := NewTokenHelper([]byte(cfg.Secret), "auth.prc",
		time.Duration(cfg.AccessExpireSeconds)*time.Second, time.Duration(cfg.RefreshExpireSeconds)*time.Second)

	claims, err := tokenHelper.Parse(in.GetRefreshToken())
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if claims.TokenType != "refresh" {
		return nil, errors.New("invalid refresh token")
	}
	jti := claims.ID
	sub := claims.Subject
	if jti == "" || sub == "" {
		return nil, errors.New("invalid refresh token")
	}

	refreshKey := util.RedisKey(l.svcCtx.Key, util.RedisKeyTypeRefresh, jti)
	uid, err := l.svcCtx.Redis.Get(refreshKey)
	if err != nil || uid == "" {
		return &auth.LogoutResp{
			Ok:      true,
			Message: fmt.Sprintf("May expired, as idempotent: %v", err),
		}, nil
	}
	if uid != sub {
		return nil, errors.New("subject mismatch")
	}
	if _, err := l.svcCtx.Redis.Del(refreshKey); err != nil {
		l.Logger.Errorf("failed to delete refresh token: %w", err)
		return nil, err
	}
	//set reuse flag
	_ = l.svcCtx.Redis.Set(util.RedisKey(l.svcCtx.Key, util.RedisKeyTypeReuse, jti), "1")

	//todo: clear jti from sid family
	return &auth.LogoutResp{Ok: true, Message: "log out"}, nil
}

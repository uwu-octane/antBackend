package logic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uwu-octane/antBackend/api/v1/auth"
	"github.com/uwu-octane/antBackend/auth/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRefreshLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshLogic {
	return &RefreshLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RefreshLogic) Refresh(in *auth.RefreshReq) (*auth.LoginResp, error) {
	if in.GetRefreshToken() == "" {
		return nil, errors.New("refresh token is required")
	}

	cfg := l.svcCtx.Config.JwtAuth
	tokenHelper := NewTokenHelper([]byte(cfg.Secret), "auth.prc",
		time.Duration(cfg.AccessExpireSeconds)*time.Second, time.Duration(cfg.RefreshExpireSeconds)*time.Second)

	claims, err := tokenHelper.Parse(in.GetRefreshToken())
	if err != nil {
		logx.Errorf("invalid token: %v", err)
		return nil, err
	}

	typ := claims.TokenType
	if typ != "refresh" {
		return nil, errors.New("invalid token type, got " + typ)
	}

	jti := claims.ID
	sub := claims.Subject

	if jti == "" || sub == "" {
		return nil, errors.New("invalid token, got jti: " + jti + " and sub: " + sub)
	}

	//* Delete the old refresh token
	oldKey := fmt.Sprintf("%srefresh:%s", l.svcCtx.Key, jti)
	uid, err := l.svcCtx.Redis.Get(oldKey)
	if err != nil || uid == "" {
		return nil, errors.New("refresh token revoked or expired")
	}

	if uid != sub {
		return nil, errors.New("invalid token, got uid: " + uid + " and sub: " + sub)
	}

	_, err = l.svcCtx.Redis.Del(oldKey)
	if err != nil {
		return nil, errors.New("failed to delete refresh token")
	}
	//* Generate new refresh and access tokens
	newRefreshJti := uuid.NewString()
	newAccessJti := uuid.NewString()

	access, accessExpireSeconds, err := tokenHelper.SignAccess(sub, newAccessJti)
	if err != nil {
		return nil, err
	}

	newRefresh, refreshExpireSeconds, err := tokenHelper.SignRefresh(sub, newRefreshJti)
	if err != nil {
		return nil, err
	}

	newKey := fmt.Sprintf("%srefresh:%s", l.svcCtx.Key, newRefreshJti)
	if err := l.svcCtx.Redis.Setex(newKey, sub, int(refreshExpireSeconds)); err != nil {
		return nil, err
	}

	return &auth.LoginResp{
		AccessToken:  access,
		RefreshToken: newRefresh,
		ExpiresIn:    accessExpireSeconds,
		TokenType:    "bearer",
	}, nil
}

//todo
// 	1.	Redis 值存 JSON：{uid, device, ip, lastRefresh, ver}；
// 2.	并发保护：SETNX / Lua 原子旋转；
// 3.	重用检测：旧 jti 再来 → 记审计 & 封禁该会话；
// 4.	滑动过期：每次刷新把 Redis TTL 重置为“会话最长不活跃时间”。

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

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogic) Login(in *auth.LoginReq) (*auth.LoginResp, error) {
	//dev only
	if in.GetUsername() == "" || in.GetPassword() == "" {
		return nil, errors.New("username and password are required")
	}

	//todo: call user rpc, then replace with userID
	if in.GetUsername() != "admin" || in.GetPassword() != "admin" {
		return nil, errors.New("invalid credentials")
	}
	userID := in.GetUsername()

	//* call token helper to sign tokens
	cfg := l.svcCtx.Config.JwtAuth
	tokenHelper := NewTokenHelper([]byte(cfg.Secret), "auth.prc",
		time.Duration(cfg.AccessExpireSeconds)*time.Second, time.Duration(cfg.RefreshExpireSeconds)*time.Second)

	refreshJti := uuid.NewString()
	accessJti := uuid.NewString()

	accessToken, accessExpireSeconds, err := tokenHelper.SignAccess(userID, accessJti)

	if err != nil {
		return nil, err
	}

	refreshToken, refreshExpireSeconds, err := tokenHelper.SignRefresh(userID, refreshJti)
	if err != nil {
		return nil, err
	}

	//* Redis：auth:refresh:<jti> = userID（Or JSON），TTL=RefreshExpireSeconds
	key := fmt.Sprintf("%srefresh:%s", l.svcCtx.Key, refreshJti)
	if err := l.svcCtx.Redis.Setex(key, userID, int(refreshExpireSeconds)); err != nil {
		return nil, err
	}

	return &auth.LoginResp{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    accessExpireSeconds,
		TokenType:    "bearer",
	}, nil
}

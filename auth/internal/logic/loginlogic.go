package logic

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/uwu-octane/antBackend/api/v1/auth"
	"github.com/uwu-octane/antBackend/auth/internal/svc"
	"github.com/uwu-octane/antBackend/auth/internal/util"

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
	refreshJti := uuid.NewString()
	accessJti := uuid.NewString()

	accessToken, accessExpireSeconds, err := l.svcCtx.TokenHelper.SignAccess(userID, accessJti)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshExpireSeconds, err := l.svcCtx.TokenHelper.SignRefresh(userID, refreshJti)
	if err != nil {
		return nil, err
	}

	sid := uuid.NewString()
	//* user->sid
	if _, err := l.svcCtx.Redis.Sadd(util.UserSidsKey(l.svcCtx.Key, userID), sid); err != nil {
		return nil, err
	}
	//* sid->jit put the first refresh jti in to the sid collection
	if _, err := l.svcCtx.Redis.Sadd(util.SidSetKey(l.svcCtx.Key, sid), refreshJti); err != nil {
		return nil, err
	}
	//* jti ->sid (index of jti to sid)
	if err := l.svcCtx.Redis.Setex(util.JtiSidKey(l.svcCtx.Key, refreshJti), sid, int(refreshExpireSeconds)); err != nil {
		return nil, err
	}

	//* Redis：auth:refresh:<jti> = userID（Or JSON），TTL=RefreshExpireSeconds
	key := util.RedisKey(l.svcCtx.Key, util.RedisKeyTypeRefresh, refreshJti)
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

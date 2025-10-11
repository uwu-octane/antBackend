package logic

import (
	"context"
	"errors"

	"github.com/uwu-octane/antBackend/api/v1/auth"
	"github.com/uwu-octane/antBackend/auth/internal/svc"
	"github.com/uwu-octane/antBackend/auth/internal/util"

	"github.com/zeromicro/go-zero/core/logx"
)

type LogoutAllLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLogoutAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutAllLogic {
	return &LogoutAllLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LogoutAllLogic) LogoutAll(in *auth.LogoutAllReq) (*auth.LogoutResp, error) {
	if in.GetRefreshToken() == "" {
		return nil, errors.New("refresh token is required")
	}
	claims, err := l.svcCtx.TokenHelper.ValidateRefreshToken(in.GetRefreshToken())
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	uid := claims.Subject
	key := l.svcCtx.Key

	userSidsKey := util.UserSidsKey(key, uid)
	sids, err := l.svcCtx.Redis.Smembers(userSidsKey)
	if err != nil {
		return nil, errors.New("failed to get sids")
	}

	totalJtis := 0
	for _, sid := range sids {
		sidKey := util.SidSetKey(key, sid)
		jtis, err := l.svcCtx.Redis.Smembers(sidKey)
		if err != nil {
			logx.WithContext(l.ctx).Errorf("failed to get jtis for sid %s: %v", sid, err)
			continue
		}
		totalJtis += len(jtis)
		for _, jti := range jtis {
			// delete active refresh
			refreshKey := util.RedisKey(key, util.RedisKeyTypeRefresh, jti)
			if _, err := l.svcCtx.Redis.Del(refreshKey); err != nil {
				logx.WithContext(l.ctx).Errorf("LogoutAll: failed to delete refresh token for jti %s: %v", jti, err)
			}
			// delete jti ->sid index
			if _, err := l.svcCtx.Redis.Del(util.JtiSidKey(key, jti)); err != nil {
				logx.WithContext(l.ctx).Errorf("LogoutAll: failed to delete jti ->sid index for jti %s: %v", jti, err)
			}
			//add reuse flag
			_ = l.svcCtx.Redis.Set(util.RedisKey(key, util.RedisKeyTypeReuse, jti), "logoutAll")
			// remove jti from sid
			if _, err := l.svcCtx.Redis.Srem(sidKey, jti); err != nil {
				logx.WithContext(l.ctx).Errorf("LogoutAll: failed to remove jti %s from sid %s: %v", jti, sid, err)
			}
		}
		// empty sid set
		if _, err := l.svcCtx.Redis.Del(sidKey); err != nil {
			logx.WithContext(l.ctx).Errorf("LogoutAll: fail to del sid set %s: %v", sid, err)
		}

	}

	// delete user sids
	if _, err := l.svcCtx.Redis.Del(userSidsKey); err != nil {
		logx.WithContext(l.ctx).Errorf("LogoutAll: fail to del user sids %s: %v", uid, err)
	}

	return &auth.LogoutResp{Ok: true, Message: "Log out all success"}, nil
}

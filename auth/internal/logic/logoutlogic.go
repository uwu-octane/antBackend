package logic

import (
	"context"
	"errors"
	"fmt"

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
	sid := in.GetSessionId()
	if sid == "" {
		return nil, errors.New("logout: session id is required")
	}

	uid, err := l.revokeOneSid(l.svcCtx.Key, sid)
	if err != nil {
		return nil, fmt.Errorf("logout: failed to revoke sid: %s, %v", sid, err)
	}

	if in.GetAll() && uid != "" {
		//get all sids of the user
		userSidsKey := util.UserSidsKey(l.svcCtx.Key, uid)
		allSids, _ := l.svcCtx.Redis.Smembers(userSidsKey)
		//revoke all sids of the user
		for _, s := range allSids {
			if _, err := l.revokeOneSid(l.svcCtx.Key, s); err != nil {
				logx.WithContext(l.ctx).Errorf("logout: revoke sid in all failed uid=%s sid=%s err=%v", uid, s, err)
			}
		}
		//delete user sids set
		_, _ = l.svcCtx.Redis.Del(userSidsKey)
	}

	return &auth.LogoutResp{Ok: true, Message: "logged out"}, nil
}

// delete all refresh jti associated with the sid
// clear index
// return uid
func (l *LogoutLogic) revokeOneSid(key, sid string) (string, error) {
	sidKey := util.SidSetKey(key, sid)
	jtis, err := l.svcCtx.Redis.Smembers(sidKey)
	if err != nil {
		return "", fmt.Errorf("logout: failed to get jtis: %w", err)
	}
	if len(jtis) == 0 {
		_, _ = l.svcCtx.Redis.Del(sidKey)
		return "", nil
	}
	var uid string
	for _, jti := range jtis {
		if uid == "" {
			if v, _ := l.svcCtx.Redis.Get(util.RedisKey(key, util.RedisKeyTypeRefresh, jti)); v != "" {
				uid = v
			}
		}
		//delete auth:refresh:<jti>
		_, _ = l.svcCtx.Redis.Del(util.RedisKey(key, util.RedisKeyTypeRefresh, jti))
		//set reuse flag
		_ = l.svcCtx.Redis.Set(util.RedisKey(key, util.RedisKeyTypeReuse, jti), "1")
		//delete auth:jti_sid:<jti>
		_, _ = l.svcCtx.Redis.Del(util.JtiSidKey(key, jti))
		//delete jti from sid set
		_, _ = l.svcCtx.Redis.Srem(sidKey, jti)
	}
	//delete sid set
	_, _ = l.svcCtx.Redis.Del(sidKey)

	//delete auth:user:<uid>:sids
	if uid != "" {
		_, _ = l.svcCtx.Redis.Srem(util.UserSidsKey(key, uid), sid)
	}
	return uid, nil
}

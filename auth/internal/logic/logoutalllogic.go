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
	perSid := make(map[string]int)
	for _, sid := range sids {
		jtis, err := l.svcCtx.Redis.Smembers(util.SidSetKey(key, sid))
		if err != nil {
			logx.WithContext(l.ctx).Errorf("failed to get jtis for sid %s: %v", sid, err)
			continue
		}
		totalJtis += len(jtis)
		perSid[sid] = len(jtis)
	}

	logx.WithContext(l.ctx).Infof("logoutAll.dryrun uid=%s sidCount=%d totalJtis=%d detail=%v",
		uid, len(sids), totalJtis, perSid)

	return &auth.LogoutResp{Ok: true, Message: "Log out all success"}, nil
}

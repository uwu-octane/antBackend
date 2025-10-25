// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package svc

import (
	"context"

	"github.com/uwu-octane/antBackend/auth/authservice"
	"github.com/uwu-octane/antBackend/gateway/internal/config"
	"github.com/uwu-octane/antBackend/user/userservice"

	"github.com/uwu-octane/antBackend/gateway/internal/upstream/consulmanager"
	"github.com/zeromicro/go-zero/core/limit"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config        config.Config
	AuthRpc       authservice.AuthService
	UserRpc       userservice.UserService
	LoginLimiter  *limit.PeriodLimit
	ConsulManager *consulmanager.Manager
	Targets       map[string]*consulmanager.Target
}

func NewServiceContext(c config.Config) *ServiceContext {
	s := &ServiceContext{
		Config:  c,
		AuthRpc: authservice.NewAuthService(zrpc.MustNewClient(c.AuthRpc)),
		UserRpc: userservice.NewUserService(zrpc.MustNewClient(c.UserRpc)),
	}
	if c.RateLimit.Enable {
		store := redis.MustNewRedis(c.RateLimit.RateLimitRedis.RedisConf)
		LoginLimiter := limit.NewPeriodLimit(c.RateLimit.WindowSeconds,
			c.RateLimit.MaxAttempts, store, "login:limit")
		s.LoginLimiter = LoginLimiter
	}

	mgr, targets, err := makeManager(c)
	if err != nil {
		logx.Errorw("make manager failed", logx.Field("error", err))
	} else {
		s.ConsulManager = mgr
		s.Targets = targets
	}
	return s
}

func makeManager(c config.Config) (*consulmanager.Manager, map[string]*consulmanager.Target, error) {
	mgr, err := consulmanager.NewManager(&consulmanager.ManagerOption{
		Address:    c.Consul.Address,
		Scheme:     c.Consul.Scheme,
		Datacenter: c.Consul.Datacenter,
		Token:      c.Consul.Token,
		WaitTime:   c.Consul.WaitTime,
	})
	if err != nil {
		logx.Errorw("new consul manager failed", logx.Field("error", err))
		return nil, nil, err
	}
	targets := make(map[string]*consulmanager.Target)
	ctx := context.Background()
	for _, upstream := range c.Upstreams {
		tgt := mgr.Watch(ctx, upstream.Service)
		targets[upstream.Name] = tgt
		logx.Infow("watch upstream", logx.Field("name", upstream.Name), logx.Field("target", tgt.String()))
	}
	return mgr, targets, nil
}

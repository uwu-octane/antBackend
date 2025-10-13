// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package svc

import (
	"github.com/uwu-octane/antBackend/auth/authservice"
	"github.com/uwu-octane/antBackend/gateway/internal/config"
	"github.com/uwu-octane/antBackend/user/userservice"

	"github.com/zeromicro/go-zero/core/limit"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config       config.Config
	AuthRpc      authservice.AuthService
	UserRpc      userservice.UserService
	LoginLimiter *limit.PeriodLimit
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
	return s
}

package svc

import (
	"github.com/uwu-octane/antBackend/auth/internal/config"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type ServiceContext struct {
	Config config.Config
	Redis  *redis.Redis
	Key    string
}

func NewServiceContext(c config.Config) *ServiceContext {
	redis := redis.MustNewRedis(c.AuthRedis.RedisConf)

	return &ServiceContext{
		Config: c,
		Redis:  redis,
		Key:    c.AuthRedis.Key,
	}
}

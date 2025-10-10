package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type JwtAuthConfig struct {
	Secret               string `json:",optional"`
	AccessExpireSeconds  int64  `json:",default=3600"`
	RefreshExpireSeconds int64  `json:",default=604800"`
}

type Config struct {
	zrpc.RpcServerConf
	JwtAuth   JwtAuthConfig
	AuthRedis redis.RedisKeyConf
}

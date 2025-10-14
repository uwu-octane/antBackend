package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	UserDatabase DatabaseConfig
	UserRedis    redis.RedisKeyConf
}

type DatabaseConfig struct {
	Driver     string
	MasterDSN  string
	ReplicaDSN string
}

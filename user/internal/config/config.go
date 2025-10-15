package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	UserDatabase     UserDatabase
	UserRedis        redis.RedisKeyConf
	UserReadStrategy UserReadStrategy
}

type UserDatabase struct {
	Driver     string
	MasterDSN  string
	ReplicaDSN string
}

type UserReadStrategy struct {
	FromReplica                 bool
	FallbackToMasterOnReadError bool
}

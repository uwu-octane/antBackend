package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type JwtAuthConfig struct {
	Secret               string
	AccessExpireSeconds  int64
	RefreshExpireSeconds int64
}

type Config struct {
	zrpc.RpcServerConf
	JwtAuth          JwtAuthConfig
	AuthRedis        redis.RedisKeyConf
	AuthDatabase     AuthDatabase
	AuthReadStrategy AuthReadStrategy
}

type AuthDatabase struct {
	Driver     string
	MasterDSN  string
	ReplicaDSN string
}

type AuthReadStrategy struct {
	FromReplica                 bool
	FallbackToMasterOnReadError bool
}

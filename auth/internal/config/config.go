package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
	"github.com/zeromicro/zero-contrib/zrpc/registry/consul"
)

type JwtAuthConfig struct {
	Secret               string
	AccessExpireSeconds  int64
	RefreshExpireSeconds int64
}

type Config struct {
	zrpc.RpcServerConf
	Consul           consul.Conf
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

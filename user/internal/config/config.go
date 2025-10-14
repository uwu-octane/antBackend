package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	Log       LogConfig
	Database  DatabaseConfig
	UserRedis redis.RedisKeyConf
}

type LogConfig struct {
	Encoding string `json:"Encoding"`
	Level    string `json:"Level"`
}

type DatabaseConfig struct {
	Driver     string `json:"Driver"`
	MasterDSN  string `json:"MasterDSN"`
	ReplicaDSN string `json:"ReplicaDSN"`
}

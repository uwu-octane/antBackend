package config

import (
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
	"github.com/zeromicro/zero-contrib/zrpc/registry/consul"
)

type Config struct {
	zrpc.RpcServerConf
	Consul           consul.Conf
	UserDatabase     UserDatabase
	UserRedis        redis.RedisKeyConf
	UserReadStrategy UserReadStrategy

	Kafka        KafkaConf
	KqUserEvents kq.KqConf
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

type KafkaConf struct {
	Env     string
	Brokers []string
}

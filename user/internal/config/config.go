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

	Kafka             KafkaConf
	KqUserEvents      kq.KqConf
	KafkaUserProducer KafkaProducerConf
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

type KafkaProducerSASL struct {
	Enable    bool
	Mechanism string // plain / scram-sha256 / scram-sha512
	Username  string
	Password  string
}

type KafkaProducerTLS struct {
	Enable bool
}

type KafkaProducerConf struct {
	Acks             string // all/local
	Idempotent       bool
	RetryMax         int
	Compression      string // none/snappy/lz4/zstd/gzip
	FlushBytes       int
	FlushMessages    int
	FlushFrequencyMs int
	MaxMessageBytes  int
	SASL             KafkaProducerSASL
	TLS              KafkaProducerTLS
}

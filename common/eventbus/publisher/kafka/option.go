package kafka

import "time"

// 供各服务层适配使用的通用 producer 选项
type ProducerOptions struct {
	Brokers         []string
	Acks            string // "all" | "local"
	Idempotent      bool
	RetryMax        int
	Compression     string // "none"|"snappy"|"lz4"|"zstd"|"gzip"
	FlushBytes      int
	FlushMessages   int
	FlushFrequency  time.Duration // 直接用 time.Duration，服务层自转毫秒
	MaxMessageBytes int
	EnableSASL      bool
	SASLMechanism   string // "plain" | "scram-sha256" | "scram-sha512"
	SASLUsername    string
	SASLPassword    string
	EnableTLS       bool
	KafkaVersion    string // "3.6.0" 等，可留空用默认
}

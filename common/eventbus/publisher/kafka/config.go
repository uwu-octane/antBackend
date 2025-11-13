package kafka

import (
	"fmt"
	"strings"
	"time"

	"github.com/Shopify/sarama"
)

func BuildSaramaConfig(c *ProducerOptions) (*sarama.Config, error) {
	if c == nil {
		c = &ProducerOptions{}
	}

	cfg := sarama.NewConfig()

	switch strings.ToLower(c.Acks) {
	case "all":
		cfg.Producer.RequiredAcks = sarama.WaitForAll
	case "local":
		cfg.Producer.RequiredAcks = sarama.WaitForLocal
	case "":
		cfg.Producer.RequiredAcks = sarama.WaitForAll
	default:
		return nil, fmt.Errorf("unsupported kafka producer acks: %s", c.Acks)
	}

	cfg.Producer.Idempotent = c.Idempotent
	cfg.Producer.Retry.Max = c.RetryMax
	cfg.Producer.Return.Successes = true

	if cfg.Producer.Idempotent {
		cfg.Producer.RequiredAcks = sarama.WaitForAll
		if cfg.Producer.Retry.Max <= 0 {
			cfg.Producer.Retry.Max = 3
		}
		cfg.Net.MaxOpenRequests = 1
	}

	switch strings.ToLower(c.Compression) {
	case "snappy":
		cfg.Producer.Compression = sarama.CompressionSnappy
	case "lz4":
		cfg.Producer.Compression = sarama.CompressionLZ4
	case "zstd":
		cfg.Producer.Compression = sarama.CompressionZSTD
	case "gzip":
		cfg.Producer.Compression = sarama.CompressionGZIP
	case "", "none":
		cfg.Producer.Compression = sarama.CompressionNone
	default:
		return nil, fmt.Errorf("unsupported kafka producer compression: %s", c.Compression)
	}

	cfg.Producer.Flush.Bytes = c.FlushBytes
	cfg.Producer.Flush.Messages = c.FlushMessages
	cfg.Producer.Flush.Frequency = c.FlushFrequency
	cfg.Producer.MaxMessageBytes = c.MaxMessageBytes

	cfg.Producer.Retry.Backoff = 300 * time.Millisecond
	cfg.Metadata.Retry.Max = 10
	cfg.Metadata.Retry.Backoff = 250 * time.Millisecond
	cfg.Metadata.RefreshFrequency = 30 * time.Second
	cfg.Net.DialTimeout = 10 * time.Second
	cfg.Net.ReadTimeout = 10 * time.Second
	cfg.Net.WriteTimeout = 10 * time.Second

	if c.EnableSASL {
		if err := applySASL(cfg, c); err != nil {
			return nil, err
		}
	}

	if c.EnableTLS {
		cfg.Net.TLS.Enable = true
	}

	if strings.TrimSpace(c.KafkaVersion) != "" {
		version, err := sarama.ParseKafkaVersion(c.KafkaVersion)
		if err != nil {
			return nil, fmt.Errorf("parse kafka version: %w", err)
		}
		cfg.Version = version
	}

	return cfg, nil
}

func applySASL(cfg *sarama.Config, opts *ProducerOptions) error {
	mechanism := strings.ToLower(opts.SASLMechanism)
	if mechanism == "" {
		mechanism = "plain"
	}

	switch mechanism {
	case "plain":
		cfg.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	case "scram-sha256":
		cfg.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
	case "scram-sha512":
		cfg.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
	default:
		cfg.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	}

	if opts.EnableSASL {
		cfg.Net.SASL.Enable = true
		cfg.Net.SASL.User = opts.SASLUsername
		cfg.Net.SASL.Password = opts.SASLPassword
		if cfg.Net.SASL.User == "" || cfg.Net.SASL.Password == "" {
			return fmt.Errorf("sasl username/password must be set when sasl is enabled")
		}
	}
	if opts.EnableTLS {
		cfg.Net.TLS.Enable = true
	}
	return nil
}

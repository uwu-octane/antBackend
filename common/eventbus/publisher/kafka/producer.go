package kafka

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/uwu-octane/antBackend/common/eventbus/publisher"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/Shopify/sarama"
)

type SaramaPublisher struct {
	producer sarama.SyncProducer
}

func NewSaramaPublisher(opts *ProducerOptions) (*SaramaPublisher, error) {
	cfg, err := BuildSaramaConfig(opts)
	if err != nil {
		logx.Errorw("build sarama config failed", logx.Field("error", err))
		return nil, err
	}
	p, err := sarama.NewSyncProducer(opts.Brokers, cfg)
	if err != nil {
		logx.Errorw("create sarama producer failed", logx.Field("error", err))
		return nil, err
	}
	logx.Infow("create sarama producer", logx.Field("brokers", opts.Brokers))
	return &SaramaPublisher{producer: p}, nil
}

func (s *SaramaPublisher) Publish(ctx context.Context, topic string, data []byte, opts *publisher.PublishOptions) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(data),
	}
	if opts != nil && len(opts.PartitionKey) > 0 {
		msg.Key = sarama.ByteEncoder(opts.PartitionKey)
		for k, v := range opts.Headers {
			msg.Headers = append(msg.Headers, sarama.RecordHeader{
				Key:   []byte(k),
				Value: []byte(v),
			})
		}
	}
	const maxRetries = 3
	backoff := 150 * time.Millisecond
	for i := 0; i < maxRetries; i++ {
		_, _, err := s.producer.SendMessage(msg)
		if err == nil {
			return nil
		}
		var ke sarama.KError
		if errors.As(err, &ke) {
			switch ke {
			case sarama.ErrOffsetsLoadInProgress,
				sarama.ErrLeaderNotAvailable,
				sarama.ErrNotEnoughReplicas,
				sarama.ErrNotEnoughReplicasAfterAppend:
				logx.Infow("kafka message send failed, retrying", logx.Field("error", err))
			default:
				return err
			}
		} else {
			continue
		}
		if i < maxRetries-1 {
			select {
			case <-time.After(backoff):
				backoff *= 2
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	return fmt.Errorf("failed to send message after %d retries", maxRetries)
}

func (s *SaramaPublisher) Close() error {
	return s.producer.Close()
}

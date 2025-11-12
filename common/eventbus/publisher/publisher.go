package publisher

import (
	"context"
)

type PublishOptions struct {
	// PartitionKey 决定消息分区与局部顺序（例如 user_id / order_id）
	PartitionKey []byte
	Headers      map[string]string
}

// Publisher 定义对上游业务暴露的统一发布接口
type Publisher interface {
	Publish(ctx context.Context, topic string, data []byte, opts *PublishOptions) error
	Close() error
}

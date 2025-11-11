package eventbus

import (
	"time"

	"github.com/oklog/ulid"
)

const (
	EventTypeOrderCreated   = "order.created"
	EventTypeUserRegistered = "user.registered"
	EventTypeUserUpdated    = "user.updated"
	EventTypeUserDeleted    = "user.deleted"
)

type Envelope[T any] struct {
	// 事件类型：语义化标识
	EventType string `json:"event_type"`
	// Schema 版本：从 1 开始，变更 payload 时递增
	EventVersion int `json:"event_version"`
	// 全局唯一 ID：用来做幂等、防重
	EventID string `json:"event_id"`
	// 事件发生时间（业务时间，不是写入 Kafka 的时间）
	OccurredAt time.Time `json:"occurred_at"`
	// 生产者标识：事件来源
	Producer string `json:"producer"`
	// 追踪 ID：用于关联事件和上游事件
	TraceID string `json:"trace_id,omitempty"`
	// 事件数据：具体事件内容
	Data T `json:"data"`
}

func NewEnvelope[T any](eventType string, eventVersion int, producer string, traceID string, data T) *Envelope[T] {
	return &Envelope[T]{
		EventType:    eventType,
		EventVersion: eventVersion,
		EventID:      ulid.MustNew(1, nil).String(),
		OccurredAt:   time.Now().UTC(),
		Producer:     producer,
		TraceID:      traceID,
		Data:         data,
	}
}

package publisher

import (
	"context"
	"errors"

	"github.com/uwu-octane/antBackend/common/eventbus/codec"
	"github.com/uwu-octane/antBackend/common/eventbus/event"
)

// EventBusPublisher 面向业务的语义化发布器（强类型泛型）
type EventBusPublisher struct {
	Pub    Publisher
	Topics event.TopicSet
}

func NewEventBusPublisher(pub Publisher, topics event.TopicSet) *EventBusPublisher {
	return &EventBusPublisher{Pub: pub, Topics: topics}
}

// SendUserEvent 发送用户流上的任意 Envelope
func Send[T any](ctx context.Context, p *EventBusPublisher, envelope *event.Envelope[T], key []byte, headers map[string]string) error {
	if p == nil || p.Pub == nil {
		return errors.New("publisher or pub is nil")
	}
	if envelope == nil {
		return errors.New("envelope is nil")
	}
	if p.Topics.UserEvents == "" {
		return errors.New("topic not set")
	}
	body, err := codec.MarshalEnvelope(envelope)
	if err != nil {
		return err
	}
	return p.Pub.Publish(ctx, p.Topics.UserEvents, body, &PublishOptions{
		PartitionKey: key,
		Headers:      headers,
	})
}

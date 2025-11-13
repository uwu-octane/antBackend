package codec

import (
	"encoding/json"
	"errors"

	"github.com/uwu-octane/antBackend/common/eventbus/event"
)

func MarshalEnvelope[T any](env *event.Envelope[T]) ([]byte, error) {
	if env == nil {
		return nil, errors.New("envelope is nil")
	}
	return json.Marshal(env)
}

func UnmarshalEnvelope[T any](b []byte, out *event.Envelope[T]) error {
	if out == nil {
		return errors.New("envelope is nil")
	}
	return json.Unmarshal(b, out)
}

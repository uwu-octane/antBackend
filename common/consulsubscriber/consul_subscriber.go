package consulsubscriber

import (
	"context"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/zeromicro/go-zero/core/configcenter/subscriber"
	"github.com/zeromicro/go-zero/core/logx"
)

// ConsulSubscriber implements subscriber.Subscriber for Consul KV
type ConsulSubscriber struct {
	client    *api.Client
	key       string
	lastValue string
	mu        sync.RWMutex
	listeners []func()
	cancel    context.CancelFunc
}

// NewConsulSubscriber creates a new subscriber instance.
func NewConsulSubscriber(addr string, key string) (subscriber.Subscriber, error) {
	cfg := api.DefaultConfig()
	cfg.Address = addr
	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	sub := &ConsulSubscriber{
		client:    client,
		key:       key,
		listeners: make([]func(), 0),
	}

	ctx, cancel := context.WithCancel(context.Background())
	sub.cancel = cancel
	go sub.watchLoop(ctx)

	return sub, nil
}

// MustNewConsulSubscriber creates subscriber and panics on error
func MustNewConsulSubscriber(addr string, key string) subscriber.Subscriber {
	sub, err := NewConsulSubscriber(addr, key)
	if err != nil {
		panic(err)
	}
	return sub
}

// Value returns the current stored configuration string (e.g. YAML or JSON)
func (s *ConsulSubscriber) Value() (string, error) {
	// fast‐path if we already have a value
	s.mu.RLock()
	if s.lastValue != "" {
		val := s.lastValue
		s.mu.RUnlock()
		return val, nil
	}
	s.mu.RUnlock()

	// initial load
	pair, _, err := s.client.KV().Get(s.key, nil)
	if err != nil {
		return "", err
	}
	if pair == nil {
		return "", nil
	}
	val := string(pair.Value)
	s.mu.Lock()
	s.lastValue = val
	s.mu.Unlock()
	return val, nil
}

// AddListener registers a change‐notification listener
func (s *ConsulSubscriber) AddListener(listener func()) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.listeners = append(s.listeners, listener)
	return nil
}

// watchLoop runs in background, listening for changes via blocking query
func (s *ConsulSubscriber) watchLoop(ctx context.Context) {
	var lastIndex uint64
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		pair, meta, err := s.client.KV().Get(s.key, &api.QueryOptions{
			WaitIndex: lastIndex,
			WaitTime:  10 * time.Second,
		})
		if err != nil {
			logx.Errorf("ConsulSubscriber watch error on key %s: %v", s.key, err)
			time.Sleep(5 * time.Second)
			continue
		}
		if pair == nil {
			// key not exist
			time.Sleep(5 * time.Second)
			continue
		}
		if meta.LastIndex > lastIndex {
			val := string(pair.Value)
			s.mu.Lock()
			s.lastValue = val
			listenersCopy := make([]func(), len(s.listeners))
			copy(listenersCopy, s.listeners)
			s.mu.Unlock()

			for _, l := range listenersCopy {
				go l()
			}
			lastIndex = meta.LastIndex
		}
	}
}

// Close stops the background watcher (optional)
func (s *ConsulSubscriber) Close() {
	if s.cancel != nil {
		s.cancel()
	}
}

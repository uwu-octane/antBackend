package consulmanager

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	consul "github.com/hashicorp/consul/api"
	"github.com/zeromicro/go-zero/core/logx"
)

// Target 内部用 atomic.Value，统一存 *url.URL（允许为已类型化的 nil）
type Target struct{ v atomic.Value }

// write type-safe nil-able *url.URL, avoid uninialized panic
func NewTarget() *Target {
	t := &Target{}
	t.v.Store((*url.URL)(nil)) // 已类型化的 nil
	return t
}

func (t *Target) LoadOK() (*url.URL, bool) {
	u := t.v.Load().(*url.URL)
	return u, u != nil
}

func (t *Target) Store(u *url.URL) {
	t.v.Store(u)
}

func (t *Target) String() string {
	if u, ok := t.LoadOK(); ok {
		return u.String()
	}
	return "<nil>"
}

type ManagerOption struct {
	Address    string
	Scheme     string
	Datacenter string
	Token      string
	WaitTime   time.Duration
}

type Manager struct {
	client   *consul.Client
	waitTime time.Duration
	targets  sync.Map // service name -> *Target
}

func NewManager(opt *ManagerOption) (*Manager, error) {
	cfg := consul.DefaultConfig()
	addr := opt.Address
	if addr != "" {
		if _, _, err := net.SplitHostPort(addr); err != nil {
			cfg.Address = addr
			cfg.Scheme = opt.Scheme
		} else {
			cfg.Address = addr
		}
	}
	if opt.Datacenter != "" {
		cfg.Datacenter = opt.Datacenter
	}
	if opt.Token != "" {
		cfg.Token = opt.Token
	}
	if opt.Scheme != "" {
		cfg.Scheme = opt.Scheme
	}
	c, err := consul.NewClient(cfg)
	if err != nil {
		logx.Errorw("new consul client failed err: %v", logx.Field("error", err))
		return nil, err
	}
	return &Manager{client: c, waitTime: opt.WaitTime}, nil

}

func (m *Manager) Watch(ctx context.Context, service string) *Target {
	actual, loaded := m.targets.LoadOrStore(service, NewTarget())
	t := actual.(*Target)
	if !loaded {
		go m.watchOne(ctx, service, t)
	}
	return t

}

func (m *Manager) GetTarget(service string) *url.URL {
	if v, ok := m.targets.Load(service); ok {
		if u, ok2 := v.(*Target).LoadOK(); ok2 {
			return u
		}
		return nil
	}
	return nil
}

func (m *Manager) ListServices() map[string]string {
	out := map[string]string{}
	m.targets.Range(func(key, value any) bool {
		svc := key.(string)
		if u, ok := value.(*Target).LoadOK(); ok {
			out[svc] = u.String()
		} else {
			out[svc] = ""
		}
		return true
	})
	return out
}

func (m *Manager) watchOne(ctx context.Context, service string, tgt *Target) {
	logx.Infof("[consulmgr] watch one service: %s", service)
	health := m.client.Health()
	var waitIndex uint64
	backoff := time.Second
	for {
		select {
		case <-ctx.Done():
			logx.Infof("[consulmgr] watch stop: %s done", service)
			return
		default:
		}
		q := &consul.QueryOptions{
			WaitIndex: waitIndex,
			WaitTime:  m.waitTime,
		}
		entries, meta, err := health.Service(service, "", true, q)
		if err != nil {
			logx.Errorf("[consulmgr] watch one service: %s err: %v retrying in %v", service, err, backoff)
			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
			}
			backoff *= 2
			if backoff > 30*time.Second {
				backoff = 30 * time.Second
			}
			continue
		}
		backoff = time.Second
		if meta != nil {
			waitIndex = meta.LastIndex
		}

		if len(entries) == 0 {
			tgt.Store(nil)
			logx.Infof("[consulmgr] watch one service: %s no passing instances, cleared", service)
			continue
		}
		addr := entries[0].Service.Address
		if addr == "" {
			addr = entries[0].Node.Address
		}
		u := &url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("%s:%d", addr, entries[0].Service.Port),
		}
		tgt.Store(u)
		logx.Infof("[consulmgr] watch one service: %s updated to: %s", service, u.String())
	}
}

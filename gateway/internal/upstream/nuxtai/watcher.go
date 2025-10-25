package nuxtai

import (
	"net/url"
	"sync/atomic"
	"time"

	"github.com/hashicorp/consul/api"
)

type Target atomic.Value

func (t *Target) Store(u *url.URL) { (*atomic.Value)(t).Store(u) }
func (t *Target) Load() *url.URL {
	v := (*atomic.Value)(t).Load()
	if v == nil {
		return nil
	}
	return v.(*url.URL)
}

type WatcherOption struct {
	Address     string
	Scheme      string
	Datencenter string
	Token       string
	WaitTime    time.Duration
	Service     string
}

type Watcher struct {
	consul   *api.Client
	service  string
	waitTime time.Duration
	target   *Target
}

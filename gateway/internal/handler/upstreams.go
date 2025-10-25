package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/uwu-octane/antBackend/gateway/internal/svc"
	"github.com/uwu-octane/antBackend/gateway/internal/upstream/consulmanager"
	"github.com/zeromicro/go-zero/core/logx"
)

// upstreamProxy 保存一个 prefix -> proxy handler
type upstreamProxy struct {
	Prefix  string
	Handler http.Handler
	Name    string
}

var (
	svctx           *svc.ServiceContext
	upstreamProxies []upstreamProxy
)

// 初始化 upstream proxies
func InitUpstreamProxies(s *svc.ServiceContext) {
	svctx = s
	var proxies []upstreamProxy

	for _, up := range s.Config.Upstreams {
		target := s.Targets[up.Name]
		if target == nil {
			logx.Errorf("upstream %s not found in ServiceContext.Targets", up.Name)
			continue
		}

		timeout := time.Duration(up.TimeoutMS) * time.Millisecond
		if timeout == 0 {
			timeout = 5 * time.Second
		}

		strip := up.StripPrefix
		// 不再自动设置 strip 为 PathPrefix，保持原始配置

		getTarget := func() *url.URL {
			if u, ok := target.LoadOK(); ok {
				return u
			}
			return nil
		}

		proxy := consulmanager.NewDynamicProxy(getTarget, &consulmanager.ProxyOption{
			StripPrefix: strip,
			PassHeaders: up.PassHeaders,
			Timeout:     timeout,
		})

		prefix := up.PathPrefix
		if !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		if !strings.HasPrefix(prefix, "/") {
			prefix = "/" + prefix
		}

		proxies = append(proxies, upstreamProxy{
			Prefix:  prefix,
			Handler: proxy,
			Name:    up.Name,
		})

		logx.Infof("[gateway] wired upstream: name=%s service=%s prefix=%s strip=%s",
			up.Name, up.Service, prefix, strip)
	}

	// 优先匹配更长的前缀（防止 /api/ 抢到 /api/ai/）
	sort.Slice(proxies, func(i, j int) bool {
		return len(proxies[i].Prefix) > len(proxies[j].Prefix)
	})

	upstreamProxies = proxies

	for _, p := range upstreamProxies {
		logx.Infof("[upstream] wired prefix=%s name=%s", p.Prefix, p.Name)
	}
}

// 在 NotFoundHandler 中调用
func UpstreamEntry(w http.ResponseWriter, r *http.Request) bool {
	path := r.URL.Path
	logx.Infof("[upstream] entry path=%s", path)
	for _, u := range upstreamProxies {
		if strings.HasPrefix(path, u.Prefix) {
			target := svctx.Targets[u.Name]
			if target == nil {
				http.Error(w, fmt.Sprintf("no target for %s", u.Name), http.StatusServiceUnavailable)
				return true
			}
			if _, ok := target.LoadOK(); !ok {
				http.Error(w, fmt.Sprintf("no healthy instance for %s", u.Name), http.StatusServiceUnavailable)
				return true
			}
			u.Handler.ServeHTTP(w, r)
			return true
		}
	}
	logx.Infof("[upstream] no prefix matched for path=%s", path)
	return false
}

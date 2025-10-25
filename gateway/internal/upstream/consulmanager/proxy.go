package consulmanager

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProxyOption struct {
	StripPrefix string
	PassHeaders []string
	Timeout     time.Duration
}

func NewDynamicProxy(getTarget func() *url.URL, opt *ProxyOption) http.Handler {
	proxy := &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			target := getTarget()
			if target == nil {
				logx.Errorf("proxy: target not found")
				return
			}
			r.URL.Scheme = target.Scheme
			r.URL.Host = target.Host

			if opt.StripPrefix != "" && strings.HasPrefix(r.URL.Path, opt.StripPrefix) {
				r.URL.Path = strings.TrimPrefix(r.URL.Path, opt.StripPrefix)
				if r.URL.Path == "" {
					r.URL.Path = "/"
				}
			}

			pass := map[string]struct{}{}
			for _, h := range opt.PassHeaders {
				pass[strings.ToLower(h)] = struct{}{}
			}
			for k, v := range r.Header {
				if _, ok := pass[strings.ToLower(k)]; !ok {
					continue
				}
				r.Header[k] = v
			}

			if r.Header.Get("X-Forwarded-Host") == "" {
				r.Header.Set("X-Forwarded-Host", r.Host)
			}
			if r.Header.Get("X-Forwarded-Proto") == "" {
				r.Header.Set("X-Forwarded-Proto", r.URL.Scheme)
			}
		},

		FlushInterval: 100 * time.Millisecond,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			logx.Errorf("proxy error: %v", err)
			http.Error(w, "proxy error", http.StatusInternalServerError)
		},
	}
	proxy.Transport = &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return proxy
}

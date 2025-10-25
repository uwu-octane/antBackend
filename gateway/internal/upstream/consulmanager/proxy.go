package consulmanager

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"net"

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

			// Set X-Forwarded-* headers
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			if ip == "" {
				ip = r.RemoteAddr
			}
			if xff := r.Header.Get("X-Forwarded-For"); xff == "" {
				r.Header.Set("X-Forwarded-For", ip)
			} else {
				r.Header.Set("X-Forwarded-For", xff+", "+ip)
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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 只对可能有请求体的方法做探针
		switch r.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch:
			if r.Body != nil {
				b, _ := io.ReadAll(r.Body)
				_ = r.Body.Close()

				// 打印长度与前 128 字节（避免刷屏）
				frag := b
				if len(frag) > 128 {
					frag = frag[:128]
				}
				logx.Infof("[proxy probe] method=%s path=%s len=%d frag=%q",
					r.Method, r.URL.Path, len(b), string(frag))

				// 复位给下游
				r.Body = io.NopCloser(bytes.NewReader(b))
				r.ContentLength = int64(len(b))
				if len(b) == 0 {
					// 没长度则交给 chunked
					r.ContentLength = -1
					r.TransferEncoding = []string{"chunked"}
				}
			}
		}

		// 请求级超时（可选）
		if opt.Timeout > 0 {
			ctx, cancel := context.WithTimeout(r.Context(), opt.Timeout)
			defer cancel()
			r = r.WithContext(ctx)
		}

		proxy.ServeHTTP(w, r)
	})
}

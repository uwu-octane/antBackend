package handler

import (
	"net/http"
	"net/url"
	"time"

	"github.com/uwu-octane/antBackend/gateway/internal/svc"
	"github.com/uwu-octane/antBackend/gateway/internal/upstream/consulmanager"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func RegisterRoutesUpstream(server *rest.Server, serverCtx *svc.ServiceContext) {
	if serverCtx.ConsulManager == nil {
		logx.Errorw("[routesupstream] Consul manager not initialized, skip dynamic upstream routes")
		return
	}

	server.AddRoutes(
		[]rest.Route{
			{
				Method: http.MethodGet,
				Path:   "/internal/upstreams",
				Handler: func(w http.ResponseWriter, r *http.Request) {
					httpx.OkJsonCtx(r.Context(), w, map[string]any{
						"services": serverCtx.ConsulManager.ListServices(), // serviceName -> target string
					})
				},
			},
		},
	)
	for _, upstream := range serverCtx.Config.Upstreams {
		tgt := serverCtx.Targets[upstream.Name]
		if tgt == nil {
			continue
		}

		timeout := time.Duration(upstream.TimeoutMS) * time.Millisecond
		proxy := consulmanager.NewDynamicProxy(
			func() *url.URL {
				u, ok := tgt.LoadOK()
				if !ok {
					return nil
				}
				return u
			}, &consulmanager.ProxyOption{
				StripPrefix: upstream.StripPrefix,
				PassHeaders: upstream.PassHeaders,
				Timeout:     timeout,
			},
		)

		methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodOptions, http.MethodHead}

		for _, method := range methods {
			server.AddRoutes(
				[]rest.Route{
					{
						Method:  method,
						Path:    upstream.PathPrefix + "*any",
						Handler: proxy.ServeHTTP,
					},
				},
			)
		}
	}
}

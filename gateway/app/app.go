package app

import (
	"fmt"
	"net/http"

	"github.com/uwu-octane/antBackend/gateway/internal/config"
	"github.com/uwu-octane/antBackend/gateway/internal/handler"
	"github.com/uwu-octane/antBackend/gateway/internal/middleware"
	"github.com/uwu-octane/antBackend/gateway/internal/svc"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
)

func BuildGatewayServer(configFile string) (service.Service, func(), error) {
	var c config.Config
	conf.MustLoad(configFile, &c, conf.UseEnv())

	server := rest.MustNewServer(
		c.RestConf,
		rest.WithCors(),
		rest.WithNotFoundHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if handler.UpstreamEntry(w, r) {
				return
			}
			http.NotFound(w, r)
		})),
		rest.WithFileServer("/schema", http.Dir("./docs/openapi")),
	)

	ctx := svc.NewServiceContext(c)
	server.Use(middleware.NewRequestID().Handle)
	server.Use(middleware.NewJwt(ctx).Handle)
	server.Use(middleware.NewPathNormalize(c.ApiPrefix, c.ApiCanonicalPrefix).Handle)
	handler.RegisterHandlers(server, ctx)
	handler.InitUpstreamProxies(ctx)

	//handler.RegisterRoutesUpstream(server, ctx)
	fmt.Println(server.Routes())
	cleanup := func() {}

	return server, cleanup, nil
}

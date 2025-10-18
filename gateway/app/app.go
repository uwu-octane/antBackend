package app

import (
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"

	"github.com/uwu-octane/antBackend/gateway/internal/config"
	"github.com/uwu-octane/antBackend/gateway/internal/handler"
	"github.com/uwu-octane/antBackend/gateway/internal/middleware"
	"github.com/uwu-octane/antBackend/gateway/internal/svc"
)

func BuildGatewayServer(configFile string) (service.Service, func(), error) {
	var c config.Config
	conf.MustLoad(configFile, &c, conf.UseEnv())

	server := rest.MustNewServer(
		c.RestConf,
		rest.WithCors(),
	)

	ctx := svc.NewServiceContext(c)
	server.Use(middleware.NewRequestID().Handle)
	server.Use(middleware.NewJwt(ctx).Handle)
	server.Use(middleware.NewPathNormalize(c.ApiPrefix, c.ApiCanonicalPrefix).Handle)

	handler.RegisterHandlers(server, ctx)

	cleanup := func() {}

	return server, cleanup, nil
}

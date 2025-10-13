// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package main

import (
	"flag"
	"fmt"

	"github.com/uwu-octane/antBackend/gateway/internal/config"
	"github.com/uwu-octane/antBackend/gateway/internal/handler"
	"github.com/uwu-octane/antBackend/gateway/internal/middleware"
	"github.com/uwu-octane/antBackend/gateway/internal/svc"

	"github.com/uwu-octane/antBackend/common/envloader"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/gateway-api.yaml", "the config file")

func main() {
	flag.Parse()
	envloader.Load()
	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	server := rest.MustNewServer(c.RestConf,
		rest.WithCors(),
	)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	server.Use(middleware.NewRequestID().Handle)
	server.Use(middleware.NewJwt(ctx).Handle)
	server.Use(middleware.NewPathNormalize(c.ApiPrefix, c.ApiCanonicalPrefix).Handle)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}

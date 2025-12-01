package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/uwu-octane/antBackend/gateway/internal/config"
	"github.com/uwu-octane/antBackend/gateway/internal/handler"
	"github.com/uwu-octane/antBackend/gateway/internal/middleware"
	"github.com/uwu-octane/antBackend/gateway/internal/svc"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
	"go.opentelemetry.io/otel"
)

func BuildGatewayServer(configFile string) (service.Service, func(), error) {
	var c config.Config
	conf.MustLoad(configFile, &c, conf.UseEnv())
	logx.Infof("Telemetry config: %+v", c.Telemetry)
	go func() {
		// 稍微等一下，确保 trace.StartAgent 已经执行
		time.Sleep(2 * time.Second)

		tracer := otel.Tracer("gateway-test")
		ctx, span := tracer.Start(context.Background(), "gateway-startup-test")
		// 模拟一点点耗时
		time.Sleep(200 * time.Millisecond)
		span.End()

		logx.Info("sent test span: gateway-startup-test")
		_ = ctx
	}()
	server := rest.MustNewServer(
		c.RestConf,
		rest.WithNotFoundHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if handler.UpstreamEntry(w, r) {
				return
			}
			http.NotFound(w, r)
		})),
		rest.WithFileServer("/schema", http.Dir("./docs/openapi")),
	)
	cors := middleware.NewCors(
		middleware.WithAllowCredentials(), // ① 允许带 Cookie
		middleware.WithAllowedOrigins([]string{
			"http://localhost:5173", // ② 允许的前端域名
		}),
		middleware.WithAllowedMethods([]string{
			"GET", "POST", "PUT", "DELETE", "OPTIONS",
		}),
		middleware.WithAllowedHeaders([]string{
			"Content-Type", "Authorization",
		}),
	)

	server.Use(cors.Handle)

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

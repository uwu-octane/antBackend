// auth/app/app.go
package app

import (
	"log"

	pb "github.com/uwu-octane/antBackend/api/v1/auth"
	"github.com/uwu-octane/antBackend/auth/internal/config"
	"github.com/uwu-octane/antBackend/auth/internal/server"
	"github.com/uwu-octane/antBackend/auth/internal/svc"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/zeromicro/zero-contrib/zrpc/registry/consul"
)

// BuildAuthRpcServer 只负责构造 zrpc.Server，不调用 Start()。
// 返回值：service.Service（可被 ServiceGroup 管理）、cleanup 回调用于资源释放。
func BuildAuthRpcServer(configFile string) (service.Service, func(), error) {
	var c config.Config
	// 开启环境变量覆盖（.env + UseEnv）
	conf.MustLoad(configFile, &c, conf.UseEnv())

	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		// 注册 gRPC 服务实现
		pb.RegisterAuthServiceServer(grpcServer, server.NewAuthServiceServer(ctx))

		// 本地开发/测试暴露 gRPC 反射
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})

	if err := consul.RegisterService(c.ListenOn, c.Consul); err != nil {
		log.Printf("consul register failed: %v", err)
		return nil, nil, err
	}

	cleanup := func() {
	}

	return s, cleanup, nil
}

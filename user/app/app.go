package app

import (
	"context"

	"github.com/uwu-octane/antBackend/api/v1/user"
	"github.com/uwu-octane/antBackend/user/internal/config"
	"github.com/uwu-octane/antBackend/user/internal/logic"
	"github.com/uwu-octane/antBackend/user/internal/server"
	"github.com/uwu-octane/antBackend/user/internal/svc"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"github.com/zeromicro/zero-contrib/zrpc/registry/consul"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func BuildUserRpcServer(configFile string) (service.Service, func(), error) {
	var c config.Config
	conf.MustLoad(configFile, &c, conf.UseEnv())

	ctx := svc.NewServiceContext(c)
	_ = logic.NewDebugPubLogic(context.Background(), ctx).Test()
	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		user.RegisterUserServiceServer(grpcServer, server.NewUserServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})

	if err := consul.RegisterService(c.ListenOn, c.Consul); err != nil {
		return nil, nil, err
	}

	cleanup := func() {

	}

	return s, cleanup, nil
}

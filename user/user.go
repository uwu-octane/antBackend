package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/uwu-octane/antBackend/api/v1/user"
	"github.com/uwu-octane/antBackend/user/internal/config"
	"github.com/uwu-octane/antBackend/user/internal/server"
	"github.com/uwu-octane/antBackend/user/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/zeromicro/zero-contrib/zrpc/registry/consul"
)

var configFile = flag.String("f", "etc/user.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		user.RegisterUserServiceServer(grpcServer, server.NewUserServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})

	if err := consul.RegisterService(c.ListenOn, c.Consul); err != nil {
		log.Fatal(err)
	}
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}

package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/uwu-octane/antBackend/api/v1/auth"
	"github.com/uwu-octane/antBackend/auth/internal/config"
	"github.com/uwu-octane/antBackend/auth/internal/server"
	"github.com/uwu-octane/antBackend/auth/internal/svc"
	"github.com/uwu-octane/antBackend/common/envloader"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/uwu-octane/antBackend/common/consulsubscriber"
	configcenter "github.com/zeromicro/go-zero/core/configcenter"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/zero-contrib/zrpc/registry/consul"
)

var configFile = flag.String("f", "etc/auth.yaml", "the config file")

func main() {
	envloader.Load()
	flag.Parse()

	// Load minimal static config to get config center info
	var staticConf struct {
		ConfigCenter struct {
			Type string `json:"type" yaml:"type"`
			Host string `json:"host" yaml:"host"`
			Key  string `json:"key" yaml:"key"`
		} `json:"configCenter" yaml:"configCenter"`
	}
	conf.MustLoad(*configFile, &staticConf, conf.UseEnv())

	if staticConf.ConfigCenter.Type != "consul" {
		panic("unsupported config center type: " + staticConf.ConfigCenter.Type)
	}

	sub := consulsubscriber.MustNewConsulSubscriber(staticConf.ConfigCenter.Host, staticConf.ConfigCenter.Key)
	cc := configcenter.MustNewConfigCenter[config.Config](configcenter.Config{
		Type: "yaml",
	}, sub)

	cfg, err := cc.GetConfig()
	if err != nil {
		panic(fmt.Sprintf("load config from consul failed: %v", err))
	}
	logx.Infof("Loaded config: %+v", cfg)

	cc.AddListener(func() {
		newCfg, err2 := cc.GetConfig()
		if err2 != nil {
			logx.Error("config reload failed:", err2)
			return
		}
		logx.Infof("Config changed: %+v", newCfg)
		// You can add custom hot‐reload logic here, e.g. reset Redis connection etc.
	})

	ctx := svc.NewServiceContext(cfg)

	s := zrpc.MustNewServer(cfg.RpcServerConf, func(grpcServer *grpc.Server) {
		auth.RegisterAuthServiceServer(grpcServer, server.NewAuthServiceServer(ctx))

		if cfg.Mode == service.DevMode || cfg.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})

	if err := consul.RegisterService(cfg.ListenOn, cfg.Consul); err != nil {
		log.Fatal(err)
	}
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", cfg.ListenOn)
	s.Start()
}

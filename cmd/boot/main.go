// cmd/boot/main.go
package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"

	"github.com/uwu-octane/antBackend/common/envloader"

	// 导入三个 BuildXxxServer
	antAuth "github.com/uwu-octane/antBackend/auth/app"
	antGateway "github.com/uwu-octane/antBackend/gateway/app" // BuildGatewayServer 在 package main 时可用别名导入
	antUser "github.com/uwu-octane/antBackend/user/app"
)

var (
	gatewayConf = flag.String("gateway", "gateway/etc/gateway-api.yaml", "gateway config file")
	authConf    = flag.String("auth", "auth/etc/auth.yaml", "auth rpc config file")
	userConf    = flag.String("user", "user/etc/user.yaml", "user rpc config file")
)

func mustFile(path string) string {
	if path == "" {
		logx.Errorw("config file is required", logx.Field("path", path))
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		logx.Errorw("failed to get absolute path", logx.Field("path", path), logx.Field("error", err))
	}
	st, err := os.Stat(abs)
	if err != nil || st.IsDir() {
		logx.Errorw("config file is not a file", logx.Field("path", path))
	}
	return abs
}

func main() {
	envloader.Load()
	flag.Parse()

	configs := map[string]string{}
	gatewayCfg := mustFile(*gatewayConf)
	authCfg := mustFile(*authConf)
	userCfg := mustFile(*userConf)

	configs["gateway"] = gatewayCfg
	configs["auth"] = authCfg
	configs["user"] = userCfg
	group := service.NewServiceGroup()

	// 1) gateway
	gatewaySrv, gatewayCleanup, err := antGateway.BuildGatewayServer(gatewayCfg)
	if err != nil {
		log.Fatalf("build gateway: %v", err)
	}
	defer safeCleanup("gateway", gatewayCleanup)
	group.Add(gatewaySrv)

	// 2) auth rpc
	authSrv, authCleanup, err := antAuth.BuildAuthRpcServer(authCfg)
	if err != nil {
		log.Fatalf("build auth rpc: %v", err)
	}
	defer safeCleanup("auth", authCleanup)
	group.Add(authSrv)

	// 3) user rpc
	userSrv, userCleanup, err := antUser.BuildUserRpcServer(userCfg)
	if err != nil {
		log.Fatalf("build user rpc: %v", err)
	}
	defer safeCleanup("user", userCleanup)
	group.Add(userSrv)

	//start and stop
	printServerInfo(configs)
	defer group.Stop()
	group.Start()
}

func safeCleanup(name string, cleanup func()) {
	if cleanup == nil {
		return
	}
	defer func() {
		if err := recover(); err != nil {
			logx.Errorw("failed to cleanup", logx.Field("name", name), logx.Field("error", err))
		}
	}()
	cleanup()
}

func printServerInfo(configs map[string]string) {
	for name, cfg := range configs {
		logx.Infow("server info", logx.Field("name", name), logx.Field("config", cfg))
		logx.Infow("server info", logx.Field("name", name), logx.Field("config", cfg))
		logx.Infow("server info", logx.Field("name", name), logx.Field("config", cfg))
	}
}

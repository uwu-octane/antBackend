// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package svc

import (
	"auth/authservice"
	"gateway/internal/config"
	"user/userservice"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config  config.Config
	AuthRpc authservice.AuthService
	UserRpc userservice.UserService
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:  c,
		AuthRpc: authservice.NewAuthService(zrpc.MustNewClient(c.AuthRpc)),
		UserRpc: userservice.NewUserService(zrpc.MustNewClient(c.UserRpc)),
	}
}

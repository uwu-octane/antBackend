// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	AuthRpc zrpc.RpcClientConf `json:"AuthRpc"`
	UserRpc zrpc.RpcClientConf `json:"UserRpc"`
	Auth    AuthConfig         `json:"Auth"`
	JwtAuth JwtAuthConfig      `json:"JwtAuth"`

	Cors               []string `json:"Cors"`
	ApiPrefix          []string `json:"ApiPrefix"`
	ApiCanonicalPrefix string   `json:"ApiCanonicalPrefix"`
}

type AuthConfig struct {
	Strict       bool
	TokenLookup  string
	IgnoreRoutes []string
}

type JwtAuthConfig struct {
	Secret        string
	Issuer        string
	LeewaySeconds int64
}

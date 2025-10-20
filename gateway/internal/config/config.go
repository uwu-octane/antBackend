// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	AuthRpc zrpc.RpcClientConf `json:"AuthRpc"`
	UserRpc zrpc.RpcClientConf `json:"UserRpc"`
	Auth    AuthConfig         `json:"Auth"`
	//JwtAuth   JwtAuthConfig      `json:"JwtAuth"`
	RateLimit RateLimitConfig `json:"RateLimit"`

	Cors               []string `json:"Cors"`
	ApiPrefix          []string `json:"ApiPrefix"`
	ApiCanonicalPrefix string   `json:"ApiCanonicalPrefix"`
}

type AuthConfig struct {
	Strict        bool
	TokenLookup   string
	AccessSecret  string
	AccessExpire  int64
	Issuer        string
	LeewaySeconds int64
	IgnoreRoutes  []string
}

// type JwtAuthConfig struct {
// 	AccessSecret  string
// 	Issuer        string
// 	LeewaySeconds int64
// }

type RateLimitConfig struct {
	Enable         bool
	WindowSeconds  int
	MaxAttempts    int
	By             string
	RateLimitRedis redis.RedisKeyConf
}

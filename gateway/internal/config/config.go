// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package config

import (
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	GatewayMode string             `json:"GatewayMode"`
	AuthRpc     zrpc.RpcClientConf `json:"AuthRpc"`
	UserRpc     zrpc.RpcClientConf `json:"UserRpc"`
	Auth        AuthConfig         `json:"Auth"`
	Upstreams   []UpstreamConfig   `json:"Upstreams"`
	Consul      ConsulConf         `json:"Consul"`
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

type UpstreamConfig struct {
	Name        string
	Service     string
	PathPrefix  string
	StripPrefix string
	TimeoutMS   int
	PassHeaders []string
}

type ConsulConf struct {
	Address    string        `json:",optional"` //nolint:staticcheck // go-zero specific JSON tag
	Scheme     string        `json:",optional"` //nolint:staticcheck // go-zero specific JSON tag
	Datacenter string        `json:",optional"` //nolint:staticcheck // go-zero specific JSON tag
	Token      string        `json:",optional"` //nolint:staticcheck // go-zero specific JSON tag
	WaitTime   time.Duration `json:",optional"` //nolint:staticcheck // go-zero specific JSON tag
}

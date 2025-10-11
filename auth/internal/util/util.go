package util

import "fmt"

type RedisKeyType string

const (
	RedisKeyTypeRefresh RedisKeyType = "refresh"
	RedisKeyTypeReuse   RedisKeyType = "reuse"
	RedisKeyTypeAccess  RedisKeyType = "access"
)

func RedisKey(key string, keyType RedisKeyType, jti string) string {
	return fmt.Sprintf("%s:%s:%s", key, keyType, jti)
}

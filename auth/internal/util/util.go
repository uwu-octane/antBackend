package util

import (
	"strings"
)

type RedisKeyType string

const (
	RedisKeyTypeRefresh RedisKeyType = "refresh"
	RedisKeyTypeReuse   RedisKeyType = "reuse"
	RedisKeyTypeAccess  RedisKeyType = "access"

	RedisKeyTypeUserSids RedisKeyType = "user" // user:<uid>:sids
	RedisKeyTypeSidSet   RedisKeyType = "sid"  //  sid:<sid>
	RedisKeyTypeJtiSid   RedisKeyType = "jti_sid"
)

func NormalizePrefix(p string) string {
	if p == "" {
		return ""
	}
	if strings.HasSuffix(p, ":") {
		return p
	}
	return p + ":"
}

func RedisKey(prefix string, typ RedisKeyType, id string) string {
	p := NormalizePrefix(prefix)
	switch typ {
	case RedisKeyTypeRefresh:
		return p + "refresh:" + id
	case RedisKeyTypeReuse:
		return p + "reuse:" + id
	case RedisKeyTypeJtiSid:
		return p + "jti_sid:" + id
	default:
		return p + string(typ) + ":" + id
	}
}

func UserSidsKey(prefix, uid string) string { // auth:user:<uid>:sids
	return NormalizePrefix(prefix) + "user:" + uid + ":sids"
}
func SidSetKey(prefix, sid string) string { // auth:sid:<sid>
	return NormalizePrefix(prefix) + "sid:" + sid
}
func JtiSidKey(prefix, jti string) string { // auth:jti_sid:<jti>
	return NormalizePrefix(prefix) + "jti_sid:" + jti
}

package util

import (
	"net"
	"net/http"
	"strings"
)

func ClientIP(r *http.Request) string {
	// X-Forwarded-For: client, proxy1, proxy2
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if ip := strings.TrimSpace(parts[0]); ip != "" {
			return ip
		}
	}
	if rip := r.Header.Get("X-Real-IP"); rip != "" {
		return rip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	// If SplitHostPort fails, return the whole RemoteAddr
	return r.RemoteAddr
}

func MakeLoginLimitKey(by, username, ip string) string {
	u := strings.ToUpper(strings.TrimSpace(username))
	i := strings.TrimSpace(ip)
	switch strings.ToLower(strings.TrimSpace(by)) {
	case "username":
		if u == "" {
			return "u::" + i
		}
		return "u:" + u
	case "ip+username":
		if u == "" {
			return "iu:" + i + "::"
		}
		return "iu:" + i + ":" + u
	default: // ip
		return "i:" + i
	}
}

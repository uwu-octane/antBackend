package middleware

import (
	"net/http"
	"strings"

	"google.golang.org/grpc/metadata"
)

const (
	cookieRefreshName = "refresh"
	mdRefreshName     = "x-refresh-token"
)

func NewGrpcMetaMiddleware() func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			refresh := readCookie(r, cookieRefreshName)
			if refresh != "" {
				inject := metadata.Pairs(mdRefreshName, strings.TrimSpace(refresh))
				if old, ok := metadata.FromOutgoingContext(r.Context()); ok {
					inject = metadata.Join(old, inject)
				}

				ctx := metadata.NewOutgoingContext(r.Context(), inject)
				r = r.WithContext(ctx)
			}
			next(w, r)
		}
	}
}

func readCookie(r *http.Request, name string) string {
	c, err := r.Cookie(name)
	if err != nil || c == nil {
		return ""
	}
	return c.Value
}

package middleware

import (
	"context"
	"gateway/internal/svc"
	"net/http"
	"strings"
)

type Jwt struct {
	svcCtx *svc.ServiceContext
}

func NewJwt(svcCtx *svc.ServiceContext) *Jwt {
	return &Jwt{
		svcCtx: svcCtx,
	}
}

// ctxKey is used to avoid collisions in context values.
// Consumers can use TokenFromContext to read the token injected by this middleware.
type ctxKey string

const ctxKeyToken ctxKey = "jwtToken"

func TokenFromContext(ctx context.Context) string {
	val, _ := ctx.Value(ctxKeyToken).(string)
	return val
}

func (m *Jwt) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		for _, ignorePath := range m.svcCtx.Config.Auth.IgnoreRoutes {
			if strings.HasPrefix(path, ignorePath) {
				next(w, r)
				return
			}
		}

		token := m.lookupToken(r)
		if token == "" {
			if m.svcCtx.Config.Auth.Strict {
				http.Error(w, "missing or invalid token", http.StatusUnauthorized)
				return
			}
			// Non-strict mode: continue without token
			next(w, r)
			return
		}

		// Attach token to request context for downstream usage
		ctx := context.WithValue(r.Context(), ctxKeyToken, token)
		next(w, r.WithContext(ctx))
	}
}

func (m *Jwt) lookupToken(r *http.Request) string {
	lookup := strings.TrimSpace(m.svcCtx.Config.Auth.TokenLookup)
	if lookup == "" {
		// default to Authorization header bearer token
		return extractBearer(r.Header.Get("Authorization"))
	}

	parts := strings.SplitN(lookup, ":", 2)
	if len(parts) != 2 {
		return ""
	}
	source := strings.TrimSpace(strings.ToLower(parts[0]))
	name := strings.TrimSpace(parts[1])

	switch source {
	case "header":
		if strings.EqualFold(name, "Authorization") {
			return extractBearer(r.Header.Get(name))
		}
		return r.Header.Get(name)
	case "query":
		return r.URL.Query().Get(name)
	case "cookie":
		if c, err := r.Cookie(name); err == nil {
			return c.Value
		}
	}
	return ""
}

func extractBearer(h string) string {
	if h == "" {
		return ""
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1]
	}
	return ""
}

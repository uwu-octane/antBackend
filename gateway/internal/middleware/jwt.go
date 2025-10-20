package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/uwu-octane/antBackend/gateway/internal/svc"
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

const (
	ctxKeyToken ctxKey = "jwtToken"
	ctxUID      ctxKey = "uid"
	ctxJTI      ctxKey = "jti"
	ctxIAT      ctxKey = "iat"
)

func TokenFromContext(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(ctxKeyToken).(string)
	return val, ok && val != ""
}

func UIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ctxUID).(string)
	return v, ok && v != ""
}

func JTIFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ctxJTI).(string)
	return v, ok && v != ""
}

func IATFromContext(ctx context.Context) (int64, bool) {
	v, ok := ctx.Value(ctxIAT).(int64)
	return v, ok
}

type accessCalims struct {
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
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

		tokenStr := m.lookupToken(r)
		if tokenStr == "" {
			if m.svcCtx.Config.Auth.Strict {
				http.Error(w, "missing or invalid token", http.StatusUnauthorized)
				return
			}
			// Non-strict mode: continue without token
			next(w, r)
			return
		}

		parser := jwt.NewParser(
			jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
			jwt.WithIssuer(m.svcCtx.Config.Auth.Issuer),
			jwt.WithLeeway(time.Duration(m.svcCtx.Config.Auth.LeewaySeconds)*time.Second),
		)
		var accessClaims accessCalims
		token, err := parser.ParseWithClaims(tokenStr, &accessClaims, func(token *jwt.Token) (any, error) {
			return []byte(m.svcCtx.Config.Auth.AccessSecret), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		if accessClaims.TokenType != "access" {
			http.Error(w, "wrong token type", http.StatusUnauthorized)
			return
		}

		if accessClaims.Subject == "" || accessClaims.ID == "" {
			http.Error(w, "missing subject or id", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ctxKeyToken, tokenStr)
		ctx = context.WithValue(ctx, ctxUID, accessClaims.Subject)
		ctx = context.WithValue(ctx, ctxJTI, accessClaims.ID)
		if accessClaims.IssuedAt != nil {
			ctx = context.WithValue(ctx, ctxIAT, accessClaims.IssuedAt.Unix())
		}

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

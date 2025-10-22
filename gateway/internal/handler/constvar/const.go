package constvar

const (
	CookieSidName     = "sid"
	CookieRefreshName = "refresh"
	CookiePath        = "/"
)

const (
	HeaderRefreshToken = "x-refresh-token"
)

// ctxKey is used to avoid collisions in context values.
// Consumers can use TokenFromContext to read the token injected by this middleware.

type ctxKey string

const (
	CtxKeyToken ctxKey = "jwtToken"
	CtxUID      ctxKey = "uid"
	CtxJTI      ctxKey = "jti"
	CtxIAT      ctxKey = "iat"
)

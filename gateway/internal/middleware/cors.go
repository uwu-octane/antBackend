package middleware

import (
	"fmt"
	"net/http"
	"strings"
)

type Cors struct {
	allowedOrigins   []string
	allowedMethods   []string
	allowedHeaders   []string
	allowCredentials bool
}

type CorsOption func(*Cors)

func WithAllowCredentials() CorsOption {
	return func(c *Cors) {
		c.allowCredentials = true
	}
}

func WithAllowedOrigins(origins []string) CorsOption {
	return func(c *Cors) {
		c.allowedOrigins = origins
	}
}

func WithAllowedMethods(methods []string) CorsOption {
	return func(c *Cors) {
		c.allowedMethods = methods
	}
}

func WithAllowedHeaders(headers []string) CorsOption {
	return func(c *Cors) {
		c.allowedHeaders = headers
	}
}

func NewCors(opts ...CorsOption) *Cors {
	c := &Cors{
		allowedOrigins:   []string{"*"},
		allowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		allowedHeaders:   []string{"Content-Type", "Authorization"},
		allowCredentials: false,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (m *Cors) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		fmt.Println(">>> CORS Origin:", origin)
		// Check if origin is allowed
		allowed := false
		if len(m.allowedOrigins) > 0 {
			for _, allowedOrigin := range m.allowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}
		}

		if allowed && origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		if m.allowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if len(m.allowedMethods) > 0 {
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(m.allowedMethods, ", "))
		}

		if len(m.allowedHeaders) > 0 {
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(m.allowedHeaders, ", "))
		}

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

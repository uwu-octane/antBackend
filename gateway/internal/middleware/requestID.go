package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type ctxKey string

const RequestIDKey ctxKey = "requestID"

type RequestID struct{}

func NewRequestID() *RequestID { return &RequestID{} }

func (m *RequestID) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get("X-Request-Id")
		if rid == "" {
			rid = uuid.NewString()
			r.Header.Set("X-Request-Id", rid)
		}
		w.Header().Set("X-Request-Id", rid)
		ctx := context.WithValue(r.Context(), RequestIDKey, rid)
		next(w, r.WithContext(ctx))
	}
}

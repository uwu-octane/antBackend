package middleware

import (
	"net/http"

	"github.com/google/uuid"
)

type RequestID struct{}

func NewRequestID() *RequestID { return &RequestID{} }

func (m *RequestID) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get("X-Request-Id")
		if rid == "" {
			rid = uuid.NewString()
		}
		w.Header().Set("X-Request-Id", rid)
		next(w, r.WithContext(r.Context()))
	}
}

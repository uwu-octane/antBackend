package middleware

import (
	"net/http"
	"strings"
)

type PathNormalize struct {
	Prefixes        []string // e.g. ["/api", "/dev-api"]
	CanonicalPrefix string   // e.g. "/api"
}

func NewPathNormalize(prefixes []string, canonical string) *PathNormalize {
	return &PathNormalize{Prefixes: prefixes, CanonicalPrefix: canonical}
}

func (m *PathNormalize) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path

		// 1) 如果命中任一已知前缀，统一转成 CanonicalPrefix
		for _, pre := range m.Prefixes {
			if pre != "" && pre != m.CanonicalPrefix && strings.HasPrefix(p, pre+"/") {
				// 去掉 pre，换成 CanonicalPrefix
				p = m.CanonicalPrefix + strings.TrimPrefix(p, pre)
				r.URL.Path = p
				break
			}
		}

		// 2) 如果直接以 /v1/ 开头（说明前缀被 rewrite 过），补上 CanonicalPrefix
		if strings.HasPrefix(p, "/v1/") {
			r.URL.Path = m.CanonicalPrefix + p // -> /api/v1/...
		}

		next(w, r)
	}
}

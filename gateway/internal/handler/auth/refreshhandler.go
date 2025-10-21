// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package auth

import (
	"net/http"

	"errors"

	"github.com/uwu-octane/antBackend/gateway/internal/logic/auth"
	"github.com/uwu-octane/antBackend/gateway/internal/response"
	"github.com/uwu-octane/antBackend/gateway/internal/svc"
)

const (
	cookieSidName     = "sid"
	cookieRefreshName = "refresh"
)

func RefreshHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sid, _ := readCookie(r, cookieSidName)
		if sid == "" {
			response.FromError(w, errors.New("session id or refresh token is required"))
			return
		}

		l := auth.NewRefreshLogic(r.Context(), svcCtx)
		resp, err := l.Refresh(sid)
		if err != nil {
			response.FromError(w, err)
			return
		} else {
			response.Ok(w, resp)
		}
	}
}

func readCookie(r *http.Request, name string) (string, bool) {
	c, err := r.Cookie(name)
	if err != nil || c == nil {
		return "", false
	}
	return c.Value, true
}

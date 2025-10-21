// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package auth

import (
	"net/http"

	"errors"

	"github.com/uwu-octane/antBackend/gateway/internal/handler/constvar"
	"github.com/uwu-octane/antBackend/gateway/internal/logic/auth"
	"github.com/uwu-octane/antBackend/gateway/internal/response"
	"github.com/uwu-octane/antBackend/gateway/internal/svc"
	"github.com/uwu-octane/antBackend/gateway/util"
	"github.com/zeromicro/go-zero/core/logx"
)

func RefreshHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sid := util.ReadCookie(r, constvar.CookieSidName)
		refresh := util.ReadCookie(r, constvar.CookieRefreshName)
		logx.Infof("refresh session id: %s", sid)
		logx.Infof("refresh refresh token: %s", refresh)

		if sid == "" || refresh == "" {
			response.FromError(w, errors.New("session id or refresh token is required"))
			return
		}

		l := auth.NewRefreshLogic(r.Context(), svcCtx)
		resp, header, err := l.Refresh(sid, refresh)
		if err != nil {
			response.FromError(w, err)
			return
		}
		newRefresh := refresh
		if header != nil {
			vals := header.Get(constvar.HeaderRefreshToken)
			if len(vals) > 0 {
				newRefresh = vals[0]
			}
		}
		logx.Infof("refresh new refresh token: %s", newRefresh)
		SetAuthCookies(w, sid, newRefresh, svcCtx.Config.GatewayMode != "DEV")
		response.Ok(w, resp)
	}
}

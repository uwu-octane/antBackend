// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package handler

import (
	"net/http"

	"github.com/uwu-octane/antBackend/gateway/internal/logic"
	"github.com/uwu-octane/antBackend/gateway/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func PingHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewPingLogic(r.Context(), svcCtx)
		err := l.Ping()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.Ok(w)
		}
	}
}

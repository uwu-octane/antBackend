// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package api

import (
	"net/http"

	"github.com/uwu-octane/antBackend/gateway/internal/logic/api"
	"github.com/uwu-octane/antBackend/gateway/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetUserInfoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := api.NewGetUserInfoLogic(r.Context(), svcCtx)
		resp, err := l.GetUserInfo()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

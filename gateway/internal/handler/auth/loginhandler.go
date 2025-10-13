// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

<<<<<<<< HEAD:gateway/internal/handler/api/refreshhandler.go
package api
========
package auth
>>>>>>>> integrationAuth:gateway/internal/handler/auth/loginhandler.go

import (
	"net/http"

<<<<<<<< HEAD:gateway/internal/handler/api/refreshhandler.go
	"github.com/uwu-octane/antBackend/gateway/internal/logic/api"
========
	"github.com/uwu-octane/antBackend/gateway/internal/grpcerr"
	"github.com/uwu-octane/antBackend/gateway/internal/logic/auth"
>>>>>>>> integrationAuth:gateway/internal/handler/auth/loginhandler.go
	"github.com/uwu-octane/antBackend/gateway/internal/svc"
	"github.com/uwu-octane/antBackend/gateway/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

<<<<<<<< HEAD:gateway/internal/handler/api/refreshhandler.go
func RefreshHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RefreshReq
========
func LoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LoginReq
>>>>>>>> integrationAuth:gateway/internal/handler/auth/loginhandler.go
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

<<<<<<<< HEAD:gateway/internal/handler/api/refreshhandler.go
		l := api.NewRefreshLogic(r.Context(), svcCtx)
		resp, err := l.Refresh(&req)
========
		l := auth.NewLoginLogic(r.Context(), svcCtx)
		resp, err := l.Login(&req)
>>>>>>>> integrationAuth:gateway/internal/handler/auth/loginhandler.go
		if err != nil {
			grpcerr.WriteGrpcError(r, w, err)
			return
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

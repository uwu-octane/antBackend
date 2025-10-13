// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

<<<<<<<< HEAD:gateway/internal/handler/api/loginhandler.go
package api
========
package auth
>>>>>>>> integrationAuth:gateway/internal/handler/auth/refreshhandler.go

import (
	"net/http"

<<<<<<<< HEAD:gateway/internal/handler/api/loginhandler.go
	"github.com/uwu-octane/antBackend/gateway/internal/logic/api"
========
	"github.com/uwu-octane/antBackend/gateway/internal/grpcerr"
	"github.com/uwu-octane/antBackend/gateway/internal/logic/auth"
>>>>>>>> integrationAuth:gateway/internal/handler/auth/refreshhandler.go
	"github.com/uwu-octane/antBackend/gateway/internal/svc"
	"github.com/uwu-octane/antBackend/gateway/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

<<<<<<<< HEAD:gateway/internal/handler/api/loginhandler.go
func LoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LoginReq
========
func RefreshHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RefreshReq
>>>>>>>> integrationAuth:gateway/internal/handler/auth/refreshhandler.go
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

<<<<<<<< HEAD:gateway/internal/handler/api/loginhandler.go
		l := api.NewLoginLogic(r.Context(), svcCtx)
		resp, err := l.Login(&req)
========
		l := auth.NewRefreshLogic(r.Context(), svcCtx)
		resp, err := l.Refresh(&req)
>>>>>>>> integrationAuth:gateway/internal/handler/auth/refreshhandler.go
		if err != nil {
			grpcerr.WriteGrpcError(r, w, err)
			return
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package auth

import (
	"net/http"

	"github.com/uwu-octane/antBackend/gateway/internal/handler/constvar"
	"github.com/uwu-octane/antBackend/gateway/internal/logic/auth"
	"github.com/uwu-octane/antBackend/gateway/internal/response"
	"github.com/uwu-octane/antBackend/gateway/internal/svc"
	"github.com/uwu-octane/antBackend/gateway/util"
	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func LogoutAllHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sid := util.ReadCookie(r, constvar.CookieSidName)
		if sid == "" {
			response.FromError(w, status.Error(codes.InvalidArgument, "logout all: session id is required"))
			return
		}
		l := auth.NewLogoutAllLogic(r.Context(), svcCtx)
		resp, err := l.LogoutAll(sid)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

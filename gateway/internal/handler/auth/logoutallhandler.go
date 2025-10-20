// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package auth

import (
	"net/http"

	"github.com/uwu-octane/antBackend/gateway/internal/logic/auth"
	"github.com/uwu-octane/antBackend/gateway/internal/response"
	"github.com/uwu-octane/antBackend/gateway/internal/svc"
	"github.com/uwu-octane/antBackend/gateway/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func LogoutAllHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LogoutAllReq
		if err := httpx.Parse(r, &req); err != nil {
			response.FromError(w, status.Error(codes.InvalidArgument, "invalid request body"))
			return
		}

		l := auth.NewLogoutAllLogic(r.Context(), svcCtx)
		resp, err := l.LogoutAll(&req)
		if err != nil {
			response.FromError(w, err)
		} else {
			response.Ok(w, resp)
		}
	}
}

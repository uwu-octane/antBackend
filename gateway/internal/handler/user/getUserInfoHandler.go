// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package user

import (
	"net/http"

	"github.com/uwu-octane/antBackend/gateway/internal/logic/user"
	"github.com/uwu-octane/antBackend/gateway/internal/response"
	"github.com/uwu-octane/antBackend/gateway/internal/svc"
)

func GetUserInfoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		l := user.NewGetUserInfoLogic(r.Context(), svcCtx)
		resp, err := l.GetUserInfo()
		if err != nil {
			response.FromError(w, err)
		} else {
			response.Ok(w, resp)
		}
	}
}

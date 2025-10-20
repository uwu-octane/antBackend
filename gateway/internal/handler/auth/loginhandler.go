// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package auth

import (
	"net/http"
	"strconv"

	"github.com/uwu-octane/antBackend/gateway/internal/logic/auth"
	"github.com/uwu-octane/antBackend/gateway/internal/response"
	"github.com/uwu-octane/antBackend/gateway/internal/svc"
	"github.com/uwu-octane/antBackend/gateway/internal/types"
	"github.com/uwu-octane/antBackend/gateway/util"
	"github.com/zeromicro/go-zero/core/limit"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func LoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LoginReq
		if err := httpx.Parse(r, &req); err != nil {
			response.FromError(w, status.Error(codes.InvalidArgument, "invalid request body"))
			return
		}

		// limit login attempts
		if limiter := svcCtx.LoginLimiter; limiter != nil && svcCtx.Config.RateLimit.Enable {
			key := util.MakeLoginLimitKey(svcCtx.Config.RateLimit.By, req.Username, util.ClientIP(r))
			code, err := limiter.TakeCtx(r.Context(), key)
			if err != nil {
				logx.WithContext(r.Context()).Errorf("rate limit check failed: %v", err)
				response.FromError(w, status.Error(codes.Internal, "rate limit check failed"))
				return
			}
			if code == limit.OverQuota {
				retryAfter := svcCtx.Config.RateLimit.WindowSeconds
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				response.FromError(w, status.Error(codes.ResourceExhausted, "too many login attempts"))
				return
			}
		}

		l := auth.NewLoginLogic(r.Context(), svcCtx)
		resp, err := l.Login(&req)
		if err != nil {
			response.FromError(w, err)
			return
		} else {
			response.Ok(w, resp)
		}
	}
}

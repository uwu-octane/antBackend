package grpcerr

import (
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const StatusClientClosedRequest = 499

func WriteGrpcError(r *http.Request, w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		logx.Error("WriteGrpcError: failed to get status from error", "error", err)
		httpx.WriteJsonCtx(r.Context(), w, http.StatusInternalServerError, map[string]any{
			"message": err.Error(),
		})
		return
	}

	code := httpStatusFromGrpc(st.Code())

	httpx.WriteJsonCtx(r.Context(), w, code, map[string]any{
		"message": st.Message(),
	})

}

func httpStatusFromGrpc(c codes.Code) int {
	switch c {
	case codes.InvalidArgument, codes.FailedPrecondition, codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists, codes.Aborted:
		return http.StatusConflict
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.Canceled:
		return StatusClientClosedRequest
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Unavailable:
		return http.StatusBadGateway
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}

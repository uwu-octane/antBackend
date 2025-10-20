package response

import (
	"net/http"

	"github.com/uwu-octane/antBackend/gateway/internal/grpcerr"
	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/status"
)

type Body[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data,omitempty"`
}

func Ok[T any](w http.ResponseWriter, data *T) {
	httpx.OkJson(w, &Body[T]{
		Code: 0,
		Msg:  "success",
		Data: *data,
	})
}

func Fail(w http.ResponseWriter, code int, msg string) {
	httpx.OkJson(w, &Body[any]{
		Code: code,
		Msg:  msg,
		Data: nil,
	})
}

func FromError(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		Fail(w, 10000, "internal error")
		return
	}
	code := grpcerr.AppCodeFromGrpc(st.Code())
	Fail(w, code, st.Message())
}

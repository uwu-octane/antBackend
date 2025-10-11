package dao

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/uwu-octane/antBackend/auth/internal/util"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

//go:embed refresh_rotate.lua
var luaRefreshRotate string

type RotateCode int

const (
	RotateCodeOldNotFound RotateCode = 0
	RotateCodeMismatch    RotateCode = -1
	RotateCodeReused      RotateCode = 2
	RotateCodeOK          RotateCode = 1
)

type RefreshRotateResult struct {
	Code    RotateCode
	Message string
}

func RefreshRotate(ctx context.Context, r *redis.Redis, oldJti string, newJti string, expectUserId string, newTtlSeconds int64, key string) (RefreshRotateResult, error) {
	oldKey := util.RedisKey(key, util.RedisKeyTypeRefresh, oldJti)
	reuseKey := util.RedisKey(key, util.RedisKeyTypeReuse, oldJti)
	newKey := util.RedisKey(key, util.RedisKeyTypeRefresh, newJti)

	replay, err := r.EvalCtx(ctx, luaRefreshRotate, []string{oldKey, reuseKey, newKey},
		[]any{expectUserId, fmt.Sprintf("%d", newTtlSeconds)})

	if err != nil {
		logx.Errorf("refresh rotate failed: %v", err)
		return RefreshRotateResult{}, err
	}

	codeInt, _ := replay.(int64)
	code := RotateCode(codeInt)

	switch code {
	case RotateCodeOK:
		return RefreshRotateResult{Code: RotateCodeOK, Message: "refresh rotate success"}, nil
	case RotateCodeOldNotFound:
		return RefreshRotateResult{Code: RotateCodeOldNotFound, Message: "old jti not found"}, nil
	case RotateCodeMismatch:
		return RefreshRotateResult{Code: RotateCodeMismatch, Message: "user id mismatch"}, nil
	case RotateCodeReused:
		return RefreshRotateResult{Code: RotateCodeReused, Message: "old jti reused"}, nil
	default:
		return RefreshRotateResult{Code: code, Message: "unknown error"}, nil
	}

}

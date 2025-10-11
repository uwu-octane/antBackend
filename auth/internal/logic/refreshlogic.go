package logic

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/uwu-octane/antBackend/api/v1/auth"
	"github.com/uwu-octane/antBackend/auth/internal/dao"
	"github.com/uwu-octane/antBackend/auth/internal/svc"
	"github.com/uwu-octane/antBackend/auth/internal/util"

	"github.com/zeromicro/go-zero/core/logx"
)

var (
	ErrRefreshNotFound = errors.New("refresh token revoked or expired")
	ErrRefreshReused   = errors.New("refresh token reused (possible replay)")
	ErrUserMismatch    = errors.New("refresh user mismatch")
	ErrUnknown         = errors.New("unknown error")
)

type RefreshLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRefreshLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshLogic {
	return &RefreshLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RefreshLogic) SimpleRefresh(in *auth.RefreshReq) (*auth.LoginResp, error) {
	if in.GetRefreshToken() == "" {
		return nil, errors.New("refresh token is required")
	}

	cfg := l.svcCtx.Config.JwtAuth
	tokenHelper := NewTokenHelper([]byte(cfg.Secret), "auth.prc",
		time.Duration(cfg.AccessExpireSeconds)*time.Second, time.Duration(cfg.RefreshExpireSeconds)*time.Second)

	claims, err := tokenHelper.Parse(in.GetRefreshToken())
	if err != nil {
		logx.Errorf("invalid token: %v", err)
		return nil, err
	}

	typ := claims.TokenType
	if typ != "refresh" {
		return nil, errors.New("invalid token type, got " + typ)
	}

	jti := claims.ID
	sub := claims.Subject

	if jti == "" || sub == "" {
		return nil, errors.New("invalid token, got jti: " + jti + " and sub: " + sub)
	}

	//* Delete the old refresh token
	oldKey := util.RedisKey(l.svcCtx.Key, util.RedisKeyTypeRefresh, jti)
	uid, err := l.svcCtx.Redis.Get(oldKey)
	if err != nil || uid == "" {
		return nil, errors.New("refresh token revoked or expired")
	}

	if uid != sub {
		return nil, errors.New("invalid token, got uid: " + uid + " and sub: " + sub)
	}

	_, err = l.svcCtx.Redis.Del(oldKey)
	if err != nil {
		return nil, errors.New("failed to delete refresh token")
	}
	//* Generate new refresh and access tokens
	newRefreshJti := uuid.NewString()
	newAccessJti := uuid.NewString()

	access, accessExpireSeconds, err := tokenHelper.SignAccess(sub, newAccessJti)
	if err != nil {
		return nil, err
	}

	newRefresh, refreshExpireSeconds, err := tokenHelper.SignRefresh(sub, newRefreshJti)
	if err != nil {
		return nil, err
	}

	newKey := util.RedisKey(l.svcCtx.Key, util.RedisKeyTypeRefresh, newRefreshJti)
	if err := l.svcCtx.Redis.Setex(newKey, sub, int(refreshExpireSeconds)); err != nil {
		return nil, err
	}

	return &auth.LoginResp{
		AccessToken:  access,
		RefreshToken: newRefresh,
		ExpiresIn:    accessExpireSeconds,
		TokenType:    "bearer",
	}, nil
}

// single flight refresh
func (l *RefreshLogic) Refresh(in *auth.RefreshReq) (*auth.LoginResp, error) {
	if in.GetRefreshToken() == "" {
		return nil, errors.New("refresh token is required")
	}

	cfg := l.svcCtx.Config.JwtAuth
	tokenHelper := NewTokenHelper([]byte(cfg.Secret), "auth.prc",
		time.Duration(cfg.AccessExpireSeconds)*time.Second, time.Duration(cfg.RefreshExpireSeconds)*time.Second)

	claims, err := tokenHelper.Parse(in.GetRefreshToken())
	if err != nil {
		logx.Errorf("invalid token: %v", err)
		return nil, err
	}

	typ := claims.TokenType
	if typ != "refresh" {
		return nil, errors.New("invalid token type, got " + typ)
	}

	jti := claims.ID
	sub := claims.Subject

	if jti == "" || sub == "" {
		return nil, errors.New("invalid token, got jti: " + jti + " and sub: " + sub)
	}

	resAny, runErr, _ := l.svcCtx.RfGroup.Do(jti, func() (any, error) {
		const maxRetries = 2
		const redisTimeout = 150 * time.Millisecond

		var lastErr error
		for attempt := 0; attempt < maxRetries; attempt++ {
			// Create short timeout context for Redis operation
			timeoutCtx, cancel := context.WithTimeout(l.ctx, redisTimeout)

			// Generate new JTIs
			newRefreshJti := uuid.NewString()
			newAccessJti := uuid.NewString()

			// Execute Redis rotation with timeout
			// RefreshRotate(
			// 	ctx context.Context,
			// 	r *redis.Redis,
			// 	keyPrefix string,
			// 	oldJti string,
			// 	newJti string,
			// 	expectUserId string,
			// 	newTtlSeconds int,
			// )
			rot, err := dao.RefreshRotate(
				timeoutCtx,
				l.svcCtx.Redis,
				l.svcCtx.Key,
				jti,
				newRefreshJti,
				sub,
				int(cfg.RefreshExpireSeconds),
			)
			cancel()

			if err != nil {
				lastErr = err
				// Check if it's a temporary error (timeout, network, etc.)
				if isTemporaryError(err) && attempt < maxRetries-1 {
					logx.WithContext(l.ctx).Infof("refresh retry attempt=%d user=%s jti=%s err=%v",
						attempt+1, sub, jti, err)
					time.Sleep(10 * time.Millisecond) // Small backoff
					continue
				}
				// Non-retryable error or max retries reached
				return nil, err
			}

			// Check rotate result - these are business errors (non-retryable)
			switch rot.Code {
			case dao.RotateCodeOldNotFound:
				return nil, ErrRefreshNotFound
			case dao.RotateCodeMismatch:
				return nil, ErrUserMismatch
			case dao.RotateCodeReused:
				return nil, ErrRefreshReused
			case dao.RotateCodeOK:
				// Success - generate tokens
			default:
				return nil, ErrUnknown
			}

			// Generate new tokens
			access, accessExpireSeconds, err := tokenHelper.SignAccess(sub, newAccessJti)
			if err != nil {
				return nil, err
			}
			newRefresh, _, err := tokenHelper.SignRefresh(sub, newRefreshJti)
			if err != nil {
				return nil, err
			}

			return &auth.LoginResp{
				AccessToken:  access,
				RefreshToken: newRefresh,
				ExpiresIn:    accessExpireSeconds,
				TokenType:    "bearer",
			}, nil
		}
		return nil, lastErr
	})

	if runErr != nil {
		// Classify errors for observability
		isBusinessError := errors.Is(runErr, ErrRefreshNotFound) ||
			errors.Is(runErr, ErrUserMismatch) ||
			errors.Is(runErr, ErrRefreshReused)

		if isBusinessError {
			// Deterministic business failure - don't forget, just log
			logx.WithContext(l.ctx).Infof("refresh-business-error user=%s jti=%s err=%v", sub, jti, runErr)
		} else {
			// Temporary/infrastructure error - forget to allow retry
			l.svcCtx.RfGroup.Forget(jti)
			logx.WithContext(l.ctx).Errorf("refresh-infra-error user=%s jti=%s err=%v (forgotten)", sub, jti, runErr)
		}
		return nil, runErr
	}
	return resAny.(*auth.LoginResp), nil
}

// isTemporaryError checks if an error is temporary and retryable
func isTemporaryError(err error) bool {
	if err == nil {
		return false
	}

	// Context timeout/cancellation
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return true
	}

	// Check error message for common temporary patterns
	errMsg := err.Error()
	temporaryPatterns := []string{
		"connection refused",
		"connection reset",
		"i/o timeout",
		"timeout",
		"broken pipe",
		"no route to host",
		"network is unreachable",
		"temporary failure",
	}

	for _, pattern := range temporaryPatterns {
		if strings.Contains(strings.ToLower(errMsg), pattern) {
			return true
		}
	}

	return false
}

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
	// Validate refresh token
	claims, err := l.svcCtx.TokenHelper.ValidateRefreshToken(in.GetRefreshToken())
	if err != nil {
		return nil, err
	}

	jti := claims.ID
	username := claims.Subject

	// Verify token in Redis and delete old one
	oldKey := util.RedisKey(l.svcCtx.Key, util.RedisKeyTypeRefresh, jti)
	uid, err := l.svcCtx.Redis.Get(oldKey)
	if err != nil || uid == "" {
		return nil, ErrRefreshNotFound
	}

	if uid != username {
		return nil, ErrUserMismatch
	}

	if _, err = l.svcCtx.Redis.Del(oldKey); err != nil {
		return nil, errors.New("failed to delete refresh token")
	}

	// Generate new token pair
	newRefreshJti := uuid.NewString()
	newAccessJti := uuid.NewString()

	resp, err := l.svcCtx.TokenHelper.GenerateTokenPair(username, newAccessJti, newRefreshJti)
	if err != nil {
		return nil, err
	}

	// Store new refresh token in Redis
	cfg := l.svcCtx.Config.JwtAuth
	newKey := util.RedisKey(l.svcCtx.Key, util.RedisKeyTypeRefresh, newRefreshJti)
	if err := l.svcCtx.Redis.Setex(newKey, username, int(cfg.RefreshExpireSeconds)); err != nil {
		return nil, err
	}

	return resp, nil
}

// Refresh performs single-flight token refresh with retry logic
func (l *RefreshLogic) Refresh(in *auth.RefreshReq) (*auth.LoginResp, error) {
	// Validate refresh token
	claims, err := l.svcCtx.TokenHelper.ValidateRefreshToken(in.GetRefreshToken())
	if err != nil {
		return nil, err
	}

	jti := claims.ID
	username := claims.Subject

	// Execute refresh in single-flight group to prevent concurrent refreshes of the same token

	resAny, runErr, _ := l.svcCtx.RfGroup.Do(jti, func() (any, error) {
		resp, newJti, err := l.executeRefreshWithRetry(jti, username)
		l.takeCareOfSid(l.svcCtx.Key, jti, username, newJti)
		return resp, err
	})

	if runErr != nil {
		l.handleRefreshError(username, jti, runErr)
		return nil, runErr
	}

	return resAny.(*auth.LoginResp), nil
}

// executeRefreshWithRetry executes the refresh operation with retry logic
func (l *RefreshLogic) executeRefreshWithRetry(oldJti, username string) (*auth.LoginResp, string, error) {
	const maxRetries = 2
	const redisTimeout = 150 * time.Millisecond

	cfg := l.svcCtx.Config.JwtAuth
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Create short timeout context for Redis operation
		timeoutCtx, cancel := context.WithTimeout(l.ctx, redisTimeout)

		// Generate new JTIs
		newRefreshJti := uuid.NewString()
		newAccessJti := uuid.NewString()

		// Execute Redis rotation with timeout
		rot, err := dao.RefreshRotate(
			timeoutCtx,
			l.svcCtx.Redis,
			l.svcCtx.Key,
			oldJti,
			newRefreshJti,
			username,
			int(cfg.RefreshExpireSeconds),
		)
		cancel()

		if err != nil {
			lastErr = err
			// Check if it's a temporary error (timeout, network, etc.)
			if isTemporaryError(err) && attempt < maxRetries-1 {
				logx.WithContext(l.ctx).Infof("refresh retry attempt=%d user=%s jti=%s err=%v",
					attempt+1, username, oldJti, err)
				time.Sleep(10 * time.Millisecond) // Small backoff
				continue
			}
			// Non-retryable error or max retries reached
			return nil, newRefreshJti, err
		}

		// Check rotate result - these are business errors (non-retryable)
		switch rot.Code {
		case dao.RotateCodeOldNotFound:
			return nil, newRefreshJti, ErrRefreshNotFound
		case dao.RotateCodeMismatch:
			return nil, newRefreshJti, ErrUserMismatch
		case dao.RotateCodeReused:
			return nil, newRefreshJti, ErrRefreshReused
		case dao.RotateCodeOK:
			// Success - generate tokens
		default:
			return nil, newRefreshJti, ErrUnknown
		}

		// Generate new token pair
		resp, err := l.svcCtx.TokenHelper.GenerateTokenPair(username, newAccessJti, newRefreshJti)
		return resp, newRefreshJti, err
	}

	return nil, "", lastErr
}

// handleRefreshError classifies and logs refresh errors
func (l *RefreshLogic) handleRefreshError(username, jti string, err error) {
	isBusinessError := errors.Is(err, ErrRefreshNotFound) ||
		errors.Is(err, ErrUserMismatch) ||
		errors.Is(err, ErrRefreshReused)

	if isBusinessError {
		// Deterministic business failure - don't forget, just log
		logx.WithContext(l.ctx).Infof("refresh-business-error user=%s jti=%s err=%v", username, jti, err)
	} else {
		// Temporary/infrastructure error - forget to allow retry
		l.svcCtx.RfGroup.Forget(jti)
		logx.WithContext(l.ctx).Errorf("refresh-infra-error user=%s jti=%s err=%v (forgotten)", username, jti, err)
	}
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

func (l *RefreshLogic) takeCareOfSid(key string, oldJti string, uid string, newJti string) error {
	sid, _ := l.svcCtx.Redis.Get(util.JtiSidKey(key, oldJti))

	//if no sid history, generate a new one
	if sid == "" {
		sid = uuid.NewString()
		//user ->sids
		if _, err := l.svcCtx.Redis.Sadd(util.UserSidsKey(key, uid), sid); err != nil {
			logx.WithContext(l.ctx).Errorf("sid backfill Sadd failed uid=%s sid=%s err=%v", uid, sid, err)
		}
	}

	//take care of sid -> jtis: remove old jti, add new jti
	_, _ = l.svcCtx.Redis.Srem(util.SidSetKey(key, sid), oldJti)
	_, _ = l.svcCtx.Redis.Sadd(util.SidSetKey(key, sid), newJti)

	//write new jti index jti-> sid(TTL = refresh expire seconds)
	if err := l.svcCtx.Redis.Setex(util.JtiSidKey(key, newJti), sid, int(l.svcCtx.Config.JwtAuth.RefreshExpireSeconds)); err != nil {
		logx.WithContext(l.ctx).Errorf("jti sid index Setex failed jti=%s sid=%s err=%v", newJti, sid, err)
	}

	//remove old jti index jti-> sid(TTL = refresh expire seconds)
	_, _ = l.svcCtx.Redis.Del(util.JtiSidKey(key, oldJti))

	return nil
}

package logic

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/uwu-octane/antBackend/api/v1/auth"
	"github.com/uwu-octane/antBackend/auth/internal/config"
	"github.com/uwu-octane/antBackend/auth/internal/svc"
	"github.com/uwu-octane/antBackend/auth/internal/util"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

func createTestServiceContext(t *testing.T) (*svc.ServiceContext, *miniredis.Miniredis) {
	mr := miniredis.RunT(t)

	cfg := config.Config{
		JwtAuth: config.JwtAuthConfig{
			Secret:               "test-secret-key-for-testing",
			AccessExpireSeconds:  3600,
			RefreshExpireSeconds: 604800,
		},
	}

	redisClient := redis.MustNewRedis(redis.RedisConf{
		Host: mr.Addr(),
		Type: "node",
	})

	return &svc.ServiceContext{
		Config: cfg,
		Redis:  redisClient,
		Key:    "test",
	}, mr
}

func TestLogout_Success(t *testing.T) {
	ctx := context.Background()
	svcCtx, mr := createTestServiceContext(t)
	defer mr.Close()

	// Create test token
	tokenHelper := NewTokenHelper(
		[]byte(svcCtx.Config.JwtAuth.Secret),
		"auth.prc",
		time.Duration(svcCtx.Config.JwtAuth.AccessExpireSeconds)*time.Second,
		time.Duration(svcCtx.Config.JwtAuth.RefreshExpireSeconds)*time.Second,
	)

	userId := "user-123"
	jti := "jti-456"
	refreshToken, _, err := tokenHelper.SignRefresh(userId, jti)
	if err != nil {
		t.Fatalf("Failed to create refresh token: %v", err)
	}

	// Set up Redis with the token
	refreshKey := util.RedisKey(svcCtx.Key, util.RedisKeyTypeRefresh, jti)
	err = mr.Set(refreshKey, userId)
	if err != nil {
		t.Fatalf("Failed to set refresh token in Redis: %v", err)
	}

	// Test logout
	logic := NewLogoutLogic(ctx, svcCtx)
	resp, err := logic.Logout(&auth.LogoutReq{
		RefreshToken: refreshToken,
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if resp == nil {
		t.Fatal("Expected response, got nil")
	}

	if !resp.Ok {
		t.Errorf("Expected Ok=true, got Ok=%v", resp.Ok)
	}

	if resp.Message != "log out" {
		t.Errorf("Expected message 'log out', got: %s", resp.Message)
	}

	// Verify reuse flag was set
	reuseKey := util.RedisKey(svcCtx.Key, util.RedisKeyTypeReuse, jti)
	if !mr.Exists(reuseKey) {
		t.Error("Expected reuse flag to be set")
	}

	// Verify refresh token was deleted
	if mr.Exists(refreshKey) {
		t.Error("Expected refresh token to be deleted from Redis")
	}
}

func TestLogout_MissingRefreshToken(t *testing.T) {
	ctx := context.Background()
	svcCtx, mr := createTestServiceContext(t)
	defer mr.Close()

	logic := NewLogoutLogic(ctx, svcCtx)
	resp, err := logic.Logout(&auth.LogoutReq{
		RefreshToken: "",
	})

	if err == nil {
		t.Fatal("Expected error for missing refresh token, got nil")
	}

	if err.Error() != "refresh token is required" {
		t.Errorf("Expected 'refresh token is required', got: %s", err.Error())
	}

	if resp != nil {
		t.Errorf("Expected nil response, got: %v", resp)
	}
}

func TestLogout_InvalidRefreshToken(t *testing.T) {
	ctx := context.Background()
	svcCtx, mr := createTestServiceContext(t)
	defer mr.Close()

	logic := NewLogoutLogic(ctx, svcCtx)
	resp, err := logic.Logout(&auth.LogoutReq{
		RefreshToken: "invalid-token",
	})

	if err == nil {
		t.Fatal("Expected error for invalid token, got nil")
	}

	if resp != nil {
		t.Errorf("Expected nil response, got: %v", resp)
	}
}

func TestLogout_WrongTokenType(t *testing.T) {
	ctx := context.Background()
	svcCtx, mr := createTestServiceContext(t)
	defer mr.Close()

	// Create an access token instead of refresh token
	tokenHelper := NewTokenHelper(
		[]byte(svcCtx.Config.JwtAuth.Secret),
		"auth.prc",
		time.Duration(svcCtx.Config.JwtAuth.AccessExpireSeconds)*time.Second,
		time.Duration(svcCtx.Config.JwtAuth.RefreshExpireSeconds)*time.Second,
	)

	userId := "user-123"
	jti := "jti-456"
	accessToken, _, err := tokenHelper.SignAccess(userId, jti)
	if err != nil {
		t.Fatalf("Failed to create access token: %v", err)
	}

	logic := NewLogoutLogic(ctx, svcCtx)
	resp, err := logic.Logout(&auth.LogoutReq{
		RefreshToken: accessToken,
	})

	if err == nil {
		t.Fatal("Expected error for wrong token type, got nil")
	}

	if err.Error() != "invalid refresh token" {
		t.Errorf("Expected 'invalid refresh token', got: %s", err.Error())
	}

	if resp != nil {
		t.Errorf("Expected nil response, got: %v", resp)
	}
}

func TestLogout_TokenNotFoundInRedis(t *testing.T) {
	ctx := context.Background()
	svcCtx, mr := createTestServiceContext(t)
	defer mr.Close()

	// Create a valid refresh token
	tokenHelper := NewTokenHelper(
		[]byte(svcCtx.Config.JwtAuth.Secret),
		"auth.prc",
		time.Duration(svcCtx.Config.JwtAuth.AccessExpireSeconds)*time.Second,
		time.Duration(svcCtx.Config.JwtAuth.RefreshExpireSeconds)*time.Second,
	)

	userId := "user-123"
	jti := "jti-456"
	refreshToken, _, err := tokenHelper.SignRefresh(userId, jti)
	if err != nil {
		t.Fatalf("Failed to create refresh token: %v", err)
	}

	// Don't set the token in Redis, simulating expired/missing token

	logic := NewLogoutLogic(ctx, svcCtx)
	resp, err := logic.Logout(&auth.LogoutReq{
		RefreshToken: refreshToken,
	})

	// Should succeed as idempotent operation
	if err != nil {
		t.Fatalf("Expected no error for idempotent operation, got: %v", err)
	}

	if resp == nil {
		t.Fatal("Expected response, got nil")
	}

	if !resp.Ok {
		t.Errorf("Expected Ok=true, got Ok=%v", resp.Ok)
	}

	// Message should indicate it may have expired
	if resp.Message == "" {
		t.Error("Expected message indicating token may be expired")
	}
}

func TestLogout_SubjectMismatch(t *testing.T) {
	ctx := context.Background()
	svcCtx, mr := createTestServiceContext(t)
	defer mr.Close()

	// Create test token
	tokenHelper := NewTokenHelper(
		[]byte(svcCtx.Config.JwtAuth.Secret),
		"auth.prc",
		time.Duration(svcCtx.Config.JwtAuth.AccessExpireSeconds)*time.Second,
		time.Duration(svcCtx.Config.JwtAuth.RefreshExpireSeconds)*time.Second,
	)

	userId := "user-123"
	jti := "jti-456"
	refreshToken, _, err := tokenHelper.SignRefresh(userId, jti)
	if err != nil {
		t.Fatalf("Failed to create refresh token: %v", err)
	}

	// Set up Redis with different user ID
	refreshKey := util.RedisKey(svcCtx.Key, util.RedisKeyTypeRefresh, jti)
	err = mr.Set(refreshKey, "different-user-789")
	if err != nil {
		t.Fatalf("Failed to set refresh token in Redis: %v", err)
	}

	logic := NewLogoutLogic(ctx, svcCtx)
	resp, err := logic.Logout(&auth.LogoutReq{
		RefreshToken: refreshToken,
	})

	if err == nil {
		t.Fatal("Expected error for subject mismatch, got nil")
	}

	if err.Error() != "subject mismatch" {
		t.Errorf("Expected 'subject mismatch', got: %s", err.Error())
	}

	if resp != nil {
		t.Errorf("Expected nil response, got: %v", resp)
	}

	// Verify token was NOT deleted
	if !mr.Exists(refreshKey) {
		t.Error("Expected refresh token to still exist after subject mismatch")
	}
}

func TestLogout_MultipleLogouts(t *testing.T) {
	ctx := context.Background()
	svcCtx, mr := createTestServiceContext(t)
	defer mr.Close()

	// Create test token
	tokenHelper := NewTokenHelper(
		[]byte(svcCtx.Config.JwtAuth.Secret),
		"auth.prc",
		time.Duration(svcCtx.Config.JwtAuth.AccessExpireSeconds)*time.Second,
		time.Duration(svcCtx.Config.JwtAuth.RefreshExpireSeconds)*time.Second,
	)

	userId := "user-123"
	jti := "jti-456"
	refreshToken, _, err := tokenHelper.SignRefresh(userId, jti)
	if err != nil {
		t.Fatalf("Failed to create refresh token: %v", err)
	}

	// Set up Redis with the token
	refreshKey := util.RedisKey(svcCtx.Key, util.RedisKeyTypeRefresh, jti)
	err = mr.Set(refreshKey, userId)
	if err != nil {
		t.Fatalf("Failed to set refresh token in Redis: %v", err)
	}

	logic := NewLogoutLogic(ctx, svcCtx)

	// First logout should succeed
	resp1, err1 := logic.Logout(&auth.LogoutReq{
		RefreshToken: refreshToken,
	})

	if err1 != nil {
		t.Fatalf("First logout failed: %v", err1)
	}

	if !resp1.Ok {
		t.Error("First logout should succeed")
	}

	// Second logout should be idempotent
	resp2, err2 := logic.Logout(&auth.LogoutReq{
		RefreshToken: refreshToken,
	})

	if err2 != nil {
		t.Fatalf("Second logout should be idempotent: %v", err2)
	}

	if !resp2.Ok {
		t.Error("Second logout should be idempotent and return Ok=true")
	}
}

func TestLogout_WithRealRedisOperations(t *testing.T) {
	ctx := context.Background()
	svcCtx, mr := createTestServiceContext(t)
	defer mr.Close()

	// Create test token
	tokenHelper := NewTokenHelper(
		[]byte(svcCtx.Config.JwtAuth.Secret),
		"auth.prc",
		time.Duration(svcCtx.Config.JwtAuth.AccessExpireSeconds)*time.Second,
		time.Duration(svcCtx.Config.JwtAuth.RefreshExpireSeconds)*time.Second,
	)

	userId := "user-123"
	jti := "jti-456"
	refreshToken, _, err := tokenHelper.SignRefresh(userId, jti)
	if err != nil {
		t.Fatalf("Failed to create refresh token: %v", err)
	}

	// Set up Redis with the token using the actual Redis client
	refreshKey := util.RedisKey(svcCtx.Key, util.RedisKeyTypeRefresh, jti)
	err = svcCtx.Redis.Set(refreshKey, userId)
	if err != nil {
		t.Fatalf("Failed to set refresh token in Redis: %v", err)
	}

	// Verify token is in Redis
	val, err := svcCtx.Redis.Get(refreshKey)
	if err != nil {
		t.Fatalf("Failed to get refresh token from Redis: %v", err)
	}
	if val != userId {
		t.Fatalf("Expected userId %s, got %s", userId, val)
	}

	// Test logout
	logic := NewLogoutLogic(ctx, svcCtx)
	resp, err := logic.Logout(&auth.LogoutReq{
		RefreshToken: refreshToken,
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !resp.Ok {
		t.Errorf("Expected Ok=true, got Ok=%v", resp.Ok)
	}

	// Verify token was deleted using Redis client
	val, err = svcCtx.Redis.Get(refreshKey)
	if err == nil && val != "" {
		t.Errorf("Expected empty value for deleted token, got: %s", val)
	}

	// Verify reuse flag was set
	reuseKey := util.RedisKey(svcCtx.Key, util.RedisKeyTypeReuse, jti)
	val, err = svcCtx.Redis.Get(reuseKey)
	if err != nil {
		t.Errorf("Expected reuse flag to be set: %v", err)
	}
	if val != "1" {
		t.Errorf("Expected reuse flag value '1', got '%s'", val)
	}
}

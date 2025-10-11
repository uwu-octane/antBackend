package logic

import (
	"context"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/uwu-octane/antBackend/api/v1/auth"
	"github.com/uwu-octane/antBackend/auth/internal/svc"
	"github.com/uwu-octane/antBackend/auth/internal/util"
)

func TestLogoutAll_SingleSession(t *testing.T) {
	ctx := context.Background()
	svcCtx, mr := createTestServiceContext(t)
	defer mr.Close()

	userId := "test-user-1"

	// Create a single session with one refresh token
	sid := uuid.NewString()
	refreshJti := uuid.NewString()

	// Set up the data structure
	setupSession(t, mr, svcCtx, userId, sid, refreshJti)

	// Create the refresh token
	refreshToken, _, err := svcCtx.TokenHelper.SignRefresh(userId, refreshJti)
	if err != nil {
		t.Fatalf("Failed to create refresh token: %v", err)
	}

	// Call LogoutAll
	logic := NewLogoutAllLogic(ctx, svcCtx)
	resp, err := logic.LogoutAll(&auth.LogoutAllReq{
		RefreshToken: refreshToken,
	})

	// Verify response
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if resp == nil {
		t.Fatal("Expected response, got nil")
	}

	if !resp.Ok {
		t.Errorf("Expected Ok=true, got Ok=%v", resp.Ok)
	}

	if resp.Message != "Log out all success" {
		t.Errorf("Expected message 'Log out all success', got: %s", resp.Message)
	}

	// Verify Redis data
	userSidsKey := util.UserSidsKey(svcCtx.Key, userId)
	sids, err := mr.SMembers(userSidsKey)
	if err != nil {
		t.Fatalf("Failed to get sids: %v", err)
	}

	if len(sids) != 1 {
		t.Errorf("Expected 1 sid, got %d", len(sids))
	}

	if !contains(sids, sid) {
		t.Errorf("Expected sid %s to be in sids list", sid)
	}

	// Verify sid->jtis mapping
	sidSetKey := util.SidSetKey(svcCtx.Key, sid)
	jtis, err := mr.SMembers(sidSetKey)
	if err != nil {
		t.Fatalf("Failed to get jtis for sid %s: %v", sid, err)
	}

	if len(jtis) != 1 {
		t.Errorf("Expected 1 jti for sid %s, got %d", sid, len(jtis))
	}
}

func TestLogoutAll_MultipleSessions(t *testing.T) {
	ctx := context.Background()
	svcCtx, mr := createTestServiceContext(t)
	defer mr.Close()

	userId := "test-user-2"

	// Create 3 different sessions (simulating logins from different devices)
	sessions := []struct {
		sid  string
		jtis []string
	}{
		{sid: uuid.NewString(), jtis: []string{uuid.NewString()}},
		{sid: uuid.NewString(), jtis: []string{uuid.NewString()}},
		{sid: uuid.NewString(), jtis: []string{uuid.NewString()}},
	}

	// Set up all sessions
	for _, session := range sessions {
		for _, jti := range session.jtis {
			setupSession(t, mr, svcCtx, userId, session.sid, jti)
		}
	}

	// Use the refresh token from the first session
	refreshToken, _, err := svcCtx.TokenHelper.SignRefresh(userId, sessions[0].jtis[0])
	if err != nil {
		t.Fatalf("Failed to create refresh token: %v", err)
	}

	// Call LogoutAll
	logic := NewLogoutAllLogic(ctx, svcCtx)
	resp, err := logic.LogoutAll(&auth.LogoutAllReq{
		RefreshToken: refreshToken,
	})

	// Verify response
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !resp.Ok {
		t.Error("Expected Ok=true")
	}

	// Verify Redis data matches what we set up
	userSidsKey := util.UserSidsKey(svcCtx.Key, userId)
	sids, err := mr.SMembers(userSidsKey)
	if err != nil {
		t.Fatalf("Failed to get sids: %v", err)
	}

	if len(sids) != 3 {
		t.Errorf("Expected 3 sids, got %d", len(sids))
	}

	// Verify each sid has the correct number of jtis
	for _, session := range sessions {
		sidSetKey := util.SidSetKey(svcCtx.Key, session.sid)
		jtis, err := mr.SMembers(sidSetKey)
		if err != nil {
			t.Fatalf("Failed to get jtis for sid %s: %v", session.sid, err)
		}

		if len(jtis) != len(session.jtis) {
			t.Errorf("Expected %d jtis for sid %s, got %d", len(session.jtis), session.sid, len(jtis))
		}
	}
}

func TestLogoutAll_MultipleRefreshesPerSession(t *testing.T) {
	ctx := context.Background()
	svcCtx, mr := createTestServiceContext(t)
	defer mr.Close()

	userId := "test-user-3"

	// Create 2 sessions, each with multiple refresh tokens (simulating refresh operations)
	sessions := []struct {
		sid  string
		jtis []string
	}{
		{
			sid:  uuid.NewString(),
			jtis: []string{uuid.NewString(), uuid.NewString(), uuid.NewString()},
		},
		{
			sid:  uuid.NewString(),
			jtis: []string{uuid.NewString(), uuid.NewString()},
		},
	}

	// Set up all sessions with multiple jtis
	for _, session := range sessions {
		for _, jti := range session.jtis {
			setupSession(t, mr, svcCtx, userId, session.sid, jti)
		}
	}

	// Use any refresh token
	refreshToken, _, err := svcCtx.TokenHelper.SignRefresh(userId, sessions[0].jtis[0])
	if err != nil {
		t.Fatalf("Failed to create refresh token: %v", err)
	}

	// Call LogoutAll
	logic := NewLogoutAllLogic(ctx, svcCtx)
	resp, err := logic.LogoutAll(&auth.LogoutAllReq{
		RefreshToken: refreshToken,
	})

	// Verify response
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !resp.Ok {
		t.Error("Expected Ok=true")
	}

	// Verify total JTI count
	totalJtis := 0
	for _, session := range sessions {
		sidSetKey := util.SidSetKey(svcCtx.Key, session.sid)
		jtis, err := mr.SMembers(sidSetKey)
		if err != nil {
			t.Fatalf("Failed to get jtis for sid %s: %v", session.sid, err)
		}
		totalJtis += len(jtis)

		if len(jtis) != len(session.jtis) {
			t.Errorf("Expected %d jtis for sid %s, got %d", len(session.jtis), session.sid, len(jtis))
		}
	}

	expectedTotal := 5 // 3 + 2
	if totalJtis != expectedTotal {
		t.Errorf("Expected %d total jtis, got %d", expectedTotal, totalJtis)
	}
}

func TestLogoutAll_MissingRefreshToken(t *testing.T) {
	ctx := context.Background()
	svcCtx, mr := createTestServiceContext(t)
	defer mr.Close()

	logic := NewLogoutAllLogic(ctx, svcCtx)
	resp, err := logic.LogoutAll(&auth.LogoutAllReq{
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

func TestLogoutAll_InvalidRefreshToken(t *testing.T) {
	ctx := context.Background()
	svcCtx, mr := createTestServiceContext(t)
	defer mr.Close()

	logic := NewLogoutAllLogic(ctx, svcCtx)
	resp, err := logic.LogoutAll(&auth.LogoutAllReq{
		RefreshToken: "invalid-token",
	})

	if err == nil {
		t.Fatal("Expected error for invalid token, got nil")
	}

	if err.Error() != "invalid refresh token" {
		t.Errorf("Expected 'invalid refresh token', got: %s", err.Error())
	}

	if resp != nil {
		t.Errorf("Expected nil response, got: %v", resp)
	}
}

func TestLogoutAll_WrongTokenType(t *testing.T) {
	ctx := context.Background()
	svcCtx, mr := createTestServiceContext(t)
	defer mr.Close()

	// Create an access token instead of refresh token
	userId := "test-user-4"
	jti := uuid.NewString()
	accessToken, _, err := svcCtx.TokenHelper.SignAccess(userId, jti)
	if err != nil {
		t.Fatalf("Failed to create access token: %v", err)
	}

	logic := NewLogoutAllLogic(ctx, svcCtx)
	resp, err := logic.LogoutAll(&auth.LogoutAllReq{
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

func TestLogoutAll_NoSessions(t *testing.T) {
	ctx := context.Background()
	svcCtx, mr := createTestServiceContext(t)
	defer mr.Close()

	userId := "test-user-5"
	jti := uuid.NewString()

	// Create refresh token but no sessions in Redis
	refreshToken, _, err := svcCtx.TokenHelper.SignRefresh(userId, jti)
	if err != nil {
		t.Fatalf("Failed to create refresh token: %v", err)
	}

	logic := NewLogoutAllLogic(ctx, svcCtx)
	resp, err := logic.LogoutAll(&auth.LogoutAllReq{
		RefreshToken: refreshToken,
	})

	// Should succeed even with no sessions
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !resp.Ok {
		t.Error("Expected Ok=true")
	}
}

func TestLogoutAll_RealWorldScenario(t *testing.T) {
	ctx := context.Background()
	svcCtx, mr := createTestServiceContext(t)
	defer mr.Close()

	userId := "test-user-6"

	// Simulate real-world scenario:
	// - User logs in from web browser (session 1)
	// - User logs in from mobile app (session 2)
	// - User logs in from tablet (session 3)
	// - Each device does some refreshes

	webSession := uuid.NewString()
	mobileSession := uuid.NewString()
	tabletSession := uuid.NewString()

	// Web browser: initial login + 2 refreshes
	webJtis := []string{uuid.NewString(), uuid.NewString(), uuid.NewString()}
	for _, jti := range webJtis {
		setupSession(t, mr, svcCtx, userId, webSession, jti)
	}

	// Mobile app: initial login + 1 refresh
	mobileJtis := []string{uuid.NewString(), uuid.NewString()}
	for _, jti := range mobileJtis {
		setupSession(t, mr, svcCtx, userId, mobileSession, jti)
	}

	// Tablet: initial login only
	tabletJtis := []string{uuid.NewString()}
	for _, jti := range tabletJtis {
		setupSession(t, mr, svcCtx, userId, tabletSession, jti)
	}

	// User calls LogoutAll from mobile device
	refreshToken, _, err := svcCtx.TokenHelper.SignRefresh(userId, mobileJtis[0])
	if err != nil {
		t.Fatalf("Failed to create refresh token: %v", err)
	}

	logic := NewLogoutAllLogic(ctx, svcCtx)
	resp, err := logic.LogoutAll(&auth.LogoutAllReq{
		RefreshToken: refreshToken,
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !resp.Ok {
		t.Error("Expected Ok=true")
	}

	// Verify all sessions are enumerated correctly
	userSidsKey := util.UserSidsKey(svcCtx.Key, userId)
	sids, err := mr.SMembers(userSidsKey)
	if err != nil {
		t.Fatalf("Failed to get sids: %v", err)
	}

	if len(sids) != 3 {
		t.Errorf("Expected 3 sids (web, mobile, tablet), got %d", len(sids))
	}

	// Verify JTI counts per session
	verifyJtiCount(t, mr, svcCtx, webSession, 3, "web session")
	verifyJtiCount(t, mr, svcCtx, mobileSession, 2, "mobile session")
	verifyJtiCount(t, mr, svcCtx, tabletSession, 1, "tablet session")

	// Verify total JTI count
	totalJtis := 0
	for _, sid := range []string{webSession, mobileSession, tabletSession} {
		sidSetKey := util.SidSetKey(svcCtx.Key, sid)
		jtis, err := mr.SMembers(sidSetKey)
		if err != nil {
			t.Fatalf("Failed to get jtis for sid %s: %v", sid, err)
		}
		totalJtis += len(jtis)
	}

	expectedTotal := 6 // 3 + 2 + 1
	if totalJtis != expectedTotal {
		t.Errorf("Expected %d total jtis, got %d", expectedTotal, totalJtis)
	}
}

func TestLogoutAll_WithRealRedisOperations(t *testing.T) {
	ctx := context.Background()
	svcCtx, mr := createTestServiceContext(t)
	defer mr.Close()

	userId := "test-user-7"
	sid1 := uuid.NewString()
	sid2 := uuid.NewString()
	jti1 := uuid.NewString()
	jti2 := uuid.NewString()
	jti3 := uuid.NewString()

	// Use the actual Redis client (which talks to miniredis)
	userSidsKey := util.UserSidsKey(svcCtx.Key, userId)

	// Add sids for the user
	_, err := svcCtx.Redis.Sadd(userSidsKey, sid1)
	if err != nil {
		t.Fatalf("Failed to add sid1: %v", err)
	}
	_, err = svcCtx.Redis.Sadd(userSidsKey, sid2)
	if err != nil {
		t.Fatalf("Failed to add sid2: %v", err)
	}

	// Add jtis for sid1
	sid1SetKey := util.SidSetKey(svcCtx.Key, sid1)
	_, err = svcCtx.Redis.Sadd(sid1SetKey, jti1)
	if err != nil {
		t.Fatalf("Failed to add jti1 to sid1: %v", err)
	}
	_, err = svcCtx.Redis.Sadd(sid1SetKey, jti2)
	if err != nil {
		t.Fatalf("Failed to add jti2 to sid1: %v", err)
	}

	// Add jtis for sid2
	sid2SetKey := util.SidSetKey(svcCtx.Key, sid2)
	_, err = svcCtx.Redis.Sadd(sid2SetKey, jti3)
	if err != nil {
		t.Fatalf("Failed to add jti3 to sid2: %v", err)
	}

	// Create refresh token
	refreshToken, _, err := svcCtx.TokenHelper.SignRefresh(userId, jti1)
	if err != nil {
		t.Fatalf("Failed to create refresh token: %v", err)
	}

	// Call LogoutAll
	logic := NewLogoutAllLogic(ctx, svcCtx)
	resp, err := logic.LogoutAll(&auth.LogoutAllReq{
		RefreshToken: refreshToken,
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !resp.Ok {
		t.Error("Expected Ok=true")
	}

	// Verify data using Redis client
	sids, err := svcCtx.Redis.Smembers(userSidsKey)
	if err != nil {
		t.Fatalf("Failed to get sids: %v", err)
	}

	if len(sids) != 2 {
		t.Errorf("Expected 2 sids, got %d", len(sids))
	}

	// Verify jtis for each sid
	sid1Jtis, err := svcCtx.Redis.Smembers(sid1SetKey)
	if err != nil {
		t.Fatalf("Failed to get jtis for sid1: %v", err)
	}
	if len(sid1Jtis) != 2 {
		t.Errorf("Expected 2 jtis for sid1, got %d", len(sid1Jtis))
	}

	sid2Jtis, err := svcCtx.Redis.Smembers(sid2SetKey)
	if err != nil {
		t.Fatalf("Failed to get jtis for sid2: %v", err)
	}
	if len(sid2Jtis) != 1 {
		t.Errorf("Expected 1 jti for sid2, got %d", len(sid2Jtis))
	}
}

func TestLogoutAll_PartialRedisFailure(t *testing.T) {
	ctx := context.Background()
	svcCtx, mr := createTestServiceContext(t)
	defer mr.Close()

	userId := "test-user-8"
	sid1 := uuid.NewString()
	sid2 := uuid.NewString()
	jti1 := uuid.NewString()

	// Set up user sids
	userSidsKey := util.UserSidsKey(svcCtx.Key, userId)
	_, err := svcCtx.Redis.Sadd(userSidsKey, sid1)
	if err != nil {
		t.Fatalf("Failed to add sid1: %v", err)
	}
	_, err = svcCtx.Redis.Sadd(userSidsKey, sid2)
	if err != nil {
		t.Fatalf("Failed to add sid2: %v", err)
	}

	// Only set up jtis for sid1, not sid2 (simulating a partial failure or missing data)
	sid1SetKey := util.SidSetKey(svcCtx.Key, sid1)
	_, err = svcCtx.Redis.Sadd(sid1SetKey, jti1)
	if err != nil {
		t.Fatalf("Failed to add jti1 to sid1: %v", err)
	}

	// Create refresh token
	refreshToken, _, err := svcCtx.TokenHelper.SignRefresh(userId, jti1)
	if err != nil {
		t.Fatalf("Failed to create refresh token: %v", err)
	}

	// Call LogoutAll - should handle missing sid gracefully
	logic := NewLogoutAllLogic(ctx, svcCtx)
	resp, err := logic.LogoutAll(&auth.LogoutAllReq{
		RefreshToken: refreshToken,
	})

	// Should succeed despite sid2 having no jtis
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !resp.Ok {
		t.Error("Expected Ok=true")
	}
}

// Helper functions

func setupSession(t *testing.T, mr *miniredis.Miniredis, svcCtx *svc.ServiceContext, userId, sid, jti string) {
	// Add sid to user's sid set
	userSidsKey := util.UserSidsKey(svcCtx.Key, userId)
	_, err := mr.SAdd(userSidsKey, sid)
	if err != nil {
		t.Fatalf("Failed to add sid %s to user sids: %v", sid, err)
	}

	// Add jti to sid's jti set
	sidSetKey := util.SidSetKey(svcCtx.Key, sid)
	_, err = mr.SAdd(sidSetKey, jti)
	if err != nil {
		t.Fatalf("Failed to add jti %s to sid %s: %v", jti, sid, err)
	}

	// Add jti->sid mapping
	jtiSidKey := util.JtiSidKey(svcCtx.Key, jti)
	err = mr.Set(jtiSidKey, sid)
	if err != nil {
		t.Fatalf("Failed to set jti->sid mapping: %v", err)
	}

	// Add refresh token
	refreshKey := util.RedisKey(svcCtx.Key, util.RedisKeyTypeRefresh, jti)
	err = mr.Set(refreshKey, userId)
	if err != nil {
		t.Fatalf("Failed to set refresh token: %v", err)
	}
}

func verifyJtiCount(t *testing.T, mr *miniredis.Miniredis, svcCtx *svc.ServiceContext, sid string, expectedCount int, sessionName string) {
	sidSetKey := util.SidSetKey(svcCtx.Key, sid)
	jtis, err := mr.SMembers(sidSetKey)
	if err != nil {
		t.Fatalf("Failed to get jtis for %s: %v", sessionName, err)
	}

	if len(jtis) != expectedCount {
		t.Errorf("Expected %d jtis for %s, got %d", expectedCount, sessionName, len(jtis))
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

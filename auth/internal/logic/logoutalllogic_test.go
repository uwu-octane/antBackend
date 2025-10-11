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

	// Verify Redis data exists BEFORE LogoutAll
	userSidsKey := util.UserSidsKey(svcCtx.Key, userId)
	sidsBefore, err := mr.SMembers(userSidsKey)
	if err != nil {
		t.Fatalf("Failed to get sids before LogoutAll: %v", err)
	}

	if len(sidsBefore) != 1 {
		t.Errorf("Expected 1 sid before LogoutAll, got %d", len(sidsBefore))
	}

	if !contains(sidsBefore, sid) {
		t.Errorf("Expected sid %s to be in sids list before LogoutAll", sid)
	}

	// Verify sid->jtis mapping exists before LogoutAll
	sidSetKey := util.SidSetKey(svcCtx.Key, sid)
	jtisBefore, err := mr.SMembers(sidSetKey)
	if err != nil {
		t.Fatalf("Failed to get jtis for sid %s before LogoutAll: %v", sid, err)
	}

	if len(jtisBefore) != 1 {
		t.Errorf("Expected 1 jti for sid %s before LogoutAll, got %d", sid, len(jtisBefore))
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

	// Verify Redis data is DELETED after LogoutAll
	if mr.Exists(userSidsKey) {
		t.Error("Expected user sids key to be deleted after LogoutAll")
	}

	if mr.Exists(sidSetKey) {
		t.Error("Expected sid set key to be deleted after LogoutAll")
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

	// Verify Redis data exists BEFORE LogoutAll
	userSidsKey := util.UserSidsKey(svcCtx.Key, userId)
	sidsBefore, err := mr.SMembers(userSidsKey)
	if err != nil {
		t.Fatalf("Failed to get sids before LogoutAll: %v", err)
	}

	if len(sidsBefore) != 3 {
		t.Errorf("Expected 3 sids before LogoutAll, got %d", len(sidsBefore))
	}

	// Verify each sid has the correct number of jtis before LogoutAll
	for _, session := range sessions {
		sidSetKey := util.SidSetKey(svcCtx.Key, session.sid)
		jtisBefore, err := mr.SMembers(sidSetKey)
		if err != nil {
			t.Fatalf("Failed to get jtis for sid %s before LogoutAll: %v", session.sid, err)
		}

		if len(jtisBefore) != len(session.jtis) {
			t.Errorf("Expected %d jtis for sid %s before LogoutAll, got %d", len(session.jtis), session.sid, len(jtisBefore))
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

	// Verify all Redis data is DELETED after LogoutAll
	if mr.Exists(userSidsKey) {
		t.Error("Expected user sids key to be deleted after LogoutAll")
	}

	for _, session := range sessions {
		sidSetKey := util.SidSetKey(svcCtx.Key, session.sid)
		if mr.Exists(sidSetKey) {
			t.Errorf("Expected sid set key to be deleted for sid %s after LogoutAll", session.sid)
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

	// Verify total JTI count BEFORE LogoutAll
	totalJtisBefore := 0
	for _, session := range sessions {
		sidSetKey := util.SidSetKey(svcCtx.Key, session.sid)
		jtisBefore, err := mr.SMembers(sidSetKey)
		if err != nil {
			t.Fatalf("Failed to get jtis for sid %s before LogoutAll: %v", session.sid, err)
		}
		totalJtisBefore += len(jtisBefore)

		if len(jtisBefore) != len(session.jtis) {
			t.Errorf("Expected %d jtis for sid %s before LogoutAll, got %d", len(session.jtis), session.sid, len(jtisBefore))
		}
	}

	expectedTotal := 5 // 3 + 2
	if totalJtisBefore != expectedTotal {
		t.Errorf("Expected %d total jtis before LogoutAll, got %d", expectedTotal, totalJtisBefore)
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

	// Verify all data is DELETED after LogoutAll
	for _, session := range sessions {
		sidSetKey := util.SidSetKey(svcCtx.Key, session.sid)
		if mr.Exists(sidSetKey) {
			t.Errorf("Expected sid set key to be deleted for sid %s after LogoutAll", session.sid)
		}
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

	// Verify all sessions exist BEFORE LogoutAll
	userSidsKey := util.UserSidsKey(svcCtx.Key, userId)
	sidsBefore, err := mr.SMembers(userSidsKey)
	if err != nil {
		t.Fatalf("Failed to get sids before LogoutAll: %v", err)
	}

	if len(sidsBefore) != 3 {
		t.Errorf("Expected 3 sids (web, mobile, tablet) before LogoutAll, got %d", len(sidsBefore))
	}

	// Verify JTI counts per session BEFORE LogoutAll
	verifyJtiCount(t, mr, svcCtx, webSession, 3, "web session before LogoutAll")
	verifyJtiCount(t, mr, svcCtx, mobileSession, 2, "mobile session before LogoutAll")
	verifyJtiCount(t, mr, svcCtx, tabletSession, 1, "tablet session before LogoutAll")

	// Verify total JTI count BEFORE LogoutAll
	totalJtisBefore := 0
	for _, sid := range []string{webSession, mobileSession, tabletSession} {
		sidSetKey := util.SidSetKey(svcCtx.Key, sid)
		jtisBefore, err := mr.SMembers(sidSetKey)
		if err != nil {
			t.Fatalf("Failed to get jtis for sid %s before LogoutAll: %v", sid, err)
		}
		totalJtisBefore += len(jtisBefore)
	}

	expectedTotal := 6 // 3 + 2 + 1
	if totalJtisBefore != expectedTotal {
		t.Errorf("Expected %d total jtis before LogoutAll, got %d", expectedTotal, totalJtisBefore)
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

	// Verify all data is DELETED after LogoutAll
	if mr.Exists(userSidsKey) {
		t.Error("Expected user sids key to be deleted after LogoutAll")
	}

	for _, sid := range []string{webSession, mobileSession, tabletSession} {
		sidSetKey := util.SidSetKey(svcCtx.Key, sid)
		if mr.Exists(sidSetKey) {
			t.Errorf("Expected sid set key to be deleted for sid %s after LogoutAll", sid)
		}
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

	// Verify data using Redis client BEFORE LogoutAll
	sidsBefore, err := svcCtx.Redis.Smembers(userSidsKey)
	if err != nil {
		t.Fatalf("Failed to get sids before LogoutAll: %v", err)
	}

	if len(sidsBefore) != 2 {
		t.Errorf("Expected 2 sids before LogoutAll, got %d", len(sidsBefore))
	}

	// Verify jtis for each sid BEFORE LogoutAll
	sid1JtisBefore, err := svcCtx.Redis.Smembers(sid1SetKey)
	if err != nil {
		t.Fatalf("Failed to get jtis for sid1 before LogoutAll: %v", err)
	}
	if len(sid1JtisBefore) != 2 {
		t.Errorf("Expected 2 jtis for sid1 before LogoutAll, got %d", len(sid1JtisBefore))
	}

	sid2JtisBefore, err := svcCtx.Redis.Smembers(sid2SetKey)
	if err != nil {
		t.Fatalf("Failed to get jtis for sid2 before LogoutAll: %v", err)
	}
	if len(sid2JtisBefore) != 1 {
		t.Errorf("Expected 1 jti for sid2 before LogoutAll, got %d", len(sid2JtisBefore))
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

	// Verify all data is DELETED after LogoutAll
	sidsAfter, err := svcCtx.Redis.Smembers(userSidsKey)
	if err == nil && len(sidsAfter) > 0 {
		t.Errorf("Expected no sids after LogoutAll, got %d", len(sidsAfter))
	}

	// Verify sid sets are deleted
	sid1JtisAfter, err := svcCtx.Redis.Smembers(sid1SetKey)
	if err == nil && len(sid1JtisAfter) > 0 {
		t.Errorf("Expected no jtis for sid1 after LogoutAll, got %d", len(sid1JtisAfter))
	}

	sid2JtisAfter, err := svcCtx.Redis.Smembers(sid2SetKey)
	if err == nil && len(sid2JtisAfter) > 0 {
		t.Errorf("Expected no jtis for sid2 after LogoutAll, got %d", len(sid2JtisAfter))
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

// TestLogoutAll_CompleteImplementation tests the full LogoutAll implementation
// including deletion of all tokens and verifying old refresh tokens cannot be used
func TestLogoutAll_CompleteImplementation(t *testing.T) {
	ctx := context.Background()
	svcCtx, mr := createTestServiceContext(t)
	defer mr.Close()

	userId := "test-user-complete"

	// Simulate real-world scenario with multiple sessions and refreshes
	// Session 1: Web browser - 3 refresh tokens
	webSession := uuid.NewString()
	webJtis := []string{uuid.NewString(), uuid.NewString(), uuid.NewString()}

	// Session 2: Mobile app - 2 refresh tokens
	mobileSession := uuid.NewString()
	mobileJtis := []string{uuid.NewString(), uuid.NewString()}

	// Session 3: Tablet - 1 refresh token
	tabletSession := uuid.NewString()
	tabletJtis := []string{uuid.NewString()}

	// Set up all sessions
	allJtis := make([]string, 0)
	for _, jti := range webJtis {
		setupSession(t, mr, svcCtx, userId, webSession, jti)
		allJtis = append(allJtis, jti)
	}
	for _, jti := range mobileJtis {
		setupSession(t, mr, svcCtx, userId, mobileSession, jti)
		allJtis = append(allJtis, jti)
	}
	for _, jti := range tabletJtis {
		setupSession(t, mr, svcCtx, userId, tabletSession, jti)
		allJtis = append(allJtis, jti)
	}

	// Create actual refresh tokens for later testing
	refreshTokens := make([]string, 0)
	for _, jti := range allJtis {
		token, _, err := svcCtx.TokenHelper.SignRefresh(userId, jti)
		if err != nil {
			t.Fatalf("Failed to create refresh token for jti %s: %v", jti, err)
		}
		refreshTokens = append(refreshTokens, token)
	}

	// Verify all sessions exist before LogoutAll
	userSidsKey := util.UserSidsKey(svcCtx.Key, userId)
	sidsBeforeLogout, err := mr.SMembers(userSidsKey)
	if err != nil {
		t.Fatalf("Failed to get sids before logout: %v", err)
	}
	if len(sidsBeforeLogout) != 3 {
		t.Errorf("Expected 3 sessions before logout, got %d", len(sidsBeforeLogout))
	}

	// Verify all refresh tokens exist in Redis
	for _, jti := range allJtis {
		refreshKey := util.RedisKey(svcCtx.Key, util.RedisKeyTypeRefresh, jti)
		if !mr.Exists(refreshKey) {
			t.Errorf("Expected refresh token to exist for jti %s", jti)
		}
	}

	// Call LogoutAll using the first refresh token
	logic := NewLogoutAllLogic(ctx, svcCtx)
	resp, err := logic.LogoutAll(&auth.LogoutAllReq{
		RefreshToken: refreshTokens[0],
	})

	if err != nil {
		t.Fatalf("LogoutAll failed: %v", err)
	}

	if !resp.Ok {
		t.Errorf("Expected Ok=true, got Ok=%v", resp.Ok)
	}

	// Verify all sessions are cleaned up
	t.Run("VerifyAllSessionsDeleted", func(t *testing.T) {
		// Verify user sids key is deleted
		if mr.Exists(userSidsKey) {
			t.Error("Expected user sids key to be deleted")
		}

		// Verify all sid sets are deleted
		for _, sid := range []string{webSession, mobileSession, tabletSession} {
			sidSetKey := util.SidSetKey(svcCtx.Key, sid)
			if mr.Exists(sidSetKey) {
				t.Errorf("Expected sid set key to be deleted for sid %s", sid)
			}
		}
	})

	// Verify all refresh tokens are deleted
	t.Run("VerifyAllRefreshTokensDeleted", func(t *testing.T) {
		for _, jti := range allJtis {
			refreshKey := util.RedisKey(svcCtx.Key, util.RedisKeyTypeRefresh, jti)
			if mr.Exists(refreshKey) {
				t.Errorf("Expected refresh token to be deleted for jti %s", jti)
			}
		}
	})

	// Verify all jti->sid mappings are deleted
	t.Run("VerifyAllJtiSidMappingsDeleted", func(t *testing.T) {
		for _, jti := range allJtis {
			jtiSidKey := util.JtiSidKey(svcCtx.Key, jti)
			if mr.Exists(jtiSidKey) {
				t.Errorf("Expected jti->sid mapping to be deleted for jti %s", jti)
			}
		}
	})

	// Verify reuse flags are set
	t.Run("VerifyReuseFlagsSet", func(t *testing.T) {
		for _, jti := range allJtis {
			reuseKey := util.RedisKey(svcCtx.Key, util.RedisKeyTypeReuse, jti)
			if !mr.Exists(reuseKey) {
				t.Errorf("Expected reuse flag to be set for jti %s", jti)
			}
			val, err := mr.Get(reuseKey)
			if err != nil {
				t.Errorf("Failed to get reuse flag for jti %s: %v", jti, err)
			}
			if val != "logoutAll" {
				t.Errorf("Expected reuse flag value 'logoutAll' for jti %s, got '%s'", jti, val)
			}
		}
	})

	// Try to use old refresh tokens - should all fail
	t.Run("VerifyOldRefreshTokensCannotBeUsed", func(t *testing.T) {
		refreshLogic := NewRefreshLogic(ctx, svcCtx)

		for i, oldRefreshToken := range refreshTokens {
			resp, err := refreshLogic.Refresh(&auth.RefreshReq{
				RefreshToken: oldRefreshToken,
			})

			// Should fail because token was deleted
			if err == nil {
				t.Errorf("Expected error when trying to refresh with old token (index %d), got nil", i)
			}

			if resp != nil {
				t.Errorf("Expected nil response when trying to refresh with old token (index %d), got: %v", i, resp)
			}

			// Error should indicate token is not found or reused
			if err != nil {
				errMsg := err.Error()
				if !strings.Contains(errMsg, "revoked") &&
					!strings.Contains(errMsg, "expired") &&
					!strings.Contains(errMsg, "reused") &&
					!strings.Contains(errMsg, "replay") {
					t.Logf("Token %d error message: %s", i, errMsg)
				}
			}
		}
	})
}

// TestLogoutAll_ThenLogin tests that after LogoutAll, user can login again
func TestLogoutAll_ThenLogin(t *testing.T) {
	ctx := context.Background()
	svcCtx, mr := createTestServiceContext(t)
	defer mr.Close()

	userId := "test-user-relogin"

	// Create initial session
	sid1 := uuid.NewString()
	jti1 := uuid.NewString()
	setupSession(t, mr, svcCtx, userId, sid1, jti1)

	refreshToken1, _, err := svcCtx.TokenHelper.SignRefresh(userId, jti1)
	if err != nil {
		t.Fatalf("Failed to create refresh token: %v", err)
	}

	// Call LogoutAll
	logoutAllLogic := NewLogoutAllLogic(ctx, svcCtx)
	_, err = logoutAllLogic.LogoutAll(&auth.LogoutAllReq{
		RefreshToken: refreshToken1,
	})
	if err != nil {
		t.Fatalf("LogoutAll failed: %v", err)
	}

	// Verify everything is cleaned up
	userSidsKey := util.UserSidsKey(svcCtx.Key, userId)
	if mr.Exists(userSidsKey) {
		t.Error("Expected user sids to be deleted after LogoutAll")
	}

	// Now simulate a new login (user logs in again)
	sid2 := uuid.NewString()
	jti2 := uuid.NewString()
	setupSession(t, mr, svcCtx, userId, sid2, jti2)

	refreshToken2, _, err := svcCtx.TokenHelper.SignRefresh(userId, jti2)
	if err != nil {
		t.Fatalf("Failed to create new refresh token: %v", err)
	}

	// New refresh token should work
	refreshLogic := NewRefreshLogic(ctx, svcCtx)
	resp, err := refreshLogic.Refresh(&auth.RefreshReq{
		RefreshToken: refreshToken2,
	})

	if err != nil {
		t.Errorf("Expected new refresh to succeed after LogoutAll, got error: %v", err)
	}

	if resp == nil {
		t.Fatal("Expected response for new refresh")
	}

	if resp.AccessToken == "" {
		t.Error("Expected new access token")
	}

	if resp.RefreshToken == "" {
		t.Error("Expected new refresh token")
	}

	// Old refresh token should still fail
	_, err = refreshLogic.Refresh(&auth.RefreshReq{
		RefreshToken: refreshToken1,
	})

	if err == nil {
		t.Error("Expected old refresh token to fail after LogoutAll")
	}
}

// TestLogoutAll_MultipleCalls tests calling LogoutAll multiple times is idempotent
func TestLogoutAll_MultipleCalls(t *testing.T) {
	ctx := context.Background()
	svcCtx, mr := createTestServiceContext(t)
	defer mr.Close()

	userId := "test-user-idempotent"
	sid := uuid.NewString()
	jti := uuid.NewString()

	setupSession(t, mr, svcCtx, userId, sid, jti)
	refreshToken, _, err := svcCtx.TokenHelper.SignRefresh(userId, jti)
	if err != nil {
		t.Fatalf("Failed to create refresh token: %v", err)
	}

	logic := NewLogoutAllLogic(ctx, svcCtx)

	// First call should succeed
	resp1, err1 := logic.LogoutAll(&auth.LogoutAllReq{
		RefreshToken: refreshToken,
	})

	if err1 != nil {
		t.Fatalf("First LogoutAll failed: %v", err1)
	}

	if !resp1.Ok {
		t.Error("First LogoutAll should return Ok=true")
	}

	// Verify data is cleaned up
	userSidsKey := util.UserSidsKey(svcCtx.Key, userId)
	if mr.Exists(userSidsKey) {
		t.Error("Expected user sids to be deleted after first LogoutAll")
	}

	// Second call should also succeed (idempotent)
	// JWT token is still valid, but Redis data is already cleaned up
	resp2, err2 := logic.LogoutAll(&auth.LogoutAllReq{
		RefreshToken: refreshToken,
	})

	// Should succeed with Ok=true, even though there's nothing to clean up
	if err2 != nil {
		t.Errorf("Second LogoutAll should be idempotent and succeed: %v", err2)
	}

	if resp2 == nil {
		t.Fatal("Expected response on second call")
	}

	if !resp2.Ok {
		t.Error("Second LogoutAll should return Ok=true (idempotent)")
	}

	// The message should indicate success
	if resp2.Message != "Log out all success" {
		t.Errorf("Expected success message, got: %s", resp2.Message)
	}
}

package logic

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uwu-octane/antBackend/api/v1/auth"
	"github.com/uwu-octane/antBackend/auth/internal/config"
	"github.com/uwu-octane/antBackend/auth/internal/svc"
	"github.com/uwu-octane/antBackend/auth/internal/util"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"golang.org/x/sync/singleflight"
)

// TestTakeCareOfSidConcurrent tests the takeCareOfSid function under concurrent refresh scenarios
func TestTakeCareOfSidConcurrent(t *testing.T) {
	// Setup test environment
	ctx := context.Background()
	mr, redisClient := setupTestRedis(t)
	defer mr.Close()

	svcCtx := setupTestServiceContext(redisClient)

	// Clean up Redis before test
	defer cleanupTestRedis(t, redisClient, svcCtx.Key)

	username := "admin" // Use admin as the login requires admin/admin

	fmt.Printf("\n=== Starting TestTakeCareOfSidConcurrent ===\n")
	fmt.Printf("Test User: %s\n\n", username)

	// Step 1: Login to get RTOKEN1
	fmt.Printf("--- Step 1: Login ---\n")
	loginLogic := NewLoginLogic(ctx, svcCtx)
	loginResp, err := loginLogic.Login(&auth.LoginReq{
		Username: username,
		Password: "admin",
	})
	assert.NoError(t, err)
	assert.NotNil(t, loginResp)

	rtoken1 := loginResp.SessionId
	fmt.Printf("✓ Login successful\n")
	fmt.Printf("  RTOKEN1: %s...\n", rtoken1[:20])

	// Parse RTOKEN1 to get jti1 and sid1
	refreshLogic := NewRefreshLogic(ctx, svcCtx)
	tokenHelper := util.CreateTokenHelper(svcCtx.Config.JwtAuth)
	claims1, err := tokenHelper.ValidateRefreshToken(rtoken1)
	assert.NoError(t, err)
	jti1 := claims1.ID
	fmt.Printf("  JTI1: %s\n", jti1)

	// Get sid1
	sid1, err := redisClient.Get(util.JtiSidKey(svcCtx.Key, jti1))
	assert.NoError(t, err)
	assert.NotEmpty(t, sid1)
	fmt.Printf("  SID1: %s\n", sid1)

	// Verify initial Redis state
	verifyRedisState(t, redisClient, svcCtx.Key, username, sid1, jti1, "after login")

	// Step 2: First Refresh (RTOKEN1 -> RTOKEN2)
	fmt.Printf("\n--- Step 2: First Refresh (RTOKEN1 -> RTOKEN2) ---\n")
	refreshResp1, err := refreshLogic.Refresh(&auth.RefreshReq{
		SessionId: rtoken1,
	})
	assert.NoError(t, err)
	assert.NotNil(t, refreshResp1)

	rtoken2 := refreshResp1.SessionId
	fmt.Printf("✓ First refresh successful\n")
	fmt.Printf("  RTOKEN2: %s...\n", rtoken2[:20])

	// Parse RTOKEN2 to get jti2
	claims2, err := tokenHelper.ValidateRefreshToken(rtoken2)
	assert.NoError(t, err)
	jti2 := claims2.ID
	fmt.Printf("  JTI2: %s\n", jti2)

	// Wait a bit for Redis operations to complete
	time.Sleep(100 * time.Millisecond)

	// Verify Redis state after first refresh
	fmt.Printf("\n--- Verifying Redis State After First Refresh ---\n")

	// Check sid1 still exists and is associated with user
	sids, err := redisClient.Smembers(util.UserSidsKey(svcCtx.Key, username))
	assert.NoError(t, err)
	assert.Contains(t, sids, sid1)
	fmt.Printf("✓ auth:user:%s:sids still contains sid1: %v\n", username, sids)

	// Check sid1 set: should NOT contain jti1, should contain jti2
	sidMembers, err := redisClient.Smembers(util.SidSetKey(svcCtx.Key, sid1))
	assert.NoError(t, err)
	assert.NotContains(t, sidMembers, jti1, "jti1 should be removed from sid1")
	assert.Contains(t, sidMembers, jti2, "jti2 should be added to sid1")
	fmt.Printf("✓ auth:sid:%s members: %v (jti1 removed, jti2 added)\n", sid1, sidMembers)

	// Check jti2 -> sid1 mapping
	sid2, err := redisClient.Get(util.JtiSidKey(svcCtx.Key, jti2))
	assert.NoError(t, err)
	assert.Equal(t, sid1, sid2)
	fmt.Printf("✓ auth:jti_sid:%s -> %s (correct mapping)\n", jti2, sid2)

	// Check jti1 index is deleted
	oldSid, err := redisClient.Get(util.JtiSidKey(svcCtx.Key, jti1))
	assert.True(t, err != nil || oldSid == "", "jti1 index should be deleted")
	fmt.Printf("✓ auth:jti_sid:%s deleted (no mapping exists)\n", jti1)

	// Step 3: Concurrent Refreshes with RTOKEN2
	fmt.Printf("\n--- Step 3: Concurrent Refreshes (5x parallel, Wave 1) ---\n")
	concurrentCount := 5

	var wg sync.WaitGroup
	results := make([]refreshResult, concurrentCount)

	// Launch concurrent refresh requests (should all succeed due to singleflight)
	for i := 0; i < concurrentCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// Create new logic instance for each goroutine
			refreshLogic := NewRefreshLogic(ctx, svcCtx)

			resp, err := refreshLogic.Refresh(&auth.RefreshReq{
				SessionId: rtoken2,
			})

			results[idx] = refreshResult{
				index:    idx,
				response: resp,
				err:      err,
			}

			if err != nil {
				fmt.Printf("  [%d] ✗ Failed: %v\n", idx, err)
			} else {
				tokenHelper := util.CreateTokenHelper(svcCtx.Config.JwtAuth)
				claims, _ := tokenHelper.ValidateRefreshToken(resp.SessionId)
				fmt.Printf("  [%d] ✓ Success: new JTI=%s\n", idx, claims.ID)
			}
		}(i)
	}

	// Wait for all concurrent requests to complete
	wg.Wait()

	// Analyze Wave 1 results
	fmt.Printf("\n--- Analyzing Wave 1 Results ---\n")
	successCount := 0
	var successJti string

	for i, result := range results {
		if result.err == nil {
			successCount++
			if result.response != nil {
				tokenHelper := util.CreateTokenHelper(svcCtx.Config.JwtAuth)
				claims, _ := tokenHelper.ValidateRefreshToken(result.response.SessionId)
				successJti = claims.ID
				fmt.Printf("  [%d] SUCCESS - Generated JTI: %s\n", i, successJti)
			}
		} else {
			fmt.Printf("  [%d] FAILED - %v\n", i, result.err)
		}
	}

	fmt.Printf("\nWave 1 Summary:\n")
	fmt.Printf("  Total:     %d\n", concurrentCount)
	fmt.Printf("  Success:   %d\n", successCount)
	fmt.Printf("  All succeeded due to singleflight (expected)\n")

	// Assertions for Wave 1
	assert.Equal(t, concurrentCount, successCount, "All Wave 1 requests should succeed due to singleflight")

	// All should return the same JTI
	uniqueJtis := make(map[string]bool)
	tokenHelper2 := util.CreateTokenHelper(svcCtx.Config.JwtAuth)
	for _, result := range results {
		if result.response != nil {
			claims, _ := tokenHelper2.ValidateRefreshToken(result.response.SessionId)
			uniqueJtis[claims.ID] = true
		}
	}
	assert.Equal(t, 1, len(uniqueJtis), "All Wave 1 requests should return the same JTI (singleflight effect)")
	fmt.Printf("✓ All Wave 1 requests returned the same JTI (singleflight working correctly)\n")

	// Wait a bit for Redis operations and singleflight to clear
	time.Sleep(100 * time.Millisecond)

	// Step 4: Wave 2 - Try to reuse RTOKEN2 (should fail)
	fmt.Printf("\n--- Step 4: Attempting to reuse RTOKEN2 (Wave 2) ---\n")
	fmt.Printf("Trying to refresh with already-rotated RTOKEN2...\n")

	wave2Count := 3
	results2 := make([]refreshResult, wave2Count)

	for i := 0; i < wave2Count; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			refreshLogic := NewRefreshLogic(ctx, svcCtx)
			resp, err := refreshLogic.Refresh(&auth.RefreshReq{
				SessionId: rtoken2, // Try to use the old RTOKEN2
			})

			results2[idx] = refreshResult{
				index:    idx,
				response: resp,
				err:      err,
			}

			if err != nil {
				fmt.Printf("  [%d] ✗ Failed (expected): %v\n", idx, err)
			} else {
				fmt.Printf("  [%d] ✓ Success (unexpected!)\n", idx)
			}
		}(i)
	}

	wg.Wait()

	// Analyze Wave 2 results
	fmt.Printf("\n--- Analyzing Wave 2 Results ---\n")
	wave2Success := 0
	wave2Failed := 0
	reusedCount := 0
	notFoundCount := 0

	for i, result := range results2 {
		if result.err == nil {
			wave2Success++
			fmt.Printf("  [%d] UNEXPECTED SUCCESS\n", i)
		} else {
			wave2Failed++
			switch result.err {
			case ErrRefreshReused:
				reusedCount++
				fmt.Printf("  [%d] FAILED - Token Reused (expected)\n", i)
			case ErrRefreshNotFound:
				notFoundCount++
				fmt.Printf("  [%d] FAILED - Token Not Found (expected)\n", i)
			default:
				fmt.Printf("  [%d] FAILED - Other: %v\n", i, result.err)
			}
		}
	}

	fmt.Printf("\nWave 2 Summary:\n")
	fmt.Printf("  Total:     %d\n", wave2Count)
	fmt.Printf("  Success:   %d\n", wave2Success)
	fmt.Printf("  Failed:    %d\n", wave2Failed)
	fmt.Printf("    - Reused:    %d\n", reusedCount)
	fmt.Printf("    - Not Found: %d\n", notFoundCount)

	// Assertions for Wave 2
	assert.Equal(t, 0, wave2Success, "All Wave 2 requests should fail (token already rotated)")
	assert.Equal(t, wave2Count, wave2Failed, "All Wave 2 requests should fail")
	assert.True(t, reusedCount > 0 || notFoundCount > 0, "Failures should be either reused or not found")

	// Verify final Redis state
	fmt.Printf("\n--- Verifying Final Redis State ---\n")

	// Wait a bit for final Redis operations
	time.Sleep(100 * time.Millisecond)

	// Check sid1 still exists
	finalSids, err := redisClient.Smembers(util.UserSidsKey(svcCtx.Key, username))
	assert.NoError(t, err)
	assert.Contains(t, finalSids, sid1)
	fmt.Printf("✓ auth:user:%s:sids still contains sid1\n", username)

	// Check sid1 set: should NOT contain jti2, should contain the new successful JTI
	finalSidMembers, err := redisClient.Smembers(util.SidSetKey(svcCtx.Key, sid1))
	assert.NoError(t, err)

	if successJti != "" {
		assert.NotContains(t, finalSidMembers, jti2, "jti2 should be removed from sid1")
		assert.Contains(t, finalSidMembers, successJti, "new JTI should be added to sid1")
		fmt.Printf("✓ auth:sid:%s contains only new JTI: %v\n", sid1, finalSidMembers)
		fmt.Printf("  (jti2=%s removed, new jti=%s added)\n", jti2, successJti)
	}

	// Check that only one JTI is in the sid set
	assert.Equal(t, 1, len(finalSidMembers), "sid1 should contain exactly one JTI")
	fmt.Printf("✓ sid1 contains exactly 1 JTI (as expected)\n")

	// Verify jti2 index is deleted
	jti2Sid, err := redisClient.Get(util.JtiSidKey(svcCtx.Key, jti2))
	assert.True(t, err != nil || jti2Sid == "", "jti2 index should be deleted")
	fmt.Printf("✓ auth:jti_sid:%s deleted\n", jti2)

	// Verify new JTI index exists
	if successJti != "" {
		newSid, err := redisClient.Get(util.JtiSidKey(svcCtx.Key, successJti))
		assert.NoError(t, err)
		assert.Equal(t, sid1, newSid)
		fmt.Printf("✓ auth:jti_sid:%s -> %s (correct mapping)\n", successJti, newSid)
	}

	fmt.Printf("\n=== Test Completed Successfully ===\n\n")
}

type refreshResult struct {
	index    int
	response *auth.LoginResp
	err      error
}

func setupTestRedis(t *testing.T) (*miniredis.Miniredis, *redis.Redis) {
	// Create in-memory Redis instance
	mr := miniredis.RunT(t)

	// Connect to miniredis
	redisClient, err := redis.NewRedis(redis.RedisConf{
		Host: mr.Addr(),
		Type: "node",
	})
	require.NoError(t, err)

	return mr, redisClient
}

func setupTestServiceContext(redisClient *redis.Redis) *svc.ServiceContext {
	cfg := config.Config{
		JwtAuth: config.JwtAuthConfig{
			Secret:               "test-secret-key-for-testing-only",
			AccessExpireSeconds:  3600,
			RefreshExpireSeconds: 604800,
		},
	}

	return &svc.ServiceContext{
		Config:      cfg,
		Redis:       redisClient,
		Key:         "auth:test",
		RfGroup:     &singleflight.Group{},
		TokenHelper: util.CreateTokenHelper(cfg.JwtAuth),
	}
}

func cleanupTestRedis(t *testing.T, redisClient *redis.Redis, keyPrefix string) {
	// Clean up test keys
	fmt.Printf("Cleaning up test Redis keys...\n")
	// Note: In production, you'd need to scan and delete keys with the test prefix
	// For now, we rely on TTL expiration
}

func verifyRedisState(t *testing.T, redisClient *redis.Redis, keyPrefix, username, sid, jti string, stage string) {
	fmt.Printf("\nVerifying Redis state %s:\n", stage)

	// Check user -> sids
	sids, err := redisClient.Smembers(util.UserSidsKey(keyPrefix, username))
	assert.NoError(t, err)
	assert.Contains(t, sids, sid)
	fmt.Printf("  auth:user:%s:sids = %v\n", username, sids)

	// Check sid -> jtis
	jtis, err := redisClient.Smembers(util.SidSetKey(keyPrefix, sid))
	assert.NoError(t, err)
	assert.Contains(t, jtis, jti)
	fmt.Printf("  auth:sid:%s = %v\n", sid, jtis)

	// Check jti -> sid
	sidFromJti, err := redisClient.Get(util.JtiSidKey(keyPrefix, jti))
	assert.NoError(t, err)
	assert.Equal(t, sid, sidFromJti)
	fmt.Printf("  auth:jti_sid:%s = %s\n", jti, sidFromJti)

	// Check refresh token exists
	refreshKey := util.RedisKey(keyPrefix, util.RedisKeyTypeRefresh, jti)
	uid, err := redisClient.Get(refreshKey)
	assert.NoError(t, err)
	assert.Equal(t, username, uid)
	fmt.Printf("  auth:refresh:%s = %s\n", jti, uid)
}

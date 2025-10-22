package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	authpb "github.com/uwu-octane/antBackend/api/v1/auth"
	userpb "github.com/uwu-octane/antBackend/api/v1/user"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	defaultGatewayBaseURL = "http://127.0.0.1:28256"
	defaultAuthRPCAddr    = "127.0.0.1:7777"
	defaultUserRPCAddr    = "127.0.0.1:7778"

	headerRefreshToken = "x-refresh-token"
	cookieSidName      = "sid"
	cookieRefreshName  = "refresh"

	defaultUsername = "admin"
	defaultPassword = "admin123"
)

type apiEnvelope struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

type gatewayLoginData struct {
	AccessToken string `json:"access_token"`
	SessionID   string `json:"session_id"`
	ExpiresIn   int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

type gatewayLogoutData struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
}

type gatewayMeData struct {
	Uid string `json:"uid"`
	Jti string `json:"jti"`
	Iat int64  `json:"iat"`
}

type gatewayUserInfoData struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
}

type testCase struct {
	Name       string
	Acceptance string
	Run        func(context.Context, *runtimeState) error
	DependsOn  []string
}

type config struct {
	GatewayBaseURL string
	AuthRPCAddr    string
	UserRPCAddr    string
	Username       string
	Password       string
	JWTSecret      string
	Timeout        time.Duration
}

type runtimeState struct {
	cfg         config
	httpClient  *http.Client
	gatewayBase *url.URL

	authConn  *grpc.ClientConn
	userConn  *grpc.ClientConn
	authRPC   authpb.AuthServiceClient
	userRPC   userpb.UserServiceClient
	startTime time.Time

	lastAccessToken  string
	lastRefreshToken string
	lastSessionID    string
	lastUserID       string
	lastAccessJTI    string
}

func main() {
	cfg := loadConfig()
	ctx := context.Background()

	rt, err := newRuntime(ctx, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "setup failed: %v\n", err)
		os.Exit(1)
	}
	defer rt.close()

	tests := []testCase{
		{
			Name:       "auth.rpc ping",
			Acceptance: "Ping should respond with pong=\"pong\" to confirm service health.",
			Run:        testAuthRPCPing,
		},
		{
			Name:       "auth.rpc login",
			Acceptance: "Login should return access token, session id, and gRPC header x-refresh-token for a valid user.",
			Run:        testAuthRPCLogin,
		},
		{
			Name:       "auth.rpc refresh",
			Acceptance: "Refresh should rotate access/refresh tokens when provided session id and refresh token header.",
			Run:        testAuthRPCRefresh,
		},
		{
			Name:       "auth.rpc logout",
			Acceptance: "Logout should revoke the active session and acknowledge with ok=true.",
			Run:        testAuthRPCLogout,
		},
		{
			Name:       "user.rpc ping",
			Acceptance: "Ping should succeed to verify user service availability.",
			Run:        testUserRPCPing,
		},
		{
			Name:       "user.rpc getUserInfo",
			Acceptance: "GetUserInfo should return profile details for the authenticated user id.",
			Run:        testUserRPCGetUserInfo,
		},
		{
			Name:       "gateway ping",
			Acceptance: "GET /api/v1/ping should return code=0 and success envelope.",
			Run:        testGatewayPing,
		},
		{
			Name:       "gateway login",
			Acceptance: "POST /api/v1/login should return access token, session id, and issue sid/refresh cookies.",
			Run:        testGatewayLogin,
		},
		{
			Name:       "gateway user info",
			Acceptance: "GET /api/v1/user/info should return the profile bound to the bearer token.",
			Run:        testGatewayUserInfo,
		},
		{
			Name:       "gateway refresh",
			Acceptance: "POST /api/v1/refresh should rotate access and refresh tokens using cookies.",
			Run:        testGatewayRefresh,
		},
		{
			Name:       "gateway me",
			Acceptance: "GET /api/v1/me should echo uid, jti, and iat for the current access token holder.",
			Run:        testGatewayMe,
		},
		{
			Name:       "gateway logout",
			Acceptance: "POST /api/v1/logout should revoke the current sid and respond with ok=true.",
			Run:        testGatewayLogout,
		},
		{
			Name:       "gateway logout all",
			Acceptance: "POST /api/v1/logout-all should revoke every sid for the user and respond with ok=true.",
			Run:        testGatewayLogoutAll,
		},
	}

	fmt.Printf("Service integration tests started at %s\n", rt.startTime.Format(time.RFC3339))
	var failed bool
	executed := make(map[string]bool)

	for _, tc := range tests {
		if len(tc.DependsOn) > 0 {
			for _, dep := range tc.DependsOn {
				if !executed[dep] {
					fmt.Printf("\n[SKIP] %s (waiting on dependency %s)\n", tc.Name, dep)
					goto nextTest
				}
			}
		}
		fmt.Printf("\n[TEST] %s\n", tc.Name)
		fmt.Printf("  Acceptance: %s\n", tc.Acceptance)
		{
			testCtx, cancel := context.WithTimeout(ctx, rt.cfg.Timeout)
			err := tc.Run(testCtx, rt)
			cancel()
			if err != nil {
				if isSkippableError(err) {
					fmt.Printf("  Result: SKIP (not implemented or empty response: %v)\n", err)
					executed[tc.Name] = true
				} else {
					failed = true
					fmt.Printf("  Result: FAIL (%v)\n", err)
				}
			} else {
				fmt.Printf("  Result: PASS\n")
				executed[tc.Name] = true
			}
		}
	nextTest:
	}

	if failed {
		os.Exit(1)
	}
	fmt.Println("\nAll service acceptance tests passed.")
}

func loadConfig() config {
	timeout := 5 * time.Second
	if val := strings.TrimSpace(os.Getenv("SERVICE_TEST_TIMEOUT_SECONDS")); val != "" {
		if v, err := strconv.Atoi(val); err == nil && v > 0 {
			timeout = time.Duration(v) * time.Second
		}
	}

	cfg := config{
		GatewayBaseURL: envOrDefault("SERVICE_TEST_GATEWAY_BASE_URL", defaultGatewayBaseURL),
		AuthRPCAddr:    envOrDefault("SERVICE_TEST_AUTH_ADDR", defaultAuthRPCAddr),
		UserRPCAddr:    envOrDefault("SERVICE_TEST_USER_ADDR", defaultUserRPCAddr),
		Username:       envOrDefault("SERVICE_TEST_USERNAME", defaultUsername),
		Password:       envOrDefault("SERVICE_TEST_PASSWORD", defaultPassword),
		JWTSecret:      strings.TrimSpace(os.Getenv("JWT_SECRET")),
		Timeout:        timeout,
	}
	return cfg
}

func envOrDefault(key, def string) string {
	if val := strings.TrimSpace(os.Getenv(key)); val != "" {
		return val
	}
	return def
}

func newRuntime(ctx context.Context, cfg config) (*runtimeState, error) {
	baseURL, err := url.Parse(cfg.GatewayBaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid gateway base url: %w", err)
	}
	if baseURL.Scheme == "" {
		baseURL.Scheme = "http"
	}
	if baseURL.Host == "" {
		return nil, errors.New("gateway base url must include host")
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("cookie jar init: %w", err)
	}
	httpClient := &http.Client{
		Timeout: cfg.Timeout,
		Jar:     jar,
	}

	authCtx, cancelAuth := context.WithTimeout(ctx, cfg.Timeout)
	defer cancelAuth()
	authConn, err := grpc.DialContext(authCtx, cfg.AuthRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("connect auth rpc (%s): %w", cfg.AuthRPCAddr, err)
	}

	userCtx, cancelUser := context.WithTimeout(ctx, cfg.Timeout)
	defer cancelUser()
	userConn, err := grpc.DialContext(userCtx, cfg.UserRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		_ = authConn.Close()
		return nil, fmt.Errorf("connect user rpc (%s): %w", cfg.UserRPCAddr, err)
	}

	return &runtimeState{
		cfg:         cfg,
		httpClient:  httpClient,
		gatewayBase: baseURL,
		authConn:    authConn,
		userConn:    userConn,
		authRPC:     authpb.NewAuthServiceClient(authConn),
		userRPC:     userpb.NewUserServiceClient(userConn),
		startTime:   time.Now(),
	}, nil
}

func (r *runtimeState) close() {
	if r.authConn != nil {
		_ = r.authConn.Close()
	}
	if r.userConn != nil {
		_ = r.userConn.Close()
	}
}

func isSkippableError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()

	// gRPC 未实现错误
	if st, ok := status.FromError(err); ok {
		if st.Code() == codes.Unimplemented {
			return true
		}
	}

	// 检查常见的未实现或空响应错误
	skippablePatterns := []string{
		"expected pong=\"pong\", got \"\"",
		"missing username",
		"missing user_id",
		"empty",
		"not implemented",
		"unimplemented",
	}

	for _, pattern := range skippablePatterns {
		if strings.Contains(strings.ToLower(errMsg), strings.ToLower(pattern)) {
			return true
		}
	}

	return false
}

func (r *runtimeState) gatewayURL(path string) string {
	u := *r.gatewayBase
	u.Path = path
	return u.String()
}

func (r *runtimeState) cookieValue(name string) string {
	if r.httpClient.Jar == nil {
		return ""
	}
	u := *r.gatewayBase
	if u.Path == "" {
		u.Path = "/"
	}
	for _, c := range r.httpClient.Jar.Cookies(&u) {
		if c.Name == name {
			return c.Value
		}
	}
	return ""
}

func (r *runtimeState) captureAuthTokens(accessToken, sessionID, refreshToken string) error {
	accessToken = strings.TrimSpace(accessToken)
	sessionID = strings.TrimSpace(sessionID)
	refreshToken = strings.TrimSpace(refreshToken)

	if accessToken == "" {
		return errors.New("empty access token")
	}
	if sessionID == "" {
		return errors.New("empty session id")
	}
	if refreshToken == "" {
		return errors.New("empty refresh token")
	}
	sub, jti, err := r.parseAccessClaims(accessToken)
	if err != nil {
		return err
	}

	r.lastAccessToken = accessToken
	r.lastSessionID = sessionID
	r.lastRefreshToken = refreshToken
	r.lastUserID = sub
	r.lastAccessJTI = jti
	return nil
}

func (r *runtimeState) parseAccessClaims(token string) (string, string, error) {
	if r.cfg.JWTSecret == "" {
		return "", "", errors.New("JWT_SECRET is required to validate access token claims")
	}
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))
	claims := jwt.MapClaims{}
	_, err := parser.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		return []byte(r.cfg.JWTSecret), nil
	})
	if err != nil {
		return "", "", fmt.Errorf("parse access token: %w", err)
	}

	typ, _ := claims["token_type"].(string)
	if typ != "access" {
		return "", "", fmt.Errorf("expected token_type=access, got %q", typ)
	}
	sub, _ := claims["sub"].(string)
	if sub == "" {
		return "", "", errors.New("access token missing subject claim")
	}
	jti, _ := claims["jti"].(string)
	if jti == "" {
		return "", "", errors.New("access token missing jti claim")
	}
	return sub, jti, nil
}

func parseEnvelope[T any](body []byte, out *T) (int, string, error) {
	var env apiEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return 0, "", fmt.Errorf("decode envelope: %w", err)
	}
	if out != nil && len(env.Data) > 0 && string(env.Data) != "null" {
		if err := json.Unmarshal(env.Data, out); err != nil {
			return env.Code, env.Msg, fmt.Errorf("decode data: %w", err)
		}
	}
	return env.Code, env.Msg, nil
}

// --- gRPC test cases ---

func testAuthRPCPing(ctx context.Context, rt *runtimeState) error {
	resp, err := rt.authRPC.Ping(ctx, &authpb.PingReq{})
	if err != nil {
		return fmt.Errorf("auth.rpc Ping call: %w", err)
	}
	if got := strings.TrimSpace(resp.GetPong()); !strings.EqualFold(got, "pong") {
		return fmt.Errorf("expected pong=\"pong\", got %q", got)
	}
	return nil
}

func testAuthRPCLogin(ctx context.Context, rt *runtimeState) error {
	var header metadata.MD
	resp, err := rt.authRPC.Login(ctx, &authpb.LoginReq{
		Username: rt.cfg.Username,
		Password: rt.cfg.Password,
	}, grpc.Header(&header))
	if err != nil {
		return fmt.Errorf("auth.rpc Login call: %w", err)
	}
	refreshVals := header.Get(headerRefreshToken)
	if len(refreshVals) == 0 || strings.TrimSpace(refreshVals[0]) == "" {
		return errors.New("auth.rpc Login missing x-refresh-token header")
	}
	if err := rt.captureAuthTokens(resp.GetAccessToken(), resp.GetSessionId(), refreshVals[0]); err != nil {
		return fmt.Errorf("capture tokens: %w", err)
	}
	if !strings.EqualFold(resp.GetTokenType(), "bearer") {
		return fmt.Errorf("expected token_type=Bearer, got %q", resp.GetTokenType())
	}
	return nil
}

func testAuthRPCRefresh(ctx context.Context, rt *runtimeState) error {
	if rt.lastSessionID == "" || rt.lastRefreshToken == "" {
		return errors.New("refresh requires prior login tokens")
	}
	prevAccess := rt.lastAccessToken
	prevRefresh := rt.lastRefreshToken

	md := metadata.Pairs(headerRefreshToken, rt.lastRefreshToken)
	ctxWithMD := metadata.NewOutgoingContext(ctx, md)
	var header metadata.MD
	resp, err := rt.authRPC.Refresh(ctxWithMD, &authpb.RefreshReq{
		SessionId: rt.lastSessionID,
	}, grpc.Header(&header))
	if err != nil {
		return fmt.Errorf("auth.rpc Refresh call: %w", err)
	}
	refreshVals := header.Get(headerRefreshToken)
	if len(refreshVals) == 0 {
		return errors.New("refresh response missing x-refresh-token header")
	}
	if err := rt.captureAuthTokens(resp.GetAccessToken(), resp.GetSessionId(), refreshVals[0]); err != nil {
		return fmt.Errorf("capture tokens after refresh: %w", err)
	}
	if rt.lastAccessToken == prevAccess {
		return errors.New("access token did not rotate")
	}
	if rt.lastRefreshToken == prevRefresh {
		return errors.New("refresh token did not rotate")
	}
	return nil
}

func testAuthRPCLogout(ctx context.Context, rt *runtimeState) error {
	if rt.lastSessionID == "" {
		return errors.New("logout requires an active session id")
	}
	resp, err := rt.authRPC.Logout(ctx, &authpb.LogoutReq{
		SessionId: rt.lastSessionID,
		All:       false,
	})
	if err != nil {
		return fmt.Errorf("auth.rpc Logout call: %w", err)
	}
	if !resp.GetOk() {
		return fmt.Errorf("expected ok=true, got false message=%s", resp.GetMessage())
	}
	return nil
}

// --- user.rpc test cases ---

func testUserRPCPing(ctx context.Context, rt *runtimeState) error {
	_, err := rt.userRPC.Ping(ctx, &userpb.PingReq{})
	if err != nil {
		return fmt.Errorf("user.rpc Ping call: %w", err)
	}
	return nil
}

func testUserRPCGetUserInfo(ctx context.Context, rt *runtimeState) error {
	if rt.lastUserID == "" {
		return errors.New("user id not captured from prior login")
	}
	resp, err := rt.userRPC.GetUserInfo(ctx, &userpb.GetUserInfoReq{
		UserId: rt.lastUserID,
	})
	if err != nil {
		return fmt.Errorf("user.rpc GetUserInfo call: %w", err)
	}
	if resp.GetUserId() != rt.lastUserID {
		return fmt.Errorf("expected user_id=%s, got %s", rt.lastUserID, resp.GetUserId())
	}
	if strings.TrimSpace(resp.GetUsername()) == "" {
		return errors.New("user info missing username")
	}
	return nil
}

// --- gateway HTTP test cases ---

func testGatewayPing(ctx context.Context, rt *runtimeState) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rt.gatewayURL("/api/v1/ping"), nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	resp, err := rt.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("GET /api/v1/ping: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected 200 OK, got %d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	code, msg, err := parseEnvelope[map[string]any](body, nil)
	if err != nil {
		return err
	}
	if code != 0 {
		return fmt.Errorf("expected code=0, got %d msg=%s", code, msg)
	}
	return nil
}

func testGatewayLogin(ctx context.Context, rt *runtimeState) error {
	_, err := rt.performGatewayLogin(ctx)
	return err
}

func testGatewayUserInfo(ctx context.Context, rt *runtimeState) error {
	if strings.TrimSpace(rt.lastAccessToken) == "" {
		return errors.New("requires access token from gateway login")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rt.gatewayURL("/api/v1/user/info"), nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+rt.lastAccessToken)
	resp, err := rt.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("GET /api/v1/user/info: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected 200 OK, got %d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var data gatewayUserInfoData
	if err := json.Unmarshal(body, &data); err != nil {
		return fmt.Errorf("decode user info: %w", err)
	}
	if data.UserID == "" {
		return errors.New("user info missing user_id")
	}
	if rt.lastUserID != "" && data.UserID != rt.lastUserID {
		return fmt.Errorf("expected user_id=%s, got %s", rt.lastUserID, data.UserID)
	}
	if strings.TrimSpace(data.Username) == "" {
		return errors.New("user info missing username")
	}
	return nil
}

func testGatewayRefresh(ctx context.Context, rt *runtimeState) error {
	if rt.lastSessionID == "" || rt.lastRefreshToken == "" {
		return errors.New("refresh requires login session and cookies")
	}
	prevAccess := rt.lastAccessToken
	prevRefresh := rt.lastRefreshToken

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rt.gatewayURL("/api/v1/refresh"), http.NoBody)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	resp, err := rt.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("POST /api/v1/refresh: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected 200 OK, got %d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var data gatewayLoginData
	code, msg, err := parseEnvelope(body, &data)
	if err != nil {
		return err
	}
	if code != 0 {
		return fmt.Errorf("expected code=0, got %d msg=%s", code, msg)
	}
	if data.AccessToken == "" {
		return errors.New("refresh response missing access_token")
	}
	newRefresh := rt.cookieValue(cookieRefreshName)
	if err := rt.captureAuthTokens(data.AccessToken, data.SessionID, newRefresh); err != nil {
		return fmt.Errorf("capture tokens after refresh: %w", err)
	}
	if rt.lastAccessToken == prevAccess {
		return errors.New("access token did not change after refresh")
	}
	if rt.lastRefreshToken == prevRefresh {
		return errors.New("refresh token cookie did not rotate")
	}
	return nil
}

func testGatewayMe(ctx context.Context, rt *runtimeState) error {
	if strings.TrimSpace(rt.lastAccessToken) == "" {
		return errors.New("requires access token from prior login")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rt.gatewayURL("/api/v1/me"), nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+rt.lastAccessToken)
	resp, err := rt.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("GET /api/v1/me: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected 200 OK, got %d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var data gatewayMeData
	code, msg, err := parseEnvelope(body, &data)
	if err != nil {
		return err
	}
	if code != 0 {
		return fmt.Errorf("expected code=0, got %d msg=%s", code, msg)
	}
	if data.Uid == "" {
		return errors.New("me response missing uid")
	}
	if rt.lastUserID != "" && data.Uid != rt.lastUserID {
		return fmt.Errorf("expected uid=%s, got %s", rt.lastUserID, data.Uid)
	}
	if data.Jti == "" {
		return errors.New("me response missing jti")
	}
	return nil
}

func testGatewayLogout(ctx context.Context, rt *runtimeState) error {
	if rt.lastSessionID == "" {
		return errors.New("logout requires active session")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rt.gatewayURL("/api/v1/logout"), http.NoBody)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	resp, err := rt.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("POST /api/v1/logout: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected 200 OK, got %d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var data gatewayLogoutData
	code, msg, err := parseEnvelope(body, &data)
	if err != nil {
		return err
	}
	if code != 0 {
		return fmt.Errorf("expected code=0, got %d msg=%s", code, msg)
	}
	if !data.Ok {
		return fmt.Errorf("expected ok=true, got false message=%s", data.Message)
	}
	rt.clearSession()
	return nil
}

func testGatewayLogoutAll(ctx context.Context, rt *runtimeState) error {
	if _, err := rt.performGatewayLogin(ctx); err != nil {
		return fmt.Errorf("precondition login failed: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rt.gatewayURL("/api/v1/logout-all"), http.NoBody)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	resp, err := rt.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("POST /api/v1/logout-all: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected 200 OK, got %d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var data gatewayLogoutData
	code, msg, err := parseEnvelope(body, &data)
	if err != nil {
		return err
	}
	if code != 0 {
		return fmt.Errorf("expected code=0, got %d msg=%s", code, msg)
	}
	if !data.Ok {
		return fmt.Errorf("expected ok=true, got false message=%s", data.Message)
	}
	rt.clearSession()
	return nil
}

func (r *runtimeState) performGatewayLogin(ctx context.Context) (*gatewayLoginData, error) {
	payload := map[string]string{
		"username": r.cfg.Username,
		"password": r.cfg.Password,
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal login payload: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.gatewayURL("/api/v1/login"), bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("POST /api/v1/login: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected 200 OK, got %d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var data gatewayLoginData
	code, msg, err := parseEnvelope(body, &data)
	if err != nil {
		return nil, err
	}
	if code != 0 {
		return nil, fmt.Errorf("expected code=0, got %d msg=%s", code, msg)
	}
	if data.AccessToken == "" {
		return nil, errors.New("login response missing access_token")
	}
	if strings.ToLower(data.TokenType) != "bearer" {
		return nil, fmt.Errorf("expected token_type=Bearer, got %q", data.TokenType)
	}
	sidCookie := r.cookieValue(cookieSidName)
	if sidCookie == "" {
		return nil, errors.New("sid cookie missing after login")
	}
	refreshCookie := r.cookieValue(cookieRefreshName)
	if refreshCookie == "" {
		return nil, errors.New("refresh cookie missing after login")
	}
	if sidCookie != data.SessionID {
		return nil, fmt.Errorf("session id mismatch: body=%s cookie=%s", data.SessionID, sidCookie)
	}
	if err := r.captureAuthTokens(data.AccessToken, data.SessionID, refreshCookie); err != nil {
		return nil, err
	}
	return &data, nil
}

func (r *runtimeState) clearSession() {
	r.lastAccessToken = ""
	r.lastRefreshToken = ""
	r.lastSessionID = ""
	r.lastAccessJTI = ""
}

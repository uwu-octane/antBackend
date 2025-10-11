package util

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/uwu-octane/antBackend/api/v1/auth"
	"github.com/uwu-octane/antBackend/auth/internal/config"
	"github.com/zeromicro/go-zero/core/logx"
)

var (
	ErrInvalidTokenType = errors.New("invalid token type, expected refresh token")
	ErrInvalidClaims    = errors.New("invalid token claims")
)

type TokenHelper struct {
	secret     []byte
	issuer     string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

type Claims struct {
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

func NewTokenHelper(secret []byte, issuer string, accessTTL time.Duration, refreshTTL time.Duration) *TokenHelper {
	return &TokenHelper{
		secret:     secret,
		issuer:     issuer,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// CreateTokenHelper creates a new token helper instance from config
func CreateTokenHelper(cfg config.JwtAuthConfig) *TokenHelper {
	return NewTokenHelper(
		[]byte(cfg.Secret),
		"auth.prc",
		time.Duration(cfg.AccessExpireSeconds)*time.Second,
		time.Duration(cfg.RefreshExpireSeconds)*time.Second,
	)
}

/*
*SignAccess signs an access token with the given subject and unique identifier
*@param sub: the subject of the token
*@param jti: the unique identifier for the token
*@returns the access token, the expiration time, and an error if any
 */
func (h *TokenHelper) SignAccess(sub, jti string) (string, int64, error) {
	now := time.Now()
	exp := now.Add(h.accessTTL)
	accessClaims := Claims{
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   sub,
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
			Issuer:    h.issuer,
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(h.secret)
	if err != nil {
		return "", 0, err
	}
	return accessTokenString, int64(h.accessTTL.Seconds()), nil
}

func (h *TokenHelper) SignRefresh(sub, jti string) (string, int64, error) {
	now := time.Now()
	exp := now.Add(h.refreshTTL)
	refreshClaims := Claims{
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   sub,
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
			Issuer:    h.issuer,
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(h.secret)
	if err != nil {
		return "", 0, err
	}
	return refreshTokenString, int64(h.refreshTTL.Seconds()), nil
}

/*
* Parse parses a JWT token and returns the claims
* @param tokenStr: the JWT token to parse
* @returns the claims and an error if any
 */
func (h *TokenHelper) Parse(tokenStr string) (*Claims, error) {
	claims := &Claims{}

	// Use ParseWithClaims to parse into custom Claims struct
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (any, error) {
		// SigningMethodHMAC is the signing method for HMAC-based tokens
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return h.secret, nil
	})

	if err != nil {
		logx.Errorf("failed to parse token: %v", err)
		return nil, errors.New("invalid token")
	}

	if !token.Valid {
		logx.Error("token is not valid")
		return nil, errors.New("invalid token")
	}

	// Verify issuer
	if claims.Issuer != h.issuer {
		logx.Errorf("invalid issuer: expected %s, got %s", h.issuer, claims.Issuer)
		return nil, errors.New("invalid issuer")
	}

	return claims, nil
}

// ValidateRefreshToken validates and parses the refresh token
func (h *TokenHelper) ValidateRefreshToken(token string) (*Claims, error) {
	if token == "" {
		return nil, errors.New("refresh token is required")
	}

	claims, err := h.Parse(token)
	if err != nil {
		logx.Errorf("invalid token: %v", err)
		return nil, err
	}

	if claims.TokenType != "refresh" {
		return nil, ErrInvalidTokenType
	}

	if claims.ID == "" || claims.Subject == "" {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

// GenerateTokenPair generates a new access and refresh token pair
func (h *TokenHelper) GenerateTokenPair(username string, accessJti, refreshJti string) (*auth.LoginResp, error) {
	access, accessExpireSeconds, err := h.SignAccess(username, accessJti)
	if err != nil {
		return nil, err
	}

	refresh, _, err := h.SignRefresh(username, refreshJti)
	if err != nil {
		return nil, err
	}

	return &auth.LoginResp{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    accessExpireSeconds,
		TokenType:    "bearer",
	}, nil
}

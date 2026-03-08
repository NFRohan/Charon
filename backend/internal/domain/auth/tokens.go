package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenManager struct {
	accessSecret  []byte
	refreshPepper []byte
	issuer        string
	accessTTL     time.Duration
	refreshTTL    time.Duration
	now           func() time.Time
}

type TokenConfig struct {
	AccessSecret  string
	RefreshPepper string
	Issuer        string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

type AccessClaims struct {
	Role      Role   `json:"role"`
	SessionID string `json:"sid"`
	Type      string `json:"typ"`
	jwt.RegisteredClaims
}

func NewTokenManager(cfg TokenConfig) (*TokenManager, error) {
	if len(cfg.AccessSecret) < 32 {
		return nil, errors.New("access token secret must be at least 32 bytes")
	}
	if len(cfg.RefreshPepper) < 32 {
		return nil, errors.New("refresh token pepper must be at least 32 bytes")
	}
	if cfg.Issuer == "" {
		return nil, errors.New("jwt issuer is required")
	}
	if cfg.AccessTTL <= 0 {
		return nil, errors.New("access token ttl must be positive")
	}
	if cfg.RefreshTTL <= 0 {
		return nil, errors.New("refresh token ttl must be positive")
	}

	return &TokenManager{
		accessSecret:  []byte(cfg.AccessSecret),
		refreshPepper: []byte(cfg.RefreshPepper),
		issuer:        cfg.Issuer,
		accessTTL:     cfg.AccessTTL,
		refreshTTL:    cfg.RefreshTTL,
		now:           time.Now,
	}, nil
}

func (m *TokenManager) IssueAccessToken(user User, session Session) (string, time.Time, error) {
	now := m.now().UTC()
	expiresAt := now.Add(m.accessTTL)

	claims := AccessClaims{
		Role:      user.Role,
		SessionID: session.ID,
		Type:      "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   user.ID,
			Audience:  []string{"charon-clients"},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.accessSecret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign access token: %w", err)
	}

	return signed, expiresAt, nil
}

func (m *TokenManager) ParseAccessToken(token string) (AccessClaims, error) {
	parsedToken, err := jwt.ParseWithClaims(
		token,
		&AccessClaims{},
		func(parsedToken *jwt.Token) (any, error) {
			if parsedToken.Method == nil || parsedToken.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, ErrAccessTokenInvalid
			}
			return m.accessSecret, nil
		},
		jwt.WithIssuer(m.issuer),
		jwt.WithAudience("charon-clients"),
		jwt.WithTimeFunc(m.now),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return AccessClaims{}, ErrAccessTokenExpired
		}

		return AccessClaims{}, ErrAccessTokenInvalid
	}

	claims, ok := parsedToken.Claims.(*AccessClaims)
	if !ok || !parsedToken.Valid || claims.Type != "access" || !claims.Role.IsValid() || claims.Subject == "" || claims.SessionID == "" {
		return AccessClaims{}, ErrAccessTokenInvalid
	}

	return *claims, nil
}

func (m *TokenManager) NewRefreshToken() (plain string, hash string, expiresAt time.Time, err error) {
	raw := make([]byte, 32)
	if _, err = rand.Read(raw); err != nil {
		return "", "", time.Time{}, fmt.Errorf("generate refresh token: %w", err)
	}

	plain = base64.RawURLEncoding.EncodeToString(raw)
	hash = m.HashRefreshToken(plain)
	expiresAt = m.now().UTC().Add(m.refreshTTL)
	return plain, hash, expiresAt, nil
}

func (m *TokenManager) HashRefreshToken(token string) string {
	mac := hmac.New(sha256.New, m.refreshPepper)
	_, _ = mac.Write([]byte(token))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

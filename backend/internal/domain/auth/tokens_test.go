package auth

import (
	"testing"
	"time"
)

func TestTokenManagerIssueAndParse(t *testing.T) {
	t.Parallel()

	manager, err := NewTokenManager(TokenConfig{
		AccessSecret:  "01234567890123456789012345678901",
		RefreshPepper: "abcdefghijklmnopqrstuvwxyz123456",
		Issuer:        "charon",
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    24 * time.Hour,
	})
	if err != nil {
		t.Fatalf("new token manager: %v", err)
	}

	token, expiresAt, err := manager.IssueAccessToken(
		User{ID: "usr_1", Role: RoleStudent},
		Session{ID: "sess_1"},
	)
	if err != nil {
		t.Fatalf("issue access token: %v", err)
	}

	if expiresAt.IsZero() {
		t.Fatal("expected non-zero access token expiry")
	}

	claims, err := manager.ParseAccessToken(token)
	if err != nil {
		t.Fatalf("parse access token: %v", err)
	}

	if claims.Subject != "usr_1" {
		t.Fatalf("expected subject usr_1, got %s", claims.Subject)
	}
	if claims.SessionID != "sess_1" {
		t.Fatalf("expected session id sess_1, got %s", claims.SessionID)
	}
	if claims.Role != RoleStudent {
		t.Fatalf("expected role %s, got %s", RoleStudent, claims.Role)
	}
}

func TestTokenManagerRejectsExpiredToken(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 8, 8, 0, 0, 0, time.UTC)
	manager, err := NewTokenManager(TokenConfig{
		AccessSecret:  "01234567890123456789012345678901",
		RefreshPepper: "abcdefghijklmnopqrstuvwxyz123456",
		Issuer:        "charon",
		AccessTTL:     time.Minute,
		RefreshTTL:    24 * time.Hour,
	})
	if err != nil {
		t.Fatalf("new token manager: %v", err)
	}

	manager.now = func() time.Time { return now }
	token, _, err := manager.IssueAccessToken(
		User{ID: "usr_1", Role: RoleStudent},
		Session{ID: "sess_1"},
	)
	if err != nil {
		t.Fatalf("issue access token: %v", err)
	}

	manager.now = func() time.Time { return now.Add(2 * time.Minute) }
	_, err = manager.ParseAccessToken(token)
	if err != ErrAccessTokenExpired {
		t.Fatalf("expected ErrAccessTokenExpired, got %v", err)
	}
}

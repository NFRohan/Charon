package auth

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
)

type stubRepository struct {
	findUserResult          User
	findUserErr             error
	createSessionResult     Session
	createSessionErr        error
	sessionByRefreshResult  SessionWithUser
	sessionByRefreshErr     error
	sessionByIDResult       SessionWithUser
	sessionByIDErr          error
	revokeSessionErr        error
	createdSessionInput     Session
	touchedSessionID        string
	touchedRefreshedAt      time.Time
	revokedRefreshTokenHash string
	revokedAt               time.Time
}

func (s *stubRepository) FindUserByLoginID(_ context.Context, _ string) (User, error) {
	return s.findUserResult, s.findUserErr
}

func (s *stubRepository) CreateSession(_ context.Context, session Session) (Session, error) {
	s.createdSessionInput = session
	return s.createSessionResult, s.createSessionErr
}

func (s *stubRepository) GetSessionByRefreshTokenHash(_ context.Context, _ string) (SessionWithUser, error) {
	return s.sessionByRefreshResult, s.sessionByRefreshErr
}

func (s *stubRepository) GetSessionByID(_ context.Context, _ string) (SessionWithUser, error) {
	return s.sessionByIDResult, s.sessionByIDErr
}

func (s *stubRepository) TouchSession(_ context.Context, sessionID string, refreshedAt time.Time) error {
	s.touchedSessionID = sessionID
	s.touchedRefreshedAt = refreshedAt
	return nil
}

func (s *stubRepository) RevokeSessionByRefreshTokenHash(_ context.Context, refreshTokenHash string, revokedAt time.Time) error {
	s.revokedRefreshTokenHash = refreshTokenHash
	s.revokedAt = revokedAt
	return s.revokeSessionErr
}

func TestServiceLoginSuccess(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 8, 8, 0, 0, 0, time.UTC)
	hasher := NewArgon2idHasher(DefaultArgon2idParams())
	passwordHash, err := hasher.Hash("ChangeMe123!")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	repo := &stubRepository{
		findUserResult: User{
			ID:           "usr_123",
			Role:         RoleStudent,
			Name:         "Student Demo",
			Status:       UserStatusActive,
			PasswordHash: passwordHash,
		},
		createSessionResult: Session{
			ID:              "sess_123",
			UserID:          "usr_123",
			CreatedAt:       now,
			LastRefreshedAt: now,
			ExpiresAt:       now.Add(24 * time.Hour),
		},
	}

	service, tokenManager := newTestService(t, repo, now)
	tokenManager.now = func() time.Time { return now }
	service.now = func() time.Time { return now }

	result, err := service.Login(context.Background(), LoginInput{
		LoginID:  "220041234",
		Password: "ChangeMe123!",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	if result.User.ID != "usr_123" {
		t.Fatalf("expected user usr_123, got %s", result.User.ID)
	}
	if result.RefreshToken == "" {
		t.Fatal("expected refresh token to be issued")
	}
	if repo.createdSessionInput.UserID != "usr_123" {
		t.Fatalf("expected session to be created for usr_123, got %s", repo.createdSessionInput.UserID)
	}
	if repo.createdSessionInput.RefreshTokenHash == "" {
		t.Fatal("expected refresh token hash to be stored")
	}
	if repo.createdSessionInput.RefreshTokenHash == result.RefreshToken {
		t.Fatal("refresh token hash should not equal the raw refresh token")
	}
}

func TestServiceRefreshSuccess(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 8, 8, 0, 0, 0, time.UTC)
	repo := &stubRepository{}
	service, tokenManager := newTestService(t, repo, now)
	tokenManager.now = func() time.Time { return now }
	service.now = func() time.Time { return now }

	refreshToken, refreshTokenHash, refreshExpiresAt, err := tokenManager.NewRefreshToken()
	if err != nil {
		t.Fatalf("new refresh token: %v", err)
	}

	repo.sessionByRefreshResult = SessionWithUser{
		Session: Session{
			ID:               "sess_123",
			UserID:           "usr_123",
			RefreshTokenHash: refreshTokenHash,
			ExpiresAt:        refreshExpiresAt,
		},
		User: User{
			ID:     "usr_123",
			Role:   RoleStudent,
			Name:   "Student Demo",
			Status: UserStatusActive,
		},
	}

	result, err := service.Refresh(context.Background(), refreshToken)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}

	if result.RefreshToken != refreshToken {
		t.Fatalf("expected stable refresh token, got %q", result.RefreshToken)
	}
	if repo.touchedSessionID != "sess_123" {
		t.Fatalf("expected session sess_123 to be touched, got %s", repo.touchedSessionID)
	}
	if repo.touchedRefreshedAt.IsZero() {
		t.Fatal("expected refresh timestamp to be recorded")
	}
}

func TestServiceLogoutInvalidRefreshToken(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 8, 8, 0, 0, 0, time.UTC)
	repo := &stubRepository{revokeSessionErr: sql.ErrNoRows}
	service, tokenManager := newTestService(t, repo, now)
	tokenManager.now = func() time.Time { return now }
	service.now = func() time.Time { return now }

	err := service.Logout(context.Background(), "missing-token")
	if !errors.Is(err, ErrRefreshTokenInvalid) {
		t.Fatalf("expected ErrRefreshTokenInvalid, got %v", err)
	}
}

func TestServiceAuthenticateAccessTokenRejectsSuspendedAccount(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 8, 8, 0, 0, 0, time.UTC)
	repo := &stubRepository{}
	service, tokenManager := newTestService(t, repo, now)
	tokenManager.now = func() time.Time { return now }
	service.now = func() time.Time { return now }

	accessToken, _, err := tokenManager.IssueAccessToken(
		User{ID: "usr_123", Role: RoleStudent},
		Session{ID: "sess_123"},
	)
	if err != nil {
		t.Fatalf("issue access token: %v", err)
	}

	repo.sessionByIDResult = SessionWithUser{
		Session: Session{
			ID:        "sess_123",
			UserID:    "usr_123",
			ExpiresAt: now.Add(24 * time.Hour),
		},
		User: User{
			ID:     "usr_123",
			Role:   RoleStudent,
			Status: UserStatusSuspended,
		},
	}

	_, err = service.AuthenticateAccessToken(context.Background(), accessToken)
	if !errors.Is(err, ErrAccountDisabled) {
		t.Fatalf("expected ErrAccountDisabled, got %v", err)
	}
}

func newTestService(t *testing.T, repo Repository, now time.Time) (*Service, *TokenManager) {
	t.Helper()

	hasher := NewArgon2idHasher(DefaultArgon2idParams())
	tokenManager, err := NewTokenManager(TokenConfig{
		AccessSecret:  "01234567890123456789012345678901",
		RefreshPepper: "abcdefghijklmnopqrstuvwxyz123456",
		Issuer:        "charon",
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    24 * time.Hour,
	})
	if err != nil {
		t.Fatalf("new token manager: %v", err)
	}

	service, err := NewService(repo, hasher, tokenManager)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	tokenManager.now = func() time.Time { return now }
	service.now = func() time.Time { return now }
	return service, tokenManager
}

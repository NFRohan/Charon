package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Service struct {
	repo      Repository
	hasher    PasswordHasher
	tokens    *TokenManager
	dummyHash string
	now       func() time.Time
}

type LoginInput struct {
	LoginID  string
	Password string
}

type LoginResult struct {
	AccessToken           string
	AccessTokenExpiresAt  time.Time
	RefreshToken          string
	RefreshTokenExpiresAt time.Time
	User                  User
}

type RefreshResult struct {
	AccessToken           string
	AccessTokenExpiresAt  time.Time
	RefreshToken          string
	RefreshTokenExpiresAt time.Time
}

func NewService(repo Repository, hasher PasswordHasher, tokens *TokenManager) (*Service, error) {
	dummyHash, err := hasher.Hash("charon-auth-dummy-password")
	if err != nil {
		return nil, fmt.Errorf("hash dummy password: %w", err)
	}

	return &Service{
		repo:      repo,
		hasher:    hasher,
		tokens:    tokens,
		dummyHash: dummyHash,
		now:       time.Now,
	}, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (LoginResult, error) {
	user, err := s.repo.FindUserByLoginID(ctx, input.LoginID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_ = s.hasher.Verify(s.dummyHash, input.Password)
			return LoginResult{}, ErrInvalidCredentials
		}
		return LoginResult{}, err
	}

	if err := s.hasher.Verify(user.PasswordHash, input.Password); err != nil {
		return LoginResult{}, ErrInvalidCredentials
	}

	if user.Status == UserStatusSuspended {
		return LoginResult{}, ErrAccountDisabled
	}

	refreshToken, refreshTokenHash, refreshExpiresAt, err := s.tokens.NewRefreshToken()
	if err != nil {
		return LoginResult{}, err
	}

	session, err := s.repo.CreateSession(ctx, Session{
		UserID:           user.ID,
		RefreshTokenHash: refreshTokenHash,
		ExpiresAt:        refreshExpiresAt,
	})
	if err != nil {
		return LoginResult{}, err
	}

	accessToken, accessExpiresAt, err := s.tokens.IssueAccessToken(user, session)
	if err != nil {
		return LoginResult{}, err
	}

	return LoginResult{
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessExpiresAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshExpiresAt,
		User:                  user,
	}, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (RefreshResult, error) {
	refreshTokenHash := s.tokens.HashRefreshToken(refreshToken)
	sessionRecord, err := s.repo.GetSessionByRefreshTokenHash(ctx, refreshTokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return RefreshResult{}, ErrRefreshTokenInvalid
		}
		return RefreshResult{}, err
	}

	if sessionRecord.Session.RevokedAt != nil {
		return RefreshResult{}, ErrRefreshTokenInvalid
	}
	if s.now().UTC().After(sessionRecord.Session.ExpiresAt) {
		return RefreshResult{}, ErrRefreshTokenExpired
	}
	if sessionRecord.User.Status == UserStatusSuspended {
		return RefreshResult{}, ErrAccountDisabled
	}

	refreshedAt := s.now().UTC()
	if err := s.repo.TouchSession(ctx, sessionRecord.Session.ID, refreshedAt); err != nil {
		return RefreshResult{}, err
	}
	sessionRecord.Session.LastRefreshedAt = refreshedAt

	accessToken, accessExpiresAt, err := s.tokens.IssueAccessToken(sessionRecord.User, sessionRecord.Session)
	if err != nil {
		return RefreshResult{}, err
	}

	return RefreshResult{
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessExpiresAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: sessionRecord.Session.ExpiresAt,
	}, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	refreshTokenHash := s.tokens.HashRefreshToken(refreshToken)
	if err := s.repo.RevokeSessionByRefreshTokenHash(ctx, refreshTokenHash, s.now().UTC()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrRefreshTokenInvalid
		}
		return err
	}

	return nil
}

func (s *Service) AuthenticateAccessToken(ctx context.Context, accessToken string) (AuthenticatedIdentity, error) {
	claims, err := s.tokens.ParseAccessToken(accessToken)
	if err != nil {
		return AuthenticatedIdentity{}, err
	}

	sessionRecord, err := s.repo.GetSessionByID(ctx, claims.SessionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AuthenticatedIdentity{}, ErrAccessTokenInvalid
		}
		return AuthenticatedIdentity{}, err
	}

	if sessionRecord.Session.RevokedAt != nil {
		return AuthenticatedIdentity{}, ErrAccessTokenInvalid
	}
	if s.now().UTC().After(sessionRecord.Session.ExpiresAt) {
		return AuthenticatedIdentity{}, ErrAccessTokenExpired
	}
	if sessionRecord.User.ID != claims.Subject || sessionRecord.User.Role != claims.Role {
		return AuthenticatedIdentity{}, ErrAccessTokenInvalid
	}
	if sessionRecord.User.Status == UserStatusSuspended {
		return AuthenticatedIdentity{}, ErrAccountDisabled
	}

	return AuthenticatedIdentity{
		User:      sessionRecord.User,
		Session:   sessionRecord.Session,
		TokenType: claims.Type,
	}, nil
}

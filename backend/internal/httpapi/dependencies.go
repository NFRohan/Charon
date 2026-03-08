package httpapi

import (
	"context"
	"errors"

	"charon/backend/internal/domain/auth"
)

type AuthService interface {
	Login(ctx context.Context, input auth.LoginInput) (auth.LoginResult, error)
	Refresh(ctx context.Context, refreshToken string) (auth.RefreshResult, error)
	Logout(ctx context.Context, refreshToken string) error
	AuthenticateAccessToken(ctx context.Context, accessToken string) (auth.AuthenticatedIdentity, error)
}

type Dependencies struct {
	Auth AuthService
}

func (d Dependencies) Validate() error {
	if d.Auth == nil {
		return errors.New("http api auth dependency is required")
	}

	return nil
}

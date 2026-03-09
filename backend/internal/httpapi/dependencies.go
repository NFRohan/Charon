package httpapi

import (
	"context"
	"errors"

	"charon/backend/internal/domain/auth"
	"charon/backend/internal/domain/wallet"
)

type AuthService interface {
	Login(ctx context.Context, input auth.LoginInput) (auth.LoginResult, error)
	Refresh(ctx context.Context, refreshToken string) (auth.RefreshResult, error)
	Logout(ctx context.Context, refreshToken string) error
	AuthenticateAccessToken(ctx context.Context, accessToken string) (auth.AuthenticatedIdentity, error)
}

type WalletService interface {
	GetBalanceSummary(ctx context.Context, userID string) (wallet.BalanceSummary, error)
	ListTransactions(ctx context.Context, userID string, params wallet.ListTransactionsParams) (wallet.TransactionHistoryPage, error)
}

type Dependencies struct {
	Auth   AuthService
	Wallet WalletService
}

func (d Dependencies) Validate() error {
	if d.Auth == nil {
		return errors.New("http api auth dependency is required")
	}
	if d.Wallet == nil {
		return errors.New("http api wallet dependency is required")
	}

	return nil
}

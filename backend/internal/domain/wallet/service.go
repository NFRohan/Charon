package wallet

import (
	"context"
	"database/sql"
	"errors"
)

type Repository interface {
	GetBalanceSummary(ctx context.Context, userID string) (BalanceSummary, error)
	ListTransactions(ctx context.Context, userID string, params ListTransactionsParams) (TransactionHistoryPage, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) (*Service, error) {
	if repo == nil {
		return nil, errors.New("wallet repository is required")
	}

	return &Service{repo: repo}, nil
}

func (s *Service) GetBalanceSummary(ctx context.Context, userID string) (BalanceSummary, error) {
	summary, err := s.repo.GetBalanceSummary(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return BalanceSummary{}, ErrWalletAccountNotFound
		}

		return BalanceSummary{}, err
	}

	return summary, nil
}

func (s *Service) ListTransactions(ctx context.Context, userID string, params ListTransactionsParams) (TransactionHistoryPage, error) {
	params = normalizeListTransactionsParams(params)

	page, err := s.repo.ListTransactions(ctx, userID, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return TransactionHistoryPage{}, ErrWalletAccountNotFound
		}

		return TransactionHistoryPage{}, err
	}

	page.Limit = params.Limit
	page.Offset = params.Offset
	return page, nil
}

func normalizeListTransactionsParams(params ListTransactionsParams) ListTransactionsParams {
	if params.Limit <= 0 {
		params.Limit = DefaultPageLimit
	}
	if params.Limit > MaxPageLimit {
		params.Limit = MaxPageLimit
	}

	return params
}

package wallet

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

type stubRepository struct {
	balanceSummary  BalanceSummary
	balanceErr      error
	transactionPage TransactionHistoryPage
	transactionErr  error
	listParams      ListTransactionsParams
}

func (s *stubRepository) GetBalanceSummary(_ context.Context, _ string) (BalanceSummary, error) {
	return s.balanceSummary, s.balanceErr
}

func (s *stubRepository) ListTransactions(_ context.Context, _ string, params ListTransactionsParams) (TransactionHistoryPage, error) {
	s.listParams = params
	return s.transactionPage, s.transactionErr
}

func TestServiceGetBalanceSummaryMapsNotFound(t *testing.T) {
	t.Parallel()

	service, err := NewService(&stubRepository{balanceErr: sql.ErrNoRows})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, err = service.GetBalanceSummary(context.Background(), "usr_123")
	if !errors.Is(err, ErrWalletAccountNotFound) {
		t.Fatalf("expected ErrWalletAccountNotFound, got %v", err)
	}
}

func TestServiceListTransactionsDefaultsLimit(t *testing.T) {
	t.Parallel()

	repo := &stubRepository{}
	service, err := NewService(repo)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, err = service.ListTransactions(context.Background(), "usr_123", ListTransactionsParams{
		Offset: 5,
	})
	if err != nil {
		t.Fatalf("list transactions: %v", err)
	}

	if repo.listParams.Limit != DefaultPageLimit {
		t.Fatalf("expected default limit %d, got %d", DefaultPageLimit, repo.listParams.Limit)
	}
	if repo.listParams.Offset != 5 {
		t.Fatalf("expected offset 5, got %d", repo.listParams.Offset)
	}
}

func TestServiceListTransactionsClampsLimit(t *testing.T) {
	t.Parallel()

	repo := &stubRepository{}
	service, err := NewService(repo)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, err = service.ListTransactions(context.Background(), "usr_123", ListTransactionsParams{
		Limit:  1000,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("list transactions: %v", err)
	}

	if repo.listParams.Limit != MaxPageLimit {
		t.Fatalf("expected clamped limit %d, got %d", MaxPageLimit, repo.listParams.Limit)
	}
}

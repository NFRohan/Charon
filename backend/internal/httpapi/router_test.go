package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"charon/backend/internal/config"
	"charon/backend/internal/domain/auth"
	"charon/backend/internal/domain/wallet"
)

type stubAuthService struct {
	loginResult        auth.LoginResult
	loginErr           error
	refreshResult      auth.RefreshResult
	refreshErr         error
	logoutErr          error
	authenticatedByJWT map[string]auth.AuthenticatedIdentity
}

func (s stubAuthService) Login(_ context.Context, _ auth.LoginInput) (auth.LoginResult, error) {
	return s.loginResult, s.loginErr
}

func (s stubAuthService) Refresh(_ context.Context, _ string) (auth.RefreshResult, error) {
	return s.refreshResult, s.refreshErr
}

func (s stubAuthService) Logout(_ context.Context, _ string) error {
	return s.logoutErr
}

func (s stubAuthService) AuthenticateAccessToken(_ context.Context, accessToken string) (auth.AuthenticatedIdentity, error) {
	identity, ok := s.authenticatedByJWT[accessToken]
	if !ok {
		return auth.AuthenticatedIdentity{}, auth.ErrAccessTokenInvalid
	}

	return identity, nil
}

type stubWalletService struct {
	balanceSummary  wallet.BalanceSummary
	balanceErr      error
	transactions    wallet.TransactionHistoryPage
	transactionsErr error
}

func (s stubWalletService) GetBalanceSummary(_ context.Context, _ string) (wallet.BalanceSummary, error) {
	return s.balanceSummary, s.balanceErr
}

func (s stubWalletService) ListTransactions(_ context.Context, _ string, _ wallet.ListTransactionsParams) (wallet.TransactionHistoryPage, error) {
	return s.transactions, s.transactionsErr
}

func TestLoginRouteReturnsAuthEnvelope(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 8, 8, 0, 0, 0, time.UTC)
	router := newTestRouter(t, stubAuthService{
		loginResult: auth.LoginResult{
			AccessToken:           "access-token",
			AccessTokenExpiresAt:  now.Add(15 * time.Minute),
			RefreshToken:          "refresh-token",
			RefreshTokenExpiresAt: now.Add(24 * time.Hour),
			User: auth.User{
				ID:         "usr_123",
				Role:       auth.RoleStudent,
				Name:       "Student Demo",
				Status:     auth.UserStatusActive,
				FareExempt: false,
			},
		},
	}, stubWalletService{})

	requestBody, err := json.Marshal(map[string]string{
		"login_id": "220041234",
		"password": "ChangeMe123!",
	})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if response["role"] != string(auth.RoleStudent) {
		t.Fatalf("expected student role, got %#v", response["role"])
	}
	if response["user_id"] != "usr_123" {
		t.Fatalf("expected user_id usr_123, got %#v", response["user_id"])
	}
}

func TestProtectedStudentRouteRequiresAuth(t *testing.T) {
	t.Parallel()

	router := newTestRouter(t, stubAuthService{}, stubWalletService{})

	request := httptest.NewRequest(http.MethodGet, "/wallet/balance", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
	assertErrorCode(t, recorder.Body.Bytes(), "AUTHORIZATION_REQUIRED")
}

func TestProtectedStudentRouteAllowsStudent(t *testing.T) {
	t.Parallel()

	router := newTestRouter(t, stubAuthService{
		authenticatedByJWT: map[string]auth.AuthenticatedIdentity{
			"student-token": {
				User: auth.User{ID: "usr_123", Role: auth.RoleStudent, Status: auth.UserStatusActive},
			},
		},
	}, stubWalletService{
		balanceSummary: wallet.BalanceSummary{
			UserID:                         "usr_123",
			AccountStatus:                  wallet.AccountStatusActive,
			BalanceMinor:                   1800,
			OverdraftLimitMinor:            200,
			FareExempt:                     false,
			AvailableEmergencyVoucherCount: 1,
		},
	})

	request := httptest.NewRequest(http.MethodGet, "/wallet/balance", nil)
	request.Header.Set("Authorization", "Bearer student-token")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if response["balance_minor"] != float64(1800) {
		t.Fatalf("expected balance_minor 1800, got %#v", response["balance_minor"])
	}
}

func TestWalletTransactionsRejectsInvalidPagination(t *testing.T) {
	t.Parallel()

	router := newTestRouter(t, stubAuthService{
		authenticatedByJWT: map[string]auth.AuthenticatedIdentity{
			"student-token": {
				User: auth.User{ID: "usr_123", Role: auth.RoleStudent, Status: auth.UserStatusActive},
			},
		},
	}, stubWalletService{})

	request := httptest.NewRequest(http.MethodGet, "/wallet/transactions?limit=abc&offset=-1", nil)
	request.Header.Set("Authorization", "Bearer student-token")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
	assertErrorCode(t, recorder.Body.Bytes(), "VALIDATION_ERROR")
}

func TestWalletTransactionsReturnsPage(t *testing.T) {
	t.Parallel()

	routeCode := "A"
	busCode := "1042"
	now := time.Date(2026, 3, 10, 8, 0, 0, 0, time.UTC)
	router := newTestRouter(t, stubAuthService{
		authenticatedByJWT: map[string]auth.AuthenticatedIdentity{
			"student-token": {
				User: auth.User{ID: "usr_123", Role: auth.RoleStudent, Status: auth.UserStatusActive},
			},
		},
	}, stubWalletService{
		transactions: wallet.TransactionHistoryPage{
			Items: []wallet.TransactionHistoryItem{
				{
					TransactionID:         "tx_123",
					Type:                  "BOARDING_FARE",
					AmountMinor:           2000,
					RouteCode:             &routeCode,
					BusCode:               &busCode,
					Status:                "SUCCESS",
					ResultingBalanceMinor: 1800,
					CreatedAt:             now,
				},
			},
			Limit:  20,
			Offset: 0,
			Total:  1,
		},
	})

	request := httptest.NewRequest(http.MethodGet, "/wallet/transactions", nil)
	request.Header.Set("Authorization", "Bearer student-token")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var response struct {
		Items []map[string]any `json:"items"`
		Total int              `json:"total"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if len(response.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(response.Items))
	}
	if response.Total != 1 {
		t.Fatalf("expected total 1, got %d", response.Total)
	}
}

func TestAdminRouteRejectsStudentRole(t *testing.T) {
	t.Parallel()

	router := newTestRouter(t, stubAuthService{
		authenticatedByJWT: map[string]auth.AuthenticatedIdentity{
			"student-token": {
				User: auth.User{ID: "usr_123", Role: auth.RoleStudent, Status: auth.UserStatusActive},
			},
		},
	}, stubWalletService{})

	request := httptest.NewRequest(http.MethodGet, "/admin/students", nil)
	request.Header.Set("Authorization", "Bearer student-token")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", recorder.Code)
	}
	assertErrorCode(t, recorder.Body.Bytes(), "INSUFFICIENT_ROLE")
}

func newTestRouter(t *testing.T, authService AuthService, walletService WalletService) http.Handler {
	t.Helper()

	router, err := NewRouter(config.Config{AppEnv: config.AppEnvTest}, Dependencies{
		Auth:   authService,
		Wallet: walletService,
	})
	if err != nil {
		t.Fatalf("new router: %v", err)
	}

	return router
}

func assertErrorCode(t *testing.T, payload []byte, expected string) {
	t.Helper()

	var response map[string]any
	if err := json.Unmarshal(payload, &response); err != nil {
		t.Fatalf("unmarshal error response: %v", err)
	}

	if response["error_code"] != expected {
		t.Fatalf("expected error_code %s, got %#v", expected, response["error_code"])
	}

	traceID, ok := response["trace_id"].(string)
	if !ok || traceID == "" {
		t.Fatalf("expected trace_id, got %#v", response["trace_id"])
	}
}

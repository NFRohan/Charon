package httpapi

import (
	"errors"
	"net/http"
	"strconv"

	"charon/backend/internal/domain/auth"
	"charon/backend/internal/domain/wallet"

	"github.com/gin-gonic/gin"
)

type walletHandler struct {
	walletService WalletService
}

func registerWalletRoutes(router *gin.RouterGroup, deps Dependencies) {
	handler := walletHandler{walletService: deps.Wallet}

	router.GET("/balance", handler.balance)
	router.GET("/transactions", handler.transactions)
}

func (h walletHandler) balance(c *gin.Context) {
	identity, ok := currentIdentity(c)
	if !ok {
		respondAuthError(c, auth.ErrAuthorizationRequired)
		return
	}

	summary, err := h.walletService.GetBalanceSummary(c.Request.Context(), identity.User.ID)
	if err != nil {
		respondWalletError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":                           summary.UserID,
		"account_status":                    summary.AccountStatus,
		"balance_minor":                     summary.BalanceMinor,
		"overdraft_limit_minor":             summary.OverdraftLimitMinor,
		"fare_exempt":                       summary.FareExempt,
		"available_emergency_voucher_count": summary.AvailableEmergencyVoucherCount,
	})
}

func (h walletHandler) transactions(c *gin.Context) {
	identity, ok := currentIdentity(c)
	if !ok {
		respondAuthError(c, auth.ErrAuthorizationRequired)
		return
	}

	params, ok := parseTransactionListParams(c)
	if !ok {
		return
	}

	page, err := h.walletService.ListTransactions(c.Request.Context(), identity.User.ID, params)
	if err != nil {
		respondWalletError(c, err)
		return
	}

	items := make([]gin.H, 0, len(page.Items))
	for _, item := range page.Items {
		items = append(items, gin.H{
			"transaction_id":          item.TransactionID,
			"type":                    item.Type,
			"amount_minor":            item.AmountMinor,
			"route_code":              item.RouteCode,
			"bus_code":                item.BusCode,
			"status":                  item.Status,
			"resulting_balance_minor": item.ResultingBalanceMinor,
			"created_at":              item.CreatedAt.UTC(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"items":  items,
		"limit":  page.Limit,
		"offset": page.Offset,
		"total":  page.Total,
	})
}

func parseTransactionListParams(c *gin.Context) (wallet.ListTransactionsParams, bool) {
	params := wallet.ListTransactionsParams{}
	var fieldErrors []fieldError

	if value := c.Query("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 {
			fieldErrors = append(fieldErrors, fieldError{
				Field:   "limit",
				Message: "Limit must be a positive integer.",
			})
		} else {
			params.Limit = parsed
		}
	}

	if value := c.Query("offset"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 0 {
			fieldErrors = append(fieldErrors, fieldError{
				Field:   "offset",
				Message: "Offset must be zero or greater.",
			})
		} else {
			params.Offset = parsed
		}
	}

	if len(fieldErrors) > 0 {
		respondValidationError(c, "One or more fields are invalid.", fieldErrors...)
		return wallet.ListTransactionsParams{}, false
	}

	return params, true
}

func respondWalletError(c *gin.Context, err error) {
	status := http.StatusInternalServerError

	switch {
	case errors.Is(err, wallet.ErrWalletAccountNotFound):
		status = http.StatusNotFound
	}

	respondError(c, status, wallet.ErrorCode(err), wallet.ErrorMessage(err))
}

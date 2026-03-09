package wallet

import "time"

const (
	DefaultPageLimit = 20
	MaxPageLimit     = 100
)

type AccountStatus string

const (
	AccountStatusActive         AccountStatus = "ACTIVE"
	AccountStatusSuspended      AccountStatus = "SUSPENDED"
	AccountStatusRestrictedDebt AccountStatus = "RESTRICTED_DEBT"
)

type BalanceSummary struct {
	UserID                         string
	AccountStatus                  AccountStatus
	BalanceMinor                   int64
	OverdraftLimitMinor            int64
	FareExempt                     bool
	AvailableEmergencyVoucherCount int
}

type ListTransactionsParams struct {
	Limit  int
	Offset int
}

type TransactionHistoryItem struct {
	TransactionID         string
	Type                  string
	AmountMinor           int64
	RouteCode             *string
	BusCode               *string
	Status                string
	ResultingBalanceMinor int64
	CreatedAt             time.Time
}

type TransactionHistoryPage struct {
	Items  []TransactionHistoryItem
	Limit  int
	Offset int
	Total  int
}

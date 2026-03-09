package wallet

import (
	"context"
	"database/sql"
	"fmt"
)

type PostgresRepository struct {
	db *sql.DB
}

type walletAccountRecord struct {
	ID                    string
	AvailableBalanceMinor int64
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) GetBalanceSummary(ctx context.Context, userID string) (BalanceSummary, error) {
	const query = `
		SELECT
			u.id,
			u.status,
			u.fare_exempt,
			wa.available_balance_minor,
			wa.overdraft_limit_minor,
			COALESCE(COUNT(erp.id), 0) AS available_emergency_voucher_count
		FROM users u
		JOIN wallet_accounts wa ON wa.user_id = u.id
		LEFT JOIN emergency_ride_permits erp
			ON erp.student_id = u.id
		   AND erp.status = 'ISSUED'
		   AND erp.expires_at > NOW()
		WHERE u.id = $1
		GROUP BY
			u.id,
			u.status,
			u.fare_exempt,
			wa.available_balance_minor,
			wa.overdraft_limit_minor
	`

	var summary BalanceSummary
	var voucherCount int64
	if err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&summary.UserID,
		&summary.AccountStatus,
		&summary.FareExempt,
		&summary.BalanceMinor,
		&summary.OverdraftLimitMinor,
		&voucherCount,
	); err != nil {
		return BalanceSummary{}, err
	}

	summary.AvailableEmergencyVoucherCount = int(voucherCount)
	return summary, nil
}

func (r *PostgresRepository) ListTransactions(ctx context.Context, userID string, params ListTransactionsParams) (TransactionHistoryPage, error) {
	account, err := r.getWalletAccountRecord(ctx, userID)
	if err != nil {
		return TransactionHistoryPage{}, err
	}

	total, err := r.countTransactions(ctx, account.ID)
	if err != nil {
		return TransactionHistoryPage{}, err
	}

	items, err := r.listTransactionItems(ctx, account, params)
	if err != nil {
		return TransactionHistoryPage{}, err
	}

	return TransactionHistoryPage{
		Items: items,
		Total: total,
	}, nil
}

func (r *PostgresRepository) getWalletAccountRecord(ctx context.Context, userID string) (walletAccountRecord, error) {
	const query = `
		SELECT id, available_balance_minor
		FROM wallet_accounts
		WHERE user_id = $1
	`

	var account walletAccountRecord
	if err := r.db.QueryRowContext(ctx, query, userID).Scan(&account.ID, &account.AvailableBalanceMinor); err != nil {
		return walletAccountRecord{}, err
	}

	return account, nil
}

func (r *PostgresRepository) countTransactions(ctx context.Context, accountID string) (int, error) {
	const query = `
		SELECT COUNT(DISTINCT transaction_id)
		FROM ledger_entries
		WHERE account_id = $1
	`

	var total int64
	if err := r.db.QueryRowContext(ctx, query, accountID).Scan(&total); err != nil {
		return 0, fmt.Errorf("count wallet transactions: %w", err)
	}

	return int(total), nil
}

func (r *PostgresRepository) listTransactionItems(ctx context.Context, account walletAccountRecord, params ListTransactionsParams) ([]TransactionHistoryItem, error) {
	const query = `
		WITH transaction_effects AS (
			SELECT
				t.id AS transaction_id,
				t.type,
				t.amount_minor,
				t.status,
				t.created_at,
				r.code AS route_code,
				b.bus_code,
				COALESCE(SUM(
					CASE le.direction
						WHEN 'DEBIT' THEN -le.amount_minor
						ELSE le.amount_minor
					END
				), 0) AS account_delta
			FROM ledger_entries le
			JOIN transactions t ON t.id = le.transaction_id
			LEFT JOIN route_sessions rs ON rs.id = t.route_session_id
			LEFT JOIN routes r ON r.id = rs.route_id
			LEFT JOIN buses b ON b.id = rs.bus_id
			WHERE le.account_id = $1
			GROUP BY
				t.id,
				t.type,
				t.amount_minor,
				t.status,
				t.created_at,
				r.code,
				b.bus_code
		)
		SELECT
			transaction_id,
			type,
			amount_minor,
			route_code,
			bus_code,
			status,
			$2 - COALESCE(
				SUM(account_delta) OVER (
					ORDER BY created_at DESC, transaction_id DESC
					ROWS BETWEEN UNBOUNDED PRECEDING AND 1 PRECEDING
				),
				0
			) AS resulting_balance_minor,
			created_at
		FROM transaction_effects
		ORDER BY created_at DESC, transaction_id DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.QueryContext(ctx, query, account.ID, account.AvailableBalanceMinor, params.Limit, params.Offset)
	if err != nil {
		return nil, fmt.Errorf("list wallet transactions: %w", err)
	}
	defer rows.Close()

	items := make([]TransactionHistoryItem, 0, params.Limit)
	for rows.Next() {
		var item TransactionHistoryItem
		var routeCode sql.NullString
		var busCode sql.NullString

		if err := rows.Scan(
			&item.TransactionID,
			&item.Type,
			&item.AmountMinor,
			&routeCode,
			&busCode,
			&item.Status,
			&item.ResultingBalanceMinor,
			&item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan wallet transaction: %w", err)
		}

		if routeCode.Valid {
			item.RouteCode = &routeCode.String
		}
		if busCode.Valid {
			item.BusCode = &busCode.String
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate wallet transactions: %w", err)
	}

	return items, nil
}

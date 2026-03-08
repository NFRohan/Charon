package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Repository interface {
	FindUserByLoginID(ctx context.Context, loginID string) (User, error)
	CreateSession(ctx context.Context, session Session) (Session, error)
	GetSessionByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (SessionWithUser, error)
	GetSessionByID(ctx context.Context, sessionID string) (SessionWithUser, error)
	TouchSession(ctx context.Context, sessionID string, refreshedAt time.Time) error
	RevokeSessionByRefreshTokenHash(ctx context.Context, refreshTokenHash string, revokedAt time.Time) error
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) FindUserByLoginID(ctx context.Context, loginID string) (User, error) {
	const query = `
		SELECT id, role, name, institutional_id, status, fare_exempt, password_hash, created_at, updated_at
		FROM users
		WHERE institutional_id = $1
	`

	var user User
	if err := r.db.QueryRowContext(ctx, query, loginID).Scan(
		&user.ID,
		&user.Role,
		&user.Name,
		&user.InstitutionalID,
		&user.Status,
		&user.FareExempt,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return User{}, err
	}

	return user, nil
}

func (r *PostgresRepository) CreateSession(ctx context.Context, session Session) (Session, error) {
	const query = `
		INSERT INTO auth_sessions (user_id, refresh_token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, last_refreshed_at
	`

	created := session
	if err := r.db.QueryRowContext(
		ctx,
		query,
		session.UserID,
		session.RefreshTokenHash,
		session.ExpiresAt,
	).Scan(&created.ID, &created.CreatedAt, &created.LastRefreshedAt); err != nil {
		return Session{}, fmt.Errorf("insert auth session: %w", err)
	}

	return created, nil
}

func (r *PostgresRepository) GetSessionByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (SessionWithUser, error) {
	return r.getSessionWithUser(ctx, `
		SELECT
			s.id,
			s.user_id,
			s.refresh_token_hash,
			s.expires_at,
			s.revoked_at,
			s.created_at,
			s.last_refreshed_at,
			u.id,
			u.role,
			u.name,
			u.institutional_id,
			u.status,
			u.fare_exempt,
			u.password_hash,
			u.created_at,
			u.updated_at
		FROM auth_sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.refresh_token_hash = $1
	`, refreshTokenHash)
}

func (r *PostgresRepository) GetSessionByID(ctx context.Context, sessionID string) (SessionWithUser, error) {
	return r.getSessionWithUser(ctx, `
		SELECT
			s.id,
			s.user_id,
			s.refresh_token_hash,
			s.expires_at,
			s.revoked_at,
			s.created_at,
			s.last_refreshed_at,
			u.id,
			u.role,
			u.name,
			u.institutional_id,
			u.status,
			u.fare_exempt,
			u.password_hash,
			u.created_at,
			u.updated_at
		FROM auth_sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.id = $1
	`, sessionID)
}

func (r *PostgresRepository) TouchSession(ctx context.Context, sessionID string, refreshedAt time.Time) error {
	const query = `
		UPDATE auth_sessions
		SET last_refreshed_at = $2
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, sessionID, refreshedAt)
	if err != nil {
		return fmt.Errorf("update auth session refresh time: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("auth session rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *PostgresRepository) RevokeSessionByRefreshTokenHash(ctx context.Context, refreshTokenHash string, revokedAt time.Time) error {
	const query = `
		UPDATE auth_sessions
		SET revoked_at = $2
		WHERE refresh_token_hash = $1
		  AND revoked_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, refreshTokenHash, revokedAt)
	if err != nil {
		return fmt.Errorf("revoke auth session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("auth session rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *PostgresRepository) getSessionWithUser(ctx context.Context, query string, arg string) (SessionWithUser, error) {
	var record SessionWithUser
	var revokedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, arg).Scan(
		&record.Session.ID,
		&record.Session.UserID,
		&record.Session.RefreshTokenHash,
		&record.Session.ExpiresAt,
		&revokedAt,
		&record.Session.CreatedAt,
		&record.Session.LastRefreshedAt,
		&record.User.ID,
		&record.User.Role,
		&record.User.Name,
		&record.User.InstitutionalID,
		&record.User.Status,
		&record.User.FareExempt,
		&record.User.PasswordHash,
		&record.User.CreatedAt,
		&record.User.UpdatedAt,
	)
	if err != nil {
		return SessionWithUser{}, err
	}

	if revokedAt.Valid {
		record.Session.RevokedAt = &revokedAt.Time
	}

	if !record.User.Role.IsValid() {
		return SessionWithUser{}, errors.New("stored auth role is invalid")
	}

	return record, nil
}

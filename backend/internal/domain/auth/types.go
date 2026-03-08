package auth

import "time"

type Role string

const (
	RoleStudent        Role = "student"
	RoleDriver         Role = "driver"
	RoleCashier        Role = "cashier"
	RoleAdmin          Role = "admin"
	RoleTechnicalAdmin Role = "technical_admin"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleStudent, RoleDriver, RoleCashier, RoleAdmin, RoleTechnicalAdmin:
		return true
	default:
		return false
	}
}

type UserStatus string

const (
	UserStatusActive         UserStatus = "ACTIVE"
	UserStatusSuspended      UserStatus = "SUSPENDED"
	UserStatusRestrictedDebt UserStatus = "RESTRICTED_DEBT"
)

type User struct {
	ID              string
	Role            Role
	Name            string
	InstitutionalID string
	Status          UserStatus
	FareExempt      bool
	PasswordHash    string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Session struct {
	ID               string
	UserID           string
	RefreshTokenHash string
	ExpiresAt        time.Time
	RevokedAt        *time.Time
	CreatedAt        time.Time
	LastRefreshedAt  time.Time
}

type SessionWithUser struct {
	Session Session
	User    User
}

type AuthenticatedIdentity struct {
	User      User
	Session   Session
	TokenType string
}

package models

import (
	"database/sql"
	"time"
)

type UserRole string

const (
	// RoleUser is kept for backward compatibility with existing accounts.
	// Sprint 2 will migrate/invite explicit SPV and employee accounts before
	// role-based route enforcement is enabled.
	RoleUser     UserRole = "user"
	RoleAdmin      UserRole = "admin"
	RoleSuperAdmin UserRole = "super_admin"
	RoleSPV        UserRole = "spv"
	RoleEmployee   UserRole = "employee"
)

type User struct {
	ID                    int64          `json:"id"`
	Email                 string         `json:"email"`
	Name                  string         `json:"name"`
	Avatar                string         `json:"avatar"`
	Password              sql.NullString `json:"-"`
	Role                  UserRole       `json:"role"`
	GoogleID              sql.NullString `json:"-"`
	EmailVerified         bool           `json:"email_verified"`
	FailedLoginAttempts   int            `json:"-"`
	LockedUntil           sql.NullTime   `json:"-"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}

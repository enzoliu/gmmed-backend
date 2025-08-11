package models

import (
	"time"

	"github.com/guregu/null/v5"
)

// User 系統使用者模型
type User struct {
	ID           string     `json:"id" db:"id"`
	Username     string     `json:"username" db:"username"`
	Email        string     `json:"email" db:"email"`
	PasswordHash string     `json:"-" db:"password_hash"` // 不在 JSON 中顯示
	Role         string     `json:"role" db:"role"`
	IsActive     bool       `json:"is_active" db:"is_active"`
	LastLoginAt  *time.Time `json:"last_login_at" db:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// UserRole 使用者角色常數
type UserRole string

const (
	RoleAdmin    UserRole = "admin"
	RoleEditor   UserRole = "editor"
	RoleReadonly UserRole = "readonly"
)

type NullableUser struct {
	Username null.String `json:"username" db:"username"`
	Email    null.String `json:"email" db:"email"`
}

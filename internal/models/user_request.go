package models

import (
	"breast-implant-warranty-system/internal/entity"
	"time"

	"github.com/guregu/null/v5"
)

// CreateUserRequest 建立使用者請求
type CreateUserRequest struct {
	Username string   `json:"username" validate:"required,min=3,max=50"`
	Email    string   `json:"email" validate:"required,email"`
	Password string   `json:"password" validate:"required,min=8"`
	Role     UserRole `json:"role" validate:"required,oneof=admin editor readonly"`
}

// UpdateUserRequest 更新使用者請求
type UpdateUserRequest struct {
	Username *string   `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	Email    *string   `json:"email,omitempty" validate:"omitempty,email"`
	Password *string   `json:"password,omitempty" validate:"omitempty,min=8"`
	Role     *UserRole `json:"role,omitempty" validate:"omitempty,oneof=admin editor readonly"`
	IsActive *bool     `json:"is_active,omitempty"`
}

// LoginRequest 登入請求
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// GetMeResponse 獲取當前使用者資訊回應
type GetMeResponse struct {
	ExpiresAt time.Time `json:"expires_at"`
	User      User      `json:"user"`
}

// UserListResponse 使用者列表回應
type UserListResponse struct {
	Users      []*User `json:"users"`
	Total      int     `json:"total"`
	Page       int     `json:"page"`
	PageSize   int     `json:"page_size"`
	TotalPages int     `json:"total_pages"`
}

// UserSearchRequest 使用者搜尋請求
type UserSearchRequest struct {
	Username      null.String `query:"username"`
	Email         null.String `query:"email"`
	Role          null.String `query:"role"`
	IsActive      null.Bool   `query:"is_active"`
	SearchDeleted null.Bool   `query:"search_deleted"`
	entity.Pagination
}

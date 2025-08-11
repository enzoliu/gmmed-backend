package handlers

import (
	"net/http"
	"time"

	"breast-implant-warranty-system/internal/models"
	"breast-implant-warranty-system/internal/services"

	"github.com/labstack/echo/v4"
)

// AuthHandler 認證處理器
type AuthHandler struct {
	service *services.AuthService
}

// NewAuthHandler 建立新的認證處理器
func NewAuthHandler(service *services.AuthService) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}

// Login 登入
func (h *AuthHandler) Login(c echo.Context) error {
	var req models.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}

	// 驗證輸入
	if req.Username == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Username and password are required",
		})
	}

	// 執行登入
	statusCode, err := h.service.Login(c, req.Username, req.Password)
	if err != nil {
		return c.JSON(statusCode, map[string]string{
			"error": err.Error(),
		})
	}

	return c.NoContent(statusCode)
}

// RefreshToken 刷新令牌
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	rtCookie, err := c.Cookie("refresh_token")
	if err != nil {
		return echo.ErrUnauthorized
	}

	token := rtCookie.Value

	// 刷新令牌
	err = h.service.RefreshToken(c, token)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": err.Error(),
		})
	}

	return c.NoContent(http.StatusOK)
}

// Logout 登出
func (h *AuthHandler) Logout(c echo.Context) error {
	now := time.Now().Add(-1 * time.Hour)

	c.SetCookie(&http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Expires:  now,
	})

	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Expires:  now,
	})

	c.SetCookie(&http.Cookie{
		Name:     "csrf_token",
		Value:    "",
		Path:     "/",
		HttpOnly: false,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Expires:  now,
	})

	return c.NoContent(http.StatusOK)
}

// ChangePassword 更改密碼
func (h *AuthHandler) ChangePassword(c echo.Context) error {
	ctx := c.Request().Context()
	type ChangePasswordRequest struct {
		OldPassword string `json:"old_password" validate:"required"`
		NewPassword string `json:"new_password" validate:"required,min=8"`
	}

	var req ChangePasswordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "輸入內容有誤",
		})
	}

	// 從上下文取得使用者ID（由JWT中間件設置）
	userID := c.Get("user_id")
	if userID == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "未登入",
		})
	}

	// 更改密碼
	err := h.service.ChangePassword(ctx, userID.(string), req.OldPassword, req.NewPassword)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "密碼變更成功",
	})
}

func (h *AuthHandler) Me(c echo.Context) error {
	ctx := c.Request().Context()

	// 從上下文取得使用者ID（由JWT中間件設置）
	userID := c.Get("user_id")
	if userID == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "未登入",
		})
	}

	// 取得使用者資訊
	user, err := h.service.GetUserByID(ctx, userID.(string))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "取得使用者資訊失敗",
		})
	}
	if user == nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "使用者不存在",
		})
	}
	resp := models.GetMeResponse{
		ExpiresAt: c.Get("expires_at").(time.Time),
		User:      *user,
	}
	return c.JSON(http.StatusOK, resp)
}

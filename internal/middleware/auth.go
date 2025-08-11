package middleware

import (
	"log/slog"
	"net/http"

	"breast-implant-warranty-system/internal/utils"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type JWTMiddlewareConfigItf interface {
	JWTSecret() string
}

// JWTAuth JWT認證中間件
func JWTAuth() echo.MiddlewareFunc {
	return JWTAuthWithConfig(nil)
}

// JWTAuthWithConfig 帶設定的JWT認證中間件
func JWTAuthWithConfig(cfg JWTMiddlewareConfigItf) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie("access_token")
			if err != nil {
				slog.Error("Access token cookie not found", "error", err)
				return echo.ErrUnauthorized
			}
			// 提取令牌
			token := cookie.Value

			if token == "" {
				slog.Error("Access token is missing")
				return echo.ErrUnauthorized
			}

			var jwtSecret string
			if cfg != nil {
				jwtSecret = cfg.JWTSecret()
			}

			if jwtSecret == "" {
				slog.Error("JWT secret is not configured")
				return echo.ErrInternalServerError
			}

			// 驗證令牌
			claims, err := utils.ValidateJWT(token, jwtSecret)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "無效或過期的登入資訊，請重新登入",
				})
			}

			// 將使用者資訊設置到上下文
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("role", string(claims.Role))
			c.Set("expires_at", claims.ExpiresAt.Time)

			return next(c)
		}
	}
}

// RequireRole 要求特定角色的中間件
func RequireRole(roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userRole := c.Get("role")
			if userRole == nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "使用者角色不存在",
				})
			}

			roleStr := userRole.(string)
			for _, role := range roles {
				if roleStr == role {
					return next(c)
				}
			}

			return c.JSON(http.StatusForbidden, map[string]string{
				"error": "權限不足",
			})
		}
	}
}

// RequireAdmin 要求管理員角色的中間件
func RequireAdmin() echo.MiddlewareFunc {
	return RequireRole("admin")
}

// RequireAdminOrEditor 要求管理員或編輯者角色的中間件
func RequireAdminOrEditor() echo.MiddlewareFunc {
	return RequireRole("admin", "editor")
}

func CSRFProtection() echo.MiddlewareFunc {
	return middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup:    "header:X-CSRF-Token",
		CookieName:     "csrf_token",
		CookieMaxAge:   3600,
		CookieHTTPOnly: false,
		CookiePath:     "/",
		CookieSecure:   true,
		CookieSameSite: http.SameSiteNoneMode,
		ContextKey:     "csrf",
	})
}

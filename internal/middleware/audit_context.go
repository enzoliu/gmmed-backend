package middleware

import (
	"breast-implant-warranty-system/internal/models"

	"github.com/labstack/echo/v4"
)

// AuditContext 創建 audit context middleware
func AuditContext() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 創建 audit context
			auditCtx := &models.AuditContext{
				IPAddress: &[]string{c.RealIP()}[0],
				UserAgent: &[]string{c.Request().UserAgent()}[0],
			}

			// 獲取用戶信息（如果已認證）
			if userID := c.Get("user_id"); userID != nil {
				if uidStr, ok := userID.(string); ok {
					auditCtx.UserID = &uidStr
				}
			}

			if username := c.Get("username"); username != nil {
				if usernameStr, ok := username.(string); ok {
					auditCtx.Username = &usernameStr
				}
			}

			// 將 audit context 設置到 echo context
			c.Set("audit_context", auditCtx)

			return next(c)
		}
	}
}

// GetAuditContext 從 echo context 獲取 audit context
func GetAuditContext(c echo.Context) *models.AuditContext {
	if auditCtx := c.Get("audit_context"); auditCtx != nil {
		if ctx, ok := auditCtx.(*models.AuditContext); ok {
			return ctx
		}
	}
	return nil
}

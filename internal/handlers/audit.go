package handlers

import (
	"net/http"

	"breast-implant-warranty-system/internal/models"
	"breast-implant-warranty-system/internal/services"
	"breast-implant-warranty-system/pkg/validator"

	"github.com/labstack/echo/v4"
)

// AuditHandler 審計處理器
type AuditHandler struct {
	service *services.AuditService
}

// NewAuditHandler 建立新的審計處理器
func NewAuditHandler(service *services.AuditService) *AuditHandler {
	return &AuditHandler{
		service: service,
	}
}

// List 列出審計日誌
func (h *AuditHandler) List(c echo.Context) error {
	ctx := c.Request().Context()
	// 解析查詢參數
	var req models.AuditSearchLogRequest
	if err := validator.Load(c, &req); err != nil {
		return err
	}

	response, err := h.service.Search(ctx, &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get audit logs: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response)
}

// GetByID 根據ID取得審計日誌
func (h *AuditHandler) GetByID(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "ID is required",
		})
	}

	auditLog, err := h.service.GetByID(ctx, id)
	if err != nil {
		if err.Error() == "audit log not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Audit log not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get audit log: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, auditLog)
}

package handlers

import (
	"net/http"
	"strings"

	"breast-implant-warranty-system/internal/middleware"
	"breast-implant-warranty-system/internal/models"
	"breast-implant-warranty-system/internal/services"
	"breast-implant-warranty-system/pkg/validator"

	"github.com/labstack/echo/v4"
)

// SerialHandler 序號處理器
type SerialHandler struct {
	service *services.SerialService
}

// NewSerialHandler 建立新的序號處理器
func NewSerialHandler(service *services.SerialService) *SerialHandler {
	return &SerialHandler{
		service: service,
	}
}

// Create 建立序號
func (h *SerialHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()

	// 驗證請求
	var req services.SerialCreateRequest
	if err := validator.Load(c, &req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	// 取得審計上下文
	auditCtx := middleware.GetAuditContext(c)

	// 建立序號
	serial, err := h.service.Create(ctx, &req, auditCtx)
	if err != nil {
		if strings.Contains(err.Error(), "(SQLSTATE 23505)") {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "該序號已存在。",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, serial)
}

// GetByID 根據ID取得序號
func (h *SerialHandler) GetByID(c echo.Context) error {
	ctx := c.Request().Context()

	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "serial ID is required",
		})
	}

	serial, err := h.service.GetByID(ctx, id)
	if err != nil {
		if err.Error() == "serial not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": err.Error(),
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, serial)
}

// GetBySerialNumber 根據序號取得序號
func (h *SerialHandler) GetBySerialNumber(c echo.Context) error {
	ctx := c.Request().Context()

	serialNumber := c.QueryParam("serial_number")
	if serialNumber == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "serial number is required",
		})
	}

	serial, err := h.service.GetBySerialNumber(ctx, serialNumber)
	if err != nil {
		if err.Error() == "serial not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": err.Error(),
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, serial)
}

// Update 更新序號
func (h *SerialHandler) Update(c echo.Context) error {
	ctx := c.Request().Context()

	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "serial ID is required",
		})
	}

	// 驗證請求
	var req services.SerialUpdateRequest
	if err := validator.Load(c, &req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	// 取得審計上下文
	auditCtx := middleware.GetAuditContext(c)

	// 更新序號
	serial, err := h.service.Update(ctx, id, &req, auditCtx)
	if err != nil {
		if err.Error() == "serial not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": err.Error(),
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, serial)
}

// Delete 刪除序號
func (h *SerialHandler) Delete(c echo.Context) error {
	ctx := c.Request().Context()

	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "serial ID is required",
		})
	}

	// 取得審計上下文
	auditCtx := middleware.GetAuditContext(c)

	// 刪除序號
	err := h.service.Delete(ctx, id, auditCtx)
	if err != nil {
		if err.Error() == "serial not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": err.Error(),
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "serial deleted successfully",
	})
}

// Search 搜尋序號
func (h *SerialHandler) Search(c echo.Context) error {
	ctx := c.Request().Context()

	// 驗證請求
	var req models.SerialSearchRequest
	if err := validator.Load(c, &req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	// 處理分頁參數
	page := &req.Pagination
	if page.Page < 1 {
		page.Page = 1
	}
	if page.PageSize < 1 || page.PageSize > 100 {
		page.PageSize = 20
	}

	// 搜尋序號
	response, err := h.service.Search(ctx, &req, page)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response)
}

// BulkCreate 大量建立序號
func (h *SerialHandler) BulkCreate(c echo.Context) error {
	ctx := c.Request().Context()

	// 驗證請求
	var req models.SerialBulkImportRequest
	if err := validator.Load(c, &req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	// 取得審計上下文
	auditCtx := middleware.GetAuditContext(c)

	// 執行大量建立
	response, err := h.service.BulkCreate(ctx, &req, auditCtx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusMultiStatus, response)
}

// CheckSerialExists 檢查序號是否存在
func (h *SerialHandler) CheckSerialExists(c echo.Context) error {
	ctx := c.Request().Context()

	serialNumber := c.QueryParam("serial_number")
	if serialNumber == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "serial number is required",
		})
	}

	exists, err := h.service.CheckSerialExists(ctx, serialNumber)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	if exists {
		return c.NoContent(http.StatusOK)
	} else {
		return c.NoContent(http.StatusNotFound)
	}
}

// GetSerialsWithProduct 取得序號及其產品資訊
func (h *SerialHandler) GetSerialsWithProduct(c echo.Context) error {
	ctx := c.Request().Context()

	// 驗證請求
	var req models.SerialSearchRequest
	if err := validator.Load(c, &req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	// 處理分頁參數
	page := &req.Pagination
	if page.Page < 1 {
		page.Page = 1
	}
	if page.PageSize < 1 || page.PageSize > 100 {
		page.PageSize = 20
	}

	// 取得序號及其產品資訊
	response, err := h.service.GetSerialsWithProduct(ctx, &req, page)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response)
}

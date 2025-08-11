package handlers

import (
	"net/http"

	"breast-implant-warranty-system/internal/middleware"
	"breast-implant-warranty-system/internal/models"
	"breast-implant-warranty-system/internal/services"
	"breast-implant-warranty-system/internal/utils"
	"breast-implant-warranty-system/pkg/validator"

	"github.com/labstack/echo/v4"
)

// WarrantyHandler 保固處理器
type WarrantyHandler struct {
	service *services.WarrantyService
}

// NewWarrantyHandler 建立新的保固處理器
func NewWarrantyHandler(service *services.WarrantyService) *WarrantyHandler {
	return &WarrantyHandler{
		service: service,
	}
}

// List 列出保固
func (h *WarrantyHandler) List(c echo.Context) error {
	ctx := c.Request().Context()

	// validate
	var req models.WarrantySearchRequest
	if err := validator.Load(c, &req); err != nil {
		return err
	}

	response, err := h.service.Search(ctx, &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response)
}

// GetByID 根據ID取得保固
func (h *WarrantyHandler) GetByID(c echo.Context) error {
	ctx := c.Request().Context()
	// 取得路徑參數中的ID
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "必須提供保固ID資訊"})
	}

	// 從服務層取得保固記錄
	warranty, err := h.service.GetByID(ctx, id)
	if err != nil {
		if err.Error() == "warranty registration not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "保固記錄不存在"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "取得保固記錄失敗"})
	}

	return c.JSON(http.StatusOK, warranty)
}

// Search 搜尋保固（管理員用）
func (h *WarrantyHandler) Search(c echo.Context) error {
	ctx := c.Request().Context()

	// validate
	var req models.WarrantySearchRequest
	if err := validator.Load(c, &req); err != nil {
		return err
	}

	response, err := h.service.Search(ctx, &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response)
}

// Update 更新保固
func (h *WarrantyHandler) Update(c echo.Context) error {
	ctx := c.Request().Context()
	// 取得路徑參數中的ID
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "必須提供保固ID資訊"})
	}

	var req models.WarrantyUpdateRequest
	if err := bindAndValidateRequest(c, &req); err != nil {
		return err
	}

	auditCtx := middleware.GetAuditContext(c)
	warranty, err := h.service.Update(ctx, id, &req, auditCtx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "寫入保固記錄失敗, 原因:" + err.Error()})
	}

	return c.JSON(http.StatusOK, warranty)
}

// Delete 刪除保固
func (h *WarrantyHandler) Delete(c echo.Context) error {
	ctx := c.Request().Context()
	// 取得路徑參數中的ID
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "必須提供保固ID資訊"})
	}

	auditCtx := middleware.GetAuditContext(c)
	err := h.service.Delete(ctx, id, auditCtx)
	if err != nil {
		if err.Error() == "warranty registration not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "保固記錄不存在"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "刪除保固記錄失敗"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "保固記錄刪除成功"})
}

// ResendEmail 重新發送信件
func (h *WarrantyHandler) ResendEmail(c echo.Context) error {
	ctx := c.Request().Context()
	// 取得路徑參數中的ID
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "必須提供保固ID資訊"})
	}

	auditCtx := middleware.GetAuditContext(c)
	err := h.service.ResendConfirmationEmail(ctx, id, auditCtx)
	if err != nil {
		if err.Error() == "warranty registration not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "保固記錄不存在"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "重新發送信件失敗"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "信件重新發送成功"})
}

// ExportExcel 匯出Excel
func (h *WarrantyHandler) ExportExcel(c echo.Context) error {
	ctx := c.Request().Context()

	// validate
	var req models.WarrantySearchRequest
	if err := validator.Load(c, &req); err != nil {
		return err
	}
	req.PageSize = 10000 // 匯出大量資料

	response, err := h.service.Search(ctx, &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	// 記錄匯出操作的 audit 日誌
	go func() {
		exportData := map[string]interface{}{
			"action":       "export_excel",
			"search_query": req,
			"total_count":  response.Total,
			"success":      true,
		}
		// 這裡應該調用 audit 服務，但在 handler 層面比較複雜
		// 實際實現中可以通過 middleware 或者在服務層添加匯出方法
		_ = exportData
	}()

	// TODO: 實現Excel匯出邏輯
	// 目前返回JSON格式的資料，後續可以實現真正的Excel匯出
	c.Response().Header().Set("Content-Type", "application/json")
	c.Response().Header().Set("Content-Disposition", "attachment; filename=warranty_export.json")

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Excel匯出功能 - 目前返回JSON資料",
		"data":    response,
		"total":   response.Total,
	})
}

// GetStatistics 取得統計資料
func (h *WarrantyHandler) GetStatistics(c echo.Context) error {
	ctx := c.Request().Context()
	stats, err := h.service.GetStatistics(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, stats)
}

// UpdateExpiredWarranties 批次更新過期保固狀態
func (h *WarrantyHandler) UpdateExpiredWarranties(c echo.Context) error {
	ctx := c.Request().Context()
	auditCtx := middleware.GetAuditContext(c)
	updatedCount, err := h.service.UpdateExpiredWarranties(ctx, auditCtx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":       "過期保固更新成功",
		"updated_count": updatedCount,
	})
}

// CheckSerialNumber 檢查產品序號是否已被使用
func (h *WarrantyHandler) CheckSerialNumber(c echo.Context) error {
	ctx := c.Request().Context()
	serialNumber := c.QueryParam("serial_number")
	if serialNumber == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "必須提供產品序號資訊",
		})
	}

	exists, err := h.service.CheckSerialNumberExists(ctx, serialNumber)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	response := models.SerialNumberCheckResponse{
		Exists: exists,
	}

	if exists {
		response.Message = "產品序號已註冊"
	} else {
		response.Message = "產品序號可用"
	}

	return c.JSON(http.StatusOK, response)
}

// BatchCreate 批次創建空白保固記錄（管理員專用）
func (h *WarrantyHandler) BatchCreate(c echo.Context) error {
	ctx := c.Request().Context()
	var req models.BatchCreateRequest
	if err := bindAndValidateRequest(c, &req); err != nil {
		return err
	}

	auditCtx := middleware.GetAuditContext(c)
	ids, err := h.service.BatchCreateEmptyWarranties(ctx, req.Count, auditCtx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	response := models.BatchCreateResponse{
		Count: len(ids),
		IDs:   ids,
	}

	return c.JSON(http.StatusCreated, response)
}

// CheckWarrantyStatus 檢查保固是否已填寫
func (h *WarrantyHandler) CheckWarrantyStatus(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "必須提供保固ID資訊",
		})
	}

	canEdit, err := h.service.CheckWarrantyCanEdit(ctx, id)
	if err != nil {
		if err.Error() == "warranty not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "保固記錄不存在",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	response := models.WarrantyStatusResponse{
		CanEdit: canEdit,
	}

	if canEdit {
		response.Message = "保固可以填寫"
	} else {
		response.Message = "保固已填寫"
	}

	return c.JSON(http.StatusOK, response)
}

// RegisterByPatient 患者填寫保固（一次性，無需認證）
func (h *WarrantyHandler) RegisterByPatient(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "必須提供保固ID資訊",
		})
	}

	var req models.PatientRegistrationRequest
	if err := validator.Load(c, &req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	auditCtx := middleware.GetAuditContext(c)
	warranty, err := h.service.RegisterByPatient(ctx, id, &req, auditCtx)
	if err != nil {
		if err.Error() == "warranty has already been filled" {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "保固已填寫",
			})
		}
		if err.Error() == "warranty not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "保固記錄不存在",
			})
		}
		if err.Error() == "product serial number already registered" {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "產品序號已被註冊",
			})
		}
		if err.Error() == "second product serial number already registered" {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "第二個產品序號已被註冊",
			})
		}
		if err.Error() == "two serial numbers cannot be the same" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "兩個序號不能相同",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "註冊保固失敗, 原因:" + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, warranty)
}

// bindAndValidateRequest 綁定並驗證請求
func bindAndValidateRequest(c echo.Context, req interface{}) error {
	return utils.BindAndValidate(c, req)
}

package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"breast-implant-warranty-system/internal/middleware"
	"breast-implant-warranty-system/internal/models"
	"breast-implant-warranty-system/internal/services"
	"breast-implant-warranty-system/internal/utils"
	"breast-implant-warranty-system/pkg/validator"

	"github.com/labstack/echo/v4"
)

// WarrantyHandler 保固處理器
type WarrantyHandler struct {
	service       *services.WarrantyService
	serialService *services.SerialService
	cfg           services.WarrantyRouteConfigItf
}

// NewWarrantyHandler 建立新的保固處理器
func NewWarrantyHandler(service *services.WarrantyService, serialService *services.SerialService, cfg services.WarrantyRouteConfigItf) *WarrantyHandler {
	return &WarrantyHandler{
		service:       service,
		serialService: serialService,
		cfg:           cfg,
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
	warrantyID := c.QueryParam("warranty_id")
	if serialNumber == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "必須提供產品序號資訊",
		})
	}
	// 防止有心人士直接呼叫此API，所有錯誤都回傳403
	if warrantyID == "" {
		return c.NoContent(http.StatusForbidden)
	}
	// 如果保固已經填寫過，則回傳403（避免有心人士直接呼叫此API去try）
	step, err := h.service.GetWarrantyStatusByPatient(ctx, warrantyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}
	if step >= models.STEP_SERIAL_VERIFIED {
		return c.NoContent(http.StatusForbidden)
	}

	exists, err := h.service.CheckSerialNumberExists(ctx, serialNumber)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	// 呼叫serial service 的 CheckSerialExists
	productID, err := h.serialService.CheckSerialExists(ctx, serialNumber)
	if err != nil {
		// 在伺服器端印出錯誤, 使用slog
		slog.Error("CheckSerialNumber - CheckSerialExists", "error", err)
		return c.NoContent(http.StatusForbidden)
	}
	if productID == "" {
		slog.Error("CheckSerialNumber - CheckSerialExists", "error", "序號不存在")
		return c.NoContent(http.StatusForbidden)
	}

	response := models.SerialNumberCheckResponse{
		Exists:    exists,
		ProductID: productID,
		Message:   "產品序號可使用",
	}
	if exists {
		response.ProductID = ""
		response.Message = "產品序號已被註冊"
		return c.JSON(http.StatusConflict, response)
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

// RegisterByPatientStep1 患者填寫保固（檢查序號是否是正貨/登記手術日）
func (h *WarrantyHandler) RegisterByPatientStep1(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "必須提供保固ID資訊",
		})
	}

	var req models.PatientRegistrationRequestStep1
	if err := validator.Load(c, &req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	auditCtx := middleware.GetAuditContext(c)
	warranty, err := h.service.RegisterByPatientStep1(ctx, id, &req, auditCtx)
	if err != nil {
		if err.Error() == "warranty not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "保固記錄不存在",
			})
		}
		if err.Error() == "warranty has already been filled" {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "無法填寫保固，可能原因：狀態不正確或是無法驗證您的裝置",
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
		if err.Error() == "product serial number not valid" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "產品序號無效",
			})
		}
		if err.Error() == "surgery date cannot be in the future" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "手術日期不能在未來",
			})
		}
		if err.Error() == "product not found" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "產品不存在",
			})
		}
		if err.Error() == "product is not active" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "產品已停用",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	if warranty.Step == models.STEP_VERIFIED_WITHOUT_WARRANTY {
		return c.NoContent(http.StatusOK)
	}

	// 驗證用的 cookie，1年後過期，保固續填必須要是同個裝置，避免被有心人士利用
	encryptedStep, err := utils.EncryptAES(fmt.Sprintf("%s-%d", warranty.ID, warranty.Step), h.cfg.EncryptionKey())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}
	expiresAt := time.Now().AddDate(1, 0, 0)
	// expires at 1 year
	c.SetCookie(&http.Cookie{
		Name:     "warranty_step",
		Value:    encryptedStep,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Expires:  expiresAt,
	})

	return c.JSON(http.StatusOK, warranty)
}

// RegisterByPatientStep2 患者填寫保固（更新患者資訊）
func (h *WarrantyHandler) RegisterByPatientStep2(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "必須提供保固ID資訊",
		})
	}

	var req models.PatientRegistrationRequestStep2
	if err := validator.Load(c, &req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	// 驗證 cookie
	valid, err := h.isPatientWarrantyCookieValid(c, id, []int{models.STEP_SERIAL_VERIFIED, models.STEP_PATIENT_INFO_FILLED})
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}
	if !valid {
		return c.NoContent(http.StatusForbidden)
	}

	auditCtx := middleware.GetAuditContext(c)
	_, err = h.service.RegisterByPatientStep2(ctx, id, &req, auditCtx)
	if err != nil {
		if err.Error() == "warranty not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "保固記錄不存在",
			})
		}
		if err.Error() == "warranty can not be filled" {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "無法填寫保固，可能原因：狀態不正確或是無法驗證您的裝置。",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	// update cookie
	encryptedStep, err := utils.EncryptAES(fmt.Sprintf("%s-%d", id, models.STEP_PATIENT_INFO_FILLED), h.cfg.EncryptionKey())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}
	expiresAt := time.Now().AddDate(1, 0, 0)
	c.SetCookie(&http.Cookie{
		Name:     "warranty_step",
		Value:    encryptedStep,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Expires:  expiresAt,
	})

	return c.NoContent(http.StatusOK)
}

// RegisterByPatientStep3 患者確認保固資訊（建立保固）
func (h *WarrantyHandler) RegisterByPatientStep3(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "必須提供保固ID資訊",
		})
	}
	// 驗證 cookie
	valid, err := h.isPatientWarrantyCookieValid(c, id, []int{models.STEP_PATIENT_INFO_FILLED})
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}
	if !valid {
		return c.NoContent(http.StatusForbidden)
	}

	auditCtx := middleware.GetAuditContext(c)
	_, err = h.service.RegisterByPatientStep3(ctx, id, auditCtx)
	if err != nil {
		if err.Error() == "warranty not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "保固記錄不存在",
			})
		}
		if err.Error() == "warranty can not be confirmed" {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "無法確認保固，可能原因：狀態不正確或是無法驗證您的裝置",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}
	// remove cookie
	c.SetCookie(&http.Cookie{
		Name:     "warranty_step",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Expires:  time.Now().Add(-1 * time.Hour),
	})

	return c.NoContent(http.StatusOK)
}

// GetWarrantyByPatient 取得患者保固資訊
func (h *WarrantyHandler) GetWarrantyByPatient(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "必須提供保固ID資訊",
		})
	}

	// 驗證 cookie
	valid, err := h.isPatientWarrantyCookieValid(c, id, []int{models.STEP_SERIAL_VERIFIED, models.STEP_PATIENT_INFO_FILLED})
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}
	if !valid {
		return c.NoContent(http.StatusForbidden)
	}

	warranty, err := h.service.GetWarrantyByPatientInSteps(ctx, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, warranty)
}

func (h *WarrantyHandler) GetWarrantyStatusByPatient(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "必須提供保固ID資訊",
		})
	}

	step, err := h.service.GetWarrantyStatusByPatient(ctx, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, models.WarrantyStepsResponse{
		Step: step,
	})
}

func (h *WarrantyHandler) isPatientWarrantyCookieValid(c echo.Context, desiredID string, desiredStep []int) (bool, error) {
	// 驗證 cookie
	cookie, err := c.Cookie("warranty_step")
	if err != nil || cookie == nil {
		return false, fmt.Errorf("保固續填必須在同一設備上進行")
	}
	if cookie.Value == "" {
		return false, nil
	}
	decryptedStep, err := utils.DecryptAES(cookie.Value, h.cfg.EncryptionKey())
	if err != nil {
		return false, nil
	}
	for _, step := range desiredStep {
		if decryptedStep == fmt.Sprintf("%s-%d", desiredID, step) {
			return true, nil
		}
	}
	return true, nil
}

// bindAndValidateRequest 綁定並驗證請求
func bindAndValidateRequest(c echo.Context, req interface{}) error {
	return utils.BindAndValidate(c, req)
}

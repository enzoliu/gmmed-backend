package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"log/slog"
	"regexp"
	"slices"

	"breast-implant-warranty-system/internal/entity"
	"breast-implant-warranty-system/internal/models"
	"breast-implant-warranty-system/internal/repositories"
	"breast-implant-warranty-system/internal/utils"
	"breast-implant-warranty-system/pkg/dbutil"

	"github.com/guregu/null/v5"
)

type SerialRouteConfigItf interface {
	EncryptionKey() string
}

// SerialService 序號服務
type SerialService struct {
	db           dbutil.PgxClientItf
	cfg          SerialRouteConfigItf
	serialRepo   *repositories.SerialRepository
	productRepo  *repositories.ProductRepository
	auditService *AuditService
}

// NewSerialService 建立新的序號服務
func NewSerialService(db dbutil.PgxClientItf, cfg SerialRouteConfigItf) *SerialService {
	return &SerialService{
		db:           db,
		cfg:          cfg,
		serialRepo:   repositories.NewSerialRepository(db),
		productRepo:  repositories.NewProductRepository(db),
		auditService: NewAuditService(db),
	}
}

// SerialCreateRequest 序號建立請求
type SerialCreateRequest struct {
	SerialNumber     string `json:"serial_number" validate:"required,min=1,max=20"`
	FullSerialNumber string `json:"full_serial_number" validate:"required,min=1,max=100"`
	ProductID        string `json:"product_id" validate:"required,uuid"`
}

// SerialUpdateRequest 序號更新請求
type SerialUpdateRequest struct {
	SerialNumber     null.String `json:"serial_number" validate:"min=1,max=20"`
	FullSerialNumber null.String `json:"full_serial_number" validate:"min=1,max=100"`
	ProductID        null.String `json:"product_id" validate:"uuid"`
}

// SerialListResponse 序號列表回應
type SerialListResponse struct {
	Serials    []*models.SerialWithChecksum `json:"serials"`
	Total      int                          `json:"total"`
	Page       int                          `json:"page"`
	PageSize   int                          `json:"page_size"`
	TotalPages int                          `json:"total_pages"`
}

type SerialListResponseWithWarranty struct {
	Serials    []*models.SerialWithWarranty `json:"serials"`
	Total      int                          `json:"total"`
	Page       int                          `json:"page"`
	PageSize   int                          `json:"page_size"`
	TotalPages int                          `json:"total_pages"`
}

// Create 建立序號
func (s *SerialService) Create(ctx context.Context, req *SerialCreateRequest, auditCtx *models.AuditContext) (*models.Serial, error) {
	// 驗證輸入
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// 如果指定了產品ID，檢查產品是否存在
	if req.ProductID != "" {
		product, err := s.productRepo.GetByID(ctx, req.ProductID)
		if err != nil {
			return nil, fmt.Errorf("failed to get product: %w", err)
		}
		if product == nil {
			return nil, errors.New("product not found")
		}
	}

	// 建立序號
	serial := &models.Serial{
		SerialNumber:     utils.SanitizeString(req.SerialNumber),
		FullSerialNumber: utils.SanitizeString(req.FullSerialNumber),
		ProductID:        null.StringFrom(req.ProductID),
	}

	// 儲存到資料庫
	createReq := &models.SerialCreateRequest{
		SerialNumber:     serial.SerialNumber,
		FullSerialNumber: serial.FullSerialNumber,
		ProductID:        serial.ProductID.String,
	}

	createdSerialID, err := s.serialRepo.Create(ctx, createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create serial: %w", err)
	}

	// 取得建立的序號
	createdSerial, err := s.serialRepo.GetByID(ctx, createdSerialID)
	if err != nil {
		return nil, fmt.Errorf("failed to get created serial: %w", err)
	}

	// 記錄審計日誌
	if auditCtx != nil {
		s.recordAuditLog(ctx, auditCtx, "CREATE", createdSerial.ID, nil, createdSerial)
	}

	return createdSerial, nil
}

// GetByID 根據ID取得序號
func (s *SerialService) GetByID(ctx context.Context, id string) (*models.SerialWithChecksum, error) {
	if id == "" {
		return nil, errors.New("serial ID is required")
	}

	serial, err := s.serialRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get serial: %w", err)
	}

	if serial == nil {
		return nil, errors.New("serial not found")
	}

	return &models.SerialWithChecksum{
		Serial: models.Serial{
			ID:               serial.ID,
			SerialNumber:     serial.SerialNumber,
			FullSerialNumber: serial.FullSerialNumber,
			ProductID:        serial.ProductID,
			CreatedAt:        serial.CreatedAt,
			UpdatedAt:        serial.UpdatedAt,
		},
		// Checksum: s.generateSerialKey(serial.SerialNumber),
	}, nil
}

// Update 更新序號
func (s *SerialService) Update(ctx context.Context, id string, req *SerialUpdateRequest, auditCtx *models.AuditContext) (*models.Serial, error) {
	// 驗證輸入
	if err := s.validateUpdateRequest(req); err != nil {
		return nil, err
	}

	// 取得更新前的序號
	existingSerial, err := s.serialRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("無法取得序號: %w", err)
	}
	if existingSerial == nil {
		return nil, errors.New("序號不存在，無法更新")
	}

	// 檢查序號是否已經被保固使用
	serials, _, _ := s.serialRepo.Search(ctx, &models.SerialSearchRequest{
		ID:               null.StringFrom(id),
		IsUsedByWarranty: null.BoolFrom(true),
	}, &entity.Pagination{
		Limit: 1,
	})
	if len(serials) > 0 && req.SerialNumber.Valid && serials[0].SerialNumber != req.SerialNumber.String {
		return nil, errors.New("序號已被保固使用，無法更新")
	}

	// 如果指定了產品ID，檢查產品是否存在
	if req.ProductID.Valid {
		product, err := s.productRepo.GetByID(ctx, req.ProductID.String)
		if err != nil {
			return nil, fmt.Errorf("無法取得產品: %w", err)
		}
		if product == nil {
			return nil, errors.New("產品不存在")
		}
	}

	// 準備更新請求
	updateReq := &models.SerialUpdateRequest{}

	if req.SerialNumber.Valid {
		updateReq.SerialNumber = null.StringFrom(utils.SanitizeString(req.SerialNumber.String))
	}
	if req.FullSerialNumber.Valid {
		updateReq.FullSerialNumber = null.StringFrom(utils.SanitizeString(req.FullSerialNumber.String))
	}
	if req.ProductID.Valid {
		updateReq.ProductID = null.StringFrom(req.ProductID.String)
	}

	// 更新序號
	err = s.serialRepo.Update(ctx, id, updateReq)
	if err != nil {
		return nil, fmt.Errorf("failed to update serial: %w", err)
	}

	// 取得更新後的序號
	updatedSerial, err := s.serialRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated serial: %w", err)
	}

	// 記錄審計日誌
	if auditCtx != nil {
		s.recordAuditLog(ctx, auditCtx, "UPDATE", id, existingSerial, updatedSerial)
	}

	return updatedSerial, nil
}

// Delete 刪除序號
func (s *SerialService) Delete(ctx context.Context, id string, auditCtx *models.AuditContext) error {
	if id == "" {
		return errors.New("serial ID is required")
	}

	// 檢查序號是否存在
	existingSerial, err := s.serialRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get existing serial: %w", err)
	}
	if existingSerial == nil {
		return errors.New("serial not found")
	}

	// 檢查序號是否被保固使用
	serials, _, err := s.serialRepo.Search(ctx, &models.SerialSearchRequest{
		ID:               null.StringFrom(id),
		IsUsedByWarranty: null.BoolFrom(true),
	}, &entity.Pagination{
		Limit: 1,
	})
	if err != nil {
		return fmt.Errorf("failed to search serials: %w", err)
	}
	if len(serials) > 0 {
		return errors.New("cannot delete serial that is used by warranty")
	}

	// 刪除序號
	err = s.serialRepo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete serial: %w", err)
	}

	// 記錄審計日誌
	if auditCtx != nil {
		s.recordAuditLog(ctx, auditCtx, "DELETE", id, existingSerial, nil)
	}

	return nil
}

// Search 搜尋序號
func (s *SerialService) Search(ctx context.Context, req *models.SerialSearchRequest, page *entity.Pagination) (*SerialListResponseWithWarranty, error) {
	// 驗證分頁參數
	if page == nil {
		page = &entity.Pagination{
			Page:     1,
			PageSize: 20,
		}
	}

	if page.Page < 1 {
		page.Page = 1
	}
	if page.PageSize < 1 || page.PageSize > 100 {
		page.PageSize = 20
	}

	// 計算偏移量
	page.Offset = (page.Page - 1) * page.PageSize

	// 搜尋序號
	serials, total, err := s.serialRepo.Search(ctx, req, page)
	if err != nil {
		return nil, fmt.Errorf("failed to search serials: %w", err)
	}

	// 暫時不使用checksum
	// for _, serial := range serials {
	// 	serial.Checksum = s.generateSerialKey(serial.SerialNumber)
	// }

	// 計算總頁數
	totalPages := int(float64(total) / float64(page.PageSize))
	if total%page.PageSize > 0 {
		totalPages++
	}

	return &SerialListResponseWithWarranty{
		Serials:    serials,
		Total:      total,
		Page:       page.Page,
		PageSize:   page.PageSize,
		TotalPages: totalPages,
	}, nil
}

// BulkCreate 大量建立序號
func (s *SerialService) BulkCreate(ctx context.Context, req *models.SerialBulkImportRequest, auditCtx *models.AuditContext) (*models.SerialBulkImportResponse, error) {
	response := &models.SerialBulkImportResponse{
		SuccessCount: 0,
		FailedCount:  0,
		FailedItems:  []models.SerialImportErrorItem{},
	}

	if len(req.Serials) == 0 {
		return response, nil
	}

	// 驗證每個序號項目
	validSerials := []models.SerialImportItem{}
	for _, serial := range req.Serials {
		if err := s.validateSerialImportItem(&serial); err != nil {
			response.FailedItems = append(response.FailedItems, models.SerialImportErrorItem{
				Index:            serial.Index,
				ProductID:        serial.ProductID,
				SerialNumber:     serial.SerialNumber,
				FullSerialNumber: serial.FullSerialNumber,
				Error:            err.Error(),
			})
			response.FailedCount++
		} else {
			validSerials = append(validSerials, serial)
		}
	}

	// 檢查產品是否存在
	productIDsMap := map[string]struct{}{}
	for _, serial := range validSerials {
		if serial.ProductID != "" {
			productIDsMap[serial.ProductID] = struct{}{}
		}
	}
	productIDs := make([]string, 0, len(productIDsMap))
	for k := range productIDsMap {
		productIDs = append(productIDs, k)
	}

	validSerials2 := []models.SerialImportItem{}
	if len(productIDs) > 0 {
		nonExistingProductIDs, err := s.productRepo.CheckManyProductIDExists(ctx, productIDs)
		if err != nil {
			return nil, fmt.Errorf("無法檢查產品，原因: %s", err.Error())
		}
		if len(nonExistingProductIDs) > 0 {
			for _, serial := range validSerials {
				if slices.Contains(nonExistingProductIDs, serial.ProductID) {
					response.FailedItems = append(response.FailedItems, models.SerialImportErrorItem{
						Index:            serial.Index,
						ProductID:        serial.ProductID,
						SerialNumber:     serial.SerialNumber,
						FullSerialNumber: serial.FullSerialNumber,
						Error:            "產品不存在。",
					})
					response.FailedCount++
				} else {
					validSerials2 = append(validSerials2, serial)
				}
			}
		} else {
			validSerials2 = validSerials
		}
	}
	req.Serials = validSerials2

	// 執行大量建立
	err := s.serialRepo.BulkCreate(ctx, req, response)
	if err != nil {
		return nil, err
	}

	// 記錄審計日誌
	if auditCtx != nil {
		s.recordAuditLog(ctx, auditCtx, "BULK_CREATE", "", nil, map[string]interface{}{
			"total_count":   len(req.Serials),
			"success_count": response.SuccessCount,
			"failed_count":  response.FailedCount,
		})
	}

	return response, nil
}

func (s *SerialService) IsValidSerialNumberAndGetProductID(ctx context.Context, serialNumber, checksum string) (bool, string, error) {
	if serialNumber == "" {
		return false, "", errors.New("序號或驗證碼不正確-ERR_SERIAL_001")
	}
	re := regexp.MustCompile(`^\d{7}-\d{3}$`)
	if !re.MatchString(serialNumber) {
		return false, "", errors.New("序號或驗證碼不正確-ERR_SERIAL_002")
	}
	// 暫時不使用checksum
	// if checksum != s.generateSerialKey(serialNumber) {
	// 	return false, "", errors.New("序號或驗證碼不正確-ERR_SERIAL_003")
	// }
	valid, productID, err := s.serialRepo.IsValidSerialNumberAndGetProductID(ctx, serialNumber)
	if err != nil {
		slog.Error("IsValidSerialNumber", "serialNumber", serialNumber, "error", err)
		return false, "", errors.New("序號或驗證碼不正確-ERR_SERIAL_004")
	}
	return valid, productID, nil
}

// GetSerialsWithProduct 取得序號及其產品資訊
func (s *SerialService) GetSerialsWithProduct(ctx context.Context, req *models.SerialSearchRequest, page *entity.Pagination) (*SerialListResponse, error) {
	// 驗證分頁參數
	if page == nil {
		page = &entity.Pagination{
			Page:     1,
			PageSize: 20,
		}
	}

	if page.Page < 1 {
		page.Page = 1
	}
	if page.PageSize < 1 || page.PageSize > 100 {
		page.PageSize = 20
	}

	// 計算偏移量
	page.Offset = (page.Page - 1) * page.PageSize

	// 搜尋序號及其產品資訊
	serials, total, err := s.serialRepo.GetSerialsWithProduct(ctx, req, page)
	if err != nil {
		return nil, fmt.Errorf("failed to get serials with product: %w", err)
	}

	// 轉換為 Serial 類型
	serialModels := make([]*models.SerialWithChecksum, len(serials))
	for i, serialDetail := range serials {
		serialModels[i] = &models.SerialWithChecksum{
			Serial: models.Serial{
				ID:               serialDetail.ID,
				SerialNumber:     serialDetail.SerialNumber,
				FullSerialNumber: serialDetail.FullSerialNumber,
				ProductID:        serialDetail.ProductID,
				CreatedAt:        serialDetail.CreatedAt,
				UpdatedAt:        serialDetail.UpdatedAt,
			},
			// Checksum: s.generateSerialKey(serialDetail.SerialNumber),
		}
	}

	// 計算總頁數
	totalPages := int(float64(total) / float64(page.PageSize))
	if total%page.PageSize > 0 {
		totalPages++
	}

	return &SerialListResponse{
		Serials:    serialModels,
		Total:      total,
		Page:       page.Page,
		PageSize:   page.PageSize,
		TotalPages: totalPages,
	}, nil
}

// validateCreateRequest 驗證建立請求
func (s *SerialService) validateCreateRequest(req *SerialCreateRequest) error {
	if req.SerialNumber == "" {
		return errors.New("序號不能為空")
	}
	if req.FullSerialNumber == "" {
		return errors.New("完整序號不能為空")
	}
	if len(req.SerialNumber) > 20 {
		return errors.New("序號長度不能超過20個字元")
	}
	if len(req.FullSerialNumber) > 100 {
		return errors.New("完整序號長度不能超過100個字元")
	}
	if req.ProductID == "" {
		return errors.New("產品ID不能為空")
	}
	return nil
}

// validateUpdateRequest 驗證更新請求
func (s *SerialService) validateUpdateRequest(req *SerialUpdateRequest) error {
	if req.SerialNumber.Valid && req.SerialNumber.String == "" {
		return errors.New("序號不能為空字串")
	}
	if req.FullSerialNumber.Valid && req.FullSerialNumber.String == "" {
		return errors.New("完整序號不能為空字串")
	}
	if req.SerialNumber.Valid && len(req.SerialNumber.String) > 20 {
		return errors.New("序號長度不能超過20個字元")
	}
	if req.FullSerialNumber.Valid && len(req.FullSerialNumber.String) > 100 {
		return errors.New("完整序號長度不能超過100個字元")
	}
	return nil
}

// validateSerialImportItem 驗證序號匯入項目
func (s *SerialService) validateSerialImportItem(item *models.SerialImportItem) error {
	if item.SerialNumber == "" {
		return errors.New("序號不能為空字串")
	}
	if item.FullSerialNumber == "" {
		return errors.New("完整序號不能為空字串")
	}
	if len(item.SerialNumber) > 20 {
		return errors.New("序號長度不能超過20個字元")
	}
	if len(item.FullSerialNumber) > 100 {
		return errors.New("完整序號長度不能超過100個字元")
	}
	if item.ProductID == "" {
		return errors.New("產品ID不能為空字串")
	}
	return nil
}

// recordAuditLog 記錄審計日誌
func (s *SerialService) recordAuditLog(ctx context.Context, auditCtx *models.AuditContext, action string, recordID string, oldData, newData interface{}) {
	// 序列化舊資料
	var oldValues []byte
	if oldData != nil {
		if oldBytes, err := json.Marshal(oldData); err == nil {
			oldValues = oldBytes
		} else {
			oldValues = []byte("{}")
		}
	} else {
		oldValues = []byte("{}")
	}

	// 序列化新資料
	var newValues []byte
	if newData != nil {
		if newBytes, err := json.Marshal(newData); err == nil {
			newValues = newBytes
		} else {
			newValues = []byte("{}")
		}
	} else {
		newValues = []byte("{}")
	}

	// 建立審計日誌請求
	auditReq := &models.CreateAuditLogRequest{
		UserID:    nil,
		Action:    models.AuditAction(action),
		TableName: "serials",
		RecordID:  recordID,
		OldValues: oldValues,
		NewValues: newValues,
		IPAddress: nil,
		UserAgent: nil,
	}

	// 如果有 audit context，使用其中的資訊
	if auditCtx != nil {
		auditReq.UserID = auditCtx.UserID
		auditReq.IPAddress = auditCtx.IPAddress
		auditReq.UserAgent = auditCtx.UserAgent
	}

	// 同步記錄到資料庫（避免競爭條件）
	if err := s.auditService.Create(ctx, auditReq); err != nil {
		// 記錄錯誤但不影響主要業務流程
		fmt.Printf("Failed to create audit log: %v\n", err)
	}
}

// generateSerialKey 生成序號key (4位數字)
func (s *SerialService) generateSerialKey(serialNumber string) string {
	h := fnv.New32a()
	h.Write([]byte(s.cfg.EncryptionKey() + serialNumber))
	return fmt.Sprintf("%04d", h.Sum32()%10_000)
}

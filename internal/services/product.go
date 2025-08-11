package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"breast-implant-warranty-system/internal/models"
	"breast-implant-warranty-system/internal/repositories"
	"breast-implant-warranty-system/internal/utils"
	"breast-implant-warranty-system/pkg/dbutil"

	"github.com/guregu/null/v5"
)

// ProductService 產品服務
type ProductService struct {
	db           dbutil.PgxClientItf
	productRepo  *repositories.ProductRepository
	auditService *AuditService
}

// NewProductService 建立新的產品服務
func NewProductService(db dbutil.PgxClientItf) *ProductService {
	return &ProductService{
		db:           db,
		productRepo:  repositories.NewProductRepository(db),
		auditService: NewAuditService(db),
	}
}

// ProductCreateRequest 產品建立請求
type ProductCreateRequest struct {
	ModelNumber   string  `json:"model_number" validate:"required,min=1,max=100"`
	Brand         string  `json:"brand" validate:"required,min=1,max=100"`
	Type          string  `json:"type" validate:"required,min=1,max=50"`
	Size          *string `json:"size,omitempty" validate:"omitempty,max=50"`
	WarrantyYears int     `json:"warranty_years" validate:"min=0,max=50"`
	Description   *string `json:"description,omitempty"`
}

// ProductUpdateRequest 產品更新請求
type ProductUpdateRequest struct {
	ModelNumber   *string `json:"model_number,omitempty" validate:"omitempty,min=1,max=100"`
	Brand         *string `json:"brand,omitempty" validate:"omitempty,min=1,max=100"`
	Type          *string `json:"type,omitempty" validate:"omitempty,min=1,max=50"`
	Size          *string `json:"size,omitempty" validate:"omitempty,max=50"`
	WarrantyYears *int    `json:"warranty_years,omitempty" validate:"omitempty,min=0,max=50"`
	Description   *string `json:"description,omitempty"`
	IsActive      *bool   `json:"is_active,omitempty"`
}

// ProductListResponse 產品列表回應
type ProductListResponse struct {
	Products   []*models.Product `json:"products"`
	Total      int               `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

// Create 建立產品
func (s *ProductService) Create(ctx context.Context, req *ProductCreateRequest, auditCtx *models.AuditContext) (*models.Product, error) {
	// 驗證輸入
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// 檢查是否重複
	duplicate, err := s.productRepo.IsDuplicate(ctx, &models.Product{
		ModelNumber: utils.SanitizeString(req.ModelNumber),
		Brand:       utils.SanitizeString(req.Brand),
		Type:        utils.SanitizeString(req.Type),
		Size:        null.StringFromPtr(req.Size),
	})
	if err != nil {
		return nil, err
	}
	if duplicate {
		return nil, errors.New("product already exists")
	}

	// 建立產品
	product := &models.Product{
		ModelNumber:   utils.SanitizeString(req.ModelNumber),
		Brand:         utils.SanitizeString(req.Brand),
		Type:          utils.SanitizeString(req.Type),
		WarrantyYears: req.WarrantyYears,
		IsActive:      true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if req.Size != nil {
		size := utils.SanitizeString(*req.Size)
		product.Size = null.StringFrom(size)
	}

	if req.Description != nil {
		desc := utils.SanitizeString(*req.Description)
		product.Description = null.StringFrom(desc)
	}

	id, err := s.productRepo.Create(ctx, product)
	if err != nil {
		return nil, err
	}

	// 記錄 audit 日誌
	s.recordAuditLog(ctx, auditCtx, models.AuditActionCreate, id, nil, product)

	return product, nil
}

// GetByID 根據ID取得產品
func (s *ProductService) GetByID(ctx context.Context, id string) (*models.Product, error) {
	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New("product not found")
	}

	return product, nil
}

// Update 更新產品
func (s *ProductService) Update(ctx context.Context, id string, req *ProductUpdateRequest, auditCtx *models.AuditContext) (*models.Product, error) {
	// 取得現有產品
	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New("product not found")
	}

	// 保存舊資料用於 audit 記錄
	oldProduct := *product

	// 更新欄位
	if req.ModelNumber != nil {
		// 檢查型號是否已被其他產品使用
		duplicate, err := s.productRepo.IsDuplicate(ctx, &models.Product{
			ID:          id,
			ModelNumber: utils.SanitizeString(*req.ModelNumber),
			Brand:       utils.SanitizeString(*req.Brand),
			Type:        utils.SanitizeString(*req.Type),
			Size:        null.StringFromPtr(req.Size),
		})
		if err != nil {
			return nil, err
		}
		if duplicate {
			return nil, errors.New("other product with same settings already exists")
		}

		product.ModelNumber = utils.SanitizeString(*req.ModelNumber)
	}

	if req.Brand != nil {
		product.Brand = utils.SanitizeString(*req.Brand)
	}

	if req.Type != nil {
		product.Type = utils.SanitizeString(*req.Type)
	}

	if req.Size != nil {
		size := utils.SanitizeString(*req.Size)
		product.Size = null.StringFrom(size)
	}

	if req.WarrantyYears != nil {
		if *req.WarrantyYears < 0 || *req.WarrantyYears > 50 {
			return nil, errors.New("warranty years must be between 0 and 50")
		}
		product.WarrantyYears = *req.WarrantyYears
	}

	if req.Description != nil {
		desc := utils.SanitizeString(*req.Description)
		product.Description = null.StringFrom(desc)
	}

	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}

	product.UpdatedAt = time.Now()

	err = s.productRepo.Update(ctx, product)
	if err != nil {
		return nil, err
	}

	// 記錄 audit 日誌
	s.recordAuditLog(ctx, auditCtx, models.AuditActionUpdate, id, &oldProduct, product)

	return product, nil
}

// Delete 刪除產品（軟刪除）
func (s *ProductService) Delete(ctx context.Context, id string, auditCtx *models.AuditContext) error {
	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if product == nil {
		return errors.New("product not found")
	}

	// 保存舊資料用於 audit 記錄
	oldProduct := *product

	err = s.productRepo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// 記錄 audit 日誌
	s.recordAuditLog(ctx, auditCtx, models.AuditActionDelete, id, &oldProduct, nil)

	return nil
}

// Search 搜尋產品
func (s *ProductService) Search(ctx context.Context, req *models.ProductSearchRequest) (*ProductListResponse, error) {
	products, totalCount, err := s.productRepo.Search(ctx, req, &req.Pagination)
	if err != nil {
		return nil, err
	}

	total := 0
	totalPages := 1
	if len(products) > 0 {
		total = totalCount
		totalPages = int(math.Ceil(float64(totalCount) / float64(req.PageSize)))
	}

	return &ProductListResponse{
		Products:   products,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetOneByCondition 根據條件取得單一產品
func (s *ProductService) GetOneByCondition(ctx context.Context, req *models.ProductSearchRequest) (*models.Product, error) {
	product, err := s.productRepo.GetOneByCondition(ctx, req)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New("product not found")
	}
	return product, nil
}

// GetMetadataList 取得產品相關資訊列表（下拉選單用）
func (s *ProductService) GetMetadataList(ctx context.Context, metatype string, req *models.ProductSearchRequest) ([]string, error) {
	switch metatype {
	case "brands":
		return s.productRepo.GetAllBrands(ctx)
	case "types":
		return s.productRepo.GetAllTypes(ctx, req)
	case "model_numbers":
		return s.productRepo.GetAllModelNumbers(ctx, req)
	case "sizes":
		return s.productRepo.GetAllSizes(ctx, req)
	default:
		return nil, fmt.Errorf("unknown metadata: %s", metatype)
	}
}

// GetMetadataAll 取得所有產品元資料（品牌、型號、類型、尺寸）
func (s *ProductService) GetMetadataAll(ctx context.Context) (*models.ProductMetadataAllResponse, error) {
	return s.productRepo.GetMetadataAll(ctx)
}

// validateCreateRequest 驗證建立產品請求
func (s *ProductService) validateCreateRequest(req *ProductCreateRequest) error {
	if req.ModelNumber == "" {
		return errors.New("model number is required")
	}
	if len(req.ModelNumber) > 100 {
		return errors.New("model number too long")
	}

	if req.Brand == "" {
		return errors.New("brand is required")
	}
	if len(req.Brand) > 100 {
		return errors.New("brand name too long")
	}

	if req.Type == "" {
		return errors.New("type is required")
	}
	if len(req.Type) > 50 {
		return errors.New("type name too long")
	}

	if req.WarrantyYears < 0 || req.WarrantyYears > 50 {
		return errors.New("warranty years must be between 0 and 50")
	}

	return nil
}

// recordAuditLog 記錄審計日誌
func (s *ProductService) recordAuditLog(ctx context.Context, auditCtx *models.AuditContext, action models.AuditAction, recordID string, oldData, newData interface{}) {
	// 序列化舊資料
	var oldValues json.RawMessage
	if oldData != nil {
		if oldBytes, err := json.Marshal(oldData); err == nil {
			oldValues = oldBytes
		} else {
			oldValues = json.RawMessage("{}")
		}
	} else {
		oldValues = json.RawMessage("{}")
	}

	// 序列化新資料
	var newValues json.RawMessage
	if newData != nil {
		if newBytes, err := json.Marshal(newData); err == nil {
			newValues = newBytes
		} else {
			newValues = json.RawMessage("{}")
		}
	} else {
		newValues = json.RawMessage("{}")
	}

	// 建立審計日誌請求
	auditReq := &models.CreateAuditLogRequest{
		UserID:    nil,
		Action:    action,
		TableName: models.AuditTableProducts,
		RecordID:  recordID,
		OldValues: oldValues,
		NewValues: newValues,
		IPAddress: nil,
		UserAgent: nil,
	}

	// 如果有 audit context，使用其中的信息
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

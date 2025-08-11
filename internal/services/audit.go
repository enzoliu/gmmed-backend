package services

import (
	"context"
	"math"

	"breast-implant-warranty-system/internal/models"
	"breast-implant-warranty-system/internal/repositories"
	"breast-implant-warranty-system/pkg/dbutil"
)

// AuditService 審計服務
type AuditService struct {
	db        dbutil.PgxClientItf
	auditRepo *repositories.AuditRepository
}

// NewAuditService 建立新的審計服務
func NewAuditService(db dbutil.PgxClientItf) *AuditService {
	return &AuditService{
		db:        db,
		auditRepo: repositories.NewAuditRepository(db),
	}
}

// Search 取得審計日誌列表
func (s *AuditService) Search(ctx context.Context, req *models.AuditSearchLogRequest) (*models.AuditLogResponse, error) {
	auditlogs, totalCount, err := s.auditRepo.List(ctx, req, &req.Pagination)
	if err != nil {
		return nil, err
	}

	total := 0
	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))
	if len(auditlogs) > 0 {
		total = totalCount
		totalPages = int(math.Ceil(float64(total) / float64(req.PageSize)))
	}

	return &models.AuditLogResponse{
		AuditLogs:  auditlogs,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetByID 根據ID取得審計日誌
func (s *AuditService) GetByID(ctx context.Context, id string) (*models.AuditLog, error) {
	return s.auditRepo.GetByID(ctx, id)
}

// Create 建立審計日誌
func (s *AuditService) Create(ctx context.Context, req *models.CreateAuditLogRequest) error {
	return s.auditRepo.Create(ctx, req)
}

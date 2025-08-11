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

	"github.com/jackc/pgx/v5"
)

// UserService 使用者服務
type UserService struct {
	db           dbutil.PgxClientItf
	userRepo     *repositories.UserRepository
	auditService *AuditService
}

// NewUserService 建立新的使用者服務
func NewUserService(db dbutil.PgxClientItf) *UserService {
	return &UserService{
		db:           db,
		userRepo:     repositories.NewUserRepository(db),
		auditService: NewAuditService(db),
	}
}

// Create 建立使用者
func (s *UserService) Create(ctx context.Context, req *models.CreateUserRequest, auditCtx *models.AuditContext) (*models.User, error) {
	// 驗證輸入
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// 檢查使用者名稱是否已存在
	existingUser, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("username already exists")
	}

	// 檢查電子信箱是否已存在
	existingUser, err = s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("email already exists")
	}

	// 雜湊密碼
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// 建立使用者
	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         string(req.Role),
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	id, err := s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}
	user.ID = id

	// 記錄 audit 日誌
	userCopy := *user
	userCopy.PasswordHash = "" // 不記錄密碼雜湊
	s.recordAuditLog(ctx, auditCtx, models.AuditActionCreate, &user.ID, nil, &userCopy)

	// 清除密碼雜湊
	user.PasswordHash = ""
	return user, nil
}

// GetByID 根據ID取得使用者
func (s *UserService) GetByID(ctx context.Context, id string) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// 清除密碼雜湊
	user.PasswordHash = ""
	return user, nil
}

// Update 更新使用者
func (s *UserService) Update(ctx context.Context, id string, req *models.UpdateUserRequest, auditCtx *models.AuditContext) (*models.User, error) {
	// 取得現有使用者
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// 保存舊資料用於 audit 記錄
	oldUser := *user
	oldUser.PasswordHash = "" // 不記錄密碼雜湊

	// 更新欄位
	if req.Username != nil {
		// 檢查使用者名稱是否已被其他使用者使用
		existingUser, err := s.userRepo.GetByUsername(ctx, *req.Username)
		if err != nil {
			return nil, err
		}
		if existingUser != nil && existingUser.ID != id {
			return nil, errors.New("username already exists")
		}
		user.Username = *req.Username
	}

	if req.Email != nil {
		// 檢查電子信箱是否已被其他使用者使用
		existingUser, err := s.userRepo.GetByEmail(ctx, *req.Email)
		if err != nil {
			return nil, err
		}
		if existingUser != nil && existingUser.ID != id {
			return nil, errors.New("email already exists")
		}
		user.Email = *req.Email
	}

	if req.Password != nil {
		// 驗證密碼強度
		if !utils.ValidatePassword(*req.Password) {
			return nil, errors.New("password does not meet security requirements")
		}
		hashedPassword, err := utils.HashPassword(*req.Password)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = hashedPassword
	}

	if req.Role != nil {
		user.Role = string(*req.Role)
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	user.UpdatedAt = time.Now()

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, err
	}

	// 記錄 audit 日誌
	newUser := *user
	newUser.PasswordHash = "" // 不記錄密碼雜湊
	s.recordAuditLog(ctx, auditCtx, models.AuditActionUpdate, &id, &oldUser, &newUser)

	// 清除密碼雜湊
	user.PasswordHash = ""
	return user, nil
}

// Delete 刪除使用者
func (s *UserService) Delete(ctx context.Context, id string, auditCtx *models.AuditContext) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// 保存舊資料用於 audit 記錄
	oldUser := *user
	oldUser.PasswordHash = "" // 不記錄密碼雜湊

	err = s.userRepo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// 記錄 audit 日誌
	s.recordAuditLog(ctx, auditCtx, models.AuditActionDelete, &id, &oldUser, nil)

	return nil
}

// List 列出使用者
func (s *UserService) Search(ctx context.Context, req *models.UserSearchRequest) (*models.UserListResponse, error) {
	users, totalCount, err := s.userRepo.Search(ctx, req, &req.Pagination)
	if err != nil {
		return nil, err
	}

	// 清除所有使用者的密碼雜湊
	for i := range users {
		users[i].PasswordHash = ""
	}
	total := 0
	totalPages := 1
	if len(users) > 0 {
		total = totalCount
		totalPages = int(math.Ceil(float64(total) / float64(req.PageSize)))
	}

	return &models.UserListResponse{
		Users:      users,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// validateCreateRequest 驗證建立使用者請求
func (s *UserService) validateCreateRequest(req *models.CreateUserRequest) error {
	if req.Username == "" {
		return errors.New("username is required")
	}
	if len(req.Username) < 3 || len(req.Username) > 50 {
		return errors.New("username must be between 3 and 50 characters")
	}

	if req.Email == "" {
		return errors.New("email is required")
	}
	if !utils.ValidateEmail(req.Email) {
		return errors.New("invalid email format")
	}

	if req.Password == "" {
		return errors.New("password is required")
	}
	if !utils.ValidatePassword(req.Password) {
		return errors.New("password does not meet security requirements")
	}

	if req.Role == "" {
		return errors.New("role is required")
	}
	if req.Role != models.RoleAdmin && req.Role != models.RoleEditor && req.Role != models.RoleReadonly {
		return errors.New("invalid role")
	}

	return nil
}

// recordAuditLog 記錄審計日誌
func (s *UserService) recordAuditLog(ctx context.Context, auditCtx *models.AuditContext, action models.AuditAction, recordID *string, oldData, newData interface{}) {
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
	if recordID == nil {
		naValue := "N/A"
		recordID = &naValue
	}

	// 建立審計日誌請求
	auditReq := &models.CreateAuditLogRequest{
		UserID:    nil,
		Action:    action,
		TableName: models.AuditTableUsers,
		RecordID:  *recordID,
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

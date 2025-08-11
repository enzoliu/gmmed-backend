package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"breast-implant-warranty-system/internal/models"
	"breast-implant-warranty-system/internal/repositories"
	"breast-implant-warranty-system/internal/utils"
	"breast-implant-warranty-system/pkg/dbutil"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
)

type AuthRouteConfigItf interface {
	JWTSecret() string
	JWTRefreshSecret() string
	JWTExpireHours() int
}

// AuthService 認證服務
type AuthService struct {
	db           dbutil.PgxClientItf
	cfg          AuthRouteConfigItf
	userRepo     *repositories.UserRepository
	auditService *AuditService
}

// NewAuthService 建立新的認證服務
func NewAuthService(db dbutil.PgxClientItf, cfg AuthRouteConfigItf) *AuthService {
	return &AuthService{
		db:           db,
		cfg:          cfg,
		userRepo:     repositories.NewUserRepository(db),
		auditService: NewAuditService(db),
	}
}

// Login 使用者登入
func (s *AuthService) Login(e echo.Context, username, password string) (int, error) {
	ctx := e.Request().Context()
	// 根據使用者名稱或電子信箱查找使用者
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return http.StatusNotFound, err
	}

	if user == nil {
		// 嘗試用電子信箱查找
		user, err = s.userRepo.GetByEmail(ctx, username)
		if err != nil {
			return http.StatusNotFound, err
		}
	}

	if user == nil {
		// 記錄失敗的登入嘗試
		failedLoginData := map[string]interface{}{
			"username": username,
			"reason":   "user_not_found",
			"success":  false,
		}
		s.recordAuditLog(ctx, nil, models.AuditActionLogin, nil, nil, failedLoginData)
		return http.StatusNotFound, errors.New("invalid credentials")
	}

	// 檢查使用者是否啟用
	if !user.IsActive {
		// 記錄停用帳戶的登入嘗試
		failedLoginData := map[string]interface{}{
			"username": username,
			"user_id":  user.ID,
			"reason":   "account_disabled",
			"success":  false,
		}
		s.recordAuditLog(ctx, &user.ID, models.AuditActionLogin, &user.ID, nil, failedLoginData)
		return http.StatusUnauthorized, errors.New("user account is disabled")
	}

	// 驗證密碼
	if !utils.CheckPassword(password, user.PasswordHash) {
		// 記錄密碼錯誤的登入嘗試
		failedLoginData := map[string]interface{}{
			"username": username,
			"user_id":  user.ID,
			"reason":   "invalid_password",
			"success":  false,
		}
		s.recordAuditLog(ctx, &user.ID, models.AuditActionLogin, &user.ID, nil, failedLoginData)
		return http.StatusUnauthorized, errors.New("invalid credentials")
	}
	// 清除密碼雜湊
	user.PasswordHash = ""

	// 生成JWT令牌
	token, expiresAt, err := utils.GenerateJWT(user, s.cfg.JWTSecret(), s.cfg.JWTExpireHours())
	if err != nil {
		return http.StatusInternalServerError, err
	}
	// 生成刷新令牌
	refreshToken, refreshExpiresAt, err := utils.GenerateJWT(user, s.cfg.JWTRefreshSecret(), s.cfg.JWTExpireHours()*60)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// 更新最後登入時間
	err = s.userRepo.UpdateLastLogin(ctx, user.ID)
	if err != nil {
		slog.Error("failed to update last login time", "user_id", user.ID, "error", err)
	}

	// 記錄登入 audit 日誌
	loginData := map[string]interface{}{
		"username":   username,
		"user_id":    user.ID,
		"login_time": expiresAt,
		"success":    true,
	}
	s.recordAuditLog(ctx, &user.ID, models.AuditActionLogin, &user.ID, nil, loginData)

	// Set access_token cookie (HttpOnly)
	e.SetCookie(&http.Cookie{
		Name:     "access_token",
		Value:    token,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteNoneMode,
		Expires:  expiresAt,
	})
	// Set refresh_token cookie (HttpOnly)
	e.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteNoneMode,
		Expires:  refreshExpiresAt,
	})

	return http.StatusOK, nil
}

// ValidateToken 驗證JWT令牌
func (s *AuthService) ValidateToken(tokenString string) (*utils.JWTClaims, error) {
	return utils.ValidateJWT(tokenString, s.cfg.JWTSecret())
}

// RefreshToken 刷新JWT令牌
func (s *AuthService) RefreshToken(c echo.Context, tokenString string) error {
	ctx := c.Request().Context()
	// 驗證當前令牌
	claims, err := utils.ValidateJWT(tokenString, s.cfg.JWTRefreshSecret())
	if err != nil {
		return err
	}

	// 取得使用者資訊
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return err
	}

	if user == nil {
		return errors.New("user not found")
	}

	if !user.IsActive {
		return errors.New("user account is disabled")
	}

	// 生成新的JWT令牌
	newToken, expiresAt, err := utils.GenerateJWT(user, s.cfg.JWTSecret(), s.cfg.JWTExpireHours())
	if err != nil {
		return err
	}

	// 更新 cookies
	c.SetCookie(&http.Cookie{
		Name:     "access_token",
		Value:    newToken,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteNoneMode,
		Expires:  expiresAt,
	})

	return nil
}

// ChangePassword 更改密碼
func (s *AuthService) ChangePassword(ctx context.Context, userID string, oldPassword, newPassword string) error {
	// 取得使用者
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if user == nil {
		return errors.New("user not found")
	}

	// 驗證舊密碼
	if !utils.CheckPassword(oldPassword, user.PasswordHash) {
		// 記錄密碼變更失敗
		failedChangeData := map[string]interface{}{
			"user_id": userID,
			"reason":  "invalid_old_password",
			"success": false,
		}
		s.recordAuditLog(ctx, &userID, models.AuditActionUpdate, &userID, nil, failedChangeData)
		return errors.New("invalid old password")
	}

	// 驗證新密碼強度
	if !utils.ValidatePassword(newPassword) {
		return errors.New("password does not meet security requirements")
	}

	// 雜湊新密碼
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// 更新密碼
	user.PasswordHash = hashedPassword
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return err
	}

	// 記錄密碼變更成功
	changeData := map[string]interface{}{
		"user_id": userID,
		"action":  "password_changed",
		"success": true,
	}
	s.recordAuditLog(ctx, &userID, models.AuditActionUpdate, &userID, nil, changeData)

	return nil
}

// GetUserByID 根據ID取得使用者資訊
func (s *AuthService) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if user == nil {
		return nil, nil
	}
	// 清除密碼雜湊
	user.PasswordHash = ""
	return user, nil
}

// recordAuditLog 記錄審計日誌
func (s *AuthService) recordAuditLog(ctx context.Context, userID *string, action models.AuditAction, recordID *string, oldData, newData interface{}) {
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
		UserID:    userID,
		Action:    action,
		TableName: models.AuditTableUsers, // 認證相關操作記錄到 users 表
		RecordID:  *recordID,
		OldValues: oldValues,
		NewValues: newValues,
		IPAddress: nil, // 在服務層無法獲取IP，由middleware處理
		UserAgent: nil, // 在服務層無法獲取UserAgent，由middleware處理
	}

	// 同步記錄到資料庫（避免競爭條件）
	if err := s.auditService.Create(ctx, auditReq); err != nil {
		// 記錄錯誤但不影響主要業務流程
		fmt.Printf("Failed to create audit log: %v\n", err)
	}
}

package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"time"

	"breast-implant-warranty-system/internal/models"
	"breast-implant-warranty-system/internal/repositories"
	"breast-implant-warranty-system/internal/utils"
	"breast-implant-warranty-system/pkg/dbutil"

	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

type WarrantyRouteConfigItf interface {
	MailgunConfigItf
	EncryptionKey() string
}

// WarrantyService 保固服務
type WarrantyService struct {
	db           dbutil.PgxClientItf
	cfg          WarrantyRouteConfigItf
	warrantyRepo *repositories.WarrantyRepository
	productRepo  *repositories.ProductRepository
	serialRepo   *repositories.SerialRepository
	emailService *EmailService
	auditService *AuditService
}

// NewWarrantyService 建立新的保固服務
func NewWarrantyService(db dbutil.PgxClientItf, cfg WarrantyRouteConfigItf) *WarrantyService {
	return &WarrantyService{
		db:           db,
		cfg:          cfg,
		warrantyRepo: repositories.NewWarrantyRepository(db, cfg.EncryptionKey()),
		productRepo:  repositories.NewProductRepository(db),
		serialRepo:   repositories.NewSerialRepository(db),
		emailService: NewEmailService(cfg),
		auditService: NewAuditService(db),
	}
}

// GetByID 根據ID取得保固登記
func (s *WarrantyService) GetByID(ctx context.Context, id string) (*models.WarrantyRegistration, error) {
	warranty, err := s.warrantyRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if warranty == nil {
		return nil, errors.New("warranty registration not found")
	}

	return warranty, nil
}

// Update 更新保固登記 (管理者更新)
func (s *WarrantyService) Update(ctx context.Context, id string, req *models.WarrantyUpdateRequest, auditCtx *models.AuditContext) (*models.WarrantyRegistration, error) {
	// 取得現有保固登記
	warranty, err := s.warrantyRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if warranty == nil {
		return nil, errors.New("未知的保固記錄")
	}

	// 保存舊資料用於 audit 記錄
	oldWarranty := *warranty

	// 病患姓名
	warranty.PatientName = null.StringFrom(utils.SanitizeString(req.PatientName))

	// 病患身分證字號加密
	encryptedID, err := utils.EncryptPatientID(req.PatientID, s.cfg.EncryptionKey())
	if err != nil {
		return nil, err
	}
	warranty.PatientIDEncrypted = null.StringFrom(encryptedID)

	// 出生日期
	warranty.PatientBirthDate = null.TimeFrom(req.PatientBirthDate)

	// 手機號碼
	encryptedPhone, err := utils.EncryptPatientPhone(req.PatientPhone, s.cfg.EncryptionKey())
	if err != nil {
		return nil, err
	}
	warranty.PatientPhoneEncrypted = null.StringFrom(encryptedPhone)

	// 電子郵件
	warranty.PatientEmail = null.StringFrom(utils.SanitizeString(req.PatientEmail))

	// 醫院名稱
	warranty.HospitalName = null.StringFrom(utils.SanitizeString(req.HospitalName))

	// 醫師姓名
	warranty.DoctorName = null.StringFrom(utils.SanitizeString(req.DoctorName))

	// 保固狀態
	warranty.Status = null.StringFrom(string(req.Status))

	// 手術日期
	warranty.SurgeryDate = null.TimeFrom(req.SurgeryDate)

	// 產品ID
	warranty.ProductID = null.StringFrom(req.ProductID)

	// 產品序號1
	if req.ProductSerialNumber == "" {
		return nil, errors.New("產品序號1是必填的")
	}
	// 檢查是否已被使用
	if req.ProductSerialNumber != oldWarranty.ProductSerialNumber.String && req.ProductSerialNumber != oldWarranty.ProductSerialNumber2.String {
		exists, _ := s.CheckSerialNumberExists(ctx, req.ProductSerialNumber)
		if exists {
			return nil, errors.New("產品序號1已被使用")
		}
	}

	warranty.ProductSerialNumber = null.StringFrom(utils.SanitizeString(req.ProductSerialNumber))

	// 產品序號2
	if req.ProductSerialNumber2 != "" {
		if req.ProductSerialNumber2 == req.ProductSerialNumber {
			return nil, errors.New("產品序號2不能與產品序號1相同")
		}
		// 檢查是否已被使用，如果是同一份保固更新前使用過的序號，則不檢查
		if req.ProductSerialNumber2 != oldWarranty.ProductSerialNumber2.String && req.ProductSerialNumber2 != oldWarranty.ProductSerialNumber.String {
			exists, _ := s.CheckSerialNumberExists(ctx, req.ProductSerialNumber2)
			if exists {
				return nil, errors.New("產品序號2已被使用")
			}
		}
		warranty.ProductSerialNumber2 = null.StringFrom(utils.SanitizeString(req.ProductSerialNumber2))
	} else {
		warranty.ProductSerialNumber2 = null.StringFrom("")
	}

	// 重新計算保固期間
	// 如果 warranty.Product 為 nil，需要取得產品資訊
	if !warranty.NullableProduct.Brand.Valid && warranty.ProductID.Valid {
		product, err := s.productRepo.GetByID(ctx, warranty.ProductID.String)
		if err != nil {
			return nil, err
		}
		if !product.IsActive {
			return nil, errors.New("產品已停用")
		}
		warranty.NullableProduct = models.NullableProduct{
			ModelNumber:   null.StringFrom(product.ModelNumber),
			Brand:         null.StringFrom(product.Brand),
			Type:          null.StringFrom(product.Type),
			Size:          product.Size,
			WarrantyYears: null.IntFrom(int64(product.WarrantyYears)),
			Description:   product.Description,
			IsActive:      null.BoolFrom(product.IsActive),
		}
	}

	if warranty.NullableProduct.Brand.Valid {
		// 保固起始日期設為手術日期（確保滿足約束 surgery_date <= warranty_start_date）
		warranty.WarrantyStartDate = null.TimeFrom(req.SurgeryDate)
		// 如果 warranty_years 為 -1，表示終身保固
		switch warranty.NullableProduct.WarrantyYears.Int64 {
		case -1:
			warranty.WarrantyEndDate = null.TimeFrom(time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC))
		case 0:
			// 無保固 - 設定為手術當天結束
			warranty.WarrantyEndDate = warranty.WarrantyStartDate
		default:
			warranty.WarrantyEndDate = null.TimeFrom(req.SurgeryDate.AddDate(int(warranty.NullableProduct.WarrantyYears.Int64), 0, 0))
		}
	}

	if warranty.WarrantyEndDate.Valid {
		if warranty.WarrantyEndDate.Time.Before(time.Now()) {
			warranty.Status = null.StringFrom(string(models.StatusExpired))
		}
	}

	warranty.UpdatedAt = time.Now()

	err = s.warrantyRepo.Update(ctx, warranty)
	if err != nil {
		return nil, err
	}

	// 記錄 audit 日誌
	s.recordAuditLog(ctx, auditCtx, models.AuditActionUpdate, &id, &oldWarranty, warranty)

	return warranty, nil
}

// Delete 刪除保固登記
func (s *WarrantyService) Delete(ctx context.Context, id string, auditCtx *models.AuditContext) error {
	warranty, err := s.warrantyRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if warranty == nil {
		return errors.New("warranty registration not found")
	}

	// 保存舊資料用於 audit 記錄
	oldWarranty := *warranty

	err = s.warrantyRepo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// 記錄 audit 日誌
	s.recordAuditLog(ctx, auditCtx, models.AuditActionDelete, &id, &oldWarranty, nil)

	return nil
}

// Search 搜尋保固登記
func (s *WarrantyService) Search(ctx context.Context, req *models.WarrantySearchRequest) (*models.WarrantyListResponse, error) {
	warranties, total, err := s.warrantyRepo.Search(ctx, req, &req.Pagination)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))

	return &models.WarrantyListResponse{
		Warranties: warranties,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetStatistics 取得保固統計
func (s *WarrantyService) GetStatistics(ctx context.Context) (*models.WarrantyStatistics, error) {
	return s.warrantyRepo.GetStatistics(ctx)
}

// ResendConfirmationEmail 重新發送確認信件
func (s *WarrantyService) ResendConfirmationEmail(ctx context.Context, id string, auditCtx *models.AuditContext) error {
	warranty, err := s.warrantyRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if warranty == nil {
		return errors.New("warranty registration not found")
	}

	// 發送確認信件
	err = s.sendConfirmationEmail(ctx, warranty)
	if err != nil {
		return err
	}

	// 更新信件發送狀態
	err = s.warrantyRepo.UpdateEmailSent(ctx, id, true)
	if err != nil {
		return err
	}

	// 記錄 audit 日誌
	var patientEmail string
	if warranty.PatientEmail.Valid {
		patientEmail = warranty.PatientEmail.String
	}
	resendData := map[string]interface{}{
		"warranty_id":   id,
		"patient_email": patientEmail,
		"action":        "resend_confirmation_email",
		"success":       true,
	}
	s.recordAuditLog(ctx, auditCtx, models.AuditActionUpdate, &id, nil, resendData)

	return nil
}

// UpdateExpiredWarranties 更新所有過期的保固狀態
func (s *WarrantyService) UpdateExpiredWarranties(ctx context.Context, auditCtx *models.AuditContext) (int, error) {
	updatedCount, err := s.warrantyRepo.UpdateExpiredWarranties(ctx)
	if err != nil {
		return 0, err
	}

	// 記錄 audit 日誌
	updateData := map[string]interface{}{
		"action":        "batch_update_expired_warranties",
		"updated_count": updatedCount,
		"success":       true,
	}
	s.recordAuditLog(ctx, auditCtx, models.AuditActionUpdate, nil, nil, updateData)

	return updatedCount, nil
}

// recordAuditLog 記錄審計日誌
func (s *WarrantyService) recordAuditLog(ctx context.Context, auditCtx *models.AuditContext, action models.AuditAction, recordID *string, oldData, newData interface{}) {
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
		TableName: models.AuditTableWarrantyRegistrations,
		RecordID:  *recordID,
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

// CheckSerialNumberExists 檢查產品序號是否已被使用
func (s *WarrantyService) CheckSerialNumberExists(ctx context.Context, serialNumber string) (bool, error) {
	if serialNumber == "" {
		return false, errors.New("serial number is required")
	}
	re := regexp.MustCompile(`^\d{7}-\d{3}$`)
	if !re.MatchString(serialNumber) {
		return false, errors.New("serial number must be in the format XXXXXXX-XXX")
	}

	warranty, err := s.warrantyRepo.GetByProductSerialNumber(ctx, serialNumber)
	if err != nil && err == pgx.ErrNoRows {
		return false, nil
	}

	if warranty != nil {
		return true, nil
	}

	return false, nil
}

// BatchCreateEmptyWarranties 批次創建空白保固記錄
func (s *WarrantyService) BatchCreateEmptyWarranties(ctx context.Context, count int, auditCtx *models.AuditContext) ([]string, error) {
	if count <= 0 || count > 100 {
		return nil, errors.New("count must be between 1 and 100")
	}

	ids, err := s.warrantyRepo.BatchCreateEmpty(ctx, count)
	if err != nil {
		return nil, err
	}

	// 記錄 audit 日誌
	batchData := map[string]interface{}{
		"count": count,
		"ids":   ids,
	}
	s.recordAuditLog(ctx, auditCtx, models.AuditActionCreate, nil, nil, batchData)

	return ids, nil
}

func (s *WarrantyService) GetWarrantyStatusByPatient(ctx context.Context, id string) (int, error) {
	warranty, err := s.warrantyRepo.GetByID(ctx, id)
	if err != nil {
		return 0, err
	}
	if warranty == nil {
		return 0, errors.New("warranty not found")
	}
	return warranty.Step, nil
}

func (s *WarrantyService) RegisterByPatientStep1(ctx context.Context, id string, req *models.PatientRegistrationRequestStep1, auditCtx *models.AuditContext) (*models.WarrantyRegistration, error) {
	// 檢查保固是否存在且可編輯
	warranty, err := s.warrantyRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if warranty == nil {
		return nil, errors.New("warranty not found")
	}

	// 檢查當前保固是否可填寫
	if warranty.Step != models.STEP_BLANK {
		return nil, errors.New("warranty has already been filled")
	}

	// 保存舊資料用於 audit 記錄
	oldWarranty := *warranty

	// 驗證產品序號不重複
	exists, err := s.CheckSerialNumberExists(ctx, req.ProductSerialNumber)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("product serial number already registered")
	}
	serial, err := s.serialRepo.ExistsBySerialNumber(ctx, req.ProductSerialNumber)
	if err != nil {
		return nil, err
	}
	if !serial {
		return nil, errors.New("product serial number not valid")
	}

	// 如果有第二個序號，也要檢查
	if req.ProductSerialNumber2 != "" {
		if req.ProductSerialNumber2 == req.ProductSerialNumber {
			return nil, errors.New("two serial numbers cannot be the same")
		}
		exists, err := s.CheckSerialNumberExists(ctx, req.ProductSerialNumber2)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("second product serial number already registered")
		}
		serial, err = s.serialRepo.ExistsBySerialNumber(ctx, req.ProductSerialNumber2)
		if err != nil {
			return nil, err
		}
		if !serial {
			return nil, errors.New("product serial number not valid")
		}
	}

	// 獲取產品資訊
	product, err := s.productRepo.GetByID(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New("product not found")
	}
	if !product.IsActive {
		return nil, errors.New("product is not active")
	}

	// 驗證手術日期不能在未來
	if req.SurgeryDate.Time.After(time.Now()) {
		return nil, errors.New("surgery date cannot be in the future")
	}

	// 計算保固期間
	warranty.WarrantyStartDate = null.TimeFrom(req.SurgeryDate.Time)
	warranty.Step = models.STEP_SERIAL_VERIFIED
	switch product.WarrantyYears {
	case -1:
		// 終身保固
		warranty.WarrantyEndDate = null.TimeFrom(time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC))
	case 0:
		// 無保固 - 設定為手術當天結束
		warranty.WarrantyEndDate = warranty.WarrantyStartDate
		warranty.Step = models.STEP_VERIFIED_WITHOUT_WARRANTY
	default:
		// 有期限保固
		warranty.WarrantyEndDate = null.TimeFrom(req.SurgeryDate.Time.AddDate(int(warranty.NullableProduct.WarrantyYears.Int64), 0, 0))
	}

	warranty.SurgeryDate = null.TimeFrom(req.SurgeryDate.Time)
	warranty.ProductID = null.StringFrom(req.ProductID)
	warranty.ProductSerialNumber = null.StringFrom(utils.SanitizeString(req.ProductSerialNumber))
	if req.ProductSerialNumber2 != "" {
		serialNumber2 := utils.SanitizeString(req.ProductSerialNumber2)
		warranty.ProductSerialNumber2 = null.StringFrom(serialNumber2)
	}
	warranty.NullableProduct = models.NullableProduct{
		ModelNumber:   null.StringFrom(product.ModelNumber),
		Brand:         null.StringFrom(product.Brand),
		Type:          null.StringFrom(product.Type),
		Size:          product.Size,
		WarrantyYears: null.IntFrom(int64(product.WarrantyYears)),
		Description:   product.Description,
		IsActive:      null.BoolFrom(product.IsActive),
	}

	// 更新保固記錄
	err = s.warrantyRepo.Update(ctx, warranty)
	if err != nil {
		return nil, err
	}
	// 記錄 audit 日誌
	s.recordAuditLog(ctx, auditCtx, models.AuditActionUpdate, &id, &oldWarranty, warranty)

	return warranty, nil
}

func (s *WarrantyService) RegisterByPatientStep2(ctx context.Context, id string, req *models.PatientRegistrationRequestStep2, auditCtx *models.AuditContext) (*models.WarrantyRegistration, error) {
	// 檢查保固是否存在且可編輯
	warranty, err := s.warrantyRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if warranty == nil {
		return nil, errors.New("warranty not found")
	}

	// 檢查當前保固是否可填寫
	if warranty.Step != models.STEP_SERIAL_VERIFIED && warranty.Step != models.STEP_PATIENT_INFO_FILLED {
		return nil, errors.New("warranty can not be filled")
	}

	// 保存舊資料用於 audit 記錄
	oldWarranty := *warranty

	// 加密敏感資料
	encryptedID, err := utils.EncryptPatientID(req.PatientID, s.cfg.EncryptionKey())
	if err != nil {
		return nil, err
	}

	encryptedPhone, err := utils.EncryptPatientPhone(req.PatientPhone, s.cfg.EncryptionKey())
	if err != nil {
		return nil, err
	}

	// 填寫保固資料
	warranty.PatientName = null.StringFrom(utils.SanitizeString(req.PatientName))
	warranty.PatientIDEncrypted = null.StringFrom(encryptedID)
	warranty.PatientBirthDate = null.TimeFrom(req.PatientBirthDate.Time)
	warranty.PatientPhoneEncrypted = null.StringFrom(encryptedPhone)
	warranty.PatientEmail = null.StringFrom(utils.SanitizeString(req.PatientEmail))
	warranty.HospitalName = null.StringFrom(utils.SanitizeString(req.HospitalName))
	warranty.DoctorName = null.StringFrom(utils.SanitizeString(req.DoctorName))
	warranty.Step = models.STEP_PATIENT_INFO_FILLED

	// 更新保固記錄
	err = s.warrantyRepo.Update(ctx, warranty)
	if err != nil {
		return nil, err
	}

	// 記錄 audit 日誌
	s.recordAuditLog(ctx, auditCtx, models.AuditActionUpdate, &id, &oldWarranty, warranty)

	return warranty, nil
}

func (s *WarrantyService) RegisterByPatientStep3(ctx context.Context, id string, auditCtx *models.AuditContext) (*models.WarrantyRegistration, error) {
	// 檢查保固是否存在且可編輯
	warranty, err := s.warrantyRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if warranty == nil {
		return nil, errors.New("warranty not found")
	}

	// 檢查當前保固是否可填寫
	if warranty.Step != models.STEP_PATIENT_INFO_FILLED {
		return nil, errors.New("warranty can not be confirmed")
	}

	// 保存舊資料用於 audit 記錄
	oldWarranty := *warranty

	warranty.Step = models.STEP_WARRANTY_ESTABLISHED
	warranty.Status = null.StringFrom(string(models.StatusActive))

	// 更新保固記錄
	err = s.warrantyRepo.Update(ctx, warranty)
	if err != nil {
		return nil, err
	}

	// 發送確認信件，忽略錯誤
	_ = s.sendConfirmationEmail(ctx, warranty)

	// 記錄 audit 日誌
	s.recordAuditLog(ctx, auditCtx, models.AuditActionUpdate, &id, &oldWarranty, warranty)

	return warranty, nil
}

func (s *WarrantyService) GetWarrantyByPatientInSteps(ctx context.Context, id string) (*models.WarrantyRegistration, error) {
	// 檢查保固是否存在
	warranty, err := s.warrantyRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if warranty == nil {
		return nil, errors.New("warranty not found")
	}

	// 檢查當前保固是否可填寫
	if warranty.Step != models.STEP_SERIAL_VERIFIED && warranty.Step != models.STEP_PATIENT_INFO_FILLED {
		return nil, errors.New("warranty can not be filled")
	}

	// 移除機敏資訊
	warranty.PatientIDEncrypted = null.StringFrom("")
	warranty.PatientPhoneEncrypted = null.StringFrom("")
	warranty.PatientID = null.StringFrom("")
	warranty.PatientPhone = null.StringFrom("")

	return warranty, nil
}

// RegisterByPatient 患者填寫保固（一次性）
func (s *WarrantyService) RegisterByPatient(ctx context.Context, id string, req *models.PatientRegistrationRequest, auditCtx *models.AuditContext) (*models.WarrantyRegistration, error) {
	// 檢查保固是否存在且可編輯
	warranty, err := s.warrantyRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if warranty == nil {
		return nil, errors.New("warranty not found")
	}
	if !warranty.CreatedAt.Equal(warranty.UpdatedAt) {
		return nil, errors.New("warranty has already been filled")
	}

	// 保存舊資料用於 audit 記錄
	oldWarranty := *warranty

	// 驗證產品序號不重複
	exists, err := s.CheckSerialNumberExists(ctx, req.ProductSerialNumber)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("product serial number already registered")
	}
	serial, err := s.serialRepo.ExistsBySerialNumber(ctx, req.ProductSerialNumber)
	if err != nil {
		return nil, err
	}
	if !serial {
		return nil, errors.New("product serial number not valid")
	}

	// 如果有第二個序號，也要檢查
	if req.ProductSerialNumber2 != "" {
		if req.ProductSerialNumber2 == req.ProductSerialNumber {
			return nil, errors.New("two serial numbers cannot be the same")
		}
		exists, err := s.CheckSerialNumberExists(ctx, req.ProductSerialNumber2)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("second product serial number already registered")
		}
	}

	// 獲取產品資訊
	product, err := s.productRepo.GetByID(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New("product not found")
	}
	if !product.IsActive {
		return nil, errors.New("產品已停用")
	}

	// 加密敏感資料
	encryptedID, err := utils.EncryptPatientID(req.PatientID, s.cfg.EncryptionKey())
	if err != nil {
		return nil, err
	}

	encryptedPhone, err := utils.EncryptPatientPhone(req.PatientPhone, s.cfg.EncryptionKey())
	if err != nil {
		return nil, err
	}

	// 填寫保固資料
	warranty.PatientName = null.StringFrom(utils.SanitizeString(req.PatientName))
	warranty.PatientIDEncrypted = null.StringFrom(encryptedID)
	warranty.PatientBirthDate = null.TimeFrom(req.PatientBirthDate.Time)
	warranty.PatientPhoneEncrypted = null.StringFrom(encryptedPhone)
	warranty.PatientEmail = null.StringFrom(utils.SanitizeString(req.PatientEmail))
	warranty.HospitalName = null.StringFrom(utils.SanitizeString(req.HospitalName))
	warranty.DoctorName = null.StringFrom(utils.SanitizeString(req.DoctorName))
	warranty.SurgeryDate = null.TimeFrom(req.SurgeryDate.Time)
	warranty.ProductID = null.StringFrom(req.ProductID)
	warranty.ProductSerialNumber = null.StringFrom(utils.SanitizeString(req.ProductSerialNumber))
	if req.ProductSerialNumber2 != "" {
		serialNumber2 := utils.SanitizeString(req.ProductSerialNumber2)
		warranty.ProductSerialNumber2 = null.StringFrom(serialNumber2)
	}

	// 驗證手術日期不能在未來
	if req.SurgeryDate.Time.After(time.Now()) {
		return nil, errors.New("surgery date cannot be in the future")
	}

	// 計算保固期間
	warranty.WarrantyStartDate = warranty.SurgeryDate
	switch product.WarrantyYears {
	case -1:
		// 終身保固
		warranty.WarrantyEndDate = null.TimeFrom(time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC))
	case 0:
		// 無保固 - 設定為手術當天結束
		warranty.WarrantyEndDate = null.TimeFrom(req.SurgeryDate.Time)
	default:
		// 有期限保固
		warranty.WarrantyEndDate = null.TimeFrom(req.SurgeryDate.Time.AddDate(product.WarrantyYears, 0, 0))
	}

	warranty.Status = null.StringFrom(string(models.StatusActive))
	warranty.NullableProduct = models.NullableProduct{
		ModelNumber:   null.StringFrom(product.ModelNumber),
		Brand:         null.StringFrom(product.Brand),
		Type:          null.StringFrom(product.Type),
		Size:          product.Size,
		WarrantyYears: null.IntFrom(int64(product.WarrantyYears)),
		Description:   product.Description,
		IsActive:      null.BoolFrom(product.IsActive),
	}

	// 更新保固記錄
	err = s.warrantyRepo.Update(ctx, warranty)
	if err != nil {
		return nil, err
	}

	// 發送確認信件，忽略錯誤
	_ = s.sendConfirmationEmail(ctx, warranty)

	// 記錄 audit 日誌
	s.recordAuditLog(ctx, auditCtx, models.AuditActionUpdate, &id, &oldWarranty, warranty)

	return warranty, nil
}

// sendConfirmationEmail 發送確認信件
func (s *WarrantyService) sendConfirmationEmail(ctx context.Context, warranty *models.WarrantyRegistration) error {
	// 檢查 Mailgun 設定
	if s.cfg.MailgunDomain() == "" || s.cfg.MailgunAPIKey() == "" {
		logrus.Warn("Mailgun not configured, skipping email sending")
		return nil
	}

	// 發送患者確認信件
	err := s.emailService.SendWarrantyConfirmation(warranty)
	if err != nil {
		logrus.WithError(err).WithField("warranty_id", warranty.ID).Error("Failed to send warranty confirmation email")
		return err
	}

	// 更新信件發送狀態和時間戳
	err = s.warrantyRepo.UpdateEmailSent(ctx, warranty.ID, true)
	if err != nil {
		logrus.WithError(err).WithField("warranty_id", warranty.ID).Error("Failed to update email sent status")
		// 不返回錯誤，因為信件已經發送成功
	}

	// 發送公司通知信件（異步，不影響主流程）
	go func() {
		if err := s.emailService.SendNotificationToCompany(warranty); err != nil {
			logrus.WithError(err).WithField("warranty_id", warranty.ID).Error("Failed to send company notification email")
		}
	}()

	logrus.WithField("warranty_id", warranty.ID).Info("Warranty confirmation email sent successfully")
	return nil
}

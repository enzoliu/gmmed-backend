package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/guregu/null/v5"
)

const (
	SurgeryDateFormat          = "2006-01-02T15:04:05" // 手術日期格式
	ProductSerialNumberPattern = `^\d{7}-\d{3}$`       // 產品序號格式：7位數字-3位數字
)

// WarrantyRegistration 保固登記模型
type WarrantyRegistration struct {
	ID                    string      `json:"id" db:"id"`
	PatientName           null.String `json:"patient_name,omitempty" db:"patient_name"`
	PatientIDEncrypted    null.String `json:"-" db:"patient_id_encrypted"` // 加密的身分證字號
	PatientID             null.String `json:"patient_id,omitempty" db:"-"` // 解密後的身分證字號（僅用於回應）
	PatientBirthDate      null.Time   `json:"patient_birth_date,omitempty" db:"patient_birth_date"`
	PatientPhoneEncrypted null.String `json:"-" db:"patient_phone_encrypted"` // 加密的手機號碼
	PatientPhone          null.String `json:"patient_phone,omitempty" db:"-"` // 解密後的手機號碼（僅用於回應）
	PatientEmail          null.String `json:"patient_email,omitempty" db:"patient_email"`
	HospitalName          null.String `json:"hospital_name,omitempty" db:"hospital_name"`
	DoctorName            null.String `json:"doctor_name,omitempty" db:"doctor_name"`
	SurgeryDate           null.Time   `json:"surgery_date,omitempty" db:"surgery_date"`
	ProductID             null.String `json:"product_id,omitempty" db:"product_id"`
	ProductSerialNumber   null.String `json:"product_serial_number,omitempty" db:"product_serial_number"`
	ProductSerialNumber2  null.String `json:"product_serial_number_2,omitempty" db:"serial_number_2"`
	WarrantyStartDate     null.Time   `json:"warranty_start_date,omitempty" db:"warranty_start_date"`
	WarrantyEndDate       null.Time   `json:"warranty_end_date,omitempty" db:"warranty_end_date"`
	ConfirmationEmailSent null.Bool   `json:"confirmation_email_sent" db:"confirmation_email_sent"`
	EmailSentAt           null.Time   `json:"email_sent_at" db:"email_sent_at"`
	Status                null.String `json:"status" db:"status"`
	CreatedAt             time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time   `json:"updated_at" db:"updated_at"`

	// 關聯資料（用於查詢時填充）
	NullableProduct
}

// IsExpired 檢查保固是否已過期
func (w *WarrantyRegistration) IsExpired() bool {
	// 如果沒有保固結束日期，視為未過期（空白記錄）
	if !w.WarrantyEndDate.Valid {
		return false
	}

	// 終身保固永不過期
	if w.NullableProduct.IsLifetimeWarranty() {
		return false
	}

	// 檢查保固結束日期，如果未設置，視為未過期
	if !w.WarrantyEndDate.Valid {
		return false
	}

	return time.Now().After(w.WarrantyEndDate.Time)
}

// GetCurrentStatus 取得當前應有的狀態
func (w *WarrantyRegistration) GetCurrentStatus() string {
	// 如果狀態未設置，視為未設定
	if !w.Status.Valid {
		return string(StatusUnset)
	}

	// 如果已經是取消狀態，保持不變
	if w.Status.String == string(StatusCancelled) {
		return w.Status.String
	}

	// 檢查是否已過期
	if w.IsExpired() {
		return string(StatusExpired)
	}

	return string(StatusActive)
}

// WarrantyStatus 保固狀態常數
type WarrantyStatus string

const (
	StatusActive    WarrantyStatus = "active"
	StatusExpired   WarrantyStatus = "expired"
	StatusCancelled WarrantyStatus = "cancelled"
	StatusUnset     WarrantyStatus = "unset"
)

// WarrantyStatistics 保固統計
type WarrantyStatistics struct {
	TotalRegistrations   int                    `json:"total_registrations"`
	ActiveWarranties     int                    `json:"active_warranties"`
	ExpiredWarranties    int                    `json:"expired_warranties"`
	ExpiringSoon         int                    `json:"expiring_soon"` // 30天內到期
	HospitalStats        []*HospitalStatistic   `json:"hospital_stats"`
	ProductStats         []*ProductStatistic    `json:"product_stats"`
	MonthlyRegistrations []*MonthlyRegistration `json:"monthly_registrations"`
}

// HospitalStatistic 醫院統計
type HospitalStatistic struct {
	HospitalName        string `json:"hospital_name"`
	TotalRegistrations  int    `json:"total_registrations"`
	ActiveRegistrations int    `json:"active_registrations"`
	TotalDoctors        int    `json:"total_doctors"`
}

// ProductStatistic 產品統計
type ProductStatistic struct {
	ProductID   uuid.UUID `json:"product_id"`
	ModelNumber string    `json:"model_number"`
	Brand       string    `json:"brand"`
	TotalUsage  int       `json:"total_usage"`
	ActiveUsage int       `json:"active_usage"`
}

// MonthlyRegistration 月度登記統計
type MonthlyRegistration struct {
	Year  int `json:"year"`
	Month int `json:"month"`
	Count int `json:"count"`
}

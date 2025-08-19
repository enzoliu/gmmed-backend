package models

import (
	"time"

	"github.com/guregu/null/v5"
)

const (
	SurgeryDateFormat              = "2006-01-02T15:04:05" // 手術日期格式
	ProductSerialNumberPattern     = `^\d{7}-\d{3}$`       // 產品序號格式：7位數字-3位數字
	STEP_BLANK                     = 0
	STEP_SERIAL_VERIFIED           = 1
	STEP_PATIENT_INFO_FILLED       = 2
	STEP_WARRANTY_ESTABLISHED      = 3
	STEP_VERIFIED_WITHOUT_WARRANTY = 9
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
	ProductSerialNumber   null.String `json:"product_serial_number,omitempty" db:"product_serial_number"`
	ProductSerialNumber2  null.String `json:"product_serial_number_2,omitempty" db:"serial_number_2"`
	WarrantyStartDate     null.Time   `json:"warranty_start_date,omitempty" db:"warranty_start_date"`
	WarrantyEndDate       null.Time   `json:"warranty_end_date,omitempty" db:"warranty_end_date"`
	ConfirmationEmailSent null.Bool   `json:"confirmation_email_sent" db:"confirmation_email_sent"`
	EmailSentAt           null.Time   `json:"email_sent_at" db:"email_sent_at"`
	Status                null.String `json:"status" db:"status"`
	CreatedAt             time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time   `json:"updated_at" db:"updated_at"`
	Step                  int         `json:"step,omitempty" db:"step"`
}

// WarrantyStatus 保固狀態常數
type WarrantyStatus string

const (
	StatusActive    WarrantyStatus = "active"
	StatusExpired   WarrantyStatus = "expired"
	StatusCancelled WarrantyStatus = "cancelled"
	StatusUnset     WarrantyStatus = "unset"
)

package models

import (
	"breast-implant-warranty-system/internal/entity"
	"breast-implant-warranty-system/pkg/validator"
	"regexp"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/guregu/null/v5"
)

// WarrantyUpdateRequest 保固更新請求
type WarrantyUpdateRequest struct {
	PatientName          string         `json:"patient_name" validate:"required,min=2,max=100"`
	PatientID            string         `json:"patient_id" validate:"required,len=10"`
	PatientBirthDate     time.Time      `json:"patient_birth_date" validate:"required"`
	PatientPhone         string         `json:"patient_phone" validate:"required,min=10,max=15"`
	PatientEmail         string         `json:"patient_email" validate:"required,email"`
	HospitalName         string         `json:"hospital_name" validate:"required,min=2,max=200"`
	DoctorName           string         `json:"doctor_name" validate:"required,min=2,max=100"`
	SurgeryDate          time.Time      `json:"surgery_date" validate:"required"`
	ProductID            string         `json:"product_id" validate:"required,uuid"`
	ProductSerialNumber  string         `json:"product_serial_number" validate:"required,min=6,max=50"`
	ProductSerialNumber2 string         `json:"product_serial_number_2,omitempty" validate:"omitempty,min=6,max=50"`
	Status               WarrantyStatus `json:"status" validate:"required"`
}

// WarrantyListResponse 保固列表回應
type WarrantyListResponse struct {
	Warranties []*WarrantyRegistration `json:"warranties"`
	Total      int                     `json:"total"`
	Page       int                     `json:"page"`
	PageSize   int                     `json:"page_size"`
	TotalPages int                     `json:"total_pages"`
}

// WarrantySearchRequest 保固搜尋請求
type WarrantySearchRequest struct {
	GeneralSearch null.String `query:"q"` // 通用搜尋（患者姓名、序號、身分證等）
	PatientName   null.String `query:"patient_name"`
	SerialNumber  null.String `query:"serial_number"`
	HospitalName  null.String `query:"hospital_name"`
	DoctorName    null.String `query:"doctor_name"`
	Status        null.String `query:"status"`
	StartDate     null.String `query:"start_date"` // YYYY-MM-DD
	EndDate       null.String `query:"end_date"`   // YYYY-MM-DD
	SearchDeleted null.Bool   `query:"search_deleted"`
	entity.Pagination
}

// SerialNumberCheckResponse 序號檢查回應
type SerialNumberCheckResponse struct {
	Exists  bool   `json:"exists"`
	Message string `json:"message"`
}

// BatchCreateRequest 批次創建保固請求
type BatchCreateRequest struct {
	Count int `json:"count" validate:"required,min=1,max=100"`
}

// BatchCreateResponse 批次創建保固回應
type BatchCreateResponse struct {
	Count int      `json:"count"`
	IDs   []string `json:"ids"`
}

// WarrantyStatusResponse 保固狀態檢查回應
type WarrantyStatusResponse struct {
	CanEdit bool   `json:"can_edit"`
	Message string `json:"message"`
}

// PatientRegistrationRequest 患者填寫保固請求
type PatientRegistrationRequest struct {
	PatientName          string      `json:"patient_name"`
	IsLocalIdentity      bool        `json:"is_local_identity"`
	PatientID            string      `json:"patient_id"`
	PatientBirthDate     GoTimeSucks `json:"patient_birth_date"`
	PatientPhone         string      `json:"patient_phone"`
	PatientEmail         string      `json:"patient_email"`
	HospitalName         string      `json:"hospital_name"`
	DoctorName           string      `json:"doctor_name"`
	SurgeryDate          GoTimeSucks `json:"surgery_date"`
	ProductID            string      `json:"product_id"`
	ProductSerialNumber  string      `json:"product_serial_number"`
	ProductSerialNumber2 string      `json:"product_serial_number_2,omitempty"`
}

func (req *PatientRegistrationRequest) Validate() error {
	return validation.ValidateStruct(req,
		validation.Field(
			&req.PatientName,
			validation.Required.Error("患者姓名是必填項"),
			validation.Length(2, 100).Error("患者姓名長度必須在2到100個字之間"),
		),
		validation.Field(
			&req.PatientID,
			validation.Required.Error("患者身分證號是必填項"),
			validation.By(func(value interface{}) error {
				if req.IsLocalIdentity {
					return validator.IsValidTaiwanID(value)
				}
				// 不檢查護照號碼
				return nil
			}),
		),
		validation.Field(
			&req.PatientBirthDate,
			validation.Required.Error("患者出生日期是必填項"),
		),
		validation.Field(
			&req.PatientPhone,
			validation.Required.Error("患者電話是必填項"),
		),
		validation.Field(
			&req.PatientEmail,
			validation.Required.Error("患者電子信箱是必填項"),
			is.Email.Error("患者電子信箱格式不正確"),
		),
		validation.Field(
			&req.HospitalName,
			validation.Required.Error("醫院名稱是必填項"),
			validation.Length(2, 200).Error("醫院名稱長度必須在2到200個字之間"),
		),
		validation.Field(
			&req.DoctorName,
			validation.Required.Error("醫生姓名是必填項"),
			validation.Length(2, 100).Error("醫生姓名長度必須在2到100個字之間"),
		),
		validation.Field(
			&req.SurgeryDate,
			validation.Required.Error("手術日期是必填項"),
		),
		validation.Field(
			&req.ProductID,
			validation.Required.Error("必填選擇產品"),
			is.UUID.Error("產品ID格式不正確"),
		),
		validation.Field(
			&req.ProductSerialNumber,
			validation.Required.Error("產品序號是必填項"),
			validation.Match(regexp.MustCompile(ProductSerialNumberPattern)).Error("產品序號格式不正確"),
		),
		validation.Field(
			&req.ProductSerialNumber2,
			validation.When(len(req.ProductSerialNumber2) > 0, validation.Match(regexp.MustCompile(ProductSerialNumberPattern)).Error("產品序號2格式不正確")),
			validation.When(len(req.ProductSerialNumber2) > 0, validation.NotIn(req.ProductSerialNumber).Error("產品序號2不能與產品序號相同")),
		),
	)
}

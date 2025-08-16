package models

import (
	"breast-implant-warranty-system/internal/entity"
	"time"

	"github.com/guregu/null/v5"
)

// Serial 序號模型
type Serial struct {
	ID               string      `json:"id" db:"id"`
	SerialNumber     string      `json:"serial_number" db:"serial_number"`
	FullSerialNumber string      `json:"full_serial_number" db:"full_serial_number"`
	ProductID        null.String `json:"product_id" db:"product_id"`
	CreatedAt        time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at" db:"updated_at"`
}

// SerialImportItem 用於大量匯入的序號項目
type SerialImportItem struct {
	ProductID        string `json:"product_id" validate:"required,uuid"`
	SerialNumber     string `json:"serial_number" validate:"required,max=20"`
	FullSerialNumber string `json:"full_serial_number" validate:"required,max=100"`
}

// SerialImportRequest 大量匯入請求
type SerialImportRequest struct {
	Serials []SerialImportItem `json:"serials" validate:"required,min=1,dive"`
}

// SerialSearchRequest 序號搜尋請求
type SerialSearchRequest struct {
	ID               null.String `query:"id"`
	SerialNumber     null.String `query:"serial_number"`
	FullSerialNumber null.String `query:"full_serial_number"`
	ProductID        null.String `query:"product_id"`
	IsUsedByWarranty null.Bool   `query:"is_used_by_warranty"`
	entity.Pagination
}

// ValidateSerialNumber 驗證序號格式
func (s *Serial) ValidateSerialNumber() bool {
	return len(s.SerialNumber) > 0 && len(s.SerialNumber) <= 20
}

// ValidateFullSerialNumber 驗證完整序號格式
func (s *Serial) ValidateFullSerialNumber() bool {
	return len(s.FullSerialNumber) > 0 && len(s.FullSerialNumber) <= 100
}

// IsProductAssigned 檢查序號是否已分配給產品
func (s *Serial) IsProductAssigned() bool {
	return s.ProductID.Valid && s.ProductID.String != ""
}

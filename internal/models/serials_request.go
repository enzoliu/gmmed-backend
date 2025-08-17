package models

import (
	"breast-implant-warranty-system/internal/entity"
	"time"

	"github.com/guregu/null/v5"
)

// SerialBulkImportRequest 大量匯入序號請求
type SerialBulkImportRequest struct {
	Serials []SerialImportItem `json:"serials" validate:"required,min=1,dive"`
}

// SerialBulkImportResponse 大量匯入序號響應
type SerialBulkImportResponse struct {
	SuccessCount int64                   `json:"success_count"`
	FailedCount  int64                   `json:"failed_count"`
	FailedItems  []SerialImportErrorItem `json:"failed_items,omitempty"`
}

// SerialImportErrorItem 匯入失敗項目
type SerialImportErrorItem struct {
	Index            int    `json:"index"`
	ProductID        string `json:"product_id"`
	SerialNumber     string `json:"serial_number"`
	FullSerialNumber string `json:"full_serial_number"`
	Error            string `json:"error"`
}

// SerialExistsRequest 檢查序號存在請求
type SerialExistsRequest struct {
	SerialNumber string `query:"serial_number" validate:"required"`
}

// SerialExistsResponse 序號存在檢查響應
type SerialExistsResponse struct {
	Exists bool `json:"exists"`
}

// SerialListRequest 序號列表請求
type SerialListRequest struct {
	ID               null.String `query:"id"`
	SerialNumber     null.String `query:"serial_number"`
	FullSerialNumber null.String `query:"full_serial_number"`
	ProductID        null.String `query:"product_id"`
	entity.Pagination
}

// SerialListResponse 序號列表響應
type SerialListResponse struct {
	Data       []Serial `json:"data"`
	Total      int64    `json:"total"`
	Page       int      `json:"page"`
	PageSize   int      `json:"page_size"`
	TotalPages int      `json:"total_pages"`
}

// SerialCreateRequest 創建序號請求
type SerialCreateRequest struct {
	SerialNumber     string `json:"serial_number" validate:"required,max=20"`
	FullSerialNumber string `json:"full_serial_number" validate:"required,max=100"`
	ProductID        string `json:"product_id" validate:"required,uuid"`
}

// SerialUpdateRequest 更新序號請求
type SerialUpdateRequest struct {
	SerialNumber     null.String `json:"serial_number" validate:"omitempty,max=20"`
	FullSerialNumber null.String `json:"full_serial_number" validate:"omitempty,max=100"`
	ProductID        null.String `json:"product_id" validate:"omitempty,uuid"`
}

// SerialDetailResponse 序號詳細資訊響應
type SerialDetailResponse struct {
	ID               string      `json:"id"`
	SerialNumber     string      `json:"serial_number"`
	FullSerialNumber string      `json:"full_serial_number"`
	ProductID        null.String `json:"product_id"`
	Product          *Product    `json:"product,omitempty"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
}

// SerialStatsResponse 序號統計響應
type SerialStatsResponse struct {
	TotalSerials      int64 `json:"total_serials"`
	AssignedSerials   int64 `json:"assigned_serials"`
	UnassignedSerials int64 `json:"unassigned_serials"`
	TotalProducts     int64 `json:"total_products"`
}

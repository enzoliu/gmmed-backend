package models

import (
	"breast-implant-warranty-system/internal/entity"
	"encoding/json"

	"github.com/guregu/null/v5"
)

// AuditSearchLogRequest 審計日誌查詢請求
type AuditSearchLogRequest struct {
	UserName  null.String `query:"username"`
	Action    null.String `query:"action"`
	TableName null.String `query:"table_name"`
	StartDate null.String `query:"start_date"`
	EndDate   null.String `query:"end_date"`
	IPAddress null.String `query:"ip_address"`
	entity.Pagination
}

// AuditLogResponse 審計日誌響應
type AuditLogResponse struct {
	AuditLogs  []*AuditLog `json:"audit_logs"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// CreateAuditLogRequest 建立審計日誌請求
type CreateAuditLogRequest struct {
	UserID    *string         `json:"user_id"`
	Action    AuditAction     `json:"action"`
	TableName AuditTable      `json:"table_name"`
	RecordID  string          `json:"record_id"`
	OldValues json.RawMessage `json:"old_values"`
	NewValues json.RawMessage `json:"new_values"`
	IPAddress *string         `json:"ip_address"`
	UserAgent *string         `json:"user_agent"`
}

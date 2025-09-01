package models

import (
	"encoding/json"
	"time"

	"github.com/guregu/null/v5"
)

var (
	// TableNameMapping 表格名稱對應中文
	TableNameMapping = map[string]string{
		"users":                  "使用者",
		"products":               "產品",
		"warranty_registrations": "保固記錄",
		"audit_logs":             "審計日誌",
		"auth":                   "登入/登出",
	}
)

// AuditLog 審計日誌模型
type AuditLog struct {
	ID        string          `json:"id" db:"id"`
	UserID    null.String     `json:"user_id" db:"user_id"`
	Action    string          `json:"action" db:"action"`
	TableName string          `json:"table_name" db:"table_name"`
	RecordID  null.String     `json:"record_id" db:"record_id"`
	OldValues json.RawMessage `json:"old_values" db:"old_values"`
	NewValues json.RawMessage `json:"new_values" db:"new_values"`
	IPAddress null.String     `json:"ip_address" db:"ip_address"`
	UserAgent null.String     `json:"user_agent" db:"user_agent"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`

	// 關聯的用戶資訊（如果存在）
	NullableUser
}

// AuditAction 審計操作類型
type AuditAction string

const (
	AuditActionCreate AuditAction = "CREATE"
	AuditActionUpdate AuditAction = "UPDATE"
	AuditActionDelete AuditAction = "DELETE"
	AuditActionLogin  AuditAction = "LOGIN"
	AuditActionLogout AuditAction = "LOGOUT"
	AuditActionView   AuditAction = "VIEW"
	AuditActionExport AuditAction = "EXPORT"
	AuditActionImport AuditAction = "IMPORT"
)

// AuditTable 審計表名
type AuditTable string

const (
	AuditTableUsers                 AuditTable = "users"
	AuditTableProducts              AuditTable = "products"
	AuditTableSerials               AuditTable = "serials"
	AuditTableWarrantyRegistrations AuditTable = "warranty_registrations"
	AuditTableAuditLogs             AuditTable = "audit_logs"
	AuditTableAuth                  AuditTable = "auth"
)

// AuditContext 審計上下文，包含用戶和請求資訊
type AuditContext struct {
	UserID    *string `json:"user_id"`
	Username  *string `json:"username"`
	IPAddress *string `json:"ip_address"`
	UserAgent *string `json:"user_agent"`
}

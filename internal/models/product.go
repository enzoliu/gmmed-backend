package models

import (
	"fmt"
	"time"

	"github.com/guregu/null/v5"
)

// Product 產品模型
type Product struct {
	ID            string      `json:"id" db:"id"`
	ModelNumber   string      `json:"model_number" db:"model_number"`
	Brand         string      `json:"brand" db:"brand"`
	Type          string      `json:"type" db:"type"`
	Size          null.String `json:"size" db:"size"`
	WarrantyYears int         `json:"warranty_years" db:"warranty_years"`
	Description   null.String `json:"description" db:"description"`
	IsActive      bool        `json:"is_active" db:"is_active"`
	CreatedAt     time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at" db:"updated_at"`
}

// IsLifetimeWarranty 檢查是否為終身保固
func (p *Product) IsLifetimeWarranty() bool {
	return p.WarrantyYears == -1
}

// GetWarrantyDescription 取得保固描述
func (p *Product) GetWarrantyDescription() string {
	if p.IsLifetimeWarranty() {
		return "終身保固"
	} else if p.WarrantyYears == 0 {
		return "無保固"
	}
	return fmt.Sprintf("%d年保固", p.WarrantyYears)
}

type ProductWithSerialNumber struct {
	Product
	SerialNumber string `json:"serial_number" db:"serial_number"`
}

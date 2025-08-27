package services

import (
	"breast-implant-warranty-system/core/singleton"
	"breast-implant-warranty-system/pkg/dbutil"
)

type GMMedServiceConfigItf interface {
	singleton.ReadDBConfigItf
	singleton.WriteDBConfigItf
	AuthRouteConfigItf
	WarrantyRouteConfigItf
	MailgunConfigItf
}

// Services 包含所有服務的結構
type Services struct {
	User     *UserService
	Product  *ProductService
	Warranty *WarrantyService
	Audit    *AuditService
	Auth     *AuthService
	Email    *EmailService
	Serial   *SerialService
}

// NewServices 建立新的服務實例
func NewServices(db dbutil.PgxClientItf, cfg GMMedServiceConfigItf) *Services {
	return &Services{
		User:     NewUserService(db),
		Product:  NewProductService(db),
		Warranty: NewWarrantyService(db, cfg),
		Audit:    NewAuditService(db),
		Auth:     NewAuthService(db, cfg),
		Email:    NewEmailService(cfg),
		Serial:   NewSerialService(db, cfg),
	}
}

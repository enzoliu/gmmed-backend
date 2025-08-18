package handlers

import (
	"breast-implant-warranty-system/internal/services"
)

// Handlers 包含所有處理器的結構
type Handlers struct {
	User     *UserHandler
	Product  *ProductHandler
	Warranty *WarrantyHandler
	Audit    *AuditHandler
	Auth     *AuthHandler
	Serial   *SerialHandler
}

// NewHandlers 建立新的處理器實例
func NewHandlers(services *services.Services, cfg services.GMMedServiceConfigItf) *Handlers {
	return &Handlers{
		User:     NewUserHandler(services.User),
		Product:  NewProductHandler(services.Product),
		Warranty: NewWarrantyHandler(services.Warranty, services.Serial, cfg),
		Audit:    NewAuditHandler(services.Audit),
		Auth:     NewAuthHandler(services.Auth),
		Serial:   NewSerialHandler(services.Serial),
	}
}

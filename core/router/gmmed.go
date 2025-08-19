package router

import (
	"breast-implant-warranty-system/core/singleton"
	"breast-implant-warranty-system/internal/handlers"
	"breast-implant-warranty-system/internal/middleware"
	"breast-implant-warranty-system/internal/services"
	"context"
)

type GMMedRouteConfigItf interface {
	singleton.ReadDBConfigItf
	singleton.WriteDBConfigItf
	services.AuthRouteConfigItf
	services.WarrantyRouteConfigItf
	services.MailgunConfigItf
}

func RegisterGMMedRoutes(ctx context.Context, router RouterItf, singletonGroup *singleton.Group, cfg GMMedRouteConfigItf) {
	writeDB := singletonGroup.GetWriteDB(ctx, cfg)

	// 初始化服務
	services := services.NewServices(writeDB, cfg)

	// 初始化處理器
	h := handlers.NewHandlers(services, cfg)

	// API 版本群組
	api := router.Group("/api/v1")

	// 公開路由
	public := api.Group("")
	public.POST("/auth/login", h.Auth.Login)
	public.POST("/auth/refresh-token", h.Auth.RefreshToken)            // 刷新令牌
	public.POST("/auth/logout", h.Auth.Logout)                         // 登出
	public.GET("/product", h.Product.GetOneByCondition)                // 公開產品列表供保固註冊使用
	public.GET("/product/serial", h.Product.GetOneBySerialNumber)      // 公開產品列表供保固註冊使用
	public.GET("/products-metadata", h.Product.ListMetadata)           // 產品元資料列表（品牌、型號等）
	public.GET("/warranty/check-serial", h.Warranty.CheckSerialNumber) // 檢查序號是否已被使用
	// public.GET("/warranty/:id/status", h.Warranty.CheckWarrantyStatus) // 檢查保固是否已填寫
	// public.PUT("/warranty/:id/register", h.Warranty.RegisterByPatient) // 患者填寫保固（一次性）
	public.PUT("/warranty/:id/serial", h.Warranty.RegisterByPatientStep1)       // 患者填寫保固（檢查序號是否是正貨/登記手術日）
	public.PUT("/warranty/:id/info", h.Warranty.RegisterByPatientStep2)         // 患者填寫保固（更新患者資訊）
	public.PUT("/warranty/:id/confirmation", h.Warranty.RegisterByPatientStep3) // 患者確認保固資訊（建立保固）
	public.GET("/warranty/:id/info", h.Warranty.GetWarrantyByPatient)           // 患者取得保固資訊（只允許正在填的狀態，且有Cookie驗證）
	public.GET("/warranty/:id/step", h.Warranty.GetWarrantyStatusByPatient)     // 取得保固填寫階段

	// 需要認證的路由
	protected := api.Group("")
	protected.Use(middleware.CSRFProtection())
	protected.Use(middleware.JWTAuthWithConfig(cfg))
	protected.Use(middleware.AuditContext())

	// 認證相關功能
	auth := protected.Group("/auth")
	auth.GET("/me", h.Auth.Me) // 獲取當前使用者資訊

	// 使用者管理
	users := protected.Group("/users")
	users.GET("", h.User.List)
	users.GET("/:id", h.User.GetByID)
	users.PUT("/:id", h.User.Update, middleware.RequireAdmin())
	users.POST("", h.User.Create, middleware.RequireAdmin())
	users.DELETE("/:id", h.User.Delete, middleware.RequireAdmin())

	// 產品管理
	products := protected.Group("/products")
	products.GET("/manage", h.Product.List)                 // 管理用的產品列表
	products.GET("/metadata-all", h.Product.GetMetadataAll) // 產品元資料列表（品牌、型號等）
	products.GET("/:id", h.Product.GetByID)
	products.GET("/export", h.Product.ExportExcel)
	products.POST("", h.Product.Create, middleware.RequireAdminOrEditor())
	products.PUT("/:id", h.Product.Update, middleware.RequireAdminOrEditor())
	products.DELETE("/:id", h.Product.Delete, middleware.RequireAdmin())
	products.POST("/import", h.Product.ImportExcel, middleware.RequireAdminOrEditor())
	products.GET("/all", h.Product.GetAllProducts)

	// QR Code 相關功能已移除

	// 保固管理（需要認證的管理功能）
	warranty := protected.Group("/warranty")
	warranty.GET("", h.Warranty.List)
	warranty.GET("/search", h.Warranty.Search) // 保固搜尋
	warranty.GET("/:id", h.Warranty.GetByID)
	warranty.GET("/export", h.Warranty.ExportExcel)
	warranty.POST("/batch-create", h.Warranty.BatchCreate, middleware.RequireAdminOrEditor()) // 批次創建空白保固
	warranty.PUT("/:id", h.Warranty.Update, middleware.RequireAdminOrEditor())
	warranty.DELETE("/:id", h.Warranty.Delete, middleware.RequireAdmin())
	warranty.POST("/:id/resend-email", h.Warranty.ResendEmail, middleware.RequireAdminOrEditor())
	warranty.POST("/update-expired", h.Warranty.UpdateExpiredWarranties, middleware.RequireAdminOrEditor()) // 批次更新過期狀態

	// 審計日誌
	audit := protected.Group("/audit")
	audit.GET("", h.Audit.List)
	audit.GET("/:id", h.Audit.GetByID)

	// 序號管理
	serials := protected.Group("/serials")
	serials.GET("", h.Serial.Search)
	serials.GET("/:id", h.Serial.GetByID)
	serials.POST("", h.Serial.Create, middleware.RequireAdminOrEditor())
	serials.PUT("/:id", h.Serial.Update, middleware.RequireAdminOrEditor())
	serials.DELETE("/:id", h.Serial.Delete, middleware.RequireAdmin())
	serials.POST("/import", h.Serial.BulkCreate, middleware.RequireAdminOrEditor())
	serials.GET("/serials-with-product", h.Serial.GetSerialsWithProduct)
}

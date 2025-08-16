package handlers

import (
	"net/http"

	"breast-implant-warranty-system/internal/middleware"
	"breast-implant-warranty-system/internal/models"
	"breast-implant-warranty-system/internal/services"
	"breast-implant-warranty-system/pkg/validator"

	"github.com/labstack/echo/v4"
)

// ProductHandler 產品處理器
type ProductHandler struct {
	service *services.ProductService
}

// NewProductHandler 建立新的產品處理器
func NewProductHandler(service *services.ProductService) *ProductHandler {
	return &ProductHandler{
		service: service,
	}
}

// GetOneByCondition 根據條件取得單一產品
func (h *ProductHandler) GetOneByCondition(c echo.Context) error {
	ctx := c.Request().Context()

	// validate
	var req models.ProductSearchRequest
	if err := validator.Load(c, &req); err != nil {
		return err
	}

	product, err := h.service.GetOneByCondition(ctx, &req)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, product)
}

// List 列出產品
func (h *ProductHandler) List(c echo.Context) error {
	ctx := c.Request().Context()

	// validate
	var req models.ProductSearchRequest
	if err := validator.Load(c, &req); err != nil {
		return err
	}

	response, err := h.service.Search(ctx, &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response)
}

// ListMetadata 列出產品元資料（品牌、型號、類型、尺寸）
func (h *ProductHandler) ListMetadata(c echo.Context) error {
	ctx := c.Request().Context()

	// validate
	type RequestWithMetadataType struct {
		MetadataType string `query:"metadata_type" validate:"required,oneof=brands types model_numbers sizes"`
		Request      models.ProductSearchRequest
	}
	var req RequestWithMetadataType
	if err := validator.Load(c, &req); err != nil {
		return err
	}

	list, err := h.service.GetMetadataList(ctx, req.MetadataType, &req.Request)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": list,
	})
}

// GetMetadataAll 取得所有產品元資料（品牌、型號、類型、尺寸）
func (h *ProductHandler) GetMetadataAll(c echo.Context) error {
	ctx := c.Request().Context()

	metadata, err := h.service.GetMetadataAll(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, metadata)
}

// Create 建立產品
func (h *ProductHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()
	var req services.ProductCreateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}

	auditCtx := middleware.GetAuditContext(c)
	product, err := h.service.Create(ctx, &req, auditCtx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, product)
}

// GetByID 根據ID取得產品
func (h *ProductHandler) GetByID(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "必須提供產品ID資訊",
		})
	}

	product, err := h.service.GetByID(ctx, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, product)
}

// Update 更新產品
func (h *ProductHandler) Update(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "必須提供產品ID資訊",
		})
	}

	var req services.ProductUpdateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "輸入內容有誤",
		})
	}

	auditCtx := middleware.GetAuditContext(c)
	product, err := h.service.Update(ctx, id, &req, auditCtx)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, product)
}

// Delete 刪除產品
func (h *ProductHandler) Delete(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "必須提供產品ID資訊",
		})
	}

	auditCtx := middleware.GetAuditContext(c)
	err := h.service.Delete(ctx, id, auditCtx)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "產品刪除成功",
	})
}

// ImportExcel 匯入Excel
func (h *ProductHandler) ImportExcel(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"message": "產品匯入Excel端點"})
}

// ExportExcel 匯出Excel
func (h *ProductHandler) ExportExcel(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"message": "產品匯出Excel端點"})
}

func (h *ProductHandler) GetAllProducts(c echo.Context) error {
	ctx := c.Request().Context()

	products, err := h.service.GetAllProducts(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": products,
	})
}

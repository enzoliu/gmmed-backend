package handlers

import (
	"net/http"

	"breast-implant-warranty-system/internal/middleware"
	"breast-implant-warranty-system/internal/models"
	"breast-implant-warranty-system/internal/services"
	"breast-implant-warranty-system/pkg/validator"

	"github.com/labstack/echo/v4"
)

// UserHandler 使用者處理器
type UserHandler struct {
	service *services.UserService
}

// NewUserHandler 建立新的使用者處理器
func NewUserHandler(service *services.UserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

// List 列出使用者
func (h *UserHandler) List(c echo.Context) error {
	ctx := c.Request().Context()

	// validate
	var req models.UserSearchRequest
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

// Create 建立使用者
func (h *UserHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()
	var req models.CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "輸入內容有誤",
		})
	}

	auditCtx := middleware.GetAuditContext(c)
	user, err := h.service.Create(ctx, &req, auditCtx)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, user)
}

// GetByID 根據ID取得使用者
func (h *UserHandler) GetByID(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "必須提供使用者ID資訊",
		})
	}

	user, err := h.service.GetByID(ctx, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, user)
}

// Update 更新使用者
func (h *UserHandler) Update(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "必須提供使用者ID資訊",
		})
	}

	var req models.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "輸入內容有誤",
		})
	}

	auditCtx := middleware.GetAuditContext(c)
	user, err := h.service.Update(ctx, id, &req, auditCtx)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, user)
}

// Delete 刪除使用者
func (h *UserHandler) Delete(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "必須提供使用者ID資訊",
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
		"message": "使用者刪除成功",
	})
}

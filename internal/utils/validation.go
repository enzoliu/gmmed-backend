package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// ValidationError 驗證錯誤結構
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// ValidationErrorResponse 驗證錯誤回應
type ValidationErrorResponse struct {
	Error   string            `json:"error"`
	Details []ValidationError `json:"details"`
}

// BindAndValidate 綁定並驗證請求
func BindAndValidate(c echo.Context, req interface{}) error {
	// 先嘗試自定義綁定（處理日期格式）
	if err := customBind(c, req); err != nil {
		// 檢查是否是 JSON 語法錯誤
		if jsonErr, ok := err.(*json.SyntaxError); ok {
			return c.JSON(400, ValidationErrorResponse{
				Error: "Invalid JSON format",
				Details: []ValidationError{
					{
						Field:   "json",
						Tag:     "syntax",
						Message: fmt.Sprintf("JSON syntax error at position %d", jsonErr.Offset),
					},
				},
			})
		}

		// 檢查是否是 JSON 類型錯誤
		if jsonErr, ok := err.(*json.UnmarshalTypeError); ok {
			return c.JSON(400, ValidationErrorResponse{
				Error: "Invalid field type",
				Details: []ValidationError{
					{
						Field:   jsonErr.Field,
						Tag:     "type",
						Value:   jsonErr.Value,
						Message: fmt.Sprintf("Expected %s but got %s", jsonErr.Type, jsonErr.Value),
					},
				},
			})
		}

		// 其他綁定錯誤
		return c.JSON(400, ValidationErrorResponse{
			Error: "Request binding failed",
			Details: []ValidationError{
				{
					Field:   "request",
					Tag:     "bind",
					Message: err.Error(),
				},
			},
		})
	}

	// 使用 validator 進行驗證
	validate := validator.New()

	// 註冊自定義標籤名稱
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	if err := validate.Struct(req); err != nil {
		var validationErrors []ValidationError

		for _, err := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, ValidationError{
				Field:   err.Field(),
				Tag:     err.Tag(),
				Value:   fmt.Sprintf("%v", err.Value()),
				Message: getValidationMessage(err),
			})
		}

		return c.JSON(400, ValidationErrorResponse{
			Error:   "Validation failed",
			Details: validationErrors,
		})
	}

	return nil
}

// getValidationMessage 取得驗證錯誤資訊
func getValidationMessage(err validator.FieldError) string {
	field := err.Field()
	tag := err.Tag()
	param := err.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", field, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", field, param)
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", field, param)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, param)
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, param)
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, param)
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, param)
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, param)
	case "numeric":
		return fmt.Sprintf("%s must be numeric", field)
	case "alpha":
		return fmt.Sprintf("%s must contain only letters", field)
	case "alphanum":
		return fmt.Sprintf("%s must contain only letters and numbers", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "datetime":
		return fmt.Sprintf("%s must be a valid datetime", field)
	default:
		return fmt.Sprintf("%s failed validation for tag '%s'", field, tag)
	}
}

// ValidateWarrantyRequest 驗證保固請求的業務邏輯
func ValidateWarrantyRequest(req interface{}) []ValidationError {
	var errors []ValidationError

	// 這裡可以添加自定義的業務邏輯驗證
	// 例如：台灣身分證格式、手機號碼格式等

	return errors
}

// customBind 自定義綁定，處理日期格式
func customBind(c echo.Context, req interface{}) error {
	// 讀取原始 JSON
	body := c.Request().Body
	defer body.Close()

	var rawData map[string]interface{}
	if err := json.NewDecoder(body).Decode(&rawData); err != nil {
		return err
	}

	// 處理日期欄位
	if err := processDateFields(rawData); err != nil {
		return err
	}

	// 重新序列化並綁定
	processedJSON, err := json.Marshal(rawData)
	if err != nil {
		return err
	}

	return json.Unmarshal(processedJSON, req)
}

// processDateFields 處理日期欄位格式
func processDateFields(data map[string]interface{}) error {
	dateFields := []string{
		"patient_birth_date",
		"surgery_date",
	}

	for _, field := range dateFields {
		if value, exists := data[field]; exists {
			if strValue, ok := value.(string); ok {
				// 嘗試解析各種日期格式
				parsedTime, err := parseFlexibleDate(strValue)
				if err != nil {
					return fmt.Errorf("invalid date format for %s: %s", field, strValue)
				}
				// 轉換為 RFC3339 格式
				data[field] = parsedTime.Format(time.RFC3339)
			}
		}
	}

	return nil
}

// parseFlexibleDate 靈活解析日期格式
func parseFlexibleDate(dateStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("empty date string")
	}

	// 支援的日期格式
	formats := []string{
		"2006-01-02",           // ISO 8601 日期格式
		"2006/01/02",           // 斜線分隔格式
		"2006-1-2",             // 不補零格式
		"2006/1/2",             // 斜線不補零格式
		"2006-01-02T15:04:05Z", // 完整 ISO 8601 格式
		"2006-01-02 15:04:05",  // 空格分隔格式
		time.RFC3339,           // RFC3339 格式
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			// 如果只有日期沒有時間，設為當天的開始時間 (UTC)
			if !strings.Contains(dateStr, "T") && !strings.Contains(dateStr, " ") {
				return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
			}
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

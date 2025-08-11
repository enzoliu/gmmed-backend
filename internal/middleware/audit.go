package middleware

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// AuditLog 審計日誌中間件
func AuditLog() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// 讀取請求體（如果有的話）
			var requestBody []byte
			if c.Request().Body != nil {
				requestBody, _ = io.ReadAll(c.Request().Body)
				c.Request().Body = io.NopCloser(bytes.NewBuffer(requestBody))
			}

			// 建立響應記錄器
			rec := &responseRecorder{
				ResponseWriter: c.Response().Writer,
				statusCode:     200,
				body:           bytes.NewBuffer(nil),
			}
			c.Response().Writer = rec

			// 執行請求
			err := next(c)

			// 記錄審計日誌
			duration := time.Since(start)

			// 取得使用者資訊（如果已認證）
			userID := c.Get("user_id")
			username := c.Get("username")

			// 構建日誌欄位
			fields := logrus.Fields{
				"method":     c.Request().Method,
				"uri":        c.Request().RequestURI,
				"status":     rec.statusCode,
				"duration":   duration.Milliseconds(),
				"ip":         c.RealIP(),
				"user_agent": c.Request().UserAgent(),
				"request_id": c.Response().Header().Get(echo.HeaderXRequestID),
			}

			// 添加使用者資訊（如果有）
			if userID != nil {
				fields["user_id"] = userID
			}
			if username != nil {
				fields["username"] = username
			}

			// 對於敏感操作，記錄請求體（但要過濾敏感資訊）
			if shouldLogRequestBody(c.Request().Method, c.Path()) {
				if len(requestBody) > 0 && len(requestBody) < 1024 { // 限制大小
					// 這裡應該過濾敏感資訊如密碼等
					fields["request_body"] = string(requestBody)
				}
			}

			// 記錄日誌
			logger := logrus.WithFields(fields)
			if err != nil {
				logger.WithError(err).Error("Request failed")
			} else if rec.statusCode >= 400 {
				logger.Warn("Request completed with error status")
			} else {
				logger.Info("Request completed")
			}

			return err
		}
	}
}

// responseRecorder 用於記錄響應資訊
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// shouldLogRequestBody 判斷是否應該記錄請求體
func shouldLogRequestBody(method, path string) bool {
	// 只對POST、PUT、PATCH等修改操作記錄請求體
	if method != "POST" && method != "PUT" && method != "PATCH" {
		return false
	}

	// 排除登入等敏感端點
	sensitiveEndpoints := []string{
		"/api/v1/auth/login",
		"/api/v1/auth/change-password",
	}

	for _, endpoint := range sensitiveEndpoints {
		if path == endpoint {
			return false
		}
	}

	return true
}

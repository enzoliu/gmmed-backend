package utils

import (
	"regexp"
	"strings"
	"unicode"
)

// ValidateEmail 驗證電子信箱格式
func ValidateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// ValidateTaiwanPhone 驗證台灣手機號碼
func ValidateTaiwanPhone(phone string) bool {
	// 移除所有空格和連字符
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")

	// 台灣手機號碼格式：09xxxxxxxx (10位數字)
	phoneRegex := regexp.MustCompile(`^09\d{8}$`)
	return phoneRegex.MatchString(phone)
}

// ValidatePassword 驗證密碼強度
func ValidatePassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// 至少包含大寫字母、小寫字母、數字中的三種
	count := 0
	if hasUpper {
		count++
	}
	if hasLower {
		count++
	}
	if hasDigit {
		count++
	}
	if hasSpecial {
		count++
	}

	return count >= 3
}

// SanitizeString 清理字串，移除危險字符
func SanitizeString(input string) string {
	// 移除HTML標籤
	htmlRegex := regexp.MustCompile(`<[^>]*>`)
	input = htmlRegex.ReplaceAllString(input, "")

	// 移除SQL注入相關字符
	sqlRegex := regexp.MustCompile(`[';\"\\]`)
	input = sqlRegex.ReplaceAllString(input, "")

	return strings.TrimSpace(input)
}

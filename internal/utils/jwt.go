package utils

import (
	"errors"
	"time"

	"breast-implant-warranty-system/internal/models"

	"github.com/golang-jwt/jwt/v4"
)

// JWTClaims JWT聲明
type JWTClaims struct {
	UserID   string          `json:"user_id"`
	Username string          `json:"username"`
	Role     models.UserRole `json:"role"`
	jwt.RegisteredClaims
}

// GenerateJWT 生成JWT令牌
func GenerateJWT(user *models.User, secret string, expireHours int) (string, time.Time, error) {
	expirationTime := time.Now().Add(time.Duration(expireHours) * time.Hour)

	claims := &JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     models.UserRole(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "gmmed",
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expirationTime, nil
}

// ValidateJWT 驗證JWT令牌
func ValidateJWT(tokenString, secret string) (*JWTClaims, error) {
	claims := &JWTClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// RefreshJWT 刷新JWT令牌
func RefreshJWT(tokenString, secret string, expireHours int) (string, time.Time, error) {
	claims, err := ValidateJWT(tokenString, secret)
	if err != nil {
		return "", time.Time{}, err
	}

	// 檢查令牌是否即將過期（在30分鐘內）
	if time.Until(claims.ExpiresAt.Time) > 30*time.Minute {
		return "", time.Time{}, errors.New("token is not eligible for refresh")
	}

	// 建立新的令牌
	expirationTime := time.Now().Add(time.Duration(expireHours) * time.Hour)
	claims.ExpiresAt = jwt.NewNumericDate(expirationTime)
	claims.IssuedAt = jwt.NewNumericDate(time.Now())

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString([]byte(secret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expirationTime, nil
}

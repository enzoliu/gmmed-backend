package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
)

// parseKey 解析密鑰，支持十六進制字符串和直接字節
func parseKey(key string) ([]byte, error) {
	// 如果是十六進制字符串（64個字符 = 32字節）
	if len(key) == 64 {
		keyBytes, err := hex.DecodeString(key)
		if err != nil {
			return nil, errors.New("invalid hex key format")
		}
		if len(keyBytes) != 32 {
			return nil, errors.New("key must be 32 bytes long")
		}
		return keyBytes, nil
	}
	
	// 如果是直接的字節字符串
	keyBytes := []byte(key)
	if len(keyBytes) != 32 {
		return nil, errors.New("key must be 32 bytes long")
	}
	return keyBytes, nil
}

// EncryptAES 使用AES-GCM加密資料
func EncryptAES(plaintext, key string) (string, error) {
	keyBytes, err := parseKey(key)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptAES 使用AES-GCM解密資料
func DecryptAES(ciphertext, key string) (string, error) {
	keyBytes, err := parseKey(key)
	if err != nil {
		return "", err
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, cipherData := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// EncryptPatientID 加密患者身分證字號
func EncryptPatientID(patientID, encryptionKey string) (string, error) {
	return EncryptAES(patientID, encryptionKey)
}

// DecryptPatientID 解密患者身分證字號
func DecryptPatientID(encryptedID, encryptionKey string) (string, error) {
	return DecryptAES(encryptedID, encryptionKey)
}

// EncryptPatientPhone 加密患者手機號碼
func EncryptPatientPhone(phone, encryptionKey string) (string, error) {
	return EncryptAES(phone, encryptionKey)
}

// DecryptPatientPhone 解密患者手機號碼
func DecryptPatientPhone(encryptedPhone, encryptionKey string) (string, error) {
	return DecryptAES(encryptedPhone, encryptionKey)
}

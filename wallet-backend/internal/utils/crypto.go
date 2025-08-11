package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// GenerateRandomHex 生成指定长度的随机十六进制字符串
func GenerateRandomHex(length int) (string, error) {
	if length%2 != 0 {
		return "", fmt.Errorf("length must be even")
	}

	bytes := make([]byte, length/2)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

// GenerateUniqueID 生成唯一ID
func GenerateUniqueID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// ValidateHexString 验证十六进制字符串
func ValidateHexString(s string) bool {
	if len(s)%2 != 0 {
		return false
	}

	_, err := hex.DecodeString(s)
	return err == nil
}

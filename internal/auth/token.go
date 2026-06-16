package auth

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateToken xavfsiz tasodifiy token (32 bayt -> 64 hex belgisi) qaytaradi.
func GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GenerateSecret webhook imzosi uchun maxfiy kalit qaytaradi.
func GenerateSecret() (string, error) {
	return GenerateToken()
}

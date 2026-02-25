package utils

import (
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes plaintext password using bcrypt.
func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// VerifyPassword validates plaintext against stored hash.
// It supports bcrypt and falls back to legacy scrypt hash for backward compatibility.
func VerifyPassword(plaintext, stored string) bool {
	if strings.HasPrefix(stored, "$2a$") || strings.HasPrefix(stored, "$2b$") || strings.HasPrefix(stored, "$2y$") {
		return bcrypt.CompareHashAndPassword([]byte(stored), []byte(plaintext)) == nil
	}
	return PasswordVerify(plaintext, stored)
}

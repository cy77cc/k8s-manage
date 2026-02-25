package utils

import "testing"

func TestHashAndVerifyPassword(t *testing.T) {
	password := "P@ssw0rd-123"

	hashed, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if hashed == password {
		t.Fatalf("hashed password should not equal plaintext")
	}
	if !VerifyPassword(password, hashed) {
		t.Fatalf("VerifyPassword should accept bcrypt hash")
	}
	if VerifyPassword("wrong-pass", hashed) {
		t.Fatalf("VerifyPassword should reject wrong password")
	}
}

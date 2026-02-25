package client

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/pem"
	"testing"

	gossh "golang.org/x/crypto/ssh"
)

func TestBuildAuthMethods_WithEncryptedPrivateKeyAndPassphrase(t *testing.T) {
	privateKey, passphrase := generateEncryptedPrivateKeyForTest(t)

	methods, err := buildAuthMethods("", privateKey, passphrase)
	if err != nil {
		t.Fatalf("buildAuthMethods returned error: %v", err)
	}
	if len(methods) != 1 {
		t.Fatalf("expected 1 auth method, got %d", len(methods))
	}
}

func TestBuildAuthMethods_PrivateKeyPreferredOverPassword(t *testing.T) {
	privateKey, passphrase := generateEncryptedPrivateKeyForTest(t)

	methods, err := buildAuthMethods("pwd", privateKey, passphrase)
	if err != nil {
		t.Fatalf("buildAuthMethods returned error: %v", err)
	}
	if len(methods) != 2 {
		t.Fatalf("expected 2 auth methods, got %d", len(methods))
	}
}

func generateEncryptedPrivateKeyForTest(t *testing.T) (string, string) {
	t.Helper()

	raw, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key failed: %v", err)
	}
	passphrase := "test-passphrase"
	block, err := gossh.MarshalPrivateKeyWithPassphrase(raw, "", []byte(passphrase))
	if err != nil {
		t.Fatalf("marshal encrypted private key failed: %v", err)
	}
	return string(pem.EncodeToMemory(block)), passphrase
}

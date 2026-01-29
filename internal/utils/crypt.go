package utils

import (
	"fmt"

	"github.com/cy77cc/k8s-manage/internal/config"
	"golang.org/x/crypto/scrypt"
)

func EncryptPassword(pwd string) (string, error) {
	salt := config.CFG.Server.Salt
	dk, err := scrypt.Key([]byte(pwd), []byte(salt), 32768, 8, 1, 32)
	return fmt.Sprintf("%x", string(dk)), err
}

func PasswordVerify(password, hashedPassword string) bool {
	bk, err := EncryptPassword(password)
	if err != nil {
		return false
	}
	return bk == hashedPassword
}

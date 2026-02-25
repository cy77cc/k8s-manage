package client

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

func NewSSHClient(user, password, host string, port int, privateKey, passphrase string) (*ssh.Client, error) {
	authMethods, err := buildAuthMethods(password, privateKey, passphrase)
	if err != nil {
		return nil, err
	}
	config := &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
		Auth:            authMethods,
	}
	addr := fmt.Sprintf("%s:%d", host, port)
	return ssh.Dial("tcp", addr, config)
}

func buildAuthMethods(password, privateKey, passphrase string) ([]ssh.AuthMethod, error) {
	authMethods := make([]ssh.AuthMethod, 0, 2)
	var keyParseErr error
	trimmedKey := strings.TrimSpace(privateKey)
	if trimmedKey != "" {
		var (
			signer ssh.Signer
			err    error
		)
		if strings.TrimSpace(passphrase) != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(trimmedKey), []byte(passphrase))
		} else {
			signer, err = ssh.ParsePrivateKey([]byte(trimmedKey))
		}
		if err != nil {
			keyParseErr = err
		} else {
			// Prefer key auth first when both are configured.
			authMethods = append(authMethods, ssh.PublicKeys(signer))
		}
	}
	if strings.TrimSpace(password) != "" {
		authMethods = append(authMethods, ssh.Password(password))
	}
	if len(authMethods) == 0 {
		if keyParseErr != nil {
			return nil, keyParseErr
		}
		return nil, fmt.Errorf("no ssh auth method provided")
	}
	return authMethods, nil
}

func RunCommand(client *ssh.Client, cmd string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}

	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	return strings.TrimSpace(string(output)), err
}

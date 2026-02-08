package client

import (
	"os"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func NewSFTPClient(sshClient *ssh.Client) (*sftp.Client, error) {
	return sftp.NewClient(sshClient)
}

func UploadFile(sftpClient *sftp.Client, local, remote string) error {
	src, err := os.Open(local)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := sftpClient.Create(remote)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = dst.ReadFrom(src)
	return err
}
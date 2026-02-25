package logic

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	sshclient "github.com/cy77cc/k8s-manage/internal/client/ssh"
	"github.com/cy77cc/k8s-manage/internal/config"
	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/utils"
	golangssh "golang.org/x/crypto/ssh"
	"gorm.io/gorm"
)

type SSHKeyCreateReq struct {
	Name       string `json:"name"`
	PrivateKey string `json:"private_key"`
	Passphrase string `json:"passphrase"`
}

type SSHKeyVerifyReq struct {
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	Username string `json:"username"`
}

func (s *HostService) ListSSHKeys(ctx context.Context) ([]model.SSHKey, error) {
	var list []model.SSHKey
	err := s.svcCtx.DB.WithContext(ctx).Select("id", "name", "public_key", "fingerprint", "algorithm", "encrypted", "usage_count", "created_at", "updated_at").
		Order("id desc").
		Find(&list).Error
	return list, err
}

func (s *HostService) CreateSSHKey(ctx context.Context, req SSHKeyCreateReq) (*model.SSHKey, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, errors.New("name is required")
	}
	if strings.TrimSpace(req.PrivateKey) == "" {
		return nil, errors.New("private_key is required")
	}
	if strings.TrimSpace(config.CFG.Security.EncryptionKey) == "" {
		return nil, errors.New("security.encryption_key is required")
	}
	pub, alg, fp, err := parsePrivateKeyMeta(req.PrivateKey, req.Passphrase)
	if err != nil {
		return nil, err
	}
	cipher, err := utils.EncryptText(req.PrivateKey, config.CFG.Security.EncryptionKey)
	if err != nil {
		return nil, err
	}
	key := &model.SSHKey{
		Name:        req.Name,
		PublicKey:   pub,
		PrivateKey:  cipher,
		Passphrase:  req.Passphrase,
		Fingerprint: fp,
		Algorithm:   alg,
		Encrypted:   true,
	}
	if err := s.svcCtx.DB.WithContext(ctx).Create(key).Error; err != nil {
		return nil, err
	}
	key.PrivateKey = ""
	key.Passphrase = ""
	return key, nil
}

func (s *HostService) DeleteSSHKey(ctx context.Context, id uint64) error {
	var count int64
	if err := s.svcCtx.DB.WithContext(ctx).Model(&model.Node{}).Where("ssh_key_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("ssh key is in use by hosts")
	}
	return s.svcCtx.DB.WithContext(ctx).Delete(&model.SSHKey{}, id).Error
}

func (s *HostService) VerifySSHKey(ctx context.Context, id uint64, req SSHKeyVerifyReq) (map[string]any, error) {
	if req.Port <= 0 {
		req.Port = DefaultSSHPort
	}
	if req.Username == "" {
		req.Username = "root"
	}
	if strings.TrimSpace(req.IP) == "" {
		return nil, errors.New("ip is required")
	}
	privateKey, passphrase, err := s.loadPrivateKey(ctx, &id)
	if err != nil {
		return nil, err
	}
	cli, err := sshclient.NewSSHClient(req.Username, "", req.IP, req.Port, privateKey, passphrase)
	if err != nil {
		return map[string]any{"reachable": false, "message": err.Error()}, nil
	}
	defer cli.Close()
	out, err := sshclient.RunCommand(cli, "hostname")
	if err != nil {
		return map[string]any{"reachable": false, "message": err.Error()}, nil
	}
	_ = s.svcCtx.DB.WithContext(ctx).Model(&model.SSHKey{}).Where("id = ?", id).UpdateColumn("usage_count", gorm.Expr("usage_count + ?", 1)).Error
	return map[string]any{"reachable": true, "hostname": out}, nil
}

func parsePrivateKeyMeta(privateKey string, passphrase string) (publicKey string, algorithm string, fingerprint string, err error) {
	var signer golangssh.Signer
	if passphrase != "" {
		signer, err = golangssh.ParsePrivateKeyWithPassphrase([]byte(privateKey), []byte(passphrase))
	} else {
		signer, err = golangssh.ParsePrivateKey([]byte(privateKey))
	}
	if err != nil {
		return "", "", "", fmt.Errorf("invalid private key: %w", err)
	}
	pub := signer.PublicKey()
	pubBytes := pub.Marshal()
	hash := sha256.Sum256(pubBytes)
	fp := "SHA256:" + base64.StdEncoding.EncodeToString(hash[:])
	return strings.TrimSpace(string(golangssh.MarshalAuthorizedKey(pub))), pub.Type(), fp, nil
}

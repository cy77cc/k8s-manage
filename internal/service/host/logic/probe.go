package logic

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	sshclient "github.com/cy77cc/k8s-manage/internal/client/ssh"
	"github.com/cy77cc/k8s-manage/internal/config"
	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/utils"
	"github.com/google/uuid"
)

func (s *HostService) Probe(ctx context.Context, userID uint64, req ProbeReq) (*ProbeResp, error) {
	normalizeProbeReq(&req)
	if err := validateProbeReq(req); err != nil {
		return &ProbeResp{Reachable: false, ErrorCode: "validation_error", Message: err.Error()}, nil
	}

	start := time.Now()
	facts, warnings, privateKey, err := s.probeFacts(ctx, req)
	latency := time.Since(start).Milliseconds()
	resp := &ProbeResp{
		Reachable: err == nil,
		LatencyMS: latency,
		Facts:     facts,
		Warnings:  warnings,
		ExpiresAt: time.Now().Add(ProbeTokenTTL),
	}
	if err != nil {
		resp.ErrorCode, resp.Message = mapProbeError(err)
	}

	token := uuid.NewString()
	hash := hashToken(token)
	factsJSON, _ := json.Marshal(resp.Facts)
	warningsJSON, _ := json.Marshal(resp.Warnings)
	probe := model.HostProbeSession{
		TokenHash:      hash,
		Name:           req.Name,
		IP:             req.IP,
		Port:           req.Port,
		AuthType:       req.AuthType,
		Username:       req.Username,
		PasswordCipher: req.Password,
		Reachable:      resp.Reachable,
		LatencyMS:      resp.LatencyMS,
		FactsJSON:      string(factsJSON),
		WarningsJSON:   string(warningsJSON),
		ExpiresAt:      resp.ExpiresAt,
		CreatedBy:      userID,
	}
	if req.SSHKeyID != nil {
		probe.SSHKeyID = req.SSHKeyID
	}
	if strings.TrimSpace(privateKey) == "" && req.AuthType == "key" {
		resp.Warnings = append(resp.Warnings, "ssh key not found, key auth may fail")
	}
	if err := s.svcCtx.DB.WithContext(ctx).Create(&probe).Error; err != nil {
		return nil, err
	}

	resp.ProbeToken = token
	return resp, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func mapProbeError(err error) (string, string) {
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "timeout") || strings.Contains(msg, "deadline"):
		return "timeout_error", err.Error()
	case strings.Contains(msg, "authentication") || strings.Contains(msg, "unable to authenticate"):
		return "auth_error", err.Error()
	case strings.Contains(msg, "no ssh auth method"):
		return "validation_error", err.Error()
	default:
		return "connect_error", err.Error()
	}
}

func (s *HostService) loadPrivateKey(ctx context.Context, sshKeyID *uint64) (string, string, error) {
	if sshKeyID == nil {
		return "", "", nil
	}
	var key model.SSHKey
	if err := s.svcCtx.DB.WithContext(ctx).Select("id", "private_key", "passphrase", "encrypted").Where("id = ?", *sshKeyID).First(&key).Error; err != nil {
		return "", "", err
	}
	passphrase := strings.TrimSpace(key.Passphrase)
	if key.Encrypted {
		privateKey, err := utils.DecryptText(key.PrivateKey, config.CFG.Security.EncryptionKey)
		if err != nil {
			return "", "", err
		}
		return privateKey, passphrase, nil
	}
	return key.PrivateKey, passphrase, nil
}

func (s *HostService) probeFacts(ctx context.Context, req ProbeReq) (ProbeFacts, []string, string, error) {
	probeCtx, cancel := context.WithTimeout(ctx, ProbeTimeout)
	defer cancel()

	privateKey, passphrase, err := s.loadPrivateKey(probeCtx, req.SSHKeyID)
	if err != nil {
		return ProbeFacts{}, nil, "", err
	}
	password := req.Password
	if strings.TrimSpace(privateKey) != "" {
		password = ""
	}
	cli, err := sshclient.NewSSHClient(req.Username, password, req.IP, req.Port, privateKey, passphrase)
	if err != nil {
		return ProbeFacts{}, nil, privateKey, err
	}
	defer cli.Close()

	cmd := `
echo "hostname=$(hostname)"
echo "os=$(cat /etc/os-release | grep PRETTY_NAME | cut -d= -f2 | tr -d '"')"
echo "arch=$(uname -m)"
echo "kernel=$(uname -r)"
echo "cpu=$(nproc)"
echo "mem=$(free -m | awk '/Mem:/ {print $2}')"
echo "disk=$(df -BG / | tail -1 | awk '{print $2}' | tr -d G)"
`
	out, err := sshclient.RunCommand(cli, cmd)
	if err != nil {
		return ProbeFacts{}, nil, privateKey, err
	}

	facts := ProbeFacts{}
	warnings := make([]string, 0)
	for _, line := range strings.Split(out, "\n") {
		parts := strings.SplitN(strings.TrimSpace(line), "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		switch key {
		case "hostname":
			facts.Hostname = val
		case "os":
			facts.OS = val
		case "arch":
			facts.Arch = val
		case "kernel":
			facts.Kernel = val
		case "cpu":
			_, _ = fmt.Sscanf(val, "%d", &facts.CPUCores)
		case "mem":
			_, _ = fmt.Sscanf(val, "%d", &facts.MemoryMB)
		case "disk":
			_, _ = fmt.Sscanf(val, "%d", &facts.DiskGB)
		}
	}
	if facts.Hostname == "" {
		warnings = append(warnings, "hostname not detected")
	}
	return facts, warnings, privateKey, nil
}

func validateProbeReq(req ProbeReq) error {
	if strings.TrimSpace(req.Name) == "" {
		return errors.New("name is required")
	}
	if strings.TrimSpace(req.IP) == "" {
		return errors.New("ip is required")
	}
	if strings.TrimSpace(req.Username) == "" {
		return errors.New("username is required")
	}
	if req.AuthType != "password" && req.AuthType != "key" {
		return errors.New("auth_type must be password or key")
	}
	if req.AuthType == "password" && strings.TrimSpace(req.Password) == "" {
		return errors.New("password is required when auth_type=password")
	}
	if req.AuthType == "key" && req.SSHKeyID == nil {
		return errors.New("ssh_key_id is required when auth_type=key")
	}
	return nil
}

func normalizeProbeReq(req *ProbeReq) {
	if req.Port <= 0 {
		req.Port = DefaultSSHPort
	}
	if req.Username == "" {
		req.Username = "root"
	}
}

func normalizeCredentialReq(req *UpdateCredentialsReq) {
	if req.Port <= 0 {
		req.Port = DefaultSSHPort
	}
	if req.AuthType == "" {
		req.AuthType = "password"
	}
}

package logic

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"gorm.io/gorm"
)

const (
	DefaultSSHPort    = 22
	ProbeTokenTTL     = 10 * time.Minute
	ProbeTimeout      = 8 * time.Second
	NodeSunsetDateRFC = "Mon, 30 Jun 2026 00:00:00 GMT"
)

type HostService struct {
	svcCtx *svc.ServiceContext
}

func NewHostService(svcCtx *svc.ServiceContext) *HostService {
	return &HostService{svcCtx: svcCtx}
}

type ProbeReq struct {
	Name     string  `json:"name"`
	IP       string  `json:"ip"`
	Port     int     `json:"port"`
	AuthType string  `json:"auth_type"`
	Username string  `json:"username"`
	Password string  `json:"password"`
	SSHKeyID *uint64 `json:"ssh_key_id"`
}

type ProbeFacts struct {
	Hostname string `json:"hostname"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Kernel   string `json:"kernel"`
	CPUCores int    `json:"cpu_cores"`
	MemoryMB int    `json:"memory_mb"`
	DiskGB   int    `json:"disk_gb"`
}

type ProbeResp struct {
	ProbeToken string     `json:"probe_token"`
	Reachable  bool       `json:"reachable"`
	LatencyMS  int64      `json:"latency_ms"`
	Facts      ProbeFacts `json:"facts"`
	Warnings   []string   `json:"warnings"`
	ErrorCode  string     `json:"error_code,omitempty"`
	Message    string     `json:"message,omitempty"`
	ExpiresAt  time.Time  `json:"expires_at"`
}

type CreateReq struct {
	ProbeToken   string   `json:"probe_token"`
	Name         string   `json:"name"`
	IP           string   `json:"ip"`
	Port         int      `json:"port"`
	AuthType     string   `json:"auth_type"`
	Username     string   `json:"username"`
	Password     string   `json:"password"`
	SSHKeyID     *uint64  `json:"ssh_key_id"`
	Description  string   `json:"description"`
	Labels       []string `json:"labels"`
	Role         string   `json:"role"`
	ClusterID    uint     `json:"cluster_id"`
	Source       string   `json:"source"`
	Provider     string   `json:"provider"`
	ProviderID   string   `json:"provider_instance_id"`
	ParentHostID *uint64  `json:"parent_host_id"`
	Force        bool     `json:"force"`
	Status       string   `json:"status"`
}

type UpdateCredentialsReq struct {
	AuthType string  `json:"auth_type"`
	Username string  `json:"username"`
	Password string  `json:"password"`
	SSHKeyID *uint64 `json:"ssh_key_id"`
	Port     int     `json:"port"`
}

func (s *HostService) List(ctx context.Context) ([]model.Node, error) {
	var list []model.Node
	return list, s.svcCtx.DB.WithContext(ctx).Find(&list).Error
}

func (s *HostService) Get(ctx context.Context, id uint64) (*model.Node, error) {
	var node model.Node
	if err := s.svcCtx.DB.WithContext(ctx).First(&node, id).Error; err != nil {
		return nil, err
	}
	return &node, nil
}

func (s *HostService) Update(ctx context.Context, id uint64, patch map[string]any) (*model.Node, error) {
	if err := s.svcCtx.DB.WithContext(ctx).Model(&model.Node{}).Where("id = ?", id).Updates(patch).Error; err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}

func (s *HostService) Delete(ctx context.Context, id uint64) error {
	return s.svcCtx.DB.WithContext(ctx).Delete(&model.Node{}, id).Error
}

func (s *HostService) UpdateStatus(ctx context.Context, id uint64, status string) error {
	return s.svcCtx.DB.WithContext(ctx).Model(&model.Node{}).Where("id = ?", id).Update("status", strings.ToLower(status)).Error
}

func (s *HostService) BatchUpdateStatus(ctx context.Context, ids []uint64, status string) error {
	if len(ids) == 0 || status == "" {
		return nil
	}
	return s.svcCtx.DB.WithContext(ctx).Model(&model.Node{}).Where("id IN ?", ids).Update("status", status).Error
}

func (s *HostService) consumeProbe(ctx context.Context, userID uint64, token string) (*model.HostProbeSession, error) {
	hash := hashToken(token)
	var probe model.HostProbeSession
	if err := s.svcCtx.DB.WithContext(ctx).Where("token_hash = ?", hash).First(&probe).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("probe_not_found")
		}
		return nil, err
	}
	if probe.CreatedBy != 0 && userID != 0 && probe.CreatedBy != userID {
		return nil, errors.New("probe_not_found")
	}
	if probe.ConsumedAt != nil {
		return nil, errors.New("probe_not_found")
	}
	if time.Now().After(probe.ExpiresAt) {
		return nil, errors.New("probe_expired")
	}

	now := time.Now()
	if err := s.svcCtx.DB.WithContext(ctx).Model(&model.HostProbeSession{}).
		Where("id = ? AND consumed_at IS NULL", probe.ID).
		Update("consumed_at", &now).Error; err != nil {
		return nil, err
	}
	probe.ConsumedAt = &now
	return &probe, nil
}

func ParseLabels(labels string) []string {
	parts := strings.Split(labels, ",")
	out := make([]string, 0, len(parts))
	for _, item := range parts {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

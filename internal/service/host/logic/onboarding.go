package logic

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/cy77cc/k8s-manage/internal/model"
)

func (s *HostService) CreateWithProbe(ctx context.Context, userID uint64, isAdmin bool, req CreateReq) (*model.Node, error) {
	if req.Force && !isAdmin {
		return nil, errors.New("force create requires admin")
	}

	if strings.TrimSpace(req.ProbeToken) == "" {
		return s.createFromLegacyReq(ctx, req)
	}

	probe, err := s.consumeProbe(ctx, userID, req.ProbeToken)
	if err != nil {
		return nil, err
	}
	if !probe.Reachable && !req.Force {
		return nil, errors.New("probe failed and force is disabled")
	}

	facts := ProbeFacts{}
	_ = json.Unmarshal([]byte(probe.FactsJSON), &facts)
	labels := strings.Join(req.Labels, ",")
	node := &model.Node{
		Name:        firstNonEmpty(req.Name, probe.Name),
		Hostname:    facts.Hostname,
		Description: req.Description,
		IP:          probe.IP,
		Port:        probe.Port,
		SSHUser:     probe.Username,
		SSHPassword: probe.PasswordCipher,
		Labels:      labels,
		Status:      buildStatus(probe.Reachable),
		OS:          facts.OS,
		Arch:        facts.Arch,
		Kernel:      facts.Kernel,
		CpuCores:    facts.CPUCores,
		MemoryMB:    facts.MemoryMB,
		DiskGB:      facts.DiskGB,
		Role:        req.Role,
		ClusterID:   req.ClusterID,
		LastCheckAt: time.Now(),
	}
	if probe.SSHKeyID != nil {
		node.SSHKeyID = nodeIDPtr(*probe.SSHKeyID)
	}

	if err := s.svcCtx.DB.WithContext(ctx).Create(node).Error; err != nil {
		return nil, err
	}
	return node, nil
}

func (s *HostService) UpdateCredentials(ctx context.Context, id uint64, req UpdateCredentialsReq) (*model.Node, *ProbeResp, error) {
	normalizeCredentialReq(&req)
	if req.Username == "" {
		return nil, nil, errors.New("username is required")
	}

	node, err := s.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	backup := *node

	probeReq := ProbeReq{
		Name:     node.Name,
		IP:       node.IP,
		Port:     req.Port,
		AuthType: req.AuthType,
		Username: req.Username,
		Password: req.Password,
		SSHKeyID: req.SSHKeyID,
	}
	resp, err := s.Probe(ctx, 0, probeReq)
	if err != nil {
		return nil, nil, err
	}
	if !resp.Reachable {
		return &backup, resp, errors.New("credential probe failed")
	}

	node.Port = req.Port
	node.SSHUser = req.Username
	node.SSHPassword = req.Password
	node.LastCheckAt = time.Now()
	if req.SSHKeyID != nil {
		node.SSHKeyID = nodeIDPtr(*req.SSHKeyID)
	} else {
		node.SSHKeyID = nil
	}
	if err := s.svcCtx.DB.WithContext(ctx).Save(node).Error; err != nil {
		return nil, nil, err
	}
	return node, resp, nil
}

func (s *HostService) createFromLegacyReq(ctx context.Context, req CreateReq) (*model.Node, error) {
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.IP) == "" {
		return nil, errors.New("name and ip are required")
	}
	if req.Port <= 0 {
		req.Port = DefaultSSHPort
	}
	status := req.Status
	if status == "" {
		status = "offline"
	}
	node := &model.Node{
		Name:        req.Name,
		IP:          req.IP,
		Port:        req.Port,
		SSHUser:     firstNonEmpty(req.Username, "root"),
		SSHPassword: req.Password,
		Description: req.Description,
		Labels:      strings.Join(req.Labels, ","),
		Status:      status,
		Role:        req.Role,
		ClusterID:   req.ClusterID,
	}
	if req.SSHKeyID != nil {
		node.SSHKeyID = nodeIDPtr(*req.SSHKeyID)
	}
	if err := s.svcCtx.DB.WithContext(ctx).Create(node).Error; err != nil {
		return nil, err
	}
	return node, nil
}

func firstNonEmpty(v ...string) string {
	for _, item := range v {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func buildStatus(reachable bool) string {
	if reachable {
		return "online"
	}
	return "offline"
}

func nodeIDPtr(v uint64) *model.NodeID {
	n := model.NodeID(v)
	return &n
}

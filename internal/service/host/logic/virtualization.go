package logic

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/google/uuid"
)

type KVMPreviewReq struct {
	Name          string `json:"name"`
	CPU           int    `json:"cpu"`
	MemoryMB      int    `json:"memory_mb"`
	DiskGB        int    `json:"disk_gb"`
	NetworkBridge string `json:"network_bridge"`
	Template      string `json:"template"`
}

type KVMProvisionReq struct {
	Name          string  `json:"name"`
	CPU           int     `json:"cpu"`
	MemoryMB      int     `json:"memory_mb"`
	DiskGB        int     `json:"disk_gb"`
	NetworkBridge string  `json:"network_bridge"`
	Template      string  `json:"template"`
	IP            string  `json:"ip"`
	SSHUser       string  `json:"ssh_user"`
	Password      string  `json:"password"`
	SSHKeyID      *uint64 `json:"ssh_key_id"`
}

func (s *HostService) KVMPreview(ctx context.Context, hostID uint64, req KVMPreviewReq) (map[string]any, error) {
	host, err := s.Get(ctx, hostID)
	if err != nil {
		return nil, err
	}
	if host.Status == "offline" {
		return nil, errors.New("host is offline")
	}
	if req.Name == "" {
		return nil, errors.New("name is required")
	}
	if req.CPU <= 0 {
		req.CPU = 2
	}
	if req.MemoryMB <= 0 {
		req.MemoryMB = 4096
	}
	if req.DiskGB <= 0 {
		req.DiskGB = 50
	}
	return map[string]any{
		"host_id":    hostID,
		"hypervisor": "kvm",
		"ready":      true,
		"preview":    req,
		"message":    "mvp preview only, libvirt execution can be attached later",
	}, nil
}

func (s *HostService) KVMProvision(ctx context.Context, uid uint64, hostID uint64, req KVMProvisionReq) (*model.HostVirtualizationTask, *model.Node, error) {
	if req.Name == "" {
		return nil, nil, errors.New("name is required")
	}
	host, err := s.Get(ctx, hostID)
	if err != nil {
		return nil, nil, err
	}
	task := &model.HostVirtualizationTask{
		ID:         uuid.NewString(),
		HostID:     hostID,
		Hypervisor: "kvm",
		VMName:     req.Name,
		VMIP:       req.IP,
		Status:     "running",
		CreatedBy:  uid,
	}
	rawReq, _ := json.Marshal(req)
	task.RequestJSON = string(rawReq)
	if err := s.svcCtx.DB.WithContext(ctx).Create(task).Error; err != nil {
		return nil, nil, err
	}

	node := &model.Node{
		Name:         req.Name,
		IP:           req.IP,
		Port:         DefaultSSHPort,
		SSHUser:      firstNonEmpty(req.SSHUser, "root"),
		SSHPassword:  req.Password,
		Status:       "online",
		Source:       "kvm_provision",
		Provider:     "kvm",
		ProviderID:   task.ID,
		ParentHostID: &host.ID,
		OS:           host.OS,
		CpuCores:     req.CPU,
		MemoryMB:     req.MemoryMB,
		DiskGB:       req.DiskGB,
		LastCheckAt:  time.Now(),
	}
	if req.SSHKeyID != nil {
		node.SSHKeyID = nodeIDPtr(*req.SSHKeyID)
	}
	if err := s.svcCtx.DB.WithContext(ctx).Create(node).Error; err != nil {
		task.Status = "failed"
		task.ErrorMessage = err.Error()
		_ = s.svcCtx.DB.WithContext(ctx).Save(task).Error
		return task, nil, err
	}
	task.Status = "success"
	if err := s.svcCtx.DB.WithContext(ctx).Save(task).Error; err != nil {
		return nil, nil, err
	}
	return task, node, nil
}

func (s *HostService) GetVirtualizationTask(ctx context.Context, taskID string) (*model.HostVirtualizationTask, error) {
	var task model.HostVirtualizationTask
	if err := s.svcCtx.DB.WithContext(ctx).Where("id = ?", taskID).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cy77cc/k8s-manage/internal/config"
	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/utils"
	"github.com/google/uuid"
)

type CloudAccountReq struct {
	Provider        string `json:"provider"`
	AccountName     string `json:"account_name"`
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	RegionDefault   string `json:"region_default"`
}

type CloudQueryReq struct {
	Provider  string `json:"provider"`
	AccountID uint64 `json:"account_id"`
	Region    string `json:"region"`
	Keyword   string `json:"keyword"`
}

type CloudImportReq struct {
	Provider  string          `json:"provider"`
	AccountID uint64          `json:"account_id"`
	Instances []CloudInstance `json:"instances"`
	Role      string          `json:"role"`
	Labels    []string        `json:"labels"`
}

type CloudInstance struct {
	InstanceID string `json:"instance_id"`
	Name       string `json:"name"`
	IP         string `json:"ip"`
	Region     string `json:"region"`
	Status     string `json:"status"`
	OS         string `json:"os"`
	CPU        int    `json:"cpu"`
	MemoryMB   int    `json:"memory_mb"`
	DiskGB     int    `json:"disk_gb"`
}

func (s *HostService) CreateCloudAccount(ctx context.Context, uid uint64, req CloudAccountReq) (*model.HostCloudAccount, error) {
	if strings.TrimSpace(config.CFG.Security.EncryptionKey) == "" {
		return nil, errors.New("security.encryption_key is required")
	}
	if req.Provider == "" || req.AccountName == "" || req.AccessKeyID == "" || req.AccessKeySecret == "" {
		return nil, errors.New("provider/account_name/access_key_id/access_key_secret are required")
	}
	secretEnc, err := utils.EncryptText(req.AccessKeySecret, config.CFG.Security.EncryptionKey)
	if err != nil {
		return nil, err
	}
	acc := &model.HostCloudAccount{
		Provider:           req.Provider,
		AccountName:        req.AccountName,
		AccessKeyID:        req.AccessKeyID,
		AccessKeySecretEnc: secretEnc,
		RegionDefault:      req.RegionDefault,
		Status:             "active",
		CreatedBy:          uid,
	}
	if err := s.svcCtx.DB.WithContext(ctx).Create(acc).Error; err != nil {
		return nil, err
	}
	return acc, nil
}

func (s *HostService) ListCloudAccounts(ctx context.Context, provider string) ([]model.HostCloudAccount, error) {
	query := s.svcCtx.DB.WithContext(ctx).Model(&model.HostCloudAccount{}).
		Select("id", "provider", "account_name", "access_key_id", "region_default", "status", "created_by", "created_at", "updated_at")
	if provider != "" {
		query = query.Where("provider = ?", provider)
	}
	var list []model.HostCloudAccount
	if err := query.Order("id desc").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (s *HostService) TestCloudAccount(ctx context.Context, req CloudAccountReq) (map[string]any, error) {
	if req.Provider == "" || req.AccessKeyID == "" || req.AccessKeySecret == "" {
		return nil, errors.New("provider/access_key_id/access_key_secret are required")
	}
	if req.Provider != "alicloud" && req.Provider != "tencent" {
		return map[string]any{"ok": false, "message": "unsupported provider"}, nil
	}
	return map[string]any{"ok": true, "provider": req.Provider, "message": "credential format accepted (mvp mock verify)"}, nil
}

func (s *HostService) QueryCloudInstances(ctx context.Context, req CloudQueryReq) ([]CloudInstance, error) {
	if req.Provider != "alicloud" && req.Provider != "tencent" {
		return nil, errors.New("unsupported provider")
	}
	if req.AccountID == 0 {
		return nil, errors.New("account_id is required")
	}
	list := []CloudInstance{
		{InstanceID: fmt.Sprintf("%s-i-001", req.Provider), Name: req.Provider + "-web-01", IP: "10.0.10.11", Region: firstNonEmpty(req.Region, "cn-hangzhou"), Status: "running", OS: "Ubuntu 22.04", CPU: 4, MemoryMB: 8192, DiskGB: 100},
		{InstanceID: fmt.Sprintf("%s-i-002", req.Provider), Name: req.Provider + "-api-01", IP: "10.0.10.12", Region: firstNonEmpty(req.Region, "cn-hangzhou"), Status: "running", OS: "CentOS 7", CPU: 8, MemoryMB: 16384, DiskGB: 200},
	}
	if kw := strings.ToLower(strings.TrimSpace(req.Keyword)); kw != "" {
		filtered := make([]CloudInstance, 0, len(list))
		for _, item := range list {
			if strings.Contains(strings.ToLower(item.Name), kw) || strings.Contains(strings.ToLower(item.InstanceID), kw) || strings.Contains(item.IP, kw) {
				filtered = append(filtered, item)
			}
		}
		return filtered, nil
	}
	return list, nil
}

func (s *HostService) ImportCloudInstances(ctx context.Context, uid uint64, req CloudImportReq) (*model.HostImportTask, []model.Node, error) {
	if len(req.Instances) == 0 {
		return nil, nil, errors.New("instances is empty")
	}
	task := &model.HostImportTask{ID: uuid.NewString(), Provider: req.Provider, AccountID: req.AccountID, Status: "running", CreatedBy: uid}
	requestJSON, _ := json.Marshal(req)
	task.RequestJSON = string(requestJSON)
	if err := s.svcCtx.DB.WithContext(ctx).Create(task).Error; err != nil {
		return nil, nil, err
	}
	created := make([]model.Node, 0, len(req.Instances))
	for _, ins := range req.Instances {
		node := model.Node{
			Name:        ins.Name,
			IP:          ins.IP,
			Port:        DefaultSSHPort,
			SSHUser:     "root",
			Status:      "online",
			Role:        req.Role,
			Labels:      strings.Join(req.Labels, ","),
			OS:          ins.OS,
			CpuCores:    ins.CPU,
			MemoryMB:    ins.MemoryMB,
			DiskGB:      ins.DiskGB,
			Source:      "cloud_import",
			Provider:    req.Provider,
			ProviderID:  ins.InstanceID,
			LastCheckAt: time.Now(),
		}
		if err := s.svcCtx.DB.WithContext(ctx).Create(&node).Error; err != nil {
			task.Status = "failed"
			task.ErrorMessage = err.Error()
			_ = s.svcCtx.DB.WithContext(ctx).Save(task).Error
			return task, nil, err
		}
		created = append(created, node)
	}
	resultJSON, _ := json.Marshal(created)
	task.Status = "success"
	task.ResultJSON = string(resultJSON)
	if err := s.svcCtx.DB.WithContext(ctx).Save(task).Error; err != nil {
		return nil, nil, err
	}
	return task, created, nil
}

func (s *HostService) GetImportTask(ctx context.Context, taskID string) (*model.HostImportTask, error) {
	var task model.HostImportTask
	if err := s.svcCtx.DB.WithContext(ctx).Where("id = ?", taskID).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

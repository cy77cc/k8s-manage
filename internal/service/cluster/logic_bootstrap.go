package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	sshclient "github.com/cy77cc/k8s-manage/internal/client/ssh"
	"github.com/cy77cc/k8s-manage/internal/config"
	"github.com/cy77cc/k8s-manage/internal/httpx"
	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/utils"
	"github.com/gin-gonic/gin"
)

// BootstrapStep defines a step in the bootstrap workflow
type BootstrapStep struct {
	Name        string
	Hosts       []string // "control-plane", "workers", "all"
	Script      string
	Timeout     time.Duration
	Rollback    string
	OnFailure   string // "abort", "continue"
	EnvVars     map[string]string
}

// BootstrapPreviewReq represents a bootstrap preview request
type BootstrapPreviewReq struct {
	Name           string   `json:"name" binding:"required"`
	ControlPlaneID uint     `json:"control_plane_host_id" binding:"required"`
	WorkerIDs      []uint   `json:"worker_host_ids"`
	K8sVersion     string   `json:"k8s_version"`
	CNI            string   `json:"cni"`
	PodCIDR        string   `json:"pod_cidr"`
	ServiceCIDR    string   `json:"service_cidr"`
}

// BootstrapPreviewResp represents a bootstrap preview response
type BootstrapPreviewResp struct {
	Name           string   `json:"name"`
	ControlPlaneID uint     `json:"control_plane_host_id"`
	WorkerIDs      []uint   `json:"worker_host_ids"`
	K8sVersion     string   `json:"k8s_version"`
	CNI            string   `json:"cni"`
	PodCIDR        string   `json:"pod_cidr"`
	ServiceCIDR    string   `json:"service_cidr"`
	Steps          []string `json:"steps"`
	ExpectedEndpoint string `json:"expected_endpoint"`
}

// BootstrapApplyResp represents a bootstrap apply response
type BootstrapApplyResp struct {
	TaskID    string `json:"task_id"`
	Status    string `json:"status"`
	ClusterID *uint  `json:"cluster_id,omitempty"`
}

// Default bootstrap steps
var defaultBootstrapSteps = []BootstrapStep{
	{Name: "preflight", Hosts: []string{"all"}, Script: "cluster/common/preflight.sh", Timeout: 60 * time.Second, OnFailure: "abort"},
	{Name: "containerd", Hosts: []string{"all"}, Script: "cluster/common/containerd-install.sh", Timeout: 5 * time.Minute, Rollback: "cluster/common/containerd-install.sh uninstall", OnFailure: "abort"},
	{Name: "kubeadm-install", Hosts: []string{"all"}, Script: "cluster/kubeadm/v1.28/install.sh", Timeout: 3 * time.Minute, Rollback: "cluster/kubeadm/v1.28/install.sh uninstall", OnFailure: "abort"},
	{Name: "control-plane-init", Hosts: []string{"control-plane"}, Script: "cluster/kubeadm/v1.28/init.sh", Timeout: 10 * time.Minute, Rollback: "cluster/kubeadm/v1.28/reset.sh", OnFailure: "abort"},
	{Name: "cni-install", Hosts: []string{"control-plane"}, Script: "", Timeout: 3 * time.Minute, OnFailure: "continue"},
	{Name: "worker-join", Hosts: []string{"workers"}, Script: "cluster/kubeadm/v1.28/join.sh", Timeout: 5 * time.Minute, Rollback: "cluster/kubeadm/v1.28/reset.sh", OnFailure: "continue"},
	{Name: "fetch-kubeconfig", Hosts: []string{"control-plane"}, Script: "cluster/kubeadm/v1.28/fetch-kubeconfig.sh", Timeout: 30 * time.Second, OnFailure: "abort"},
	{Name: "sync-nodes", Hosts: []string{"control-plane"}, Script: "", Timeout: 30 * time.Second, OnFailure: "continue"},
}

// PreviewBootstrap previews a bootstrap operation
func (h *Handler) PreviewBootstrap(c *gin.Context) {
	var req BootstrapPreviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BindErr(c, err)
		return
	}

	// Load hosts to validate
	control, workers, err := h.loadBootstrapHosts(c.Request.Context(), req.ControlPlaneID, req.WorkerIDs)
	if err != nil {
		httpx.ServerErr(c, err)
		return
	}

	// Set defaults
	name := strings.TrimSpace(req.Name)
	k8sVersion := defaultIfEmpty(req.K8sVersion, "1.28.0")
	cni := defaultIfEmpty(req.CNI, "calico")
	podCIDR := defaultIfEmpty(req.PodCIDR, "10.244.0.0/16")
	serviceCIDR := defaultIfEmpty(req.ServiceCIDR, "10.96.0.0/12")

	// Build steps description
	steps := []string{
		fmt.Sprintf("1. 预检查所有节点 (%d 节点)", 1+len(workers)),
		"2. 安装 containerd 容器运行时",
		fmt.Sprintf("3. 安装 kubeadm/kubelet/kubectl v%s", k8sVersion),
		fmt.Sprintf("4. 初始化控制平面节点: %s (%s)", control.Name, control.IP),
		fmt.Sprintf("5. 安装 CNI 插件: %s", cni),
		fmt.Sprintf("6. 加入 Worker 节点 (%d 节点)", len(workers)),
		"7. 获取 kubeconfig 并存储",
		"8. 同步节点信息到数据库",
	}

	resp := BootstrapPreviewResp{
		Name:             name,
		ControlPlaneID:   req.ControlPlaneID,
		WorkerIDs:        req.WorkerIDs,
		K8sVersion:       k8sVersion,
		CNI:              cni,
		PodCIDR:          podCIDR,
		ServiceCIDR:      serviceCIDR,
		Steps:            steps,
		ExpectedEndpoint: fmt.Sprintf("https://%s:6443", control.IP),
	}

	httpx.OK(c, resp)
}

// ApplyBootstrap applies a bootstrap operation
func (h *Handler) ApplyBootstrap(c *gin.Context) {
	if !httpx.Authorize(c, h.svcCtx.DB, "cluster:write") {
		return
	}

	var req BootstrapPreviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BindErr(c, err)
		return
	}

	uid := httpx.UIDFromCtx(c)

	// Set defaults
	name := strings.TrimSpace(req.Name)
	k8sVersion := defaultIfEmpty(req.K8sVersion, "1.28.0")
	cni := defaultIfEmpty(req.CNI, "calico")
	podCIDR := defaultIfEmpty(req.PodCIDR, "10.244.0.0/16")
	serviceCIDR := defaultIfEmpty(req.ServiceCIDR, "10.96.0.0/12")

	// Create task record
	task := &model.ClusterBootstrapTask{
		ID:             fmt.Sprintf("boot-%d", time.Now().UnixNano()),
		Name:           name,
		ControlPlaneID: req.ControlPlaneID,
		WorkerIDsJSON:  toJSON(req.WorkerIDs),
		K8sVersion:     k8sVersion,
		CNI:            cni,
		PodCIDR:        podCIDR,
		ServiceCIDR:    serviceCIDR,
		Status:         "queued",
		CreatedBy:      uid,
	}

	if err := h.svcCtx.DB.Create(task).Error; err != nil {
		httpx.ServerErr(c, err)
		return
	}

	// Execute bootstrap asynchronously
	go h.executeBootstrap(context.Background(), task)

	httpx.OK(c, BootstrapApplyResp{
		TaskID: task.ID,
		Status: task.Status,
	})
}

// GetBootstrapTask returns bootstrap task status
func (h *Handler) GetBootstrapTask(c *gin.Context) {
	taskID := strings.TrimSpace(c.Param("task_id"))
	if taskID == "" {
		httpx.BindErr(c, nil)
		return
	}

	var task model.ClusterBootstrapTask
	if err := h.svcCtx.DB.Where("id = ?", taskID).First(&task).Error; err != nil {
		httpx.NotFound(c, "task not found")
		return
	}

	// Parse steps JSON
	var steps []BootstrapStepStatus
	if task.StepsJSON != "" {
		json.Unmarshal([]byte(task.StepsJSON), &steps)
	}

	detail := BootstrapTaskDetail{
		ID:           task.ID,
		Name:         task.Name,
		ClusterID:    task.ClusterID,
		K8sVersion:   task.K8sVersion,
		CNI:          task.CNI,
		PodCIDR:      task.PodCIDR,
		ServiceCIDR:  task.ServiceCIDR,
		Status:       task.Status,
		Steps:        steps,
		ErrorMessage: task.ErrorMessage,
		CreatedAt:    task.CreatedAt,
		UpdatedAt:    task.UpdatedAt,
	}

	httpx.OK(c, detail)
}

// executeBootstrap executes the bootstrap workflow
func (h *Handler) executeBootstrap(ctx context.Context, task *model.ClusterBootstrapTask) {
	// Update status to running
	task.Status = "running"
	h.svcCtx.DB.Save(task)

	// Load hosts
	control, workers, err := h.loadBootstrapHosts(ctx, task.ControlPlaneID, parseWorkerIDs(task.WorkerIDsJSON))
	if err != nil {
		h.failTask(ctx, task, err.Error())
		return
	}

	// Initialize steps tracking
	steps := make([]BootstrapStepStatus, len(defaultBootstrapSteps))
	stepsJSON, _ := json.Marshal(steps)
	task.StepsJSON = string(stepsJSON)
	h.svcCtx.DB.Save(task)

	// Prepare environment variables
	envVars := map[string]string{
		"K8S_VERSION":     task.K8sVersion,
		"POD_CIDR":       task.PodCIDR,
		"SERVICE_CIDR":   task.ServiceCIDR,
		"CNI":            task.CNI,
	}

	// Execute each step
	var lastErr error
	for i, step := range defaultBootstrapSteps {
		steps[i] = BootstrapStepStatus{
			Name:    step.Name,
			Status:  "running",
		}
		now := time.Now().UTC()
		steps[i].StartedAt = &now
		h.updateSteps(ctx, task, steps)

		// Determine target hosts
		var targetHosts []*model.Node
		for _, h := range step.Hosts {
			switch h {
			case "control-plane":
				targetHosts = append(targetHosts, control)
			case "workers":
				targetHosts = append(targetHosts, workers...)
			case "all":
				targetHosts = append(targetHosts, control)
				targetHosts = append(targetHosts, workers...)
			}
		}

		// Execute step
		if step.Script != "" {
			err := h.executeStepOnHosts(ctx, step, targetHosts, envVars)
			if err != nil {
				lastErr = err
				steps[i].Status = "failed"
				steps[i].Message = err.Error()
				finished := time.Now().UTC()
				steps[i].FinishedAt = &finished
				h.updateSteps(ctx, task, steps)

				// Execute rollback if defined
				if step.Rollback != "" {
					rollbackStep := BootstrapStep{
						Name:    step.Name + "-rollback",
						Script:  step.Rollback,
						Timeout: step.Timeout,
					}
					h.executeStepOnHosts(ctx, rollbackStep, targetHosts, envVars)
				}

				if step.OnFailure == "abort" {
					break
				}
			} else {
				steps[i].Status = "succeeded"
				finished := time.Now().UTC()
				steps[i].FinishedAt = &finished
				h.updateSteps(ctx, task, steps)
			}
		} else {
			// Special handling for CNI and sync steps
			if step.Name == "cni-install" {
				err := h.installCNI(ctx, control, task.CNI, task.PodCIDR)
				if err != nil {
					lastErr = err
					steps[i].Status = "failed"
					steps[i].Message = err.Error()
				} else {
					steps[i].Status = "succeeded"
				}
				finished := time.Now().UTC()
				steps[i].FinishedAt = &finished
				h.updateSteps(ctx, task, steps)
			} else if step.Name == "sync-nodes" {
				// This will be handled after we get kubeconfig
				steps[i].Status = "succeeded"
				finished := time.Now().UTC()
				steps[i].FinishedAt = &finished
				h.updateSteps(ctx, task, steps)
			}
		}
	}

	// Check if all critical steps succeeded
	if lastErr != nil {
		h.failTask(ctx, task, lastErr.Error())
		return
	}

	// Create cluster record
	cluster := &model.Cluster{
		Name:        task.Name,
		Source:      "platform_managed",
		Type:        "kubernetes",
		K8sVersion:  task.K8sVersion,
		Endpoint:    fmt.Sprintf("https://%s:6443", control.IP),
		PodCIDR:     task.PodCIDR,
		ServiceCIDR: task.ServiceCIDR,
		Status:      "active",
		AuthMethod:  "kubeconfig",
		CreatedBy:   fmt.Sprintf("%d", task.CreatedBy),
	}

	if err := h.svcCtx.DB.Create(cluster).Error; err != nil {
		h.failTask(ctx, task, fmt.Sprintf("failed to create cluster: %v", err))
		return
	}

	task.ClusterID = &cluster.ID
	task.Status = "succeeded"
	task.ResultJSON = toJSON(map[string]interface{}{
		"cluster_id": cluster.ID,
		"endpoint":   cluster.Endpoint,
	})
	h.svcCtx.DB.Save(task)
}

// loadBootstrapHosts loads control plane and worker hosts
func (h *Handler) loadBootstrapHosts(ctx context.Context, controlID uint, workerIDs []uint) (*model.Node, []*model.Node, error) {
	var control model.Node
	if err := h.svcCtx.DB.WithContext(ctx).First(&control, controlID).Error; err != nil {
		return nil, nil, fmt.Errorf("control plane host not found: %w", err)
	}
	if strings.TrimSpace(control.IP) == "" {
		return nil, nil, fmt.Errorf("control plane host missing IP")
	}

	workers := make([]*model.Node, 0, len(workerIDs))
	for _, id := range workerIDs {
		if id == 0 || id == controlID {
			continue
		}
		var row model.Node
		if err := h.svcCtx.DB.WithContext(ctx).First(&row, id).Error; err != nil {
			return nil, nil, fmt.Errorf("worker host %d not found", id)
		}
		workers = append(workers, &row)
	}

	return &control, workers, nil
}

// executeStepOnHosts executes a script step on target hosts
func (h *Handler) executeStepOnHosts(ctx context.Context, step BootstrapStep, hosts []*model.Node, envVars map[string]string) error {
	for _, host := range hosts {
		privateKey, passphrase, err := h.loadNodePrivateKey(ctx, host)
		if err != nil {
			return fmt.Errorf("failed to load SSH key for host %s: %w", host.Name, err)
		}

		password := strings.TrimSpace(host.SSHPassword)
		if strings.TrimSpace(privateKey) != "" {
			password = ""
		}

		cli, err := sshclient.NewSSHClient(host.SSHUser, password, host.IP, host.Port, privateKey, passphrase)
		if err != nil {
			return fmt.Errorf("SSH connection failed to %s: %w", host.Name, err)
		}

		// Build command with environment variables
		scriptPath := filepath.Join("script", step.Script)
		cmd := fmt.Sprintf("bash %s", scriptPath)

		// Add environment variables
		for k, v := range envVars {
			cmd = fmt.Sprintf("%s=%s %s", k, v, cmd)
		}
		for k, v := range step.EnvVars {
			cmd = fmt.Sprintf("%s=%s %s", k, v, cmd)
		}

		_, err = sshclient.RunCommand(cli, cmd)
		cli.Close()

		if err != nil {
			return fmt.Errorf("step %s failed on host %s: %w", step.Name, host.Name, err)
		}
	}

	return nil
}

// installCNI installs the CNI plugin
func (h *Handler) installCNI(ctx context.Context, control *model.Node, cni, podCIDR string) error {
	var scriptPath string
	switch strings.ToLower(cni) {
	case "calico":
		scriptPath = "script/cluster/cni/calico/v3.26/install.sh"
	case "flannel":
		scriptPath = "script/cluster/cni/flannel/v0.22/install.sh"
	case "cilium":
		scriptPath = "script/cluster/cni/cilium/v1.14/install.sh"
	default:
		return fmt.Errorf("unsupported CNI: %s", cni)
	}

	// Check if script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("CNI script not found: %s", scriptPath)
	}

	// Execute via SSH
	privateKey, passphrase, err := h.loadNodePrivateKey(ctx, control)
	if err != nil {
		return err
	}

	password := strings.TrimSpace(control.SSHPassword)
	if strings.TrimSpace(privateKey) != "" {
		password = ""
	}

	cli, err := sshclient.NewSSHClient(control.SSHUser, password, control.IP, control.Port, privateKey, passphrase)
	if err != nil {
		return err
	}
	defer cli.Close()

	cmd := fmt.Sprintf("POD_CIDR=%s bash %s", podCIDR, scriptPath)
	_, err = sshclient.RunCommand(cli, cmd)
	return err
}

// loadNodePrivateKey loads the SSH private key for a node
func (h *Handler) loadNodePrivateKey(ctx context.Context, node *model.Node) (string, string, error) {
	if node.SSHKeyID == nil || *node.SSHKeyID == 0 {
		return "", "", nil
	}

	var key model.SSHKey
	if err := h.svcCtx.DB.WithContext(ctx).First(&key, *node.SSHKeyID).Error; err != nil {
		return "", "", err
	}

	enc := strings.TrimSpace(config.CFG.Security.EncryptionKey)
	if enc != "" && key.Encrypted {
		decrypted, err := utils.DecryptText(key.PrivateKey, enc)
		if err != nil {
			return "", "", err
		}
		return decrypted, key.Passphrase, nil
	}

	return key.PrivateKey, key.Passphrase, nil
}

// failTask marks a task as failed
func (h *Handler) failTask(ctx context.Context, task *model.ClusterBootstrapTask, message string) {
	task.Status = "failed"
	task.ErrorMessage = message
	h.svcCtx.DB.WithContext(ctx).Save(task)
}

// updateSteps updates the steps JSON in the task
func (h *Handler) updateSteps(ctx context.Context, task *model.ClusterBootstrapTask, steps []BootstrapStepStatus) {
	stepsJSON, _ := json.Marshal(steps)
	task.StepsJSON = string(stepsJSON)
	h.svcCtx.DB.WithContext(ctx).Save(task)
}

// Helper functions
func defaultIfEmpty(value, defaultValue string) string {
	if strings.TrimSpace(value) == "" {
		return defaultValue
	}
	return value
}

func parseWorkerIDs(jsonStr string) []uint {
	var ids []uint
	json.Unmarshal([]byte(jsonStr), &ids)
	return ids
}

func toJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

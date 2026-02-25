package deployment

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	sshclient "github.com/cy77cc/k8s-manage/internal/client/ssh"
	"github.com/cy77cc/k8s-manage/internal/config"
	"github.com/cy77cc/k8s-manage/internal/model"
	projectlogic "github.com/cy77cc/k8s-manage/internal/service/project/logic"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/cy77cc/k8s-manage/internal/utils"
	"gorm.io/gorm"
)

type Logic struct {
	svcCtx *svc.ServiceContext
}

func NewLogic(svcCtx *svc.ServiceContext) *Logic { return &Logic{svcCtx: svcCtx} }

const (
	releaseStatusPendingApproval = "pending_approval"
	releaseStatusApproved        = "approved"
	releaseStatusRejected        = "rejected"
	releaseStatusExecuting       = "executing"
	releaseStatusSucceeded       = "succeeded"
	releaseStatusFailed          = "failed"
	releaseStatusRolledBack      = "rolled_back"
)

type releaseDiagnostic struct {
	Runtime string `json:"runtime"`
	Stage   string `json:"stage"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Summary string `json:"summary"`
}

func (l *Logic) ListTargets(ctx context.Context, projectID, teamID uint) ([]TargetResp, error) {
	q := l.svcCtx.DB.WithContext(ctx).Model(&model.DeploymentTarget{})
	if projectID > 0 {
		q = q.Where("project_id = ?", projectID)
	}
	if teamID > 0 {
		q = q.Where("team_id = ?", teamID)
	}
	var rows []model.DeploymentTarget
	if err := q.Order("id DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]TargetResp, 0, len(rows))
	for i := range rows {
		item, err := l.GetTarget(ctx, rows[i].ID)
		if err != nil {
			continue
		}
		out = append(out, item)
	}
	return out, nil
}

func (l *Logic) GetTarget(ctx context.Context, id uint) (TargetResp, error) {
	var row model.DeploymentTarget
	if err := l.svcCtx.DB.WithContext(ctx).First(&row, id).Error; err != nil {
		return TargetResp{}, err
	}
	resp := TargetResp{
		ID:          row.ID,
		Name:        row.Name,
		TargetType:  row.TargetType,
		RuntimeType: defaultIfEmpty(row.RuntimeType, row.TargetType),
		ClusterID:   row.ClusterID,
		ProjectID:   row.ProjectID,
		TeamID:      row.TeamID,
		Env:         row.Env,
		Status:      row.Status,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
	var nodes []model.DeploymentTargetNode
	if err := l.svcCtx.DB.WithContext(ctx).Where("target_id = ?", row.ID).Find(&nodes).Error; err == nil {
		resp.Nodes = make([]TargetNodeResp, 0, len(nodes))
		for _, n := range nodes {
			item := TargetNodeResp{HostID: n.HostID, Role: n.Role, Weight: n.Weight, Status: n.Status}
			var host model.Node
			if err := l.svcCtx.DB.WithContext(ctx).First(&host, n.HostID).Error; err == nil {
				item.Name = host.Name
				item.IP = host.IP
				item.Status = host.Status
			}
			resp.Nodes = append(resp.Nodes, item)
		}
	}
	return resp, nil
}

func (l *Logic) CreateTarget(ctx context.Context, uid uint64, req TargetUpsertReq) (TargetResp, error) {
	runtimeType := normalizedRuntime(req.TargetType, req.RuntimeType)
	row := model.DeploymentTarget{
		Name:        strings.TrimSpace(req.Name),
		TargetType:  runtimeType,
		RuntimeType: runtimeType,
		ClusterID:   req.ClusterID,
		ProjectID:   req.ProjectID,
		TeamID:      req.TeamID,
		Env:         defaultIfEmpty(req.Env, "staging"),
		Status:      "active",
		CreatedBy:   uint(uid),
	}
	if row.TargetType != "k8s" && row.TargetType != "compose" {
		return TargetResp{}, fmt.Errorf("unsupported target_type: %s", row.TargetType)
	}
	if err := l.validateTargetUpsert(ctx, row.TargetType, row.ClusterID, req.Nodes); err != nil {
		return TargetResp{}, err
	}
	if err := l.svcCtx.DB.WithContext(ctx).Create(&row).Error; err != nil {
		return TargetResp{}, err
	}
	if len(req.Nodes) > 0 {
		if err := l.ReplaceTargetNodes(ctx, row.ID, req.Nodes); err != nil {
			return TargetResp{}, err
		}
	}
	return l.GetTarget(ctx, row.ID)
}

func (l *Logic) UpdateTarget(ctx context.Context, id uint, req TargetUpsertReq) (TargetResp, error) {
	var row model.DeploymentTarget
	if err := l.svcCtx.DB.WithContext(ctx).First(&row, id).Error; err != nil {
		return TargetResp{}, err
	}
	if strings.TrimSpace(req.Name) != "" {
		row.Name = strings.TrimSpace(req.Name)
	}
	if strings.TrimSpace(req.TargetType) != "" {
		row.TargetType = normalizedRuntime(req.TargetType, req.RuntimeType)
		row.RuntimeType = row.TargetType
	} else if strings.TrimSpace(req.RuntimeType) != "" {
		row.RuntimeType = normalizedRuntime(row.TargetType, req.RuntimeType)
		row.TargetType = row.RuntimeType
	}
	if req.ClusterID > 0 || row.TargetType == "k8s" {
		row.ClusterID = req.ClusterID
	}
	if req.ProjectID > 0 {
		row.ProjectID = req.ProjectID
	}
	if req.TeamID > 0 {
		row.TeamID = req.TeamID
	}
	if strings.TrimSpace(req.Env) != "" {
		row.Env = req.Env
	}
	if err := l.validateTargetUpsert(ctx, row.TargetType, row.ClusterID, req.Nodes); err != nil {
		return TargetResp{}, err
	}
	if err := l.svcCtx.DB.WithContext(ctx).Save(&row).Error; err != nil {
		return TargetResp{}, err
	}
	if req.Nodes != nil {
		if err := l.ReplaceTargetNodes(ctx, row.ID, req.Nodes); err != nil {
			return TargetResp{}, err
		}
	}
	return l.GetTarget(ctx, row.ID)
}

func (l *Logic) DeleteTarget(ctx context.Context, id uint) error {
	return l.svcCtx.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("target_id = ?", id).Delete(&model.DeploymentTargetNode{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.DeploymentTarget{}, id).Error
	})
}

func (l *Logic) ReplaceTargetNodes(ctx context.Context, targetID uint, nodes []TargetNodeReq) error {
	var target model.DeploymentTarget
	if err := l.svcCtx.DB.WithContext(ctx).Select("id,target_type").First(&target, targetID).Error; err != nil {
		return err
	}
	if target.TargetType == "compose" && len(nodes) == 0 {
		return fmt.Errorf("compose target requires at least one host node")
	}
	return l.svcCtx.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("target_id = ?", targetID).Delete(&model.DeploymentTargetNode{}).Error; err != nil {
			return err
		}
		for _, n := range nodes {
			if n.HostID == 0 {
				continue
			}
			var host model.Node
			if err := tx.Select("id,ip,status").First(&host, n.HostID).Error; err != nil {
				return fmt.Errorf("host node %d not found", n.HostID)
			}
			if strings.TrimSpace(host.IP) == "" {
				return fmt.Errorf("host node %d has empty ip", n.HostID)
			}
			if target.TargetType == "compose" {
				status := strings.ToLower(strings.TrimSpace(host.Status))
				if status == "offline" || status == "error" || status == "inactive" {
					return fmt.Errorf("host node %d is unavailable: %s", n.HostID, host.Status)
				}
			}
			row := model.DeploymentTargetNode{TargetID: targetID, HostID: n.HostID, Role: defaultIfEmpty(n.Role, "worker"), Weight: defaultInt(n.Weight, 100), Status: "active"}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (l *Logic) PreviewClusterBootstrap(ctx context.Context, req ClusterBootstrapPreviewReq) (ClusterBootstrapPreviewResp, error) {
	control, workers, err := l.loadBootstrapHosts(ctx, req.ControlPlaneID, req.WorkerIDs)
	if err != nil {
		return ClusterBootstrapPreviewResp{}, err
	}
	name := strings.TrimSpace(req.Name)
	cni := defaultIfEmpty(req.CNI, "flannel")
	steps := []string{
		fmt.Sprintf("控制平面节点: %s(%s), 执行 kubeadm init --pod-network-cidr=10.244.0.0/16", control.Name, control.IP),
		fmt.Sprintf("安装 CNI: %s", cni),
		fmt.Sprintf("工作节点数量: %d, 执行 kubeadm join", len(workers)),
		"采集 kubeconfig 并注册集群",
		"自动创建 k8s 部署目标并绑定当前项目/团队",
	}
	return ClusterBootstrapPreviewResp{
		Name:             name,
		ControlPlaneID:   req.ControlPlaneID,
		WorkerHostIDs:    req.WorkerIDs,
		CNI:              cni,
		Steps:            steps,
		ExpectedEndpoint: fmt.Sprintf("https://%s:6443", control.IP),
	}, nil
}

func (l *Logic) ApplyClusterBootstrap(ctx context.Context, uid uint64, req ClusterBootstrapPreviewReq) (ClusterBootstrapApplyResp, error) {
	preview, err := l.PreviewClusterBootstrap(ctx, req)
	if err != nil {
		return ClusterBootstrapApplyResp{}, err
	}
	task := &model.ClusterBootstrapTask{
		ID:             fmt.Sprintf("boot-%d", time.Now().UnixNano()),
		Name:           preview.Name,
		ControlPlaneID: req.ControlPlaneID,
		WorkerIDsJSON:  toJSON(req.WorkerIDs),
		CNI:            preview.CNI,
		Status:         "running",
		CreatedBy:      uid,
	}
	if err := l.svcCtx.DB.WithContext(ctx).Create(task).Error; err != nil {
		return ClusterBootstrapApplyResp{}, err
	}
	control, _, err := l.loadBootstrapHosts(ctx, req.ControlPlaneID, req.WorkerIDs)
	if err != nil {
		task.Status = "failed"
		task.ErrorMessage = err.Error()
		_ = l.svcCtx.DB.WithContext(ctx).Save(task).Error
		return ClusterBootstrapApplyResp{TaskID: task.ID, Status: task.Status}, err
	}
	privateKey, passphrase, err := l.loadNodePrivateKey(ctx, control)
	if err != nil {
		task.Status = "failed"
		task.ErrorMessage = err.Error()
		_ = l.svcCtx.DB.WithContext(ctx).Save(task).Error
		return ClusterBootstrapApplyResp{TaskID: task.ID, Status: task.Status}, err
	}
	password := strings.TrimSpace(control.SSHPassword)
	if strings.TrimSpace(privateKey) != "" {
		password = ""
	}
	cli, err := sshclient.NewSSHClient(control.SSHUser, password, control.IP, control.Port, privateKey, passphrase)
	if err != nil {
		task.Status = "failed"
		task.ErrorMessage = err.Error()
		_ = l.svcCtx.DB.WithContext(ctx).Save(task).Error
		return ClusterBootstrapApplyResp{TaskID: task.ID, Status: task.Status}, err
	}
	defer cli.Close()
	preflightOut, preflightErr := sshclient.RunCommand(cli, "command -v kubeadm >/dev/null 2>&1 && command -v kubectl >/dev/null 2>&1 && echo ok")
	if preflightErr != nil {
		task.Status = "failed"
		task.ErrorMessage = fmt.Sprintf("preflight failed: %s", truncateText(preflightErr.Error(), 240))
		task.ResultJSON = toJSON(map[string]any{"preflight_output": truncateText(preflightOut, 1000)})
		_ = l.svcCtx.DB.WithContext(ctx).Save(task).Error
		return ClusterBootstrapApplyResp{TaskID: task.ID, Status: task.Status}, fmt.Errorf("%s", task.ErrorMessage)
	}

	cluster := model.Cluster{
		Name:       preview.Name,
		Endpoint:   preview.ExpectedEndpoint,
		Status:     "provisioning",
		Type:       "kubernetes",
		AuthMethod: "kubeconfig",
	}
	if err := l.svcCtx.DB.WithContext(ctx).Create(&cluster).Error; err != nil {
		task.Status = "failed"
		task.ErrorMessage = err.Error()
		_ = l.svcCtx.DB.WithContext(ctx).Save(task).Error
		return ClusterBootstrapApplyResp{TaskID: task.ID, Status: task.Status}, err
	}
	target := model.DeploymentTarget{
		Name:       fmt.Sprintf("%s-target", preview.Name),
		TargetType: "k8s",
		ClusterID:  cluster.ID,
		ProjectID:  1,
		TeamID:     1,
		Env:        "staging",
		Status:     "active",
		CreatedBy:  uint(uid),
	}
	if err := l.svcCtx.DB.WithContext(ctx).Create(&target).Error; err != nil {
		task.Status = "failed"
		task.ErrorMessage = err.Error()
		task.ResultJSON = toJSON(map[string]any{"cluster_id": cluster.ID})
		_ = l.svcCtx.DB.WithContext(ctx).Save(task).Error
		return ClusterBootstrapApplyResp{TaskID: task.ID, Status: task.Status}, err
	}

	task.Status = "succeeded"
	task.ResultJSON = toJSON(map[string]any{
		"cluster_id":           cluster.ID,
		"target_id":            target.ID,
		"next_manual_action":   "登录控制平面节点执行 kubeadm init 并回填 kubeconfig 到集群配置",
		"preflight_checked_at": time.Now().Format(time.RFC3339),
	})
	_ = l.svcCtx.DB.WithContext(ctx).Save(task).Error
	return ClusterBootstrapApplyResp{TaskID: task.ID, Status: task.Status, ClusterID: cluster.ID, TargetID: target.ID}, nil
}

func (l *Logic) GetClusterBootstrapTask(ctx context.Context, taskID string) (*model.ClusterBootstrapTask, error) {
	if strings.TrimSpace(taskID) == "" {
		return nil, fmt.Errorf("task_id is required")
	}
	var task model.ClusterBootstrapTask
	if err := l.svcCtx.DB.WithContext(ctx).Where("id = ?", taskID).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (l *Logic) loadBootstrapHosts(ctx context.Context, controlID uint, workerIDs []uint) (*model.Node, []model.Node, error) {
	var control model.Node
	if err := l.svcCtx.DB.WithContext(ctx).First(&control, controlID).Error; err != nil {
		return nil, nil, fmt.Errorf("control plane host not found: %w", err)
	}
	if strings.TrimSpace(control.IP) == "" {
		return nil, nil, fmt.Errorf("control plane host missing ip")
	}
	workers := make([]model.Node, 0, len(workerIDs))
	for _, id := range workerIDs {
		if id == 0 || id == controlID {
			continue
		}
		var row model.Node
		if err := l.svcCtx.DB.WithContext(ctx).First(&row, id).Error; err != nil {
			return nil, nil, fmt.Errorf("worker host %d not found", id)
		}
		workers = append(workers, row)
	}
	return &control, workers, nil
}

func (l *Logic) PreviewRelease(ctx context.Context, req ReleasePreviewReq) (ReleasePreviewResp, error) {
	svc, target, manifest, err := l.resolveReleaseContext(ctx, req)
	if err != nil {
		return ReleasePreviewResp{}, err
	}
	checks := []map[string]string{
		{"code": "target", "message": fmt.Sprintf("target=%s:%d", target.TargetType, target.ID), "level": "info"},
		{"code": "service", "message": fmt.Sprintf("service=%s", svc.Name), "level": "info"},
	}
	var warnings []map[string]string
	if target.TargetType == "compose" {
		if !strings.Contains(manifest, "services:") {
			warnings = append(warnings, map[string]string{"code": "compose_shape", "message": "manifest may not be valid docker compose schema", "level": "warning"})
		}
	}
	return ReleasePreviewResp{
		ResolvedManifest: manifest,
		Checks:           checks,
		Warnings:         warnings,
		Runtime:          target.TargetType,
	}, nil
}

func (l *Logic) ApplyRelease(ctx context.Context, uid uint64, req ReleasePreviewReq) (ReleaseApplyResp, error) {
	svc, target, manifest, err := l.resolveReleaseContext(ctx, req)
	if err != nil {
		return ReleaseApplyResp{}, err
	}
	release := &model.DeploymentRelease{
		ServiceID:          svc.ID,
		TargetID:           target.ID,
		NamespaceOrProject: defaultIfEmpty(req.Env, svc.Env),
		RuntimeType:        target.TargetType,
		Strategy:           defaultIfEmpty(req.Strategy, "rolling"),
		RevisionID:         svc.LastRevisionID,
		Status:             releaseStatusPendingApproval,
		ManifestSnapshot:   manifest,
		RuntimeContextJSON: toJSON(map[string]any{
			"runtime":   target.TargetType,
			"target_id": target.ID,
			"env":       defaultIfEmpty(req.Env, svc.Env),
			"service":   svc.Name,
		}),
		ChecksJSON:       "[]",
		WarningsJSON:     "[]",
		DiagnosticsJSON:  "[]",
		VerificationJSON: "{}",
		Operator:         uint(uid),
	}
	if err := l.svcCtx.DB.WithContext(ctx).Create(release).Error; err != nil {
		return ReleaseApplyResp{}, err
	}

	release.Status = releaseStatusApproved
	_ = l.svcCtx.DB.WithContext(ctx).Save(release).Error
	release.Status = releaseStatusExecuting
	_ = l.svcCtx.DB.WithContext(ctx).Save(release).Error
	switch target.TargetType {
	case "k8s":
		var cluster model.Cluster
		if err := l.svcCtx.DB.WithContext(ctx).First(&cluster, target.ClusterID).Error; err != nil {
			release.Status = releaseStatusFailed
			release.DiagnosticsJSON = toJSON([]releaseDiagnostic{{
				Runtime: "k8s", Stage: "validate", Code: "cluster_not_found", Message: err.Error(), Summary: "cluster binding not found",
			}})
			_ = l.svcCtx.DB.WithContext(ctx).Save(release).Error
			return ReleaseApplyResp{ReleaseID: release.ID, Status: release.Status, RuntimeType: release.RuntimeType}, err
		}
		if err := projectlogic.DeployToCluster(ctx, &cluster, manifest); err != nil {
			release.Status = releaseStatusFailed
			release.DiagnosticsJSON = toJSON([]releaseDiagnostic{{
				Runtime: "k8s", Stage: "execute", Code: "deploy_failed", Message: err.Error(), Summary: "k8s runtime apply failed",
			}})
			_ = l.svcCtx.DB.WithContext(ctx).Save(release).Error
			return ReleaseApplyResp{ReleaseID: release.ID, Status: release.Status, RuntimeType: release.RuntimeType}, err
		}
		release.VerificationJSON = toJSON(map[string]any{"runtime": "k8s", "checks": []string{"apply_succeeded"}, "passed": true})
		release.Status = releaseStatusSucceeded
	default:
		out, execErr := l.applyComposeRelease(ctx, target, release.ID, manifest)
		if execErr != nil {
			release.Status = releaseStatusFailed
			release.WarningsJSON = toJSON([]map[string]string{{"code": "compose_apply_failed", "message": truncateText(out, 1200), "level": "warning"}})
			release.DiagnosticsJSON = toJSON([]releaseDiagnostic{{
				Runtime: "compose", Stage: "execute", Code: "compose_apply_failed", Message: truncateText(execErr.Error(), 500), Summary: truncateText(out, 800),
			}})
			_ = l.svcCtx.DB.WithContext(ctx).Save(release).Error
			return ReleaseApplyResp{ReleaseID: release.ID, Status: release.Status, RuntimeType: release.RuntimeType}, execErr
		}
		release.Status = releaseStatusSucceeded
		release.VerificationJSON = toJSON(map[string]any{"runtime": "compose", "checks": []string{"docker_compose_ps"}, "passed": true})
		release.ChecksJSON = toJSON([]map[string]string{{"code": "compose_ps", "message": truncateText(out, 1200), "level": "info"}})
	}
	_ = l.svcCtx.DB.WithContext(ctx).Save(release).Error
	return ReleaseApplyResp{ReleaseID: release.ID, Status: release.Status, RuntimeType: release.RuntimeType}, nil
}

func (l *Logic) RollbackRelease(ctx context.Context, id uint, uid uint64) (ReleaseApplyResp, error) {
	var current model.DeploymentRelease
	if err := l.svcCtx.DB.WithContext(ctx).First(&current, id).Error; err != nil {
		return ReleaseApplyResp{}, err
	}
	var prev model.DeploymentRelease
	if err := l.svcCtx.DB.WithContext(ctx).
		Where("service_id = ? AND target_id = ? AND id < ?", current.ServiceID, current.TargetID, current.ID).
		Order("id DESC").First(&prev).Error; err != nil {
		return ReleaseApplyResp{}, fmt.Errorf("no previous release to rollback")
	}
	rollback := &model.DeploymentRelease{
		ServiceID:          current.ServiceID,
		TargetID:           current.TargetID,
		NamespaceOrProject: current.NamespaceOrProject,
		RuntimeType:        current.RuntimeType,
		Strategy:           "rollback",
		RevisionID:         prev.RevisionID,
		SourceReleaseID:    current.ID,
		TargetRevision:     fmt.Sprintf("%d", prev.RevisionID),
		Status:             releaseStatusExecuting,
		ManifestSnapshot:   prev.ManifestSnapshot,
		RuntimeContextJSON: toJSON(map[string]any{"runtime": current.RuntimeType, "rollback_from": current.ID}),
		ChecksJSON:         "[]",
		WarningsJSON:       "[]",
		DiagnosticsJSON:    "[]",
		VerificationJSON:   "{}",
		Operator:           uint(uid),
	}
	if err := l.svcCtx.DB.WithContext(ctx).Create(rollback).Error; err != nil {
		return ReleaseApplyResp{}, err
	}
	switch current.RuntimeType {
	case "k8s":
		var target model.DeploymentTarget
		if err := l.svcCtx.DB.WithContext(ctx).First(&target, current.TargetID).Error; err != nil {
			rollback.Status = releaseStatusFailed
			rollback.DiagnosticsJSON = toJSON([]releaseDiagnostic{{Runtime: "k8s", Stage: "rollback", Code: "target_not_found", Message: err.Error(), Summary: "rollback target missing"}})
			_ = l.svcCtx.DB.WithContext(ctx).Save(rollback).Error
			return ReleaseApplyResp{ReleaseID: rollback.ID, Status: rollback.Status, RuntimeType: rollback.RuntimeType}, err
		}
		var cluster model.Cluster
		if err := l.svcCtx.DB.WithContext(ctx).First(&cluster, target.ClusterID).Error; err != nil {
			rollback.Status = releaseStatusFailed
			rollback.DiagnosticsJSON = toJSON([]releaseDiagnostic{{Runtime: "k8s", Stage: "rollback", Code: "cluster_not_found", Message: err.Error(), Summary: "rollback cluster missing"}})
			_ = l.svcCtx.DB.WithContext(ctx).Save(rollback).Error
			return ReleaseApplyResp{ReleaseID: rollback.ID, Status: rollback.Status, RuntimeType: rollback.RuntimeType}, err
		}
		if err := projectlogic.DeployToCluster(ctx, &cluster, prev.ManifestSnapshot); err != nil {
			rollback.Status = releaseStatusFailed
			rollback.DiagnosticsJSON = toJSON([]releaseDiagnostic{{Runtime: "k8s", Stage: "rollback", Code: "rollback_apply_failed", Message: err.Error(), Summary: "k8s rollback apply failed"}})
			_ = l.svcCtx.DB.WithContext(ctx).Save(rollback).Error
			return ReleaseApplyResp{ReleaseID: rollback.ID, Status: rollback.Status, RuntimeType: rollback.RuntimeType}, err
		}
	case "compose":
		var target model.DeploymentTarget
		if err := l.svcCtx.DB.WithContext(ctx).First(&target, current.TargetID).Error; err != nil {
			rollback.Status = releaseStatusFailed
			rollback.DiagnosticsJSON = toJSON([]releaseDiagnostic{{Runtime: "compose", Stage: "rollback", Code: "target_not_found", Message: err.Error(), Summary: "rollback target missing"}})
			_ = l.svcCtx.DB.WithContext(ctx).Save(rollback).Error
			return ReleaseApplyResp{ReleaseID: rollback.ID, Status: rollback.Status, RuntimeType: rollback.RuntimeType}, err
		}
		out, err := l.applyComposeRelease(ctx, &target, rollback.ID, prev.ManifestSnapshot)
		if err != nil {
			rollback.Status = releaseStatusFailed
			rollback.DiagnosticsJSON = toJSON([]releaseDiagnostic{{Runtime: "compose", Stage: "rollback", Code: "rollback_apply_failed", Message: err.Error(), Summary: truncateText(out, 800)}})
			_ = l.svcCtx.DB.WithContext(ctx).Save(rollback).Error
			return ReleaseApplyResp{ReleaseID: rollback.ID, Status: rollback.Status, RuntimeType: rollback.RuntimeType}, err
		}
		rollback.ChecksJSON = toJSON([]map[string]string{{"code": "compose_rollback_ps", "message": truncateText(out, 1200), "level": "info"}})
	default:
		rollback.Status = releaseStatusRejected
		rollback.DiagnosticsJSON = toJSON([]releaseDiagnostic{{Runtime: current.RuntimeType, Stage: "rollback", Code: "runtime_not_supported", Message: "unsupported runtime", Summary: "rollback rejected"}})
		_ = l.svcCtx.DB.WithContext(ctx).Save(rollback).Error
		return ReleaseApplyResp{ReleaseID: rollback.ID, Status: rollback.Status, RuntimeType: rollback.RuntimeType}, fmt.Errorf("unsupported runtime: %s", current.RuntimeType)
	}
	rollback.Status = releaseStatusRolledBack
	rollback.VerificationJSON = toJSON(map[string]any{"runtime": current.RuntimeType, "rollback_succeeded": true})
	_ = l.svcCtx.DB.WithContext(ctx).Save(rollback).Error
	return ReleaseApplyResp{ReleaseID: rollback.ID, Status: rollback.Status, RuntimeType: rollback.RuntimeType}, nil
}

func (l *Logic) ListReleases(ctx context.Context, serviceID, targetID uint, runtimeType string) ([]model.DeploymentRelease, error) {
	q := l.svcCtx.DB.WithContext(ctx).Model(&model.DeploymentRelease{})
	if serviceID > 0 {
		q = q.Where("service_id = ?", serviceID)
	}
	if targetID > 0 {
		q = q.Where("target_id = ?", targetID)
	}
	if runtime := strings.TrimSpace(runtimeType); runtime != "" {
		q = q.Where("runtime_type = ?", runtime)
	}
	var rows []model.DeploymentRelease
	if err := q.Order("id DESC").Limit(200).Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (l *Logic) GetRelease(ctx context.Context, id uint) (*model.DeploymentRelease, error) {
	var row model.DeploymentRelease
	if err := l.svcCtx.DB.WithContext(ctx).First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (l *Logic) resolveReleaseContext(ctx context.Context, req ReleasePreviewReq) (*model.Service, *model.DeploymentTarget, string, error) {
	var svc model.Service
	if err := l.svcCtx.DB.WithContext(ctx).First(&svc, req.ServiceID).Error; err != nil {
		return nil, nil, "", err
	}
	var target model.DeploymentTarget
	if err := l.svcCtx.DB.WithContext(ctx).First(&target, req.TargetID).Error; err != nil {
		return nil, nil, "", err
	}
	if target.TargetType != "k8s" && target.TargetType != "compose" {
		return nil, nil, "", fmt.Errorf("unsupported runtime target")
	}
	if target.TargetType == "k8s" && target.ClusterID == 0 {
		return nil, nil, "", fmt.Errorf("k8s target missing cluster binding")
	}
	if target.TargetType == "compose" {
		var cnt int64
		if err := l.svcCtx.DB.WithContext(ctx).Model(&model.DeploymentTargetNode{}).
			Where("target_id = ? AND status = ?", target.ID, "active").Count(&cnt).Error; err != nil {
			return nil, nil, "", err
		}
		if cnt == 0 {
			return nil, nil, "", fmt.Errorf("compose target has no active host node")
		}
	}
	manifest := strings.TrimSpace(defaultIfEmpty(svc.CustomYAML, svc.YamlContent))
	if manifest == "" {
		return nil, nil, "", fmt.Errorf("empty service manifest")
	}
	for k, v := range req.Variables {
		manifest = strings.ReplaceAll(manifest, "{{"+k+"}}", v)
	}
	if strings.Contains(manifest, "{{") && strings.Contains(manifest, "}}") {
		return nil, nil, "", fmt.Errorf("manifest contains unresolved template variables")
	}
	return &svc, &target, manifest, nil
}

func (l *Logic) GetGovernance(ctx context.Context, serviceID uint, env string) (*model.ServiceGovernancePolicy, error) {
	var row model.ServiceGovernancePolicy
	err := l.svcCtx.DB.WithContext(ctx).
		Where("service_id = ? AND env = ?", serviceID, defaultIfEmpty(env, "staging")).
		First(&row).Error
	if err != nil {
		return &model.ServiceGovernancePolicy{ServiceID: serviceID, Env: defaultIfEmpty(env, "staging")}, nil
	}
	return &row, nil
}

func (l *Logic) UpsertGovernance(ctx context.Context, uid uint64, serviceID uint, req GovernanceReq) (*model.ServiceGovernancePolicy, error) {
	env := defaultIfEmpty(req.Env, "staging")
	var row model.ServiceGovernancePolicy
	err := l.svcCtx.DB.WithContext(ctx).Where("service_id = ? AND env = ?", serviceID, env).First(&row).Error
	if err != nil {
		row = model.ServiceGovernancePolicy{ServiceID: serviceID, Env: env}
	}
	row.TrafficPolicyJSON = toJSON(req.TrafficPolicy)
	row.ResiliencePolicyJSON = toJSON(req.ResiliencePolicy)
	row.AccessPolicyJSON = toJSON(req.AccessPolicy)
	row.SLOPolicyJSON = toJSON(req.SLOPolicy)
	row.UpdatedBy = uint(uid)
	if row.ID == 0 {
		if err := l.svcCtx.DB.WithContext(ctx).Create(&row).Error; err != nil {
			return nil, err
		}
	} else if err := l.svcCtx.DB.WithContext(ctx).Save(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func defaultIfEmpty(v, d string) string {
	if strings.TrimSpace(v) == "" {
		return d
	}
	return v
}

func defaultInt(v, d int) int {
	if v <= 0 {
		return d
	}
	return v
}

func toJSON(v any) string {
	if v == nil {
		return "{}"
	}
	raw, _ := json.Marshal(v)
	return string(raw)
}

func truncateText(v string, max int) string {
	s := strings.TrimSpace(v)
	if len(s) <= max || max <= 0 {
		return s
	}
	return s[:max]
}

func (l *Logic) applyComposeRelease(ctx context.Context, target *model.DeploymentTarget, releaseID uint, manifest string) (string, error) {
	node, err := l.pickComposeNode(ctx, target.ID)
	if err != nil {
		return "", err
	}
	privateKey, passphrase, err := l.loadNodePrivateKey(ctx, node)
	if err != nil {
		return "", err
	}
	password := strings.TrimSpace(node.SSHPassword)
	if strings.TrimSpace(privateKey) != "" {
		password = ""
	}
	cli, err := sshclient.NewSSHClient(node.SSHUser, password, node.IP, node.Port, privateKey, passphrase)
	if err != nil {
		return "", err
	}
	defer cli.Close()

	workDir := fmt.Sprintf("/tmp/opspilot/releases/%d", releaseID)
	composeFile := fmt.Sprintf("%s/docker-compose.yaml", workDir)
	encoded := base64.StdEncoding.EncodeToString([]byte(manifest))
	cmd := fmt.Sprintf("command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1 && mkdir -p %s && echo '%s' | base64 -d > %s && docker compose -f %s pull && docker compose -f %s up -d && docker compose -f %s ps", workDir, encoded, composeFile, composeFile, composeFile, composeFile)
	out, err := sshclient.RunCommand(cli, cmd)
	if err != nil {
		return out, err
	}
	return out, nil
}

func (l *Logic) pickComposeNode(ctx context.Context, targetID uint) (*model.Node, error) {
	var links []model.DeploymentTargetNode
	if err := l.svcCtx.DB.WithContext(ctx).
		Where("target_id = ? AND status = ?", targetID, "active").
		Order("CASE WHEN role = 'manager' THEN 0 ELSE 1 END, id ASC").
		Find(&links).Error; err != nil {
		return nil, err
	}
	if len(links) == 0 {
		return nil, fmt.Errorf("compose target has no active nodes")
	}
	var node model.Node
	if err := l.svcCtx.DB.WithContext(ctx).First(&node, links[0].HostID).Error; err != nil {
		return nil, err
	}
	return &node, nil
}

func (l *Logic) loadNodePrivateKey(ctx context.Context, node *model.Node) (string, string, error) {
	if node == nil || node.SSHKeyID == nil {
		return "", "", nil
	}
	var key model.SSHKey
	if err := l.svcCtx.DB.WithContext(ctx).
		Select("id", "private_key", "passphrase", "encrypted").
		Where("id = ?", uint64(*node.SSHKeyID)).
		First(&key).Error; err != nil {
		return "", "", err
	}
	passphrase := strings.TrimSpace(key.Passphrase)
	if !key.Encrypted {
		return strings.TrimSpace(key.PrivateKey), passphrase, nil
	}
	if strings.TrimSpace(config.CFG.Security.EncryptionKey) == "" {
		return "", "", fmt.Errorf("security.encryption_key is required")
	}
	plain, err := utils.DecryptText(strings.TrimSpace(key.PrivateKey), config.CFG.Security.EncryptionKey)
	if err != nil {
		return "", "", err
	}
	return plain, passphrase, nil
}

func (l *Logic) validateTargetUpsert(ctx context.Context, targetType string, clusterID uint, nodes []TargetNodeReq) error {
	switch targetType {
	case "k8s":
		if clusterID == 0 {
			return fmt.Errorf("cluster_id is required for k8s target")
		}
		var cluster model.Cluster
		if err := l.svcCtx.DB.WithContext(ctx).Select("id,status").First(&cluster, clusterID).Error; err != nil {
			return fmt.Errorf("cluster binding not found: %w", err)
		}
		return nil
	case "compose":
		if clusterID != 0 {
			return fmt.Errorf("compose target must not bind cluster_id")
		}
		if nodes != nil && len(nodes) == 0 {
			return fmt.Errorf("compose target requires at least one host node")
		}
		return nil
	default:
		return fmt.Errorf("unsupported target_type: %s", targetType)
	}
}

func normalizedRuntime(targetType, runtimeType string) string {
	target := strings.TrimSpace(targetType)
	if target == "" {
		target = strings.TrimSpace(runtimeType)
	}
	if target == "" {
		return "k8s"
	}
	return target
}

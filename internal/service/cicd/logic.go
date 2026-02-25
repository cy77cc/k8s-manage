package cicd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	cicdv1 "github.com/cy77cc/k8s-manage/api/cicd/v1"
	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/service/cicd/repo"
	"github.com/cy77cc/k8s-manage/internal/svc"
)

const timelineTTL = 30 * time.Second

type Logic struct {
	svcCtx *svc.ServiceContext
	repo   *repo.Repository
}

func NewLogic(svcCtx *svc.ServiceContext) *Logic {
	return &Logic{svcCtx: svcCtx, repo: repo.New(svcCtx.DB)}
}

func (l *Logic) UpsertServiceCIConfig(ctx context.Context, uid uint, serviceID uint, req UpsertServiceCIConfigReq) (*cicdv1.ServiceCIConfigResp, error) {
	mode := normalizeTriggerMode(req.TriggerMode)
	if mode == "" {
		return nil, fmt.Errorf("trigger_mode must be one of: manual, source-event, both")
	}
	if strings.TrimSpace(req.RepoURL) == "" || strings.TrimSpace(req.ArtifactTarget) == "" {
		return nil, fmt.Errorf("repo_url and artifact_target are required")
	}
	row, err := l.repo.UpsertServiceCIConfig(ctx, model.CICDServiceCIConfig{
		ServiceID:      serviceID,
		RepoURL:        strings.TrimSpace(req.RepoURL),
		Branch:         defaultIfEmpty(strings.TrimSpace(req.Branch), "main"),
		BuildStepsJSON: mustJSON(req.BuildSteps),
		ArtifactTarget: strings.TrimSpace(req.ArtifactTarget),
		TriggerMode:    mode,
		Status:         "active",
		UpdatedBy:      uid,
	})
	if err != nil {
		return nil, err
	}
	_ = l.writeAudit(ctx, serviceID, 0, 0, "ci.config.updated", uid, map[string]any{"ci_config_id": row.ID, "trigger_mode": row.TriggerMode})
	l.invalidateTimelineCache(ctx, serviceID)
	return toServiceCIConfigResp(row), nil
}

func (l *Logic) GetServiceCIConfig(ctx context.Context, serviceID uint) (*cicdv1.ServiceCIConfigResp, error) {
	row, err := l.repo.GetServiceCIConfig(ctx, serviceID)
	if err != nil {
		return nil, err
	}
	return toServiceCIConfigResp(row), nil
}

func (l *Logic) DeleteServiceCIConfig(ctx context.Context, uid uint, serviceID uint) error {
	if err := l.repo.DeleteServiceCIConfig(ctx, serviceID); err != nil {
		return err
	}
	_ = l.writeAudit(ctx, serviceID, 0, 0, "ci.config.deleted", uid, map[string]any{"service_id": serviceID})
	l.invalidateTimelineCache(ctx, serviceID)
	return nil
}

func (l *Logic) TriggerCIRun(ctx context.Context, uid uint, serviceID uint, req TriggerCIRunReq) (*cicdv1.CIRunResp, error) {
	cfg, err := l.repo.GetServiceCIConfig(ctx, serviceID)
	if err != nil {
		return nil, fmt.Errorf("service ci config not found: %w", err)
	}
	triggerType := normalizeTriggerType(req.TriggerType)
	if triggerType == "" {
		return nil, fmt.Errorf("trigger_type must be manual or source-event")
	}
	if !allowTrigger(cfg.TriggerMode, triggerType) {
		_ = l.writeAudit(ctx, serviceID, 0, 0, "ci.run.blocked", uid, map[string]any{"trigger_type": triggerType, "reason": "trigger mode mismatch"})
		return nil, fmt.Errorf("trigger mode %s does not allow trigger type %s", cfg.TriggerMode, triggerType)
	}
	run, err := l.repo.CreateCIRun(ctx, model.CICDServiceCIRun{
		ServiceID:   serviceID,
		CIConfigID:  cfg.ID,
		TriggerType: triggerType,
		Status:      "queued",
		Reason:      strings.TrimSpace(req.Reason),
		TriggeredBy: uid,
		TriggeredAt: time.Now(),
	})
	if err != nil {
		return nil, err
	}
	_ = l.writeAudit(ctx, serviceID, 0, 0, "ci.run.queued", uid, map[string]any{"ci_run_id": run.ID, "trigger_type": run.TriggerType})
	l.invalidateTimelineCache(ctx, serviceID)
	return toCIRunResp(run), nil
}

func (l *Logic) ListCIRuns(ctx context.Context, serviceID uint) ([]cicdv1.CIRunResp, error) {
	rows, err := l.repo.ListCIRuns(ctx, serviceID)
	if err != nil {
		return nil, err
	}
	out := make([]cicdv1.CIRunResp, 0, len(rows))
	for i := range rows {
		out = append(out, *toCIRunResp(&rows[i]))
	}
	return out, nil
}

func (l *Logic) UpsertDeploymentCDConfig(ctx context.Context, uid uint, deploymentID uint, req UpsertDeploymentCDConfigReq) (*cicdv1.DeploymentCDConfigResp, error) {
	strategy := normalizeStrategy(req.Strategy)
	if strategy == "" {
		return nil, fmt.Errorf("strategy must be one of: rolling, blue-green, canary")
	}
	if strategy == "canary" {
		if _, ok := req.StrategyConfig["traffic_percent"]; !ok {
			return nil, fmt.Errorf("canary strategy requires strategy_config.traffic_percent")
		}
		if _, ok := req.StrategyConfig["steps"]; !ok {
			return nil, fmt.Errorf("canary strategy requires strategy_config.steps")
		}
	}
	row, err := l.repo.UpsertDeploymentCDConfig(ctx, model.CICDDeploymentCDConfig{
		DeploymentID:       deploymentID,
		Env:                defaultIfEmpty(strings.TrimSpace(req.Env), "staging"),
		Strategy:           strategy,
		StrategyConfigJSON: mustJSON(req.StrategyConfig),
		ApprovalRequired:   req.ApprovalRequired,
		UpdatedBy:          uid,
	})
	if err != nil {
		return nil, err
	}
	_ = l.writeAudit(ctx, 0, deploymentID, 0, "cd.config.updated", uid, map[string]any{"cd_config_id": row.ID, "env": row.Env, "strategy": row.Strategy})
	return toCDConfigResp(row), nil
}

func (l *Logic) GetDeploymentCDConfig(ctx context.Context, deploymentID uint, env string) (*cicdv1.DeploymentCDConfigResp, error) {
	row, err := l.repo.GetDeploymentCDConfig(ctx, deploymentID, strings.TrimSpace(env))
	if err != nil {
		return nil, err
	}
	return toCDConfigResp(row), nil
}

func (l *Logic) TriggerRelease(ctx context.Context, uid uint, req TriggerReleaseReq) (*cicdv1.ReleaseResp, error) {
	cfg, err := l.repo.GetDeploymentCDConfig(ctx, req.DeploymentID, strings.TrimSpace(req.Env))
	if err != nil {
		return nil, fmt.Errorf("cd config not found for deployment/env: %w", err)
	}
	now := time.Now()
	status := "executing"
	startedAt := &now
	finishedAt := &now
	if cfg.ApprovalRequired {
		status = "pending_approval"
		startedAt = nil
		finishedAt = nil
	}
	release, err := l.repo.CreateRelease(ctx, model.CICDRelease{
		ServiceID:    req.ServiceID,
		DeploymentID: req.DeploymentID,
		Env:          strings.TrimSpace(req.Env),
		Version:      strings.TrimSpace(req.Version),
		Strategy:     cfg.Strategy,
		Status:       status,
		TriggeredBy:  uid,
		StartedAt:    startedAt,
		FinishedAt:   finishedAt,
	})
	if err != nil {
		return nil, err
	}
	if !cfg.ApprovalRequired {
		release.Status = "succeeded"
		_ = l.repo.SaveRelease(ctx, release)
	}
	_ = l.writeAudit(ctx, req.ServiceID, req.DeploymentID, release.ID, "release.triggered", uid, map[string]any{
		"status":            release.Status,
		"approval_required": cfg.ApprovalRequired,
		"version":           release.Version,
	})
	l.invalidateTimelineCache(ctx, req.ServiceID)
	return toReleaseResp(release), nil
}

func (l *Logic) ApproveRelease(ctx context.Context, uid uint, releaseID uint, comment string) (*cicdv1.ReleaseResp, error) {
	row, err := l.repo.GetRelease(ctx, releaseID)
	if err != nil {
		return nil, err
	}
	if row.Status != "pending_approval" {
		return nil, fmt.Errorf("release state %s cannot be approved", row.Status)
	}
	if _, err := l.repo.CreateApproval(ctx, model.CICDReleaseApproval{ReleaseID: releaseID, ApproverID: uid, Decision: "approved", Comment: strings.TrimSpace(comment)}); err != nil {
		return nil, err
	}
	now := time.Now()
	row.ApprovedBy = uid
	row.ApprovalComment = strings.TrimSpace(comment)
	row.Status = "executing"
	row.StartedAt = &now
	if err := l.repo.SaveRelease(ctx, row); err != nil {
		return nil, err
	}
	row.Status = "succeeded"
	row.FinishedAt = &now
	if err := l.repo.SaveRelease(ctx, row); err != nil {
		return nil, err
	}
	_ = l.writeAudit(ctx, row.ServiceID, row.DeploymentID, row.ID, "release.approved", uid, map[string]any{"comment": row.ApprovalComment})
	l.invalidateTimelineCache(ctx, row.ServiceID)
	return toReleaseResp(row), nil
}

func (l *Logic) RejectRelease(ctx context.Context, uid uint, releaseID uint, comment string) (*cicdv1.ReleaseResp, error) {
	row, err := l.repo.GetRelease(ctx, releaseID)
	if err != nil {
		return nil, err
	}
	if row.Status != "pending_approval" {
		return nil, fmt.Errorf("release state %s cannot be rejected", row.Status)
	}
	if _, err := l.repo.CreateApproval(ctx, model.CICDReleaseApproval{ReleaseID: releaseID, ApproverID: uid, Decision: "rejected", Comment: strings.TrimSpace(comment)}); err != nil {
		return nil, err
	}
	now := time.Now()
	row.ApprovedBy = uid
	row.ApprovalComment = strings.TrimSpace(comment)
	row.Status = "rejected"
	row.FinishedAt = &now
	if err := l.repo.SaveRelease(ctx, row); err != nil {
		return nil, err
	}
	_ = l.writeAudit(ctx, row.ServiceID, row.DeploymentID, row.ID, "release.rejected", uid, map[string]any{"comment": row.ApprovalComment})
	l.invalidateTimelineCache(ctx, row.ServiceID)
	return toReleaseResp(row), nil
}

func (l *Logic) RollbackRelease(ctx context.Context, uid uint, releaseID uint, targetVersion, comment string) (*cicdv1.ReleaseResp, error) {
	row, err := l.repo.GetRelease(ctx, releaseID)
	if err != nil {
		return nil, err
	}
	if row.Status != "failed" && row.Status != "succeeded" && row.Status != "executing" {
		return nil, fmt.Errorf("release state %s cannot be rolled back", row.Status)
	}
	now := time.Now()
	rollback, err := l.repo.CreateRelease(ctx, model.CICDRelease{
		ServiceID:             row.ServiceID,
		DeploymentID:          row.DeploymentID,
		Env:                   row.Env,
		Version:               strings.TrimSpace(targetVersion),
		Strategy:              "rollback",
		Status:                "rolled_back",
		TriggeredBy:           uid,
		ApprovedBy:            uid,
		ApprovalComment:       strings.TrimSpace(comment),
		RollbackFromReleaseID: row.ID,
		StartedAt:             &now,
		FinishedAt:            &now,
	})
	if err != nil {
		return nil, err
	}
	_ = l.writeAudit(ctx, row.ServiceID, row.DeploymentID, rollback.ID, "release.rolled_back", uid, map[string]any{
		"from_release_id": row.ID,
		"target_version":  targetVersion,
		"comment":         comment,
	})
	l.invalidateTimelineCache(ctx, row.ServiceID)
	return toReleaseResp(rollback), nil
}

func (l *Logic) ListReleases(ctx context.Context, serviceID, deploymentID uint) ([]cicdv1.ReleaseResp, error) {
	rows, err := l.repo.ListReleases(ctx, serviceID, deploymentID)
	if err != nil {
		return nil, err
	}
	out := make([]cicdv1.ReleaseResp, 0, len(rows))
	for i := range rows {
		out = append(out, *toReleaseResp(&rows[i]))
	}
	return out, nil
}

func (l *Logic) ListApprovals(ctx context.Context, releaseID uint) ([]cicdv1.ReleaseApprovalResp, error) {
	rows, err := l.repo.ListApprovals(ctx, releaseID)
	if err != nil {
		return nil, err
	}
	out := make([]cicdv1.ReleaseApprovalResp, 0, len(rows))
	for i := range rows {
		out = append(out, cicdv1.ReleaseApprovalResp{ID: rows[i].ID, ReleaseID: rows[i].ReleaseID, ApproverID: rows[i].ApproverID, Decision: rows[i].Decision, Comment: rows[i].Comment, CreatedAt: rows[i].CreatedAt})
	}
	return out, nil
}

func (l *Logic) ServiceTimeline(ctx context.Context, serviceID uint) ([]cicdv1.ReleaseTimelineEventResp, error) {
	cacheKey := fmt.Sprintf("cicd:timeline:service:%d", serviceID)
	if l.svcCtx.Rdb != nil {
		if raw, err := l.svcCtx.Rdb.Get(ctx, cacheKey).Result(); err == nil && strings.TrimSpace(raw) != "" {
			var cached []cicdv1.ReleaseTimelineEventResp
			if jerr := json.Unmarshal([]byte(raw), &cached); jerr == nil {
				return cached, nil
			}
		}
	}
	rows, err := l.repo.ListAuditEventsByService(ctx, serviceID, 200)
	if err != nil {
		return nil, err
	}
	out := make([]cicdv1.ReleaseTimelineEventResp, 0, len(rows))
	for i := range rows {
		out = append(out, cicdv1.ReleaseTimelineEventResp{
			ID:           rows[i].ID,
			ServiceID:    rows[i].ServiceID,
			DeploymentID: rows[i].DeploymentID,
			ReleaseID:    rows[i].ReleaseID,
			EventType:    rows[i].EventType,
			ActorID:      rows[i].ActorID,
			Payload:      parseAnyJSON(rows[i].PayloadJSON),
			CreatedAt:    rows[i].CreatedAt,
		})
	}
	if l.svcCtx.Rdb != nil {
		if raw, err := json.Marshal(out); err == nil {
			_ = l.svcCtx.Rdb.Set(ctx, cacheKey, string(raw), timelineTTL).Err()
		}
	}
	return out, nil
}

func (l *Logic) invalidateTimelineCache(ctx context.Context, serviceID uint) {
	if serviceID == 0 || l.svcCtx.Rdb == nil {
		return
	}
	_ = l.svcCtx.Rdb.Del(ctx, fmt.Sprintf("cicd:timeline:service:%d", serviceID)).Err()
}

func (l *Logic) writeAudit(ctx context.Context, serviceID, deploymentID, releaseID uint, eventType string, actor uint, payload any) error {
	_, err := l.repo.CreateAuditEvent(ctx, model.CICDAuditEvent{
		ServiceID:    serviceID,
		DeploymentID: deploymentID,
		ReleaseID:    releaseID,
		EventType:    eventType,
		ActorID:      actor,
		PayloadJSON:  mustJSON(payload),
	})
	return err
}

func toServiceCIConfigResp(row *model.CICDServiceCIConfig) *cicdv1.ServiceCIConfigResp {
	return &cicdv1.ServiceCIConfigResp{
		ID:             row.ID,
		ServiceID:      row.ServiceID,
		RepoURL:        row.RepoURL,
		Branch:         row.Branch,
		BuildSteps:     parseStringSliceJSON(row.BuildStepsJSON),
		ArtifactTarget: row.ArtifactTarget,
		TriggerMode:    row.TriggerMode,
		Status:         row.Status,
		UpdatedBy:      row.UpdatedBy,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}
}

func toCIRunResp(row *model.CICDServiceCIRun) *cicdv1.CIRunResp {
	return &cicdv1.CIRunResp{ID: row.ID, ServiceID: row.ServiceID, CIConfigID: row.CIConfigID, TriggerType: row.TriggerType, Status: row.Status, Reason: row.Reason, TriggeredBy: row.TriggeredBy, TriggeredAt: row.TriggeredAt, CreatedAt: row.CreatedAt}
}

func toCDConfigResp(row *model.CICDDeploymentCDConfig) *cicdv1.DeploymentCDConfigResp {
	return &cicdv1.DeploymentCDConfigResp{ID: row.ID, DeploymentID: row.DeploymentID, Env: row.Env, Strategy: row.Strategy, StrategyConfig: parseMapJSON(row.StrategyConfigJSON), ApprovalRequired: row.ApprovalRequired, UpdatedBy: row.UpdatedBy, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
}

func toReleaseResp(row *model.CICDRelease) *cicdv1.ReleaseResp {
	return &cicdv1.ReleaseResp{ID: row.ID, ServiceID: row.ServiceID, DeploymentID: row.DeploymentID, Env: row.Env, Version: row.Version, Strategy: row.Strategy, Status: row.Status, TriggeredBy: row.TriggeredBy, ApprovedBy: row.ApprovedBy, ApprovalComment: row.ApprovalComment, RollbackFromReleaseID: row.RollbackFromReleaseID, StartedAt: row.StartedAt, FinishedAt: row.FinishedAt, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
}

func normalizeTriggerMode(mode string) string {
	switch strings.TrimSpace(mode) {
	case "manual", "source-event", "both":
		return strings.TrimSpace(mode)
	default:
		return ""
	}
}

func normalizeTriggerType(triggerType string) string {
	switch strings.TrimSpace(triggerType) {
	case "manual", "source-event":
		return strings.TrimSpace(triggerType)
	default:
		return ""
	}
}

func allowTrigger(mode, triggerType string) bool {
	if mode == "both" {
		return true
	}
	return mode == triggerType
}

func normalizeStrategy(strategy string) string {
	s := strings.TrimSpace(strategy)
	switch s {
	case "rolling", "blue-green", "canary":
		return s
	default:
		return ""
	}
}

func parseStringSliceJSON(raw string) []string {
	out := make([]string, 0)
	_ = json.Unmarshal([]byte(raw), &out)
	return out
}

func parseMapJSON(raw string) map[string]any {
	out := map[string]any{}
	_ = json.Unmarshal([]byte(raw), &out)
	return out
}

func parseAnyJSON(raw string) any {
	var out any
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return map[string]any{}
	}
	return out
}

func defaultIfEmpty(v, d string) string {
	if strings.TrimSpace(v) == "" {
		return d
	}
	return v
}

func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

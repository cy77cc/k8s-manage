package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	sshclient "github.com/cy77cc/k8s-manage/internal/client/ssh"
	"github.com/cy77cc/k8s-manage/internal/config"
	"github.com/cy77cc/k8s-manage/internal/model"
	projectlogic "github.com/cy77cc/k8s-manage/internal/service/project/logic"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/cy77cc/k8s-manage/internal/utils"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

type Logic struct {
	svcCtx *svc.ServiceContext
}

func NewLogic(svcCtx *svc.ServiceContext) *Logic { return &Logic{svcCtx: svcCtx} }

func (l *Logic) Preview(req RenderPreviewReq) (RenderPreviewResp, error) {
	req.Variables = normalizeStringMap(req.Variables)
	if req.Mode == "custom" {
		diagnostics := validateCustomYAML(req.Target, req.CustomYAML)
		resolved, unresolved := resolveTemplateVars(req.CustomYAML, req.Variables, nil)
		return RenderPreviewResp{
			RenderedYAML:   req.CustomYAML,
			ResolvedYAML:   resolved,
			Diagnostics:    diagnostics,
			UnresolvedVars: unresolved,
			DetectedVars:   detectTemplateVars(req.CustomYAML),
		}, nil
	}
	resp, err := renderFromStandard(req.ServiceName, req.ServiceType, req.Target, req.StandardConfig)
	if err != nil {
		return RenderPreviewResp{}, err
	}
	resp.DetectedVars = detectTemplateVars(resp.RenderedYAML)
	resp.ResolvedYAML, resp.UnresolvedVars = resolveTemplateVars(resp.RenderedYAML, req.Variables, nil)
	resp.ASTSummary = map[string]any{
		"target": req.Target,
		"docs":   strings.Count(resp.RenderedYAML, "\n---\n") + 1,
	}
	return resp, nil
}

func (l *Logic) Transform(req TransformReq) (TransformResp, error) {
	res, err := renderFromStandard(req.ServiceName, req.ServiceType, req.Target, req.StandardConfig)
	if err != nil {
		return TransformResp{}, err
	}
	return TransformResp{
		CustomYAML:   res.RenderedYAML,
		SourceHash:   sourceHash(res.RenderedYAML),
		DetectedVars: detectTemplateVars(res.RenderedYAML),
	}, nil
}

func (l *Logic) Create(ctx context.Context, uid uint64, req ServiceCreateReq) (ServiceListItem, error) {
	normalized, rendered, err := l.normalizeAndRender(req)
	if err != nil {
		return ServiceListItem{}, err
	}
	cfg := ensureStandardConfig(normalized.StandardConfig)
	labelsJSON, _ := json.Marshal(normalized.Labels)
	standardJSON, _ := json.Marshal(cfg)
	service := &model.Service{
		ProjectID:             normalized.ProjectID,
		TeamID:                normalized.TeamID,
		OwnerUserID:           uint(uid),
		Owner:                 normalized.Owner,
		Env:                   normalized.Env,
		RuntimeType:           normalized.RuntimeType,
		ConfigMode:            normalized.ConfigMode,
		ServiceKind:           normalized.ServiceKind,
		RenderTarget:          normalized.RenderTarget,
		LabelsJSON:            string(labelsJSON),
		StandardJSON:          string(standardJSON),
		CustomYAML:            normalized.CustomYAML,
		TemplateVer:           defaultIfEmpty(normalized.SourceTemplateV, "v1"),
		TemplateEngineVersion: "v1",
		Status:                defaultIfEmpty(normalized.Status, "draft"),
		Name:                  normalized.Name,
		Type:                  normalized.ServiceType,
		Image:                 cfg.Image,
		Replicas:              cfg.Replicas,
		ServicePort:           cfg.Ports[0].ServicePort,
		ContainerPort:         cfg.Ports[0].ContainerPort,
		EnvVars:               buildLegacyEnvs(cfg),
		Resources:             buildLegacyResources(cfg),
		YamlContent:           rendered,
	}
	if err := l.svcCtx.DB.WithContext(ctx).Create(service).Error; err != nil {
		return ServiceListItem{}, err
	}
	if _, err := l.createRevisionRecord(ctx, service, uint(uid), nil); err != nil {
		return ServiceListItem{}, err
	}
	return toServiceListItem(service), nil
}

func (l *Logic) Update(ctx context.Context, id uint, req ServiceCreateReq) (ServiceListItem, error) {
	var existing model.Service
	if err := l.svcCtx.DB.WithContext(ctx).First(&existing, id).Error; err != nil {
		return ServiceListItem{}, err
	}
	if req.ProjectID == 0 {
		req.ProjectID = existing.ProjectID
	}
	if req.TeamID == 0 {
		req.TeamID = existing.TeamID
	}
	if strings.TrimSpace(req.Name) == "" {
		req.Name = existing.Name
	}
	if strings.TrimSpace(req.Env) == "" {
		req.Env = existing.Env
	}
	if strings.TrimSpace(req.Owner) == "" {
		req.Owner = existing.Owner
	}
	if strings.TrimSpace(req.RuntimeType) == "" {
		req.RuntimeType = existing.RuntimeType
	}
	if strings.TrimSpace(req.ConfigMode) == "" {
		req.ConfigMode = existing.ConfigMode
	}
	if strings.TrimSpace(req.ServiceType) == "" {
		req.ServiceType = existing.Type
	}
	if strings.TrimSpace(req.RenderTarget) == "" {
		req.RenderTarget = existing.RenderTarget
	}
	if req.StandardConfig == nil && existing.StandardJSON != "" {
		var c StandardServiceConfig
		if err := json.Unmarshal([]byte(existing.StandardJSON), &c); err == nil {
			req.StandardConfig = &c
		}
	}
	if req.CustomYAML == "" {
		req.CustomYAML = existing.CustomYAML
	}

	normalized, rendered, err := l.normalizeAndRender(req)
	if err != nil {
		return ServiceListItem{}, err
	}
	cfg := ensureStandardConfig(normalized.StandardConfig)
	labelsJSON, _ := json.Marshal(normalized.Labels)
	standardJSON, _ := json.Marshal(cfg)

	existing.ProjectID = normalized.ProjectID
	existing.TeamID = normalized.TeamID
	existing.Owner = normalized.Owner
	existing.Env = normalized.Env
	existing.RuntimeType = normalized.RuntimeType
	existing.ConfigMode = normalized.ConfigMode
	existing.ServiceKind = normalized.ServiceKind
	existing.RenderTarget = normalized.RenderTarget
	existing.LabelsJSON = string(labelsJSON)
	existing.StandardJSON = string(standardJSON)
	existing.CustomYAML = normalized.CustomYAML
	existing.TemplateVer = defaultIfEmpty(normalized.SourceTemplateV, existing.TemplateVer)
	existing.Status = defaultIfEmpty(normalized.Status, existing.Status)
	existing.Name = normalized.Name
	existing.Type = normalized.ServiceType
	existing.Image = cfg.Image
	existing.Replicas = cfg.Replicas
	existing.ServicePort = cfg.Ports[0].ServicePort
	existing.ContainerPort = cfg.Ports[0].ContainerPort
	existing.EnvVars = buildLegacyEnvs(cfg)
	existing.Resources = buildLegacyResources(cfg)
	existing.YamlContent = rendered

	if err := l.svcCtx.DB.WithContext(ctx).Save(&existing).Error; err != nil {
		return ServiceListItem{}, err
	}
	if _, err := l.createRevisionRecord(ctx, &existing, existing.OwnerUserID, nil); err != nil {
		return ServiceListItem{}, err
	}
	return toServiceListItem(&existing), nil
}

func (l *Logic) List(ctx context.Context, filters map[string]string) ([]ServiceListItem, int64, error) {
	query := l.svcCtx.DB.WithContext(ctx).Model(&model.Service{})
	if v := strings.TrimSpace(filters["project_id"]); v != "" {
		query = query.Where("project_id = ?", v)
	}
	if v := strings.TrimSpace(filters["team_id"]); v != "" {
		query = query.Where("team_id = ?", v)
	}
	if v := strings.TrimSpace(filters["runtime_type"]); v != "" {
		query = query.Where("runtime_type = ?", v)
	}
	if v := strings.TrimSpace(filters["env"]); v != "" {
		query = query.Where("env = ?", v)
	}
	if v := strings.TrimSpace(filters["q"]); v != "" {
		query = query.Where("name LIKE ? OR owner LIKE ?", "%"+v+"%", "%"+v+"%")
	}
	if v := strings.TrimSpace(filters["label_selector"]); v != "" {
		parts := strings.Split(v, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			query = query.Where("labels_json LIKE ?", "%"+p+"%")
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.Service
	if err := query.Order("updated_at DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	out := make([]ServiceListItem, 0, len(list))
	for i := range list {
		out = append(out, toServiceListItem(&list[i]))
	}
	return out, total, nil
}

func (l *Logic) Get(ctx context.Context, id uint) (ServiceListItem, error) {
	var service model.Service
	if err := l.svcCtx.DB.WithContext(ctx).First(&service, id).Error; err != nil {
		return ServiceListItem{}, err
	}
	return toServiceListItem(&service), nil
}

func (l *Logic) Delete(ctx context.Context, id uint) error {
	return l.svcCtx.DB.WithContext(ctx).Delete(&model.Service{}, id).Error
}

func (l *Logic) DeployPreview(ctx context.Context, id uint, req DeployReq) (DeployPreviewResp, error) {
	var service model.Service
	if err := l.svcCtx.DB.WithContext(ctx).First(&service, id).Error; err != nil {
		return DeployPreviewResp{}, err
	}
	target, err := l.resolveDeployTarget(ctx, id, req)
	if err != nil {
		return DeployPreviewResp{}, err
	}
	resolved, unresolved, err := l.resolveServiceTemplate(ctx, &service, req.Env, req.Variables)
	if err != nil {
		return DeployPreviewResp{}, err
	}
	checks := []RenderDiagnostic{
		{Level: "info", Code: "cluster_selected", Message: fmt.Sprintf("cluster=%d namespace=%s", target.ClusterID, target.Namespace)},
	}
	warnings := make([]RenderDiagnostic, 0)
	if len(unresolved) > 0 {
		warnings = append(warnings, RenderDiagnostic{Level: "warning", Code: "unresolved_vars", Message: strings.Join(unresolved, ",")})
	}
	return DeployPreviewResp{
		ResolvedYAML: resolved,
		Checks:       checks,
		Warnings:     warnings,
		Target:       target,
	}, nil
}

func (l *Logic) Deploy(ctx context.Context, id uint, operator uint64, req DeployReq) (uint, error) {
	var service model.Service
	if err := l.svcCtx.DB.WithContext(ctx).First(&service, id).Error; err != nil {
		return 0, err
	}
	targetResp, err := l.resolveDeployTarget(ctx, id, req)
	if err != nil {
		return 0, err
	}
	target := defaultIfEmpty(req.DeployTarget, targetResp.DeployTarget)
	resolved, unresolved, err := l.resolveServiceTemplate(ctx, &service, req.Env, req.Variables)
	if err != nil {
		return 0, err
	}
	if len(unresolved) > 0 {
		return 0, fmt.Errorf("unresolved template vars: %s", strings.Join(unresolved, ","))
	}

	rec := &model.ServiceReleaseRecord{
		ServiceID:         id,
		RevisionID:        service.LastRevisionID,
		ClusterID:         targetResp.ClusterID,
		Namespace:         targetResp.Namespace,
		Env:               defaultIfEmpty(req.Env, service.Env),
		DeployTarget:      target,
		Status:            "running",
		RenderedYAML:      resolved,
		VariablesSnapshot: mustJSON(req.Variables),
		Operator:          uint(operator),
	}
	if err := l.svcCtx.DB.WithContext(ctx).Create(rec).Error; err != nil {
		return 0, err
	}

	switch target {
	case "compose":
		out, execErr := l.applyComposeByTarget(ctx, targetResp.ClusterID, rec.ID, resolved)
		if execErr != nil {
			rec.Status = "failed"
			rec.Error = truncateStr(execErr.Error(), 500)
			_ = l.svcCtx.DB.WithContext(ctx).Save(rec).Error
			return rec.ID, fmt.Errorf("compose apply failed: %s; %s", execErr.Error(), truncateStr(out, 300))
		}
		rec.Status = "succeeded"
		_ = l.svcCtx.DB.WithContext(ctx).Save(rec).Error
		return rec.ID, nil
	case "helm":
		if err := l.deployHelm(ctx, id); err != nil {
			rec.Status = "failed"
			rec.Error = err.Error()
			_ = l.svcCtx.DB.WithContext(ctx).Save(rec).Error
			return rec.ID, err
		}
		rec.Status = "succeeded"
		_ = l.svcCtx.DB.WithContext(ctx).Save(rec).Error
		return rec.ID, nil
	default:
		var cluster model.Cluster
		if err := l.svcCtx.DB.WithContext(ctx).First(&cluster, targetResp.ClusterID).Error; err != nil {
			rec.Status = "failed"
			rec.Error = err.Error()
			_ = l.svcCtx.DB.WithContext(ctx).Save(rec).Error
			return rec.ID, err
		}
		if strings.TrimSpace(resolved) == "" {
			return rec.ID, fmt.Errorf("empty rendered yaml")
		}
		if err := projectlogic.DeployToCluster(ctx, &cluster, resolved); err != nil {
			rec.Status = "failed"
			rec.Error = err.Error()
			_ = l.svcCtx.DB.WithContext(ctx).Save(rec).Error
			return rec.ID, err
		}
		rec.Status = "succeeded"
		_ = l.svcCtx.DB.WithContext(ctx).Save(rec).Error
		return rec.ID, nil
	}
}

func (l *Logic) HelmImport(ctx context.Context, uid uint64, req HelmImportReq) (*model.ServiceHelmRelease, error) {
	rec := &model.ServiceHelmRelease{
		ServiceID:    req.ServiceID,
		ChartName:    req.ChartName,
		ChartVersion: req.ChartVersion,
		ChartRef:     req.ChartRef,
		ValuesYAML:   req.ValuesYAML,
		RenderedYAML: req.RenderedYAML,
		Status:       "imported",
		CreatedBy:    uint(uid),
	}
	if err := l.svcCtx.DB.WithContext(ctx).Create(rec).Error; err != nil {
		return nil, err
	}
	return rec, nil
}

func (l *Logic) HelmRender(ctx context.Context, req HelmRenderReq) (string, []RenderDiagnostic, error) {
	diags := make([]RenderDiagnostic, 0)
	if strings.TrimSpace(req.RenderedYAML) != "" {
		return req.RenderedYAML, diags, nil
	}
	chartRef := strings.TrimSpace(req.ChartRef)
	if chartRef == "" && req.ReleaseID > 0 {
		var release model.ServiceHelmRelease
		if err := l.svcCtx.DB.WithContext(ctx).First(&release, req.ReleaseID).Error; err != nil {
			return "", nil, err
		}
		chartRef = release.ChartRef
		if req.ValuesYAML == "" {
			req.ValuesYAML = release.ValuesYAML
		}
	}
	if chartRef == "" {
		return "", []RenderDiagnostic{{Level: "error", Code: "helm_chart_ref_required", Message: "chart_ref is required"}}, errors.New("chart_ref required")
	}
	_, err := exec.LookPath("helm")
	if err != nil {
		return "", []RenderDiagnostic{{Level: "error", Code: "helm_binary_missing", Message: "helm binary not found in PATH"}}, err
	}
	valuesFile, err := os.CreateTemp("", "helm-values-*.yaml")
	if err != nil {
		return "", nil, err
	}
	defer os.Remove(valuesFile.Name())
	if _, err := valuesFile.WriteString(req.ValuesYAML); err != nil {
		_ = valuesFile.Close()
		return "", nil, err
	}
	_ = valuesFile.Close()

	ctx2, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx2, "helm", "template", defaultIfEmpty(req.ChartName, "release"), chartRef, "-f", valuesFile.Name())
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", []RenderDiagnostic{{Level: "error", Code: "helm_template_failed", Message: string(out)}}, err
	}
	return string(out), diags, nil
}

func (l *Logic) deployHelm(ctx context.Context, serviceID uint) error {
	var release model.ServiceHelmRelease
	if err := l.svcCtx.DB.WithContext(ctx).Where("service_id = ?", serviceID).Order("id DESC").First(&release).Error; err != nil {
		return err
	}
	if strings.TrimSpace(release.RenderedYAML) == "" {
		release.RenderedYAML = "# helm release imported but not rendered\n"
	}
	release.Status = "deployed"
	return l.svcCtx.DB.WithContext(ctx).Save(&release).Error
}

func (l *Logic) createRevisionRecord(ctx context.Context, service *model.Service, createdBy uint, override []TemplateVar) (*model.ServiceRevision, error) {
	var maxRevision uint
	_ = l.svcCtx.DB.WithContext(ctx).Model(&model.ServiceRevision{}).Where("service_id = ?", service.ID).Select("COALESCE(MAX(revision_no),0)").Scan(&maxRevision).Error
	schema := override
	if len(schema) == 0 {
		schema = detectTemplateVars(defaultIfEmpty(service.CustomYAML, service.YamlContent))
	}
	schemaJSON := mustJSON(schema)
	rev := &model.ServiceRevision{
		ServiceID:      service.ID,
		RevisionNo:     maxRevision + 1,
		ConfigMode:     service.ConfigMode,
		RenderTarget:   service.RenderTarget,
		StandardConfig: service.StandardJSON,
		CustomYAML:     service.CustomYAML,
		VariableSchema: schemaJSON,
		CreatedBy:      createdBy,
	}
	if err := l.svcCtx.DB.WithContext(ctx).Create(rev).Error; err != nil {
		return nil, err
	}
	service.LastRevisionID = rev.ID
	if err := l.svcCtx.DB.WithContext(ctx).Model(service).Updates(map[string]any{
		"last_revision_id":        rev.ID,
		"template_engine_version": defaultIfEmpty(service.TemplateEngineVersion, "v1"),
	}).Error; err != nil {
		return nil, err
	}
	return rev, nil
}

func toDeployTargetResp(t *model.ServiceDeployTarget) DeployTargetResp {
	resp := DeployTargetResp{
		ID:           t.ID,
		ServiceID:    t.ServiceID,
		ClusterID:    t.ClusterID,
		Namespace:    t.Namespace,
		DeployTarget: t.DeployTarget,
		IsDefault:    t.IsDefault,
		UpdatedAt:    t.UpdatedAt,
	}
	if strings.TrimSpace(t.PolicyJSON) != "" {
		_ = json.Unmarshal([]byte(t.PolicyJSON), &resp.Policy)
	}
	return resp
}

func (l *Logic) resolveDeployTarget(ctx context.Context, serviceID uint, req DeployReq) (DeployTargetResp, error) {
	if req.ClusterID > 0 {
		if strings.EqualFold(defaultIfEmpty(req.DeployTarget, "k8s"), "compose") {
			var target model.DeploymentTarget
			if err := l.svcCtx.DB.WithContext(ctx).Where("id = ? AND target_type = ?", req.ClusterID, "compose").First(&target).Error; err != nil {
				return DeployTargetResp{}, fmt.Errorf("compose deployment target not found: %w", err)
			}
		}
		return DeployTargetResp{
			ServiceID:    serviceID,
			ClusterID:    req.ClusterID,
			Namespace:    defaultIfEmpty(req.Namespace, "default"),
			DeployTarget: defaultIfEmpty(req.DeployTarget, "k8s"),
			IsDefault:    false,
		}, nil
	}
	var row model.ServiceDeployTarget
	if err := l.svcCtx.DB.WithContext(ctx).Where("service_id = ? AND is_default = 1", serviceID).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return DeployTargetResp{}, fmt.Errorf("deploy target not configured")
		}
		return DeployTargetResp{}, err
	}
	resp := toDeployTargetResp(&row)
	if strings.TrimSpace(req.Namespace) != "" {
		resp.Namespace = req.Namespace
	}
	if strings.TrimSpace(req.DeployTarget) != "" {
		resp.DeployTarget = req.DeployTarget
	}
	if strings.EqualFold(resp.DeployTarget, "compose") {
		var target model.DeploymentTarget
		if err := l.svcCtx.DB.WithContext(ctx).Where("id = ? AND target_type = ?", resp.ClusterID, "compose").First(&target).Error; err != nil {
			return DeployTargetResp{}, fmt.Errorf("compose deployment target not found: %w", err)
		}
	}
	return resp, nil
}

func (l *Logic) resolveServiceTemplate(ctx context.Context, service *model.Service, env string, reqValues map[string]string) (string, []string, error) {
	content := defaultIfEmpty(service.CustomYAML, service.YamlContent)
	if strings.TrimSpace(content) == "" {
		return "", nil, fmt.Errorf("empty service template")
	}
	envValues := map[string]string{}
	var set model.ServiceVariableSet
	err := l.svcCtx.DB.WithContext(ctx).Where("service_id = ? AND env = ?", service.ID, defaultIfEmpty(env, service.Env)).First(&set).Error
	if err == nil && strings.TrimSpace(set.ValuesJSON) != "" {
		_ = json.Unmarshal([]byte(set.ValuesJSON), &envValues)
	}
	resolved, unresolved := resolveTemplateVars(content, normalizeStringMap(reqValues), normalizeStringMap(envValues))
	return resolved, unresolved, nil
}

func mustJSON(v any) string {
	raw, _ := json.Marshal(v)
	return string(raw)
}

func normalizeStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		out[k] = strings.TrimSpace(v)
	}
	return out
}

func (l *Logic) ExtractVariables(ctx context.Context, req VariableExtractReq) (VariableExtractResp, error) {
	if strings.TrimSpace(req.CustomYAML) != "" {
		return VariableExtractResp{Vars: detectTemplateVars(req.CustomYAML)}, nil
	}
	resp, err := renderFromStandard(req.ServiceName, req.ServiceType, req.RenderTarget, req.StandardConfig)
	if err != nil {
		return VariableExtractResp{}, err
	}
	return VariableExtractResp{Vars: detectTemplateVars(resp.RenderedYAML)}, nil
}

func (l *Logic) GetVariableSchema(ctx context.Context, serviceID uint) ([]TemplateVar, error) {
	var rev model.ServiceRevision
	if err := l.svcCtx.DB.WithContext(ctx).Where("service_id = ?", serviceID).Order("revision_no DESC").First(&rev).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			var service model.Service
			if err := l.svcCtx.DB.WithContext(ctx).First(&service, serviceID).Error; err != nil {
				return nil, err
			}
			content := defaultIfEmpty(service.CustomYAML, service.YamlContent)
			return detectTemplateVars(content), nil
		}
		return nil, err
	}
	var vars []TemplateVar
	if strings.TrimSpace(rev.VariableSchema) != "" {
		_ = json.Unmarshal([]byte(rev.VariableSchema), &vars)
	}
	return vars, nil
}

func (l *Logic) GetVariableValues(ctx context.Context, serviceID uint, env string) (VariableValuesResp, error) {
	var set model.ServiceVariableSet
	err := l.svcCtx.DB.WithContext(ctx).Where("service_id = ? AND env = ?", serviceID, defaultIfEmpty(env, "staging")).First(&set).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return VariableValuesResp{ServiceID: serviceID, Env: defaultIfEmpty(env, "staging"), Values: map[string]string{}}, nil
		}
		return VariableValuesResp{}, err
	}
	out := VariableValuesResp{
		ServiceID: serviceID,
		Env:       set.Env,
		Values:    map[string]string{},
		UpdatedAt: set.UpdatedAt,
	}
	_ = json.Unmarshal([]byte(set.ValuesJSON), &out.Values)
	_ = json.Unmarshal([]byte(set.SecretKeys), &out.SecretKeys)
	return out, nil
}

func (l *Logic) UpsertVariableValues(ctx context.Context, serviceID uint, uid uint64, req VariableValuesUpsertReq) (VariableValuesResp, error) {
	env := defaultIfEmpty(req.Env, "staging")
	req.Values = normalizeStringMap(req.Values)
	valuesJSON := mustJSON(req.Values)
	secretJSON := mustJSON(req.SecretKeys)
	var set model.ServiceVariableSet
	err := l.svcCtx.DB.WithContext(ctx).Where("service_id = ? AND env = ?", serviceID, env).First(&set).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return VariableValuesResp{}, err
		}
		set = model.ServiceVariableSet{
			ServiceID:  serviceID,
			Env:        env,
			ValuesJSON: valuesJSON,
			SecretKeys: secretJSON,
			UpdatedBy:  uint(uid),
		}
		if err := l.svcCtx.DB.WithContext(ctx).Create(&set).Error; err != nil {
			return VariableValuesResp{}, err
		}
	} else {
		set.ValuesJSON = valuesJSON
		set.SecretKeys = secretJSON
		set.UpdatedBy = uint(uid)
		if err := l.svcCtx.DB.WithContext(ctx).Save(&set).Error; err != nil {
			return VariableValuesResp{}, err
		}
	}
	return l.GetVariableValues(ctx, serviceID, env)
}

func (l *Logic) ListRevisions(ctx context.Context, serviceID uint) ([]ServiceRevisionItem, error) {
	var rows []model.ServiceRevision
	if err := l.svcCtx.DB.WithContext(ctx).Where("service_id = ?", serviceID).Order("revision_no DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]ServiceRevisionItem, 0, len(rows))
	for i := range rows {
		item := ServiceRevisionItem{
			ID:           rows[i].ID,
			ServiceID:    rows[i].ServiceID,
			RevisionNo:   rows[i].RevisionNo,
			ConfigMode:   rows[i].ConfigMode,
			RenderTarget: rows[i].RenderTarget,
			CreatedBy:    rows[i].CreatedBy,
			CreatedAt:    rows[i].CreatedAt,
		}
		if strings.TrimSpace(rows[i].VariableSchema) != "" {
			_ = json.Unmarshal([]byte(rows[i].VariableSchema), &item.VariableSchema)
		}
		out = append(out, item)
	}
	return out, nil
}

func (l *Logic) CreateRevision(ctx context.Context, serviceID uint, uid uint64, req RevisionCreateReq) (ServiceRevisionItem, error) {
	var service model.Service
	if err := l.svcCtx.DB.WithContext(ctx).First(&service, serviceID).Error; err != nil {
		return ServiceRevisionItem{}, err
	}
	if strings.TrimSpace(req.ConfigMode) != "" {
		service.ConfigMode = req.ConfigMode
	}
	if strings.TrimSpace(req.RenderTarget) != "" {
		service.RenderTarget = req.RenderTarget
	}
	if req.StandardConfig != nil {
		b, _ := json.Marshal(req.StandardConfig)
		service.StandardJSON = string(b)
	}
	if strings.TrimSpace(req.CustomYAML) != "" {
		service.CustomYAML = req.CustomYAML
		service.YamlContent = req.CustomYAML
	}
	rev, err := l.createRevisionRecord(ctx, &service, uint(uid), req.VariableSchema)
	if err != nil {
		return ServiceRevisionItem{}, err
	}
	out := ServiceRevisionItem{
		ID:           rev.ID,
		ServiceID:    rev.ServiceID,
		RevisionNo:   rev.RevisionNo,
		ConfigMode:   rev.ConfigMode,
		RenderTarget: rev.RenderTarget,
		CreatedBy:    rev.CreatedBy,
		CreatedAt:    rev.CreatedAt,
	}
	if strings.TrimSpace(rev.VariableSchema) != "" {
		_ = json.Unmarshal([]byte(rev.VariableSchema), &out.VariableSchema)
	}
	return out, nil
}

func (l *Logic) UpsertDeployTarget(ctx context.Context, serviceID uint, uid uint64, req DeployTargetUpsertReq) (DeployTargetResp, error) {
	if req.ClusterID == 0 {
		return DeployTargetResp{}, fmt.Errorf("cluster_id is required")
	}
	ns := defaultIfEmpty(req.Namespace, "default")
	deployTarget := defaultIfEmpty(req.DeployTarget, "k8s")
	if deployTarget == "compose" {
		var target model.DeploymentTarget
		if err := l.svcCtx.DB.WithContext(ctx).Where("id = ? AND target_type = ?", req.ClusterID, "compose").First(&target).Error; err != nil {
			return DeployTargetResp{}, fmt.Errorf("compose deployment target not found: %w", err)
		}
	}
	policyJSON := mustJSON(req.Policy)
	var row model.ServiceDeployTarget
	err := l.svcCtx.DB.WithContext(ctx).Where("service_id = ? AND is_default = 1", serviceID).First(&row).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return DeployTargetResp{}, err
		}
		row = model.ServiceDeployTarget{
			ServiceID:    serviceID,
			ClusterID:    req.ClusterID,
			Namespace:    ns,
			DeployTarget: deployTarget,
			PolicyJSON:   policyJSON,
			IsDefault:    true,
			UpdatedBy:    uint(uid),
		}
		if err := l.svcCtx.DB.WithContext(ctx).Create(&row).Error; err != nil {
			return DeployTargetResp{}, err
		}
	} else {
		row.ClusterID = req.ClusterID
		row.Namespace = ns
		row.DeployTarget = deployTarget
		row.PolicyJSON = policyJSON
		row.UpdatedBy = uint(uid)
		if err := l.svcCtx.DB.WithContext(ctx).Save(&row).Error; err != nil {
			return DeployTargetResp{}, err
		}
	}
	if err := l.svcCtx.DB.WithContext(ctx).Model(&model.Service{}).Where("id = ?", serviceID).Update("default_target_id", row.ID).Error; err != nil {
		return DeployTargetResp{}, err
	}
	return toDeployTargetResp(&row), nil
}

func (l *Logic) ListReleaseRecords(ctx context.Context, serviceID uint) ([]ReleaseRecordItem, error) {
	var rows []model.ServiceReleaseRecord
	if err := l.svcCtx.DB.WithContext(ctx).Where("service_id = ?", serviceID).Order("id DESC").Limit(50).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]ReleaseRecordItem, 0, len(rows))
	for i := range rows {
		out = append(out, ReleaseRecordItem{
			ID:           rows[i].ID,
			ServiceID:    rows[i].ServiceID,
			RevisionID:   rows[i].RevisionID,
			ClusterID:    rows[i].ClusterID,
			Namespace:    rows[i].Namespace,
			Env:          rows[i].Env,
			DeployTarget: rows[i].DeployTarget,
			Status:       rows[i].Status,
			Error:        rows[i].Error,
			CreatedAt:    rows[i].CreatedAt,
		})
	}
	return out, nil
}

func validateCustomYAML(target, content string) []RenderDiagnostic {
	diags := make([]RenderDiagnostic, 0)
	if strings.TrimSpace(content) == "" {
		return []RenderDiagnostic{{Level: "warning", Code: "empty_yaml", Message: "custom_yaml is empty"}}
	}
	if target == "compose" {
		var obj map[string]any
		if err := yaml.Unmarshal([]byte(content), &obj); err != nil {
			diags = append(diags, RenderDiagnostic{Level: "error", Code: "invalid_compose_yaml", Message: err.Error()})
			return diags
		}
		if _, ok := obj["services"]; !ok {
			diags = append(diags, RenderDiagnostic{Level: "warning", Code: "compose_services_missing", Message: "compose yaml missing services"})
		}
		return diags
	}
	dec := yaml.NewDecoder(strings.NewReader(content))
	for {
		var obj map[string]any
		if err := dec.Decode(&obj); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			diags = append(diags, RenderDiagnostic{Level: "error", Code: "invalid_k8s_yaml", Message: err.Error()})
			break
		}
		if len(obj) == 0 {
			continue
		}
		if _, ok := obj["kind"]; !ok {
			diags = append(diags, RenderDiagnostic{Level: "warning", Code: "k8s_kind_missing", Message: "yaml doc missing kind"})
		}
	}
	return diags
}

func (l *Logic) normalizeAndRender(req ServiceCreateReq) (ServiceCreateReq, string, error) {
	r := req
	if r.ProjectID == 0 {
		if v := strings.TrimSpace(os.Getenv("DEFAULT_PROJECT_ID")); v != "" {
			if n, _ := strconv.Atoi(v); n > 0 {
				r.ProjectID = uint(n)
			}
		}
	}
	if r.ProjectID == 0 {
		r.ProjectID = 1
	}
	r.Env = defaultIfEmpty(r.Env, "staging")
	r.Owner = defaultIfEmpty(r.Owner, "system")
	r.RuntimeType = defaultIfEmpty(r.RuntimeType, "k8s")
	r.ConfigMode = defaultIfEmpty(r.ConfigMode, "standard")
	r.RenderTarget = defaultIfEmpty(r.RenderTarget, r.RuntimeType)
	r.ServiceType = defaultIfEmpty(r.ServiceType, "stateless")
	if r.Name == "" {
		r.Name = "service"
	}

	if r.ConfigMode == "custom" {
		r.CustomYAML = defaultIfEmpty(r.CustomYAML, r.YamlContent)
		if strings.TrimSpace(r.CustomYAML) == "" {
			return r, "", fmt.Errorf("custom_yaml is required when config_mode=custom")
		}
		return r, r.CustomYAML, nil
	}

	if r.StandardConfig == nil {
		r.StandardConfig = &StandardServiceConfig{
			Image:    defaultIfEmpty(r.Image, "nginx:latest"),
			Replicas: maxInt32(r.Replicas, 1),
			Ports: []PortConfig{{
				Name:          "http",
				Protocol:      "TCP",
				ContainerPort: maxInt32(r.ContainerPort, 8080),
				ServicePort:   maxInt32(r.ServicePort, 80),
			}},
			Envs:      r.EnvVars,
			Resources: r.Resources,
		}
	}
	if strings.TrimSpace(r.StandardConfig.Image) == "" {
		r.StandardConfig.Image = defaultIfEmpty(r.Image, "nginx:latest")
	}
	if len(r.StandardConfig.Ports) == 0 {
		r.StandardConfig.Ports = []PortConfig{{
			Name:          "http",
			Protocol:      "TCP",
			ContainerPort: maxInt32(r.ContainerPort, 8080),
			ServicePort:   maxInt32(r.ServicePort, 80),
		}}
	}
	resp, err := renderFromStandard(r.Name, r.ServiceType, r.RenderTarget, r.StandardConfig)
	if err != nil {
		return r, "", err
	}
	r.CustomYAML = resp.RenderedYAML
	return r, resp.RenderedYAML, nil
}

func buildLegacyEnvs(cfg *StandardServiceConfig) string {
	if cfg == nil {
		return ""
	}
	b, _ := json.Marshal(cfg.Envs)
	return string(b)
}

func buildLegacyResources(cfg *StandardServiceConfig) string {
	if cfg == nil {
		return ""
	}
	b, _ := json.Marshal(map[string]any{"limits": cfg.Resources})
	return string(b)
}

func truncateStr(v string, max int) string {
	s := strings.TrimSpace(v)
	if len(s) <= max || max <= 0 {
		return s
	}
	return s[:max]
}

func (l *Logic) applyComposeByTarget(ctx context.Context, targetID uint, releaseID uint, manifest string) (string, error) {
	if targetID == 0 {
		return "", fmt.Errorf("compose target id is required")
	}
	var links []model.DeploymentTargetNode
	if err := l.svcCtx.DB.WithContext(ctx).
		Where("target_id = ? AND status = ?", targetID, "active").
		Order("CASE WHEN role = 'manager' THEN 0 ELSE 1 END, id ASC").
		Find(&links).Error; err != nil {
		return "", err
	}
	if len(links) == 0 {
		return "", fmt.Errorf("compose target has no active nodes")
	}
	var node model.Node
	if err := l.svcCtx.DB.WithContext(ctx).First(&node, links[0].HostID).Error; err != nil {
		return "", err
	}
	privateKey, err := l.loadNodeSSHPrivateKey(ctx, &node)
	if err != nil {
		return "", err
	}
	cli, err := sshclient.NewSSHClient(node.SSHUser, node.SSHPassword, node.IP, node.Port, privateKey)
	if err != nil {
		return "", err
	}
	defer cli.Close()
	workDir := fmt.Sprintf("/tmp/opspilot/service-releases/%d", releaseID)
	composeFile := fmt.Sprintf("%s/docker-compose.yaml", workDir)
	encoded := base64.StdEncoding.EncodeToString([]byte(manifest))
	cmd := fmt.Sprintf("mkdir -p %s && echo '%s' | base64 -d > %s && docker compose -f %s pull && docker compose -f %s up -d && docker compose -f %s ps", workDir, encoded, composeFile, composeFile, composeFile, composeFile)
	return sshclient.RunCommand(cli, cmd)
}

func (l *Logic) loadNodeSSHPrivateKey(ctx context.Context, node *model.Node) (string, error) {
	if node == nil || node.SSHKeyID == nil {
		return "", nil
	}
	var key model.SSHKey
	if err := l.svcCtx.DB.WithContext(ctx).
		Select("id", "private_key", "encrypted").
		Where("id = ?", uint64(*node.SSHKeyID)).
		First(&key).Error; err != nil {
		return "", err
	}
	if !key.Encrypted {
		return strings.TrimSpace(key.PrivateKey), nil
	}
	if strings.TrimSpace(config.CFG.Security.EncryptionKey) == "" {
		return "", fmt.Errorf("security.encryption_key is required")
	}
	return utils.DecryptText(strings.TrimSpace(key.PrivateKey), config.CFG.Security.EncryptionKey)
}

func ensureStandardConfig(cfg *StandardServiceConfig) *StandardServiceConfig {
	if cfg == nil {
		cfg = &StandardServiceConfig{
			Image:    "nginx:latest",
			Replicas: 1,
			Resources: map[string]string{
				"cpu":    "500m",
				"memory": "512Mi",
			},
		}
	}
	if strings.TrimSpace(cfg.Image) == "" {
		cfg.Image = "nginx:latest"
	}
	if cfg.Replicas <= 0 {
		cfg.Replicas = 1
	}
	if len(cfg.Ports) == 0 {
		cfg.Ports = []PortConfig{{
			Name:          "http",
			Protocol:      "TCP",
			ContainerPort: 8080,
			ServicePort:   80,
		}}
	}
	return cfg
}

func toServiceListItem(s *model.Service) ServiceListItem {
	out := ServiceListItem{
		ID:                    s.ID,
		ProjectID:             s.ProjectID,
		TeamID:                s.TeamID,
		Name:                  s.Name,
		Env:                   s.Env,
		Owner:                 s.Owner,
		RuntimeType:           s.RuntimeType,
		ConfigMode:            s.ConfigMode,
		ServiceKind:           s.ServiceKind,
		Status:                s.Status,
		LastRevisionID:        s.LastRevisionID,
		DefaultTargetID:       s.DefaultTargetID,
		TemplateEngineVersion: s.TemplateEngineVersion,
		CustomYAML:            s.CustomYAML,
		RenderedYAML:          s.YamlContent,
		CreatedAt:             s.CreatedAt,
		UpdatedAt:             s.UpdatedAt,
	}
	if strings.TrimSpace(s.LabelsJSON) != "" {
		_ = json.Unmarshal([]byte(s.LabelsJSON), &out.Labels)
	}
	if strings.TrimSpace(s.StandardJSON) != "" {
		var cfg StandardServiceConfig
		if json.Unmarshal([]byte(s.StandardJSON), &cfg) == nil {
			out.StandardConfig = &cfg
		}
	}
	return out
}

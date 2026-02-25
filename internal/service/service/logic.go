package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/cy77cc/k8s-manage/internal/model"
	projectlogic "github.com/cy77cc/k8s-manage/internal/service/project/logic"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"gopkg.in/yaml.v3"
)

type Logic struct {
	svcCtx *svc.ServiceContext
}

func NewLogic(svcCtx *svc.ServiceContext) *Logic { return &Logic{svcCtx: svcCtx} }

func (l *Logic) Preview(req RenderPreviewReq) (RenderPreviewResp, error) {
	if req.Mode == "custom" {
		diagnostics := validateCustomYAML(req.Target, req.CustomYAML)
		return RenderPreviewResp{RenderedYAML: req.CustomYAML, Diagnostics: diagnostics}, nil
	}
	return renderFromStandard(req.ServiceName, req.ServiceType, req.Target, req.StandardConfig)
}

func (l *Logic) Transform(req TransformReq) (TransformResp, error) {
	res, err := renderFromStandard(req.ServiceName, req.ServiceType, req.Target, req.StandardConfig)
	if err != nil {
		return TransformResp{}, err
	}
	return TransformResp{CustomYAML: res.RenderedYAML, SourceHash: sourceHash(res.RenderedYAML)}, nil
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
		ProjectID:     normalized.ProjectID,
		TeamID:        normalized.TeamID,
		OwnerUserID:   uint(uid),
		Owner:         normalized.Owner,
		Env:           normalized.Env,
		RuntimeType:   normalized.RuntimeType,
		ConfigMode:    normalized.ConfigMode,
		ServiceKind:   normalized.ServiceKind,
		RenderTarget:  normalized.RenderTarget,
		LabelsJSON:    string(labelsJSON),
		StandardJSON:  string(standardJSON),
		CustomYAML:    normalized.CustomYAML,
		TemplateVer:   defaultIfEmpty(normalized.SourceTemplateV, "v1"),
		Status:        defaultIfEmpty(normalized.Status, "draft"),
		Name:          normalized.Name,
		Type:          normalized.ServiceType,
		Image:         cfg.Image,
		Replicas:      cfg.Replicas,
		ServicePort:   cfg.Ports[0].ServicePort,
		ContainerPort: cfg.Ports[0].ContainerPort,
		EnvVars:       buildLegacyEnvs(cfg),
		Resources:     buildLegacyResources(cfg),
		YamlContent:   rendered,
	}
	if err := l.svcCtx.DB.WithContext(ctx).Create(service).Error; err != nil {
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

func (l *Logic) Deploy(ctx context.Context, id uint, req DeployReq) error {
	var service model.Service
	if err := l.svcCtx.DB.WithContext(ctx).First(&service, id).Error; err != nil {
		return err
	}
	target := defaultIfEmpty(req.DeployTarget, service.RuntimeType)
	switch target {
	case "compose":
		return nil // MVP: only mark deploy action accepted
	case "helm":
		return l.deployHelm(ctx, id)
	default:
		if req.ClusterID == 0 {
			req.ClusterID = 1
		}
		var cluster model.Cluster
		if err := l.svcCtx.DB.WithContext(ctx).First(&cluster, req.ClusterID).Error; err != nil {
			return err
		}
		content := service.YamlContent
		if strings.TrimSpace(content) == "" {
			content = service.CustomYAML
		}
		if strings.TrimSpace(content) == "" {
			return fmt.Errorf("empty rendered yaml")
		}
		return projectlogic.DeployToCluster(ctx, &cluster, content)
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
		ID:           s.ID,
		ProjectID:    s.ProjectID,
		TeamID:       s.TeamID,
		Name:         s.Name,
		Env:          s.Env,
		Owner:        s.Owner,
		RuntimeType:  s.RuntimeType,
		ConfigMode:   s.ConfigMode,
		ServiceKind:  s.ServiceKind,
		Status:       s.Status,
		CustomYAML:   s.CustomYAML,
		RenderedYAML: s.YamlContent,
		CreatedAt:    s.CreatedAt,
		UpdatedAt:    s.UpdatedAt,
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

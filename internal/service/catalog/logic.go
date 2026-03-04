package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"gorm.io/gorm"
)

var catalogTemplateVarPattern = regexp.MustCompile(`\{\{\s*([a-zA-Z_][a-zA-Z0-9_\.\-]*)(?:\|default:([^}]+))?\s*\}\}`)

type Logic struct {
	svcCtx *svc.ServiceContext
}

func NewLogic(svcCtx *svc.ServiceContext) *Logic {
	return &Logic{svcCtx: svcCtx}
}

func (l *Logic) InitPresetCategories(ctx context.Context) error {
	type preset struct {
		Name, DisplayName, Icon string
		SortOrder               int
	}
	presets := []preset{
		{Name: "database", DisplayName: "数据库", Icon: "DatabaseOutlined", SortOrder: 10},
		{Name: "cache", DisplayName: "缓存", Icon: "ThunderboltOutlined", SortOrder: 20},
		{Name: "message-queue", DisplayName: "消息队列", Icon: "MessageOutlined", SortOrder: 30},
		{Name: "web-server", DisplayName: "Web 服务器", Icon: "GlobalOutlined", SortOrder: 40},
		{Name: "monitoring", DisplayName: "监控告警", Icon: "MonitorOutlined", SortOrder: 50},
		{Name: "dev-tools", DisplayName: "开发工具", Icon: "CodeOutlined", SortOrder: 60},
		{Name: "custom", DisplayName: "自定义服务", Icon: "AppstoreOutlined", SortOrder: 70},
	}
	for _, p := range presets {
		cat := model.ServiceCategory{}
		err := l.svcCtx.DB.WithContext(ctx).Where("name = ?", p.Name).First(&cat).Error
		if err == nil {
			continue
		}
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		if err := l.svcCtx.DB.WithContext(ctx).Create(&model.ServiceCategory{
			Name:        p.Name,
			DisplayName: p.DisplayName,
			Icon:        p.Icon,
			SortOrder:   p.SortOrder,
			IsSystem:    true,
			CreatedBy:   0,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (l *Logic) ListCategories(ctx context.Context) ([]CategoryResponse, error) {
	rows, err := model.ListServiceCategories(l.svcCtx.DB.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	out := make([]CategoryResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, toCategoryResp(row))
	}
	return out, nil
}

func (l *Logic) CreateCategory(ctx context.Context, uid uint64, req CategoryCreateRequest) (CategoryResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return CategoryResponse{}, fmt.Errorf("name is required")
	}
	item := model.ServiceCategory{}
	if err := item.GetByName(l.svcCtx.DB.WithContext(ctx), name); err == nil {
		return CategoryResponse{}, fmt.Errorf("分类名称已存在")
	} else if err != gorm.ErrRecordNotFound {
		return CategoryResponse{}, err
	}
	row := model.ServiceCategory{
		Name:        name,
		DisplayName: strings.TrimSpace(req.DisplayName),
		Icon:        strings.TrimSpace(req.Icon),
		Description: strings.TrimSpace(req.Description),
		SortOrder:   req.SortOrder,
		CreatedBy:   uid,
	}
	if row.DisplayName == "" {
		return CategoryResponse{}, fmt.Errorf("display_name is required")
	}
	if err := row.Create(l.svcCtx.DB.WithContext(ctx)); err != nil {
		return CategoryResponse{}, err
	}
	return toCategoryResp(row), nil
}

func (l *Logic) UpdateCategory(ctx context.Context, id uint, req CategoryUpdateRequest) (CategoryResponse, error) {
	row := model.ServiceCategory{}
	if err := row.GetByID(l.svcCtx.DB.WithContext(ctx), id); err != nil {
		return CategoryResponse{}, err
	}
	payload := map[string]any{}
	if req.DisplayName != nil {
		payload["display_name"] = strings.TrimSpace(*req.DisplayName)
	}
	if req.Icon != nil {
		payload["icon"] = strings.TrimSpace(*req.Icon)
	}
	if req.Description != nil {
		payload["description"] = strings.TrimSpace(*req.Description)
	}
	if req.SortOrder != nil {
		payload["sort_order"] = *req.SortOrder
	}
	if len(payload) > 0 {
		if err := row.Update(l.svcCtx.DB.WithContext(ctx), id, payload); err != nil {
			return CategoryResponse{}, err
		}
	}
	if err := row.GetByID(l.svcCtx.DB.WithContext(ctx), id); err != nil {
		return CategoryResponse{}, err
	}
	return toCategoryResp(row), nil
}

func (l *Logic) DeleteCategory(ctx context.Context, id uint) error {
	row := model.ServiceCategory{}
	if err := row.GetByID(l.svcCtx.DB.WithContext(ctx), id); err != nil {
		return err
	}
	if row.IsSystem {
		return fmt.Errorf("系统预置分类无法删除")
	}
	var cnt int64
	if err := l.svcCtx.DB.WithContext(ctx).Model(&model.ServiceTemplate{}).Where("category_id = ?", id).Count(&cnt).Error; err != nil {
		return err
	}
	if cnt > 0 {
		return fmt.Errorf("该分类下存在模板，请先移动或删除模板")
	}
	return row.Delete(l.svcCtx.DB.WithContext(ctx), id)
}

func (l *Logic) ListTemplates(ctx context.Context, filters map[string]string) (TemplateListResponse, error) {
	qf := func(db *gorm.DB) *gorm.DB {
		q := db
		if v := strings.TrimSpace(filters["category_id"]); v != "" {
			q = q.Where("category_id = ?", v)
		}
		if v := strings.TrimSpace(filters["status"]); v != "" {
			q = q.Where("status = ?", v)
		}
		if v := strings.TrimSpace(filters["visibility"]); v != "" {
			q = q.Where("visibility = ?", v)
		}
		if v := strings.TrimSpace(filters["owner_id"]); v != "" {
			q = q.Where("owner_id = ?", v)
		}
		if v := strings.TrimSpace(filters["q"]); v != "" {
			like := "%" + v + "%"
			q = q.Where("name LIKE ? OR display_name LIKE ? OR description LIKE ? OR tags LIKE ?", like, like, like, like)
		}
		return q
	}
	rows, err := model.ListServiceTemplates(l.svcCtx.DB.WithContext(ctx), qf)
	if err != nil {
		return TemplateListResponse{}, err
	}
	out := make([]TemplateResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, toTemplateResp(row))
	}
	return TemplateListResponse{List: out, Total: int64(len(out))}, nil
}

func (l *Logic) GetTemplate(ctx context.Context, id uint) (TemplateResponse, error) {
	row := model.ServiceTemplate{}
	if err := row.GetByID(l.svcCtx.DB.WithContext(ctx), id); err != nil {
		return TemplateResponse{}, err
	}
	return toTemplateResp(row), nil
}

func (l *Logic) CreateTemplate(ctx context.Context, uid uint64, req TemplateCreateRequest) (TemplateResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return TemplateResponse{}, fmt.Errorf("name is required")
	}
	if strings.TrimSpace(req.DisplayName) == "" {
		return TemplateResponse{}, fmt.Errorf("display_name is required")
	}
	item := model.ServiceTemplate{}
	if err := item.GetByName(l.svcCtx.DB.WithContext(ctx), name); err == nil {
		return TemplateResponse{}, fmt.Errorf("模板名称已存在")
	} else if err != gorm.ErrRecordNotFound {
		return TemplateResponse{}, err
	}
	varsJSON, err := mustJSON(req.VariablesSchema)
	if err != nil {
		return TemplateResponse{}, err
	}
	tagsJSON, err := mustJSON(req.Tags)
	if err != nil {
		return TemplateResponse{}, err
	}
	row := model.ServiceTemplate{
		Name:            name,
		DisplayName:     strings.TrimSpace(req.DisplayName),
		Description:     strings.TrimSpace(req.Description),
		Icon:            strings.TrimSpace(req.Icon),
		CategoryID:      req.CategoryID,
		Version:         defaultIfEmpty(strings.TrimSpace(req.Version), "1.0.0"),
		OwnerID:         uid,
		Visibility:      defaultIfEmpty(strings.TrimSpace(req.Visibility), model.TemplateVisibilityPrivate),
		Status:          model.TemplateStatusDraft,
		K8sTemplate:     req.K8sTemplate,
		ComposeTemplate: req.ComposeTemplate,
		VariablesSchema: varsJSON,
		Readme:          req.Readme,
		Tags:            tagsJSON,
	}
	if err := row.Create(l.svcCtx.DB.WithContext(ctx)); err != nil {
		return TemplateResponse{}, err
	}
	return toTemplateResp(row), nil
}

func (l *Logic) UpdateTemplate(ctx context.Context, id uint, uid uint64, isAdmin bool, req TemplateUpdateRequest) (TemplateResponse, error) {
	row := model.ServiceTemplate{}
	if err := row.GetByID(l.svcCtx.DB.WithContext(ctx), id); err != nil {
		return TemplateResponse{}, err
	}
	if row.OwnerID != uid && !isAdmin {
		return TemplateResponse{}, fmt.Errorf("无权限编辑此模板")
	}
	payload := map[string]any{}
	if req.DisplayName != nil {
		payload["display_name"] = strings.TrimSpace(*req.DisplayName)
	}
	if req.Description != nil {
		payload["description"] = strings.TrimSpace(*req.Description)
	}
	if req.Icon != nil {
		payload["icon"] = strings.TrimSpace(*req.Icon)
	}
	if req.CategoryID != nil {
		payload["category_id"] = *req.CategoryID
	}
	if req.Version != nil {
		payload["version"] = strings.TrimSpace(*req.Version)
	}
	if req.Visibility != nil {
		payload["visibility"] = strings.TrimSpace(*req.Visibility)
	}
	if req.K8sTemplate != nil {
		payload["k8s_template"] = *req.K8sTemplate
	}
	if req.ComposeTemplate != nil {
		payload["compose_template"] = *req.ComposeTemplate
	}
	if req.VariablesSchema != nil {
		buf, err := mustJSON(*req.VariablesSchema)
		if err != nil {
			return TemplateResponse{}, err
		}
		payload["variables_schema"] = buf
	}
	if req.Readme != nil {
		payload["readme"] = *req.Readme
	}
	if req.Tags != nil {
		buf, err := mustJSON(*req.Tags)
		if err != nil {
			return TemplateResponse{}, err
		}
		payload["tags"] = buf
	}
	if len(payload) > 0 {
		if err := row.Update(l.svcCtx.DB.WithContext(ctx), id, payload); err != nil {
			return TemplateResponse{}, err
		}
	}
	if err := row.GetByID(l.svcCtx.DB.WithContext(ctx), id); err != nil {
		return TemplateResponse{}, err
	}
	return toTemplateResp(row), nil
}

func (l *Logic) DeleteTemplate(ctx context.Context, id uint, uid uint64, isAdmin bool) error {
	row := model.ServiceTemplate{}
	if err := row.GetByID(l.svcCtx.DB.WithContext(ctx), id); err != nil {
		return err
	}
	if row.OwnerID != uid && !isAdmin {
		return fmt.Errorf("无权限执行此操作")
	}
	if row.Status == model.TemplateStatusPublished {
		return fmt.Errorf("已发布的模板无法删除，请先下架")
	}
	return row.Delete(l.svcCtx.DB.WithContext(ctx), id)
}

func (l *Logic) SubmitForReview(ctx context.Context, id uint, uid uint64, isAdmin bool) (TemplateResponse, error) {
	row := model.ServiceTemplate{}
	if err := row.GetByID(l.svcCtx.DB.WithContext(ctx), id); err != nil {
		return TemplateResponse{}, err
	}
	if row.OwnerID != uid && !isAdmin {
		return TemplateResponse{}, fmt.Errorf("无权限执行此操作")
	}
	if row.Status != model.TemplateStatusDraft && row.Status != model.TemplateStatusRejected {
		return TemplateResponse{}, fmt.Errorf("当前状态不允许提交审核")
	}
	if strings.TrimSpace(row.K8sTemplate) == "" && strings.TrimSpace(row.ComposeTemplate) == "" {
		return TemplateResponse{}, fmt.Errorf("请先填写模板内容")
	}
	if err := row.Update(l.svcCtx.DB.WithContext(ctx), id, map[string]any{"status": model.TemplateStatusPendingReview}); err != nil {
		return TemplateResponse{}, err
	}
	_ = row.GetByID(l.svcCtx.DB.WithContext(ctx), id)
	return toTemplateResp(row), nil
}

func (l *Logic) PublishTemplate(ctx context.Context, id uint) (TemplateResponse, error) {
	row := model.ServiceTemplate{}
	if err := row.GetByID(l.svcCtx.DB.WithContext(ctx), id); err != nil {
		return TemplateResponse{}, err
	}
	if row.Status != model.TemplateStatusPendingReview {
		return TemplateResponse{}, fmt.Errorf("当前状态不允许发布")
	}
	if err := row.Update(l.svcCtx.DB.WithContext(ctx), id, map[string]any{
		"status":      model.TemplateStatusPublished,
		"visibility":  model.TemplateVisibilityPublic,
		"review_note": "",
	}); err != nil {
		return TemplateResponse{}, err
	}
	_ = row.GetByID(l.svcCtx.DB.WithContext(ctx), id)
	return toTemplateResp(row), nil
}

func (l *Logic) RejectTemplate(ctx context.Context, id uint, reason string) (TemplateResponse, error) {
	row := model.ServiceTemplate{}
	if err := row.GetByID(l.svcCtx.DB.WithContext(ctx), id); err != nil {
		return TemplateResponse{}, err
	}
	if row.Status != model.TemplateStatusPendingReview {
		return TemplateResponse{}, fmt.Errorf("当前状态不允许驳回")
	}
	if err := row.Update(l.svcCtx.DB.WithContext(ctx), id, map[string]any{
		"status":      model.TemplateStatusRejected,
		"review_note": strings.TrimSpace(reason),
	}); err != nil {
		return TemplateResponse{}, err
	}
	_ = row.GetByID(l.svcCtx.DB.WithContext(ctx), id)
	return toTemplateResp(row), nil
}

func (l *Logic) PreviewYAML(ctx context.Context, req PreviewRequest) (PreviewResponse, error) {
	row := model.ServiceTemplate{}
	if err := row.GetByID(l.svcCtx.DB.WithContext(ctx), req.TemplateID); err != nil {
		return PreviewResponse{}, err
	}
	source := pickTemplateContent(row, req.Target)
	if strings.TrimSpace(source) == "" {
		return PreviewResponse{}, fmt.Errorf("模板未配置目标 %s 的内容", req.Target)
	}
	resolved, unresolved, err := l.renderTemplateWithSchema(source, row.VariablesSchema, req.Variables)
	if err != nil {
		return PreviewResponse{}, err
	}
	return PreviewResponse{RenderedYAML: resolved, UnresolvedVars: unresolved}, nil
}

func (l *Logic) DeployFromTemplate(ctx context.Context, uid uint64, req DeployRequest) (DeployResponse, error) {
	preview, err := l.PreviewYAML(ctx, PreviewRequest{
		TemplateID: req.TemplateID,
		Target:     req.Target,
		Variables:  req.Variables,
	})
	if err != nil {
		return DeployResponse{}, err
	}
	if len(preview.UnresolvedVars) > 0 {
		return DeployResponse{}, fmt.Errorf("请填写所有必填字段")
	}

	tpl := model.ServiceTemplate{}
	if err := tpl.GetByID(l.svcCtx.DB.WithContext(ctx), req.TemplateID); err != nil {
		return DeployResponse{}, err
	}

	service := model.Service{
		ProjectID:     req.ProjectID,
		TeamID:        req.TeamID,
		OwnerUserID:   uint(uid),
		Owner:         "catalog",
		Env:           defaultIfEmpty(req.Environment, "staging"),
		RuntimeType:   normalizeTarget(req.Target),
		ConfigMode:    "custom",
		ServiceKind:   "catalog",
		RenderTarget:  normalizeTarget(req.Target),
		Status:        "draft",
		Name:          defaultIfEmpty(strings.TrimSpace(req.ServiceName), tpl.Name+"-instance"),
		Type:          "stateless",
		Image:         "catalog/template",
		Replicas:      1,
		ServicePort:   80,
		ContainerPort: 80,
		CustomYAML:    preview.RenderedYAML,
		YamlContent:   preview.RenderedYAML,
		TemplateVer:   tpl.Version,
	}
	if service.ProjectID == 0 {
		return DeployResponse{}, fmt.Errorf("project_id is required")
	}
	if err := l.svcCtx.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&service).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.ServiceTemplate{}).
			Where("id = ?", tpl.ID).
			Update("deploy_count", gorm.Expr("deploy_count + 1")).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return DeployResponse{}, err
	}
	_ = tpl.GetByID(l.svcCtx.DB.WithContext(ctx), tpl.ID)
	return DeployResponse{ServiceID: service.ID, TemplateID: tpl.ID, DeployCount: tpl.DeployCount}, nil
}

func (l *Logic) renderTemplateWithSchema(content, schemaJSON string, vars map[string]any) (string, []string, error) {
	schema, err := parseVariablesSchema(schemaJSON)
	if err != nil {
		return "", nil, err
	}
	values := make(map[string]string, len(vars))
	for k, v := range vars {
		values[k] = strings.TrimSpace(fmt.Sprint(v))
	}
	if err := validateVariables(schema, values); err != nil {
		return "", nil, err
	}
	resolved := catalogTemplateVarPattern.ReplaceAllStringFunc(content, func(token string) string {
		m := catalogTemplateVarPattern.FindStringSubmatch(token)
		if len(m) < 2 {
			return token
		}
		key := strings.TrimSpace(m[1])
		def := ""
		if len(m) > 2 {
			def = strings.TrimSpace(m[2])
		}
		if v, ok := values[key]; ok && v != "" {
			return v
		}
		if def != "" {
			return def
		}
		return token
	})
	unresolved := make([]string, 0, 8)
	seen := map[string]struct{}{}
	for _, m := range catalogTemplateVarPattern.FindAllStringSubmatch(resolved, -1) {
		if len(m) < 2 {
			continue
		}
		name := strings.TrimSpace(m[1])
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		unresolved = append(unresolved, name)
	}
	sort.Strings(unresolved)
	return resolved, unresolved, nil
}

func validateVariables(schema []CatalogVariableSchema, vars map[string]string) error {
	for _, item := range schema {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			continue
		}
		v := strings.TrimSpace(vars[name])
		if v == "" && item.Default != nil {
			v = strings.TrimSpace(fmt.Sprint(item.Default))
		}
		if item.Required && v == "" {
			return fmt.Errorf("请填写所有必填字段")
		}
		if v == "" {
			continue
		}
		switch strings.ToLower(strings.TrimSpace(item.Type)) {
		case "number":
			if _, err := strconv.ParseFloat(v, 64); err != nil {
				return fmt.Errorf("变量 %s 不是有效数字", name)
			}
		case "boolean":
			if _, err := strconv.ParseBool(v); err != nil {
				return fmt.Errorf("变量 %s 不是有效布尔值", name)
			}
		case "select":
			if len(item.Options) == 0 {
				continue
			}
			ok := false
			for _, opt := range item.Options {
				if strings.TrimSpace(opt) == v {
					ok = true
					break
				}
			}
			if !ok {
				return fmt.Errorf("变量 %s 不在可选值中", name)
			}
		}
	}
	return nil
}

func parseVariablesSchema(raw string) ([]CatalogVariableSchema, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	var out []CatalogVariableSchema
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, fmt.Errorf("variables_schema invalid: %w", err)
	}
	return out, nil
}

func pickTemplateContent(row model.ServiceTemplate, target string) string {
	switch normalizeTarget(target) {
	case "compose":
		if strings.TrimSpace(row.ComposeTemplate) != "" {
			return row.ComposeTemplate
		}
		return row.K8sTemplate
	default:
		if strings.TrimSpace(row.K8sTemplate) != "" {
			return row.K8sTemplate
		}
		return row.ComposeTemplate
	}
}

func normalizeTarget(target string) string {
	t := strings.ToLower(strings.TrimSpace(target))
	if t == "compose" {
		return "compose"
	}
	return "k8s"
}

func toCategoryResp(row model.ServiceCategory) CategoryResponse {
	return CategoryResponse{
		ID:          row.ID,
		Name:        row.Name,
		DisplayName: row.DisplayName,
		Icon:        row.Icon,
		Description: row.Description,
		SortOrder:   row.SortOrder,
		IsSystem:    row.IsSystem,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func toTemplateResp(row model.ServiceTemplate) TemplateResponse {
	resp := TemplateResponse{
		ID:              row.ID,
		Name:            row.Name,
		DisplayName:     row.DisplayName,
		Description:     row.Description,
		Icon:            row.Icon,
		CategoryID:      row.CategoryID,
		Version:         row.Version,
		OwnerID:         row.OwnerID,
		Visibility:      row.Visibility,
		Status:          row.Status,
		K8sTemplate:     row.K8sTemplate,
		ComposeTemplate: row.ComposeTemplate,
		Readme:          row.Readme,
		DeployCount:     row.DeployCount,
		ReviewNote:      row.ReviewNote,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
	_ = json.Unmarshal([]byte(defaultIfEmpty(row.VariablesSchema, "[]")), &resp.VariablesSchema)
	_ = json.Unmarshal([]byte(defaultIfEmpty(row.Tags, "[]")), &resp.Tags)
	return resp
}

func mustJSON(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func defaultIfEmpty(v, d string) string {
	if strings.TrimSpace(v) == "" {
		return d
	}
	return strings.TrimSpace(v)
}

package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/cy77cc/k8s-manage/internal/model"
)

type ParamHintValue struct {
	Value any    `json:"value"`
	Label string `json:"label"`
}

type ParamHintItem struct {
	Type       string           `json:"type,omitempty"`
	Required   bool             `json:"required"`
	Default    any              `json:"default,omitempty"`
	Hint       string           `json:"hint,omitempty"`
	EnumSource string           `json:"enum_source,omitempty"`
	Values     []ParamHintValue `json:"values,omitempty"`
}

type ToolParamHintsResponse struct {
	Tool   string                   `json:"tool"`
	Params map[string]ParamHintItem `json:"params"`
}

func ResolveToolParamHints(ctx context.Context, deps PlatformDeps, meta ToolMeta) ToolParamHintsResponse {
	_ = ctx
	resp := ToolParamHintsResponse{Tool: meta.Name, Params: map[string]ParamHintItem{}}
	requiredSet := map[string]struct{}{}
	for _, item := range meta.Required {
		requiredSet[item] = struct{}{}
	}
	properties := schemaProperties(meta.Schema)
	for name, prop := range properties {
		hint := ParamHintItem{}
		if typ, ok := prop["type"].(string); ok {
			hint.Type = typ
		}
		if desc, ok := prop["description"].(string); ok {
			hint.Hint = strings.TrimSpace(desc)
		}
		if v, ok := requiredSet[name]; ok && v == struct{}{} {
			hint.Required = true
		}
		if meta.ParamHints != nil {
			if text := strings.TrimSpace(meta.ParamHints[name]); text != "" {
				hint.Hint = text
			}
		}
		if meta.DefaultHint != nil {
			if val, ok := meta.DefaultHint[name]; ok {
				hint.Default = val
			}
		}
		if source := strings.TrimSpace(meta.EnumSources[name]); source != "" {
			hint.EnumSource = source
			hint.Values = enumValuesBySource(deps, source)
		}
		resp.Params[name] = hint
	}
	for name, source := range meta.EnumSources {
		if _, ok := resp.Params[name]; ok {
			continue
		}
		resp.Params[name] = ParamHintItem{
			Required:   hasRequired(meta.Required, name),
			Hint:       strings.TrimSpace(meta.ParamHints[name]),
			Default:    meta.DefaultHint[name],
			EnumSource: source,
			Values:     enumValuesBySource(deps, source),
		}
	}
	return resp
}

func hasRequired(required []string, key string) bool {
	for _, item := range required {
		if item == key {
			return true
		}
	}
	return false
}

func schemaProperties(schema map[string]any) map[string]map[string]any {
	if schema == nil {
		return map[string]map[string]any{}
	}
	raw, ok := schema["properties"].(map[string]any)
	if !ok {
		return map[string]map[string]any{}
	}
	out := make(map[string]map[string]any, len(raw))
	for k, v := range raw {
		if m, ok := v.(map[string]any); ok {
			out[k] = m
		}
	}
	return out
}

func enumValuesBySource(deps PlatformDeps, source string) []ParamHintValue {
	if deps.DB == nil {
		return nil
	}
	switch strings.TrimSpace(source) {
	case "host_list_inventory":
		var rows []model.Node
		if err := deps.DB.Select("id,name,ip").Order("id desc").Limit(100).Find(&rows).Error; err != nil {
			return nil
		}
		out := make([]ParamHintValue, 0, len(rows))
		for _, row := range rows {
			out = append(out, ParamHintValue{Value: row.ID, Label: firstNonEmptyString(row.Name, row.IP, fmt.Sprintf("host-%d", row.ID))})
		}
		return out
	case "cluster_list_inventory":
		var rows []model.Cluster
		if err := deps.DB.Select("id,name").Order("id desc").Limit(100).Find(&rows).Error; err != nil {
			return nil
		}
		out := make([]ParamHintValue, 0, len(rows))
		for _, row := range rows {
			out = append(out, ParamHintValue{Value: row.ID, Label: firstNonEmptyString(row.Name, fmt.Sprintf("cluster-%d", row.ID))})
		}
		return out
	case "service_list_inventory", "config_app_list":
		var rows []model.Service
		if err := deps.DB.Select("id,name").Order("id desc").Limit(100).Find(&rows).Error; err != nil {
			return nil
		}
		out := make([]ParamHintValue, 0, len(rows))
		for _, row := range rows {
			out = append(out, ParamHintValue{Value: row.ID, Label: firstNonEmptyString(row.Name, fmt.Sprintf("service-%d", row.ID))})
		}
		return out
	case "deployment_target_list":
		var rows []model.DeploymentTarget
		if err := deps.DB.Select("id,name").Order("id desc").Limit(100).Find(&rows).Error; err != nil {
			return nil
		}
		out := make([]ParamHintValue, 0, len(rows))
		for _, row := range rows {
			out = append(out, ParamHintValue{Value: row.ID, Label: firstNonEmptyString(row.Name, fmt.Sprintf("target-%d", row.ID))})
		}
		return out
	case "credential_list":
		var rows []model.ClusterCredential
		if err := deps.DB.Select("id,name").Order("id desc").Limit(100).Find(&rows).Error; err != nil {
			return nil
		}
		out := make([]ParamHintValue, 0, len(rows))
		for _, row := range rows {
			out = append(out, ParamHintValue{Value: row.ID, Label: firstNonEmptyString(row.Name, fmt.Sprintf("credential-%d", row.ID))})
		}
		return out
	case "cicd_pipeline_list":
		var rows []model.CICDServiceCIConfig
		if err := deps.DB.Select("id,repo_url").Order("id desc").Limit(100).Find(&rows).Error; err != nil {
			return nil
		}
		out := make([]ParamHintValue, 0, len(rows))
		for _, row := range rows {
			out = append(out, ParamHintValue{Value: row.ID, Label: firstNonEmptyString(row.RepoURL, fmt.Sprintf("pipeline-%d", row.ID))})
		}
		return out
	case "job_list":
		var rows []model.Job
		if err := deps.DB.Select("id,name").Order("id desc").Limit(100).Find(&rows).Error; err != nil {
			return nil
		}
		out := make([]ParamHintValue, 0, len(rows))
		for _, row := range rows {
			out = append(out, ParamHintValue{Value: row.ID, Label: firstNonEmptyString(row.Name, fmt.Sprintf("job-%d", row.ID))})
		}
		return out
	case "user_list":
		var rows []model.User
		if err := deps.DB.Select("id,username").Order("id desc").Limit(100).Find(&rows).Error; err != nil {
			return nil
		}
		out := make([]ParamHintValue, 0, len(rows))
		for _, row := range rows {
			out = append(out, ParamHintValue{Value: row.ID, Label: firstNonEmptyString(row.Username, fmt.Sprintf("user-%d", row.ID))})
		}
		return out
	default:
		return nil
	}
}

func firstNonEmptyString(items ...string) string {
	for _, item := range items {
		if strings.TrimSpace(item) != "" {
			return strings.TrimSpace(item)
		}
	}
	return ""
}

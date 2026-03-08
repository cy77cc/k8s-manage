package deployment

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	. "github.com/cy77cc/k8s-manage/internal/ai/tools/core"
	"github.com/cy77cc/k8s-manage/internal/model"
)

func DeploymentTargetList(ctx context.Context, deps PlatformDeps, input DeploymentTargetListInput) (ToolResult, error) {
	return RunWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:        "deployment_target_list",
			Description: "查询部署目标列表。可选参数 env/status/keyword/limit。示例: {\"env\":\"prod\",\"limit\":20}。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			DefaultHint: map[string]any{"limit": 50},
			SceneScope:  []string{"deployment:targets"},
		},
		input,
		func(in DeploymentTargetListInput) (any, string, error) {
			if deps.DB == nil {
				return nil, "db", fmt.Errorf("db unavailable")
			}
			limit := in.Limit
			if limit <= 0 {
				limit = 50
			}
			if limit > 200 {
				limit = 200
			}
			query := deps.DB.Model(&model.DeploymentTarget{})
			if env := strings.TrimSpace(in.Env); env != "" {
				query = query.Where("env = ?", env)
			}
			if status := strings.TrimSpace(in.Status); status != "" {
				query = query.Where("status = ?", status)
			}
			if kw := strings.TrimSpace(in.Keyword); kw != "" {
				pattern := "%" + kw + "%"
				query = query.Where("name LIKE ?", pattern)
			}
			var rows []model.DeploymentTarget
			if err := query.Order("id desc").Limit(limit).Find(&rows).Error; err != nil {
				return nil, "db", err
			}
			list := make([]map[string]any, 0, len(rows))
			for _, item := range rows {
				list = append(list, map[string]any{
					"id":               item.ID,
					"name":             item.Name,
					"env":              item.Env,
					"status":           item.Status,
					"target_type":      item.TargetType,
					"runtime_type":     item.RuntimeType,
					"cluster_id":       item.ClusterID,
					"credential_id":    item.CredentialID,
					"readiness_status": item.ReadinessStatus,
				})
			}
			return map[string]any{"total": len(list), "list": list}, "db", nil
		},
	)
}

func DeploymentTargetDetail(ctx context.Context, deps PlatformDeps, input DeploymentTargetDetailInput) (ToolResult, error) {
	return RunWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:        "deployment_target_detail",
			Description: "查询部署目标详情。target_id 必填。示例: {\"target_id\":12}。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			Required:    []string{"target_id"},
			EnumSources: map[string]string{"target_id": "deployment_target_list"},
			SceneScope:  []string{"deployment:targets"},
		},
		input,
		func(in DeploymentTargetDetailInput) (any, string, error) {
			if deps.DB == nil {
				return nil, "db", fmt.Errorf("db unavailable")
			}
			if in.TargetID <= 0 {
				return nil, "validation", NewMissingParam("target_id", "target_id is required")
			}
			var target model.DeploymentTarget
			if err := deps.DB.First(&target, in.TargetID).Error; err != nil {
				return nil, "db", err
			}
			var nodes []model.DeploymentTargetNode
			_ = deps.DB.Where("target_id = ?", target.ID).Order("id asc").Find(&nodes).Error
			return map[string]any{"target": target, "nodes": nodes}, "db", nil
		},
	)
}

func DeploymentBootstrapStatus(ctx context.Context, deps PlatformDeps, input DeploymentBootstrapStatusInput) (ToolResult, error) {
	return RunWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:        "deployment_bootstrap_status",
			Description: "查询部署目标环境引导状态。target_id 必填。示例: {\"target_id\":12}。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			Required:    []string{"target_id"},
			EnumSources: map[string]string{"target_id": "deployment_target_list"},
			SceneScope:  []string{"deployment:targets", "deployment:clusters"},
		},
		input,
		func(in DeploymentBootstrapStatusInput) (any, string, error) {
			if deps.DB == nil {
				return nil, "db", fmt.Errorf("db unavailable")
			}
			if in.TargetID <= 0 {
				return nil, "validation", NewMissingParam("target_id", "target_id is required")
			}
			var target model.DeploymentTarget
			if err := deps.DB.First(&target, in.TargetID).Error; err != nil {
				return nil, "db", err
			}
			result := map[string]any{
				"target_id":        target.ID,
				"target_name":      target.Name,
				"bootstrap_job_id": target.BootstrapJobID,
				"target_status":    target.Status,
				"readiness_status": target.ReadinessStatus,
			}
			if strings.TrimSpace(target.BootstrapJobID) == "" {
				return result, "db", nil
			}
			var job model.EnvironmentInstallJob
			if err := deps.DB.Where("id = ?", target.BootstrapJobID).First(&job).Error; err == nil {
				result["bootstrap_job"] = job
				var steps []model.EnvironmentInstallJobStep
				_ = deps.DB.Where("job_id = ?", job.ID).Order("id asc").Find(&steps).Error
				result["steps"] = steps
			}
			return result, "db", nil
		},
	)
}

func ConfigAppList(ctx context.Context, deps PlatformDeps, input ConfigAppListInput) (ToolResult, error) {
	return RunWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:        "config_app_list",
			Description: "查询配置应用列表。可选参数 keyword/env/limit。示例: {\"env\":\"prod\"}。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			DefaultHint: map[string]any{"limit": 50},
			SceneScope:  []string{"configcenter"},
		},
		input,
		func(in ConfigAppListInput) (any, string, error) {
			if deps.DB == nil {
				return nil, "db", fmt.Errorf("db unavailable")
			}
			limit := in.Limit
			if limit <= 0 {
				limit = 50
			}
			if limit > 200 {
				limit = 200
			}
			query := deps.DB.Model(&model.Service{})
			if kw := strings.TrimSpace(in.Keyword); kw != "" {
				pattern := "%" + kw + "%"
				query = query.Where("name LIKE ? OR owner LIKE ?", pattern, pattern)
			}
			if env := strings.TrimSpace(in.Env); env != "" {
				query = query.Where("env = ?", env)
			}
			var services []model.Service
			if err := query.Order("id desc").Limit(limit).Find(&services).Error; err != nil {
				return nil, "db", err
			}
			list := make([]map[string]any, 0, len(services))
			for _, svc := range services {
				list = append(list, map[string]any{"app_id": svc.ID, "name": svc.Name, "env": svc.Env, "owner": svc.Owner})
			}
			return map[string]any{"total": len(list), "list": list}, "db", nil
		},
	)
}

func ConfigItemGet(ctx context.Context, deps PlatformDeps, input ConfigItemGetInput) (ToolResult, error) {
	return RunWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:        "config_item_get",
			Description: "查询配置项值。app_id/key 必填，可选 env。示例: {\"app_id\":12,\"key\":\"DATABASE_URL\"}。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			Required:    []string{"app_id", "key"},
			EnumSources: map[string]string{"app_id": "config_app_list"},
			SceneScope:  []string{"configcenter"},
		},
		input,
		func(in ConfigItemGetInput) (any, string, error) {
			if deps.DB == nil {
				return nil, "db", fmt.Errorf("db unavailable")
			}
			if in.AppID <= 0 {
				return nil, "validation", NewMissingParam("app_id", "app_id is required")
			}
			key := strings.TrimSpace(in.Key)
			if key == "" {
				return nil, "validation", NewMissingParam("key", "key is required")
			}
			env := strings.TrimSpace(in.Env)
			if env == "" {
				env = "staging"
			}
			var set model.ServiceVariableSet
			if err := deps.DB.Where("service_id = ? AND env = ?", in.AppID, env).Order("updated_at desc").First(&set).Error; err != nil {
				return nil, "db", err
			}
			values := map[string]any{}
			_ = json.Unmarshal([]byte(set.ValuesJSON), &values)
			return map[string]any{"app_id": in.AppID, "env": env, "key": key, "value": values[key], "updated_at": set.UpdatedAt}, "db", nil
		},
	)
}

func ConfigDiff(ctx context.Context, deps PlatformDeps, input ConfigDiffInput) (ToolResult, error) {
	return RunWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:        "config_diff",
			Description: "对比配置差异。app_id/env_a/env_b 必填。示例: {\"app_id\":12,\"env_a\":\"staging\",\"env_b\":\"prod\"}。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			Required:    []string{"app_id", "env_a", "env_b"},
			EnumSources: map[string]string{"app_id": "config_app_list"},
			SceneScope:  []string{"configcenter"},
		},
		input,
		func(in ConfigDiffInput) (any, string, error) {
			if deps.DB == nil {
				return nil, "db", fmt.Errorf("db unavailable")
			}
			if in.AppID <= 0 {
				return nil, "validation", NewMissingParam("app_id", "app_id is required")
			}
			envA := strings.TrimSpace(in.EnvA)
			envB := strings.TrimSpace(in.EnvB)
			if envA == "" {
				return nil, "validation", NewMissingParam("env_a", "env_a is required")
			}
			if envB == "" {
				return nil, "validation", NewMissingParam("env_b", "env_b is required")
			}
			readEnv := func(env string) (map[string]any, error) {
				var set model.ServiceVariableSet
				if err := deps.DB.Where("service_id = ? AND env = ?", in.AppID, env).Order("updated_at desc").First(&set).Error; err != nil {
					return nil, err
				}
				out := map[string]any{}
				_ = json.Unmarshal([]byte(set.ValuesJSON), &out)
				return out, nil
			}
			a, err := readEnv(envA)
			if err != nil {
				return nil, "db", err
			}
			b, err := readEnv(envB)
			if err != nil {
				return nil, "db", err
			}
			diff := make([]map[string]any, 0)
			seen := map[string]struct{}{}
			for k, av := range a {
				seen[k] = struct{}{}
				bv := b[k]
				if fmt.Sprintf("%v", av) != fmt.Sprintf("%v", bv) {
					diff = append(diff, map[string]any{"key": k, "env_a": av, "env_b": bv})
				}
			}
			for k, bv := range b {
				if _, ok := seen[k]; ok {
					continue
				}
				diff = append(diff, map[string]any{"key": k, "env_a": nil, "env_b": bv})
			}
			return map[string]any{"app_id": in.AppID, "env_a": envA, "env_b": envB, "diff_count": len(diff), "diff": diff}, "db", nil
		},
	)
}

func ClusterListInventory(ctx context.Context, deps PlatformDeps, input ClusterInventoryInput) (ToolResult, error) {
	return RunWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:       "cluster_list_inventory",
			Mode:       ToolModeReadonly,
			Risk:       ToolRiskLow,
			Provider:   "local",
			Permission: "ai:tool:read",
		},
		input,
		func(in ClusterInventoryInput) (any, string, error) {
			if deps.DB == nil {
				return nil, "db", fmt.Errorf("db unavailable")
			}
			limit := in.Limit
			if limit <= 0 {
				limit = 50
			}
			if limit > 200 {
				limit = 200
			}
			query := deps.DB.Model(&model.Cluster{})
			if status := strings.TrimSpace(in.Status); status != "" {
				query = query.Where("status = ?", status)
			}
			if kw := strings.TrimSpace(in.Keyword); kw != "" {
				pattern := "%" + kw + "%"
				query = query.Where("name LIKE ? OR endpoint LIKE ?", pattern, pattern)
			}
			var rows []model.Cluster
			if err := query.Order("id desc").Limit(limit).Find(&rows).Error; err != nil {
				return nil, "db", err
			}
			list := make([]map[string]any, 0, len(rows))
			for _, item := range rows {
				list = append(list, map[string]any{
					"id":         item.ID,
					"name":       item.Name,
					"status":     item.Status,
					"type":       item.Type,
					"endpoint":   item.Endpoint,
					"version":    item.Version,
					"updated_at": item.UpdatedAt,
				})
			}
			return map[string]any{
				"total": len(list),
				"list":  list,
				"filters_applied": map[string]any{
					"status":  strings.TrimSpace(in.Status),
					"keyword": strings.TrimSpace(in.Keyword),
					"limit":   limit,
				},
			}, "db", nil
		})
}

func ServiceListInventory(ctx context.Context, deps PlatformDeps, input ServiceInventoryInput) (ToolResult, error) {
	return RunWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:       "service_list_inventory",
			Mode:       ToolModeReadonly,
			Risk:       ToolRiskLow,
			Provider:   "local",
			Permission: "ai:tool:read",
		},
		input,
		func(in ServiceInventoryInput) (any, string, error) {
			if deps.DB == nil {
				return nil, "db", fmt.Errorf("db unavailable")
			}
			limit := in.Limit
			if limit <= 0 {
				limit = 50
			}
			if limit > 200 {
				limit = 200
			}
			query := deps.DB.Model(&model.Service{})
			if status := strings.TrimSpace(in.Status); status != "" {
				query = query.Where("status = ?", status)
			}
			if env := strings.TrimSpace(in.Env); env != "" {
				query = query.Where("env = ?", env)
			}
			if runtime := strings.TrimSpace(in.RuntimeType); runtime != "" {
				query = query.Where("runtime_type = ?", runtime)
			}
			if kw := strings.TrimSpace(in.Keyword); kw != "" {
				pattern := "%" + kw + "%"
				query = query.Where("name LIKE ? OR owner LIKE ?", pattern, pattern)
			}
			var rows []model.Service
			if err := query.Order("id desc").Limit(limit).Find(&rows).Error; err != nil {
				return nil, "db", err
			}
			list := make([]map[string]any, 0, len(rows))
			for _, item := range rows {
				list = append(list, map[string]any{
					"id":            item.ID,
					"name":          item.Name,
					"status":        item.Status,
					"env":           item.Env,
					"owner":         item.Owner,
					"runtime_type":  item.RuntimeType,
					"config_mode":   item.ConfigMode,
					"render_target": item.RenderTarget,
					"updated_at":    item.UpdatedAt,
				})
			}
			return map[string]any{
				"total": len(list),
				"list":  list,
				"filters_applied": map[string]any{
					"status":       strings.TrimSpace(in.Status),
					"env":          strings.TrimSpace(in.Env),
					"runtime_type": strings.TrimSpace(in.RuntimeType),
					"keyword":      strings.TrimSpace(in.Keyword),
					"limit":        limit,
				},
			}, "db", nil
		})
}

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cy77cc/k8s-manage/internal/model"
)

func configAppList(ctx context.Context, deps PlatformDeps, input ConfigAppListInput) (ToolResult, error) {
	return runWithPolicyAndEvent(
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

func configItemGet(ctx context.Context, deps PlatformDeps, input ConfigItemGetInput) (ToolResult, error) {
	return runWithPolicyAndEvent(
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

func configDiff(ctx context.Context, deps PlatformDeps, input ConfigDiffInput) (ToolResult, error) {
	return runWithPolicyAndEvent(
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

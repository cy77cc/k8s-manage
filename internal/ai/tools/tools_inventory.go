package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/cy77cc/k8s-manage/internal/model"
)

func clusterListInventory(ctx context.Context, deps PlatformDeps, input ClusterInventoryInput) (ToolResult, error) {
	return runWithPolicyAndEvent(
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

func serviceListInventory(ctx context.Context, deps PlatformDeps, input ServiceInventoryInput) (ToolResult, error) {
	return runWithPolicyAndEvent(
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

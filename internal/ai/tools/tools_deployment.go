package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/cy77cc/k8s-manage/internal/model"
)

func deploymentTargetList(ctx context.Context, deps PlatformDeps, input DeploymentTargetListInput) (ToolResult, error) {
	return runWithPolicyAndEvent(
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

func deploymentTargetDetail(ctx context.Context, deps PlatformDeps, input DeploymentTargetDetailInput) (ToolResult, error) {
	return runWithPolicyAndEvent(
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

func deploymentBootstrapStatus(ctx context.Context, deps PlatformDeps, input DeploymentBootstrapStatusInput) (ToolResult, error) {
	return runWithPolicyAndEvent(
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

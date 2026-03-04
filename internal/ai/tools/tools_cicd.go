package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cy77cc/k8s-manage/internal/model"
)

func cicdPipelineList(ctx context.Context, deps PlatformDeps, input CICDPipelineListInput) (ToolResult, error) {
	return runWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:        "cicd_pipeline_list",
			Description: "查询 CI 流水线列表。可选参数 status/keyword/limit。示例: {\"status\":\"active\"}。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			DefaultHint: map[string]any{"limit": 50},
			SceneScope:  []string{"cicd"},
		},
		input,
		func(in CICDPipelineListInput) (any, string, error) {
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
			query := deps.DB.Model(&model.CICDServiceCIConfig{})
			if status := strings.TrimSpace(in.Status); status != "" {
				query = query.Where("status = ?", status)
			}
			if kw := strings.TrimSpace(in.Keyword); kw != "" {
				pattern := "%" + kw + "%"
				query = query.Where("repo_url LIKE ? OR branch LIKE ?", pattern, pattern)
			}
			var rows []model.CICDServiceCIConfig
			if err := query.Order("id desc").Limit(limit).Find(&rows).Error; err != nil {
				return nil, "db", err
			}
			return map[string]any{"total": len(rows), "list": rows}, "db", nil
		},
	)
}

func cicdPipelineStatus(ctx context.Context, deps PlatformDeps, input CICDPipelineStatusInput) (ToolResult, error) {
	return runWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:        "cicd_pipeline_status",
			Description: "查询流水线状态。pipeline_id 必填。示例: {\"pipeline_id\":3}。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			Required:    []string{"pipeline_id"},
			EnumSources: map[string]string{"pipeline_id": "cicd_pipeline_list"},
			SceneScope:  []string{"cicd"},
		},
		input,
		func(in CICDPipelineStatusInput) (any, string, error) {
			if deps.DB == nil {
				return nil, "db", fmt.Errorf("db unavailable")
			}
			if in.PipelineID <= 0 {
				return nil, "validation", NewMissingParam("pipeline_id", "pipeline_id is required")
			}
			var cfg model.CICDServiceCIConfig
			if err := deps.DB.First(&cfg, in.PipelineID).Error; err != nil {
				return nil, "db", err
			}
			var runs []model.CICDServiceCIRun
			_ = deps.DB.Where("ci_config_id = ?", cfg.ID).Order("id desc").Limit(10).Find(&runs).Error
			return map[string]any{"pipeline": cfg, "recent_runs": runs}, "db", nil
		},
	)
}

func cicdPipelineTrigger(ctx context.Context, deps PlatformDeps, input CICDPipelineTriggerInput) (ToolResult, error) {
	return runWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:        "cicd_pipeline_trigger",
			Description: "触发流水线构建。pipeline_id/branch 必填。示例: {\"pipeline_id\":3,\"branch\":\"main\"}。",
			Mode:        ToolModeMutating,
			Risk:        ToolRiskHigh,
			Provider:    "local",
			Permission:  "ai:tool:execute",
			Required:    []string{"pipeline_id", "branch"},
			EnumSources: map[string]string{"pipeline_id": "cicd_pipeline_list"},
			SceneScope:  []string{"cicd"},
		},
		input,
		func(in CICDPipelineTriggerInput) (any, string, error) {
			if deps.DB == nil {
				return nil, "db", fmt.Errorf("db unavailable")
			}
			if in.PipelineID <= 0 {
				return nil, "validation", NewMissingParam("pipeline_id", "pipeline_id is required")
			}
			if strings.TrimSpace(in.Branch) == "" {
				return nil, "validation", NewMissingParam("branch", "branch is required")
			}
			var cfg model.CICDServiceCIConfig
			if err := deps.DB.First(&cfg, in.PipelineID).Error; err != nil {
				return nil, "db", err
			}
			run := model.CICDServiceCIRun{
				ServiceID:   cfg.ServiceID,
				CIConfigID:  cfg.ID,
				TriggerType: "manual",
				Status:      "queued",
				Reason:      "triggered by ai tool",
				TriggeredAt: time.Now(),
			}
			if err := deps.DB.Create(&run).Error; err != nil {
				return nil, "db", err
			}
			return map[string]any{"pipeline_id": cfg.ID, "run_id": run.ID, "branch": strings.TrimSpace(in.Branch), "status": run.Status}, "db", nil
		},
	)
}

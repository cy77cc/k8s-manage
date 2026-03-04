package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cy77cc/k8s-manage/internal/model"
)

func jobList(ctx context.Context, deps PlatformDeps, input JobListInput) (ToolResult, error) {
	return runWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:        "job_list",
			Description: "查询任务列表。可选参数 status/keyword/limit。示例: {\"status\":\"running\"}。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			DefaultHint: map[string]any{"limit": 50},
			SceneScope:  []string{"jobs"},
		},
		input,
		func(in JobListInput) (any, string, error) {
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
			query := deps.DB.Model(&model.Job{})
			if status := strings.TrimSpace(in.Status); status != "" {
				query = query.Where("status = ?", status)
			}
			if kw := strings.TrimSpace(in.Keyword); kw != "" {
				pattern := "%" + kw + "%"
				query = query.Where("name LIKE ? OR type LIKE ?", pattern, pattern)
			}
			var jobs []model.Job
			if err := query.Order("id desc").Limit(limit).Find(&jobs).Error; err != nil {
				return nil, "db", err
			}
			return map[string]any{"total": len(jobs), "list": jobs}, "db", nil
		},
	)
}

func jobExecutionStatus(ctx context.Context, deps PlatformDeps, input JobExecutionStatusInput) (ToolResult, error) {
	return runWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:        "job_execution_status",
			Description: "查询任务执行状态。job_id 必填，可选 execution_id。示例: {\"job_id\":12}。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			Required:    []string{"job_id"},
			EnumSources: map[string]string{"job_id": "job_list"},
			SceneScope:  []string{"jobs"},
		},
		input,
		func(in JobExecutionStatusInput) (any, string, error) {
			if deps.DB == nil {
				return nil, "db", fmt.Errorf("db unavailable")
			}
			if in.JobID <= 0 {
				return nil, "validation", NewMissingParam("job_id", "job_id is required")
			}
			query := deps.DB.Model(&model.JobExecution{}).Where("job_id = ?", in.JobID)
			if in.ExecutionID > 0 {
				query = query.Where("id = ?", in.ExecutionID)
			}
			var rows []model.JobExecution
			if err := query.Order("id desc").Limit(20).Find(&rows).Error; err != nil {
				return nil, "db", err
			}
			return map[string]any{"total": len(rows), "list": rows}, "db", nil
		},
	)
}

func jobRun(ctx context.Context, deps PlatformDeps, input JobRunInput) (ToolResult, error) {
	return runWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:        "job_run",
			Description: "手动触发任务执行。job_id 必填。示例: {\"job_id\":12}。",
			Mode:        ToolModeMutating,
			Risk:        ToolRiskMedium,
			Provider:    "local",
			Permission:  "ai:tool:execute",
			Required:    []string{"job_id"},
			EnumSources: map[string]string{"job_id": "job_list"},
			SceneScope:  []string{"jobs"},
		},
		input,
		func(in JobRunInput) (any, string, error) {
			if deps.DB == nil {
				return nil, "db", fmt.Errorf("db unavailable")
			}
			if in.JobID <= 0 {
				return nil, "validation", NewMissingParam("job_id", "job_id is required")
			}
			var job model.Job
			if err := deps.DB.First(&job, in.JobID).Error; err != nil {
				return nil, "db", err
			}
			now := time.Now()
			exec := model.JobExecution{
				JobID:     job.ID,
				Status:    "running",
				ExitCode:  0,
				Output:    "triggered by ai tool",
				StartTime: now,
			}
			if err := deps.DB.Create(&exec).Error; err != nil {
				return nil, "db", err
			}
			_ = deps.DB.Model(&job).Updates(map[string]any{"status": "running", "last_run": now}).Error
			return map[string]any{"job_id": job.ID, "execution_id": exec.ID, "status": exec.Status}, "db", nil
		},
	)
}

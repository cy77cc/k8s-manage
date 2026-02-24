package ai

import (
	"context"
	"errors"

	"github.com/cy77cc/k8s-manage/internal/model"
)

func serviceGetDetail(ctx context.Context, deps PlatformDeps, input map[string]any) (ToolResult, error) {
	return runWithPolicyAndEvent(ctx, ToolMeta{Name: "service.get_detail", Mode: ToolModeReadonly, Risk: ToolRiskLow, Provider: "local", Permission: "ai:tool:read"}, input, func() (any, string, error) {
		sid := toInt(input["service_id"])
		if sid <= 0 {
			return nil, "db", errors.New("service_id is required")
		}
		var s model.Service
		if err := deps.DB.First(&s, sid).Error; err != nil {
			return nil, "db", err
		}
		return s, "db", nil
	})
}

func serviceDeployPreview(ctx context.Context, deps PlatformDeps, input map[string]any) (ToolResult, error) {
	return runWithPolicyAndEvent(ctx, ToolMeta{Name: "service.deploy_preview", Mode: ToolModeReadonly, Risk: ToolRiskMedium, Provider: "local", Permission: "ai:tool:read"}, input, func() (any, string, error) {
		sid := toInt(input["service_id"])
		cid := toInt(input["cluster_id"])
		if sid <= 0 || cid <= 0 {
			return nil, "preview", errors.New("service_id and cluster_id are required")
		}
		var s model.Service
		if err := deps.DB.First(&s, sid).Error; err != nil {
			return nil, "db", err
		}
		return map[string]any{"service_id": sid, "cluster_id": cid, "name": s.Name, "image": s.Image, "replicas": s.Replicas}, "preview", nil
	})
}

func serviceDeployApply(ctx context.Context, deps PlatformDeps, input map[string]any) (ToolResult, error) {
	return runWithPolicyAndEvent(ctx, ToolMeta{Name: "service.deploy_apply", Mode: ToolModeMutating, Risk: ToolRiskHigh, Provider: "local", Permission: "ai:tool:execute"}, input, func() (any, string, error) {
		sid := toInt(input["service_id"])
		cid := toInt(input["cluster_id"])
		if sid <= 0 || cid <= 0 {
			return nil, "deploy", errors.New("service_id and cluster_id are required")
		}
		var svc model.Service
		if err := deps.DB.First(&svc, sid).Error; err != nil {
			return nil, "db", err
		}
		var cluster model.Cluster
		if err := deps.DB.First(&cluster, cid).Error; err != nil {
			return nil, "db", err
		}
		_ = cluster
		return map[string]any{
			"applied":    true,
			"service_id": sid,
			"cluster_id": cid,
			"message":    "deploy apply executed in MVP mode",
			"image":      svc.Image,
		}, "deploy", nil
	})
}

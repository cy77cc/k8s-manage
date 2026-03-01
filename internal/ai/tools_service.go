package ai

import (
	"context"

	"github.com/cy77cc/k8s-manage/internal/model"
)

func serviceGetDetail(ctx context.Context, deps PlatformDeps, input ServiceDetailInput) (ToolResult, error) {
	return runWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:        "service_get_detail",
			Description: "读取服务详情。默认 service_id=0。示例: {\"service_id\":123}",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			DefaultHint: map[string]any{"service_id": 0},
		},
		input,
		func(in ServiceDetailInput) (any, string, error) {
			sid := in.ServiceID
			if sid <= 0 {
				return nil, "validation", NewMissingParam("service_id", "service_id is required")
			}
			var s model.Service
			if err := deps.DB.First(&s, sid).Error; err != nil {
				return nil, "db", err
			}
			return s, "db", nil
		})
}

func serviceDeployPreview(ctx context.Context, deps PlatformDeps, input ServiceDeployPreviewInput) (ToolResult, error) {
	return runWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:        "service_deploy_preview",
			Description: "部署服务预览。默认 service_id=0, cluster_id=0。示例: {\"service_id\":123, \"cluster_id\":456}",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskMedium,
			Provider:    "local",
			Permission:  "ai:tool:read",
			DefaultHint: map[string]any{"service_id": 0, "cluster_id": 0},
		},
		input,
		func(in ServiceDeployPreviewInput) (any, string, error) {
			sid := in.ServiceID
			cid := in.ClusterID
			if sid <= 0 {
				return nil, "preview", NewMissingParam("service_id", "service_id is required")
			}
			if cid <= 0 {
				return nil, "preview", NewMissingParam("cluster_id", "cluster_id is required")
			}
			var s model.Service
			if err := deps.DB.First(&s, sid).Error; err != nil {
				return nil, "db", err
			}
			return map[string]any{"service_id": sid, "cluster_id": cid, "name": s.Name, "image": s.Image, "replicas": s.Replicas}, "preview", nil
		})
}

func serviceDeployApply(ctx context.Context, deps PlatformDeps, input ServiceDeployApplyInput) (ToolResult, error) {
	return runWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:        "service_deploy_apply",
			Description: "部署服务应用。默认 service_id=0, cluster_id=0。示例: {\"service_id\":123, \"cluster_id\":456}",
			Mode:        ToolModeMutating,
			Risk:        ToolRiskHigh,
			Provider:    "local",
			Permission:  "ai:tool:execute",
			DefaultHint: map[string]any{"service_id": 0, "cluster_id": 0},
		},
		input,
		func(in ServiceDeployApplyInput) (any, string, error) {
			sid := in.ServiceID
			cid := in.ClusterID
			if sid <= 0 {
				return nil, "deploy", NewMissingParam("service_id", "service_id is required")
			}
			if cid <= 0 {
				return nil, "deploy", NewMissingParam("cluster_id", "cluster_id is required")
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

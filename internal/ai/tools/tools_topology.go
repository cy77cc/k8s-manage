package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cy77cc/k8s-manage/internal/model"
)

func topologyGet(ctx context.Context, deps PlatformDeps, input TopologyGetInput) (ToolResult, error) {
	return runWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:        "topology_get",
			Description: "查询服务拓扑关系。可选 service_id/depth。示例: {\"service_id\":12,\"depth\":2}。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			DefaultHint: map[string]any{"depth": 2},
			EnumSources: map[string]string{"service_id": "service_list_inventory"},
			SceneScope:  []string{"deployment:topology", "services:detail"},
		},
		input,
		func(in TopologyGetInput) (any, string, error) {
			if deps.DB == nil {
				return nil, "db", fmt.Errorf("db unavailable")
			}
			depth := in.Depth
			if depth <= 0 {
				depth = 2
			}
			services := make([]model.Service, 0)
			query := deps.DB.Model(&model.Service{})
			if in.ServiceID > 0 {
				query = query.Where("id = ?", in.ServiceID)
			}
			if err := query.Order("id desc").Limit(100).Find(&services).Error; err != nil {
				return nil, "db", err
			}
			nodes := make([]map[string]any, 0, len(services))
			edges := make([]map[string]any, 0)
			for _, svc := range services {
				nodes = append(nodes, map[string]any{"id": fmt.Sprintf("service-%d", svc.ID), "type": "service", "label": svc.Name, "service_id": svc.ID})
				var releases []model.DeploymentRelease
				_ = deps.DB.Where("service_id = ?", svc.ID).Order("id desc").Limit(depth).Find(&releases).Error
				for _, rel := range releases {
					targetNodeID := fmt.Sprintf("target-%d", rel.TargetID)
					edges = append(edges, map[string]any{"from": fmt.Sprintf("service-%d", svc.ID), "to": targetNodeID, "type": "deploy"})
				}
			}
			return map[string]any{"nodes": nodes, "edges": edges, "depth": depth}, "db", nil
		},
	)
}

func auditLogSearch(ctx context.Context, deps PlatformDeps, input AuditLogSearchInput) (ToolResult, error) {
	return runWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:        "audit_log_search",
			Description: "查询审计日志。可选 time_range/resource_type/action/user_id/limit。示例: {\"time_range\":\"24h\"}。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			DefaultHint: map[string]any{"time_range": "24h", "limit": 50},
			SceneScope:  []string{"deployment:audit", "governance:permissions"},
		},
		input,
		func(in AuditLogSearchInput) (any, string, error) {
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
			since := time.Now().Add(-parseTimeRange(strings.TrimSpace(in.TimeRange), 24*time.Hour))
			query := deps.DB.Model(&model.AuditLog{}).Where("created_at >= ?", since)
			if rt := strings.TrimSpace(in.ResourceType); rt != "" {
				query = query.Where("resource_type = ?", rt)
			}
			if action := strings.TrimSpace(in.Action); action != "" {
				query = query.Where("action_type = ?", action)
			}
			if in.UserID > 0 {
				query = query.Where("actor_id = ?", in.UserID)
			}
			var rows []model.AuditLog
			if err := query.Order("id desc").Limit(limit).Find(&rows).Error; err != nil {
				return nil, "db", err
			}
			return map[string]any{"total": len(rows), "list": rows}, "db", nil
		},
	)
}

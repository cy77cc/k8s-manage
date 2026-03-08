package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	. "github.com/cy77cc/k8s-manage/internal/ai/tools/core"
	"github.com/cy77cc/k8s-manage/internal/model"
)

func ServiceGetDetail(ctx context.Context, deps PlatformDeps, input ServiceDetailInput) (ToolResult, error) {
	return RunWithPolicyAndEvent(
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

func ServiceStatus(ctx context.Context, deps PlatformDeps, input ServiceStatusInput) (ToolResult, error) {
	return RunWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:        "service_status",
			Description: "读取服务当前状态与基础信息。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			DefaultHint: map[string]any{"service_id": 0},
		},
		input,
		func(in ServiceStatusInput) (any, string, error) {
			if in.ServiceID <= 0 {
				return nil, "validation", NewMissingParam("service_id", "service_id is required")
			}
			var svc model.Service
			if err := deps.DB.First(&svc, in.ServiceID).Error; err != nil {
				return nil, "db", err
			}
			return map[string]any{
				"service_id":   svc.ID,
				"name":         svc.Name,
				"status":       svc.Status,
				"env":          svc.Env,
				"runtime_type": svc.RuntimeType,
				"image":        svc.Image,
				"replicas":     svc.Replicas,
				"updated_at":   svc.UpdatedAt,
			}, "db", nil
		})
}

func serviceDeployPreviewData(deps PlatformDeps, sid, cid int) (any, string, error) {
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
	return map[string]any{
		"preview":    true,
		"service_id": sid,
		"cluster_id": cid,
		"name":       s.Name,
		"image":      s.Image,
		"replicas":   s.Replicas,
	}, "preview", nil
}

func ServiceDeployPreview(ctx context.Context, deps PlatformDeps, input ServiceDeployPreviewInput) (ToolResult, error) {
	return RunWithPolicyAndEvent(
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
			return serviceDeployPreviewData(deps, in.ServiceID, in.ClusterID)
		})
}

func serviceDeployApplyData(deps PlatformDeps, sid, cid int) (any, string, error) {
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
}

func ServiceDeploy(ctx context.Context, deps PlatformDeps, input ServiceDeployInput) (ToolResult, error) {
	return RunWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:        "service_deploy",
			Description: "统一服务部署工具，支持 preview/apply。",
			Mode:        ToolModeMutating,
			Risk:        ToolRiskHigh,
			Provider:    "local",
			Permission:  "ai:tool:execute",
			DefaultHint: map[string]any{"preview": true, "apply": false},
		},
		input,
		func(in ServiceDeployInput) (any, string, error) {
			if in.Apply {
				return serviceDeployApplyData(deps, in.ServiceID, in.ClusterID)
			}
			return serviceDeployPreviewData(deps, in.ServiceID, in.ClusterID)
		})
}
func ServiceDeployApply(ctx context.Context, deps PlatformDeps, input ServiceDeployApplyInput) (ToolResult, error) {
	return RunWithPolicyAndEvent(
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
			return serviceDeployApplyData(deps, in.ServiceID, in.ClusterID)
		})
}

func ServiceCatalogList(ctx context.Context, deps PlatformDeps, input ServiceCatalogListInput) (ToolResult, error) {
	return RunWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:        "service_catalog_list",
			Description: "查询服务目录列表。可选参数 keyword/category_id/limit。示例: {\"keyword\":\"payment\",\"category_id\":2,\"limit\":20}。category_id 可选 1(中间件)/2(业务)。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			DefaultHint: map[string]any{"limit": 50},
			ParamHints: map[string]string{
				"category_id": "1=middleware, 2=business",
				"keyword":     "按名称或负责人模糊匹配",
			},
			SceneScope: []string{"services:list", "services:catalog"},
		},
		input,
		func(in ServiceCatalogListInput) (any, string, error) {
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
			switch in.CategoryID {
			case 1:
				query = query.Where("service_kind = ?", "middleware")
			case 2:
				query = query.Where("service_kind = ?", "business")
			}
			var rows []model.Service
			if err := query.Order("id desc").Limit(limit).Find(&rows).Error; err != nil {
				return nil, "db", err
			}
			list := make([]map[string]any, 0, len(rows))
			for _, item := range rows {
				list = append(list, map[string]any{
					"id":           item.ID,
					"name":         item.Name,
					"owner":        item.Owner,
					"env":          item.Env,
					"service_kind": item.ServiceKind,
					"visibility":   item.Visibility,
					"deploy_count": item.DeployCount,
					"icon":         item.Icon,
				})
			}
			return map[string]any{
				"total": len(list),
				"list":  list,
				"filters_applied": map[string]any{
					"keyword":     strings.TrimSpace(in.Keyword),
					"category_id": in.CategoryID,
					"limit":       limit,
				},
			}, "db", nil
		},
	)
}

func ServiceCategoryTree(ctx context.Context, deps PlatformDeps) (ToolResult, error) {
	return RunWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:        "service_category_tree",
			Description: "查询服务分类树。当前分类来自服务类型聚合。示例: {}。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			SceneScope:  []string{"services:list", "services:catalog"},
		},
		struct{}{},
		func(_ struct{}) (any, string, error) {
			if deps.DB == nil {
				return nil, "db", fmt.Errorf("db unavailable")
			}
			type countRow struct {
				ServiceKind string
				Count       int64
			}
			var rows []countRow
			if err := deps.DB.Model(&model.Service{}).
				Select("service_kind, COUNT(1) AS count").
				Group("service_kind").
				Scan(&rows).Error; err != nil {
				return nil, "db", err
			}
			tree := []map[string]any{
				{"id": 1, "key": "middleware", "label": "中间件服务", "count": int64(0)},
				{"id": 2, "key": "business", "label": "业务服务", "count": int64(0)},
			}
			for _, row := range rows {
				switch strings.TrimSpace(row.ServiceKind) {
				case "middleware":
					tree[0]["count"] = row.Count
				case "business":
					tree[1]["count"] = row.Count
				}
			}
			return map[string]any{"tree": tree}, "db", nil
		},
	)
}

func ServiceVisibilityCheck(ctx context.Context, deps PlatformDeps, input ServiceVisibilityCheckInput) (ToolResult, error) {
	return RunWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:        "service_visibility_check",
			Description: "查询服务可见性配置。service_id 必填。示例: {\"service_id\":123}。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			Required:    []string{"service_id"},
			EnumSources: map[string]string{"service_id": "service_list_inventory"},
			ParamHints:  map[string]string{"service_id": "可从 service_list_inventory 获取"},
			SceneScope:  []string{"services:detail", "services:catalog"},
		},
		input,
		func(in ServiceVisibilityCheckInput) (any, string, error) {
			if deps.DB == nil {
				return nil, "db", fmt.Errorf("db unavailable")
			}
			if in.ServiceID <= 0 {
				return nil, "validation", NewMissingParam("service_id", "service_id is required")
			}
			var svc model.Service
			if err := deps.DB.First(&svc, in.ServiceID).Error; err != nil {
				return nil, "db", err
			}
			granted := []uint{}
			if strings.TrimSpace(svc.GrantedTeams) != "" {
				_ = json.Unmarshal([]byte(svc.GrantedTeams), &granted)
			}
			return map[string]any{
				"service_id":    svc.ID,
				"service_name":  svc.Name,
				"service_kind":  svc.ServiceKind,
				"visibility":    svc.Visibility,
				"granted_teams": granted,
				"owner_user_id": svc.OwnerUserID,
				"owner_team_id": svc.TeamID,
				"updated_at":    svc.UpdatedAt,
			}, "db", nil
		},
	)
}

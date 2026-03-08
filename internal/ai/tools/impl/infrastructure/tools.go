package infrastructure

import (
	"context"
	"fmt"
	"strings"

	. "github.com/cy77cc/k8s-manage/internal/ai/tools/core"
	"github.com/cy77cc/k8s-manage/internal/model"
)

func CredentialList(ctx context.Context, deps PlatformDeps, input CredentialListInput) (ToolResult, error) {
	return RunWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:        "credential_list",
			Description: "查询凭证列表。可选参数 type/keyword/limit。示例: {\"type\":\"k8s\",\"limit\":20}。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			DefaultHint: map[string]any{"limit": 50},
			SceneScope:  []string{"deployment:credentials"},
		},
		input,
		func(in CredentialListInput) (any, string, error) {
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
			query := deps.DB.Model(&model.ClusterCredential{})
			if t := strings.TrimSpace(in.Type); t != "" {
				query = query.Where("runtime_type = ? OR source = ?", t, t)
			}
			if kw := strings.TrimSpace(in.Keyword); kw != "" {
				pattern := "%" + kw + "%"
				query = query.Where("name LIKE ? OR endpoint LIKE ?", pattern, pattern)
			}
			var rows []model.ClusterCredential
			if err := query.Order("id desc").Limit(limit).Find(&rows).Error; err != nil {
				return nil, "db", err
			}
			list := make([]map[string]any, 0, len(rows))
			for _, item := range rows {
				list = append(list, map[string]any{
					"id":                item.ID,
					"name":              item.Name,
					"runtime_type":      item.RuntimeType,
					"source":            item.Source,
					"endpoint":          item.Endpoint,
					"status":            item.Status,
					"last_test_at":      item.LastTestAt,
					"last_test_status":  item.LastTestStatus,
					"last_test_message": item.LastTestMessage,
				})
			}
			return map[string]any{"total": len(list), "list": list}, "db", nil
		},
	)
}

func CredentialTest(ctx context.Context, deps PlatformDeps, input CredentialTestInput) (ToolResult, error) {
	return RunWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:        "credential_test",
			Description: "查询凭证连通性测试结果。credential_id 必填。示例: {\"credential_id\":5}。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			Required:    []string{"credential_id"},
			EnumSources: map[string]string{"credential_id": "credential_list"},
			SceneScope:  []string{"deployment:credentials"},
		},
		input,
		func(in CredentialTestInput) (any, string, error) {
			if deps.DB == nil {
				return nil, "db", fmt.Errorf("db unavailable")
			}
			if in.CredentialID <= 0 {
				return nil, "validation", NewMissingParam("credential_id", "credential_id is required")
			}
			var cred model.ClusterCredential
			if err := deps.DB.First(&cred, in.CredentialID).Error; err != nil {
				return nil, "db", err
			}
			return map[string]any{
				"credential_id":     cred.ID,
				"name":              cred.Name,
				"status":            cred.Status,
				"last_test_at":      cred.LastTestAt,
				"last_test_status":  cred.LastTestStatus,
				"last_test_message": cred.LastTestMessage,
			}, "db", nil
		},
	)
}

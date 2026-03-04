package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/cy77cc/k8s-manage/internal/model"
)

func userList(ctx context.Context, deps PlatformDeps, input UserListInput) (ToolResult, error) {
	return runWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:        "user_list",
			Description: "查询用户列表。可选 keyword/status/limit。示例: {\"keyword\":\"alice\"}。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			DefaultHint: map[string]any{"limit": 50},
			SceneScope:  []string{"governance:users"},
		},
		input,
		func(in UserListInput) (any, string, error) {
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
			query := deps.DB.Model(&model.User{})
			if in.Status != 0 {
				query = query.Where("status = ?", in.Status)
			}
			if kw := strings.TrimSpace(in.Keyword); kw != "" {
				pattern := "%" + kw + "%"
				query = query.Where("username LIKE ? OR email LIKE ?", pattern, pattern)
			}
			var users []model.User
			if err := query.Order("id desc").Limit(limit).Find(&users).Error; err != nil {
				return nil, "db", err
			}
			return map[string]any{"total": len(users), "list": users}, "db", nil
		},
	)
}

func roleList(ctx context.Context, deps PlatformDeps, input RoleListInput) (ToolResult, error) {
	return runWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:        "role_list",
			Description: "查询角色列表。可选 keyword/limit。示例: {\"keyword\":\"admin\"}。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			DefaultHint: map[string]any{"limit": 50},
			SceneScope:  []string{"governance:roles"},
		},
		input,
		func(in RoleListInput) (any, string, error) {
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
			query := deps.DB.Model(&model.Role{})
			if kw := strings.TrimSpace(in.Keyword); kw != "" {
				pattern := "%" + kw + "%"
				query = query.Where("name LIKE ? OR code LIKE ?", pattern, pattern)
			}
			var roles []model.Role
			if err := query.Order("id desc").Limit(limit).Find(&roles).Error; err != nil {
				return nil, "db", err
			}
			return map[string]any{"total": len(roles), "list": roles}, "db", nil
		},
	)
}

func permissionCheck(ctx context.Context, deps PlatformDeps, input PermissionCheckInput) (ToolResult, error) {
	return runWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:        "permission_check",
			Description: "检查用户权限。user_id/resource/action 必填。示例: {\"user_id\":1,\"resource\":\"service\",\"action\":\"read\"}。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			Required:    []string{"user_id", "resource", "action"},
			EnumSources: map[string]string{"user_id": "user_list"},
			SceneScope:  []string{"governance:permissions"},
		},
		input,
		func(in PermissionCheckInput) (any, string, error) {
			if deps.DB == nil {
				return nil, "db", fmt.Errorf("db unavailable")
			}
			if in.UserID <= 0 {
				return nil, "validation", NewMissingParam("user_id", "user_id is required")
			}
			resource := strings.TrimSpace(in.Resource)
			action := strings.TrimSpace(in.Action)
			if resource == "" {
				return nil, "validation", NewMissingParam("resource", "resource is required")
			}
			if action == "" {
				return nil, "validation", NewMissingParam("action", "action is required")
			}

			var roleIDs []int64
			if err := deps.DB.Model(&model.UserRole{}).Where("user_id = ?", in.UserID).Pluck("role_id", &roleIDs).Error; err != nil {
				return nil, "db", err
			}
			if len(roleIDs) == 0 {
				return map[string]any{"allowed": false, "reason": "user has no roles"}, "db", nil
			}
			var permIDs []int64
			if err := deps.DB.Model(&model.RolePermission{}).Where("role_id IN ?", roleIDs).Pluck("permission_id", &permIDs).Error; err != nil {
				return nil, "db", err
			}
			if len(permIDs) == 0 {
				return map[string]any{"allowed": false, "reason": "roles have no permissions"}, "db", nil
			}
			var perms []model.Permission
			if err := deps.DB.Where("id IN ?", permIDs).Find(&perms).Error; err != nil {
				return nil, "db", err
			}
			matched := make([]model.Permission, 0)
			for _, perm := range perms {
				if strings.EqualFold(strings.TrimSpace(perm.Resource), resource) && strings.EqualFold(strings.TrimSpace(perm.Action), action) {
					matched = append(matched, perm)
				}
			}
			return map[string]any{"allowed": len(matched) > 0, "matched_permissions": matched, "checked": map[string]any{"user_id": in.UserID, "resource": resource, "action": action}}, "db", nil
		},
	)
}

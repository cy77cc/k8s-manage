package ai

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/cy77cc/k8s-manage/internal/model"
	"gorm.io/gorm"
)

type ToolResourceRule struct {
	ToolName          string
	ResourceType      string
	ResourceIDParam   string
	WritePermission   string
	ApprovePermission string
}

// ToolResourceMapping defines how tool calls map to resources and permission codes.
type ToolResourceMapping map[string]ToolResourceRule

type PermissionChecker struct {
	db      *gorm.DB
	mapping ToolResourceMapping
}

func NewPermissionChecker(db *gorm.DB) *PermissionChecker {
	return &PermissionChecker{
		db:      db,
		mapping: defaultToolResourceMapping(),
	}
}

func defaultToolResourceMapping() ToolResourceMapping {
	return ToolResourceMapping{
		"host_batch_exec_apply": {
			ToolName:          "host_batch_exec_apply",
			ResourceType:      "host",
			ResourceIDParam:   "host_ids",
			WritePermission:   "host:write",
			ApprovePermission: "host:approve",
		},
		"service_deploy_apply": {
			ToolName:          "service_deploy_apply",
			ResourceType:      "service",
			ResourceIDParam:   "service_id",
			WritePermission:   "service:deploy",
			ApprovePermission: "service:approve",
		},
		"k8s_apply_manifest": {
			ToolName:          "k8s_apply_manifest",
			ResourceType:      "cluster",
			ResourceIDParam:   "cluster_id",
			WritePermission:   "k8s:write",
			ApprovePermission: "k8s:approve",
		},
	}
}

func (p *PermissionChecker) Rule(toolName string) (ToolResourceRule, bool) {
	if p == nil {
		return ToolResourceRule{}, false
	}
	rule, ok := p.mapping[strings.TrimSpace(toolName)]
	return rule, ok
}

func (p *PermissionChecker) CheckPermission(ctx context.Context, userID uint64, toolName string) (bool, error) {
	if p == nil || p.db == nil {
		return false, errors.New("permission checker not initialized")
	}
	if userID == 0 {
		return false, nil
	}
	rule, ok := p.Rule(toolName)
	if !ok {
		return false, nil
	}
	return p.hasPermission(ctx, userID, rule.WritePermission)
}

func (p *PermissionChecker) FindApprovers(ctx context.Context, toolName string, resourceID string) ([]uint64, error) {
	if p == nil || p.db == nil {
		return nil, errors.New("permission checker not initialized")
	}
	rule, ok := p.Rule(toolName)
	if !ok {
		return nil, fmt.Errorf("tool mapping not found: %s", toolName)
	}

	result := make([]uint64, 0, 8)
	seen := make(map[uint64]struct{}, 8)
	addUsers := func(items []uint64) {
		for _, uid := range items {
			if uid == 0 {
				continue
			}
			if _, ok := seen[uid]; ok {
				continue
			}
			seen[uid] = struct{}{}
			result = append(result, uid)
		}
	}

	withPerm, err := p.listUsersByPermission(ctx, rule.ApprovePermission)
	if err != nil {
		return nil, err
	}
	addUsers(withPerm)

	owner, err := p.findResourceOwner(ctx, rule.ResourceType, resourceID)
	if err != nil {
		return nil, err
	}
	if owner != 0 {
		addUsers([]uint64{owner})
	}

	admins, err := p.listUsersByRoleCode(ctx, "admin")
	if err != nil {
		return nil, err
	}
	addUsers(admins)

	return result, nil
}

func (p *PermissionChecker) hasPermission(ctx context.Context, userID uint64, permissionCode string) (bool, error) {
	codes, err := p.listPermissionCodes(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, code := range codes {
		if code == permissionCode || code == "*:*" {
			return true, nil
		}
		prefix := strings.Split(permissionCode, ":")[0]
		if code == prefix+":*" {
			return true, nil
		}
	}
	return false, nil
}

func (p *PermissionChecker) listPermissionCodes(ctx context.Context, userID uint64) ([]string, error) {
	type row struct {
		Code string `gorm:"column:code"`
	}
	var rows []row
	err := p.db.WithContext(ctx).
		Table("permissions").
		Select("permissions.code").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ?", userID).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, item := range rows {
		code := strings.TrimSpace(item.Code)
		if code == "" {
			continue
		}
		out = append(out, code)
	}
	return out, nil
}

func (p *PermissionChecker) listUsersByPermission(ctx context.Context, permissionCode string) ([]uint64, error) {
	type row struct {
		UserID uint64 `gorm:"column:user_id"`
	}
	var rows []row
	err := p.db.WithContext(ctx).
		Table("user_roles").
		Select("user_roles.user_id").
		Joins("JOIN role_permissions ON role_permissions.role_id = user_roles.role_id").
		Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
		Where("permissions.code = ?", permissionCode).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]uint64, 0, len(rows))
	for _, item := range rows {
		out = append(out, item.UserID)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out, nil
}

func (p *PermissionChecker) listUsersByRoleCode(ctx context.Context, roleCode string) ([]uint64, error) {
	type row struct {
		UserID uint64 `gorm:"column:user_id"`
	}
	var rows []row
	err := p.db.WithContext(ctx).
		Table("user_roles").
		Select("user_roles.user_id").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("roles.code = ?", roleCode).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]uint64, 0, len(rows))
	for _, item := range rows {
		out = append(out, item.UserID)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out, nil
}

func (p *PermissionChecker) findResourceOwner(ctx context.Context, resourceType, resourceID string) (uint64, error) {
	switch strings.TrimSpace(resourceType) {
	case "service":
		var svc model.Service
		if err := p.db.WithContext(ctx).Select("owner_user_id").Where("id = ?", resourceID).First(&svc).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return 0, nil
			}
			return 0, err
		}
		return uint64(svc.OwnerUserID), nil
	case "host":
		return 0, nil
	default:
		return 0, nil
	}
}

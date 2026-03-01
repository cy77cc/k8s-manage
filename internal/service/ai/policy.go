package ai

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/cy77cc/k8s-manage/internal/ai/tools"
	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/gin-gonic/gin"
)

func (h *handler) toolPolicy(ctx context.Context, meta tools.ToolMeta, params map[string]any) error {
	uid, approvalToken := tools.ToolUserFromContext(ctx)
	if uid == 0 {
		return errors.New("unauthorized")
	}
	if !h.hasPermission(uid, meta.Permission) {
		return errors.New("permission denied")
	}
	if meta.Mode == tools.ToolModeReadonly {
		return nil
	}
	if strings.TrimSpace(approvalToken) == "" {
		t := h.store.newApproval(uid, approvalTicket{
			Tool:   meta.Name,
			Params: params,
			Risk:   meta.Risk,
			Mode:   meta.Mode,
			Meta:   meta,
		})
		return &tools.ApprovalRequiredError{
			Token:     t.ID,
			Tool:      t.Tool,
			ExpiresAt: t.ExpiresAt,
			Message:   "approval required",
		}
	}
	t, ok := h.store.getApproval(approvalToken)
	if !ok {
		return errors.New("approval not found")
	}
	if t.Tool != meta.Name {
		return errors.New("approval tool mismatch")
	}
	if t.RequestUID != uid && !h.isAdmin(uid) {
		return errors.New("approval owner mismatch")
	}
	if time.Now().After(t.ExpiresAt) {
		return errors.New("approval expired")
	}
	if t.Status != "approved" {
		return errors.New("approval not approved")
	}
	return nil
}

func (h *handler) hasPermission(uid uint64, code string) bool {
	if uid == 0 {
		return false
	}
	if h.isAdmin(uid) {
		return true
	}
	if code == "" {
		return true
	}
	perms, err := h.fetchPermissions(uid)
	if err != nil {
		return false
	}
	parts := strings.Split(code, ":")
	resource := code
	if len(parts) > 0 {
		resource = parts[0]
	}
	for _, p := range perms {
		if p == code || p == resource+":*" || p == "*:*" {
			return true
		}
	}
	return false
}

func (h *handler) isAdmin(uid uint64) bool {
	var u model.User
	if err := h.svcCtx.DB.Select("id", "username").Where("id = ?", uid).First(&u).Error; err == nil {
		if strings.EqualFold(strings.TrimSpace(u.Username), "admin") {
			return true
		}
	}
	type roleRow struct {
		Code string `gorm:"column:code"`
	}
	var rows []roleRow
	err := h.svcCtx.DB.Table("roles").
		Select("roles.code").
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", uid).
		Scan(&rows).Error
	if err != nil {
		return false
	}
	for _, row := range rows {
		if strings.EqualFold(strings.TrimSpace(row.Code), "admin") {
			return true
		}
	}
	return false
}

func (h *handler) fetchPermissions(uid uint64) ([]string, error) {
	type row struct {
		Code string `gorm:"column:code"`
	}
	var rows []row
	err := h.svcCtx.DB.Table("permissions").
		Select("permissions.code").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ?", uid).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.Code)
	}
	return out, nil
}

func uidFromContext(c *gin.Context) (uint64, bool) {
	v, ok := c.Get("uid")
	if !ok {
		return 0, false
	}
	switch x := v.(type) {
	case uint:
		return uint64(x), true
	case uint64:
		return x, true
	case int:
		return uint64(x), true
	case int64:
		return uint64(x), true
	case float64:
		return uint64(x), true
	default:
		return 0, false
	}
}

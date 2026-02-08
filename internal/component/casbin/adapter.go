package adapter

import (
	"fmt"

	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"gorm.io/gorm"
)

type Adapter struct {
	db *gorm.DB
}

func NewAdapter(db *gorm.DB) *Adapter {
	return &Adapter{db: db}
}

// LoadPolicy loads all policy rules from the storage.
func (a *Adapter) LoadPolicy(model model.Model) error {
	// 1. Load Role Policies (p, role, resource, action)
	// Query: SELECT r.code, p.resource, p.action FROM roles r JOIN role_permissions rp ON r.id = rp.role_id JOIN permissions p ON p.id = rp.permission_id WHERE r.status = 1 AND p.status = 1
	type PolicyResult struct {
		RoleCode string
		Resource string
		Action   string
	}
	var policies []PolicyResult
	err := a.db.Table("roles").
		Select("roles.code as role_code, permissions.resource, permissions.action").
		Joins("JOIN role_permissions ON roles.id = role_permissions.role_id").
		Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
		Where("roles.status = 1 AND permissions.status = 1").
		Scan(&policies).Error
	if err != nil {
		return err
	}

	for _, policy := range policies {
		persist.LoadPolicyLine(fmt.Sprintf("p, %s, %s, %s", policy.RoleCode, policy.Resource, policy.Action), model)
	}

	// 2. Load User Role Inheritance (g, user_id, role)
	// Query: SELECT u.id, r.code FROM users u JOIN user_roles ur ON u.id = ur.user_id JOIN roles r ON r.id = ur.role_id WHERE u.status = 1 AND r.status = 1
	type GroupResult struct {
		UserID   int64
		RoleCode string
	}
	var groups []GroupResult
	err = a.db.Table("users").
		Select("users.id as user_id, roles.code as role_code").
		Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("users.status = 1 AND roles.status = 1").
		Scan(&groups).Error
	if err != nil {
		return err
	}

	for _, group := range groups {
		persist.LoadPolicyLine(fmt.Sprintf("g, %d, %s", group.UserID, group.RoleCode), model)
	}

	return nil
}

// SavePolicy saves all policy rules to the storage.
func (a *Adapter) SavePolicy(model model.Model) error {
	return fmt.Errorf("not implemented: read-only adapter")
}

// AddPolicy adds a policy rule to the storage.
func (a *Adapter) AddPolicy(sec string, ptype string, rule []string) error {
	return fmt.Errorf("not implemented: read-only adapter")
}

// RemovePolicy removes a policy rule from the storage.
func (a *Adapter) RemovePolicy(sec string, ptype string, rule []string) error {
	return fmt.Errorf("not implemented: read-only adapter")
}

// RemoveFilteredPolicy removes policy rules that match the filter from the storage.
func (a *Adapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	return fmt.Errorf("not implemented: read-only adapter")
}

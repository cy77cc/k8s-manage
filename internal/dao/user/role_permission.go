package dao

import (
	"context"
	"github.com/cy77cc/k8s-manage/internal/model"
	"gorm.io/gorm"
)

type RolePermissionDAO struct {
	db *gorm.DB
}

func NewRolePermissionDAO(db *gorm.DB) *RolePermissionDAO {
	return &RolePermissionDAO{db: db}
}

func (d *RolePermissionDAO) Create(ctx context.Context, rolePermission *model.RolePermission) error {
	return d.db.WithContext(ctx).Create(rolePermission).Error
}

func (d *RolePermissionDAO) Delete(ctx context.Context, id int64) error {
	return d.db.WithContext(ctx).Delete(&model.RolePermission{}, id).Error
}

func (d *RolePermissionDAO) GetByRoleID(ctx context.Context, roleID int64) ([]model.RolePermission, error) {
	var rolePermissions []model.RolePermission
	err := d.db.WithContext(ctx).Where("role_id = ?", roleID).Find(&rolePermissions).Error
	if err != nil {
		return nil, err
	}
	return rolePermissions, nil
}

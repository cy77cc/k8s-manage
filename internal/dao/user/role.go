package dao

import (
	"context"
	"github.com/cy77cc/k8s-manage/internal/model"
	"gorm.io/gorm"
)

type RoleDAO struct {
	db *gorm.DB
}

func NewRoleDAO(db *gorm.DB) *RoleDAO {
	return &RoleDAO{db: db}
}

func (d *RoleDAO) Create(ctx context.Context, role *model.Role) error {
	return d.db.WithContext(ctx).Create(role).Error
}

func (d *RoleDAO) Update(ctx context.Context, role *model.Role) error {
	return d.db.WithContext(ctx).Save(role).Error
}

func (d *RoleDAO) Delete(ctx context.Context, id int64) error {
	return d.db.WithContext(ctx).Delete(&model.Role{}, id).Error
}

func (d *RoleDAO) GetByID(ctx context.Context, id int64) (*model.Role, error) {
	var role model.Role
	err := d.db.WithContext(ctx).First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

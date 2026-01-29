package user

import (
	"context"

	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type RoleDAO struct {
	db    *gorm.DB
	cache *expirable.LRU[string, any]
	rdb   redis.UniversalClient
}

func NewRoleDAO(db *gorm.DB, cache *expirable.LRU[string, any], rdb redis.UniversalClient) *RoleDAO {
	return &RoleDAO{db: db, cache: cache, rdb: rdb}
}

func (d *RoleDAO) Create(ctx context.Context, role *model.Role) error {
	return d.db.WithContext(ctx).Create(role).Error
}

func (d *RoleDAO) Update(ctx context.Context, role *model.Role) error {
	return d.db.WithContext(ctx).Save(role).Error
}

func (d *RoleDAO) Delete(ctx context.Context, id model.UserID) error {
	return d.db.WithContext(ctx).Delete(&model.Role{}, id).Error
}

func (d *RoleDAO) GetByID(ctx context.Context, id model.UserID) (*model.Role, error) {
	var role model.Role
	err := d.db.WithContext(ctx).First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

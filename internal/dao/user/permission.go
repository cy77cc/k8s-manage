package user

import (
	"context"

	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/storage"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type PermissionDAO struct {
	db       *gorm.DB
	cache    *storage.Cache[string, any]
	redisCli redis.UniversalClient
}

func NewPermissionDAO(db *gorm.DB, cache *storage.Cache[string, any], redisCli redis.UniversalClient) *PermissionDAO {
	return &PermissionDAO{db: db, cache: cache, redisCli: redisCli}
}

func (d *PermissionDAO) Create(ctx context.Context, permission *model.Permission) error {
	return d.db.WithContext(ctx).Create(permission).Error
}

func (d *PermissionDAO) Update(ctx context.Context, permission *model.Permission) error {
	return d.db.WithContext(ctx).Save(permission).Error
}

func (d *PermissionDAO) Delete(ctx context.Context, id int64) error {
	return d.db.WithContext(ctx).Delete(&model.Permission{}, id).Error
}

func (d *PermissionDAO) GetByID(ctx context.Context, id int64) (*model.Permission, error) {
	var permission model.Permission
	err := d.db.WithContext(ctx).First(&permission, id).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

package user

import (
	"context"

	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type UserRoleDAO struct {
	db       *gorm.DB
	cache    *expirable.LRU[string, any]
	redisCli redis.UniversalClient
}

func NewUserRoleDAO(db *gorm.DB, cache *expirable.LRU[string, any], redisCli redis.UniversalClient) *UserRoleDAO {
	return &UserRoleDAO{db: db, cache: cache, redisCli: redisCli}
}

func (d *UserRoleDAO) Create(ctx context.Context, userRole *model.UserRole) error {
	return d.db.WithContext(ctx).Create(userRole).Error
}

func (d *UserRoleDAO) Delete(ctx context.Context, id int64) error {
	return d.db.WithContext(ctx).Delete(&model.UserRole{}, id).Error
}

func (d *UserRoleDAO) GetByUserID(ctx context.Context, userID int64) ([]model.UserRole, error) {
	var userRoles []model.UserRole
	err := d.db.WithContext(ctx).Where("user_id = ?", userID).Find(&userRoles).Error
	if err != nil {
		return nil, err
	}
	return userRoles, nil
}

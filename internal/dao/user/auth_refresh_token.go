package user

import (
	"context"

	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type AuthRefreshTokenDAO struct {
	db       *gorm.DB
	cache    *expirable.LRU[string, any]
	redisCli redis.UniversalClient
}

func NewAuthRefreshTokenDAO(db *gorm.DB, cache *expirable.LRU[string, any], redisCli redis.UniversalClient) *AuthRefreshTokenDAO {
	return &AuthRefreshTokenDAO{db: db, cache: cache, redisCli: redisCli}
}

func (d *AuthRefreshTokenDAO) Create(ctx context.Context, token *model.AuthRefreshToken) error {
	return d.db.WithContext(ctx).Create(token).Error
}

func (d *AuthRefreshTokenDAO) Update(ctx context.Context, token *model.AuthRefreshToken) error {
	return d.db.WithContext(ctx).Save(token).Error
}

func (d *AuthRefreshTokenDAO) Delete(ctx context.Context, id int64) error {
	return d.db.WithContext(ctx).Delete(&model.AuthRefreshToken{}, id).Error
}

func (d *AuthRefreshTokenDAO) GetByToken(ctx context.Context, token string) (*model.AuthRefreshToken, error) {
	var refreshToken model.AuthRefreshToken
	err := d.db.WithContext(ctx).Where("token = ?", token).First(&refreshToken).Error
	if err != nil {
		return nil, err
	}
	return &refreshToken, nil
}

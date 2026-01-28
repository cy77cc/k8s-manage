package user

import (
	"context"
	"time"

	"github.com/cy77cc/k8s-manage/internal/consts"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type BlackListDao struct {
	db    *gorm.DB
	cache *expirable.LRU[string, any]
	rdb   redis.UniversalClient
}

func NewBlackListDao(db *gorm.DB, cache *expirable.LRU[string, any], rdb redis.UniversalClient) *BlackListDao {
	return &BlackListDao{db: db, cache: cache, rdb: rdb}
}

func (b *BlackListDao) AddToBlacklist(ctx context.Context, jti string, exp time.Time) error {
	ttl := time.Until(exp)
	return b.rdb.Set(ctx, consts.JwtBlackListKey+jti, 1, ttl).Err()
}

// 检查是否在黑名单
func (b *BlackListDao) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	res, err := b.rdb.Exists(ctx, consts.JwtBlackListKey+jti).Result()
	return res > 0, err
}

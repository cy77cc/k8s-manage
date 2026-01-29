package user

import (
	"context"
	"time"

	"github.com/cy77cc/k8s-manage/internal/constants"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type WhiteListDao struct {
	db    *gorm.DB
	cache *expirable.LRU[string, any]
	rdb   redis.UniversalClient
}

func NewWhiteListDao(db *gorm.DB, cache *expirable.LRU[string, any], rdb redis.UniversalClient) *WhiteListDao {
	return &WhiteListDao{db: db, cache: cache, rdb: rdb}
}

func (w *WhiteListDao) AddToWhitelist(ctx context.Context, token string, exp time.Time) error {
	ttl := time.Until(exp)
	return w.rdb.Set(ctx, constants.JwtWhiteListKey+token, 1, ttl).Err()
}

func (w *WhiteListDao) DeleteToken(ctx context.Context, token string) error {
	return w.rdb.Del(ctx, constants.JwtWhiteListKey+token).Err()
}

// 检查是否在黑名单
func (w *WhiteListDao) IsWhitelisted(ctx context.Context, token string) (bool, error) {
	res, err := w.rdb.Exists(ctx, constants.JwtWhiteListKey+token).Result()
	return res > 0, err
}

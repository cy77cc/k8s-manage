package user

import (
	"github.com/cy77cc/k8s-manage/storage"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type BlackListDao struct {
	db *gorm.DB
	cache    *storage.Cache[string, any]
	redisCli redis.UniversalClient
}

func NewBlackListDao(db *gorm.DB, cache *storage.Cache[string, any], redisCli redis.UniversalClient) *BlackListDao {
	return &BlackListDao{db: db, cache: cache, redisCli: redisCli}
}

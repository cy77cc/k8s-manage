package user

import (
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type BlackListDao struct {
	db       *gorm.DB
	cache    *expirable.LRU[string, any]
	redisCli redis.UniversalClient
}

func NewBlackListDao(db *gorm.DB, cache *expirable.LRU[string, any], redisCli redis.UniversalClient) *BlackListDao {
	return &BlackListDao{db: db, cache: cache, redisCli: redisCli}
}

package svc

import (
	"time"

	"github.com/cy77cc/k8s-manage/storage"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"k8s.io/client-go/kubernetes"
)

type ServiceContext struct {
	Clientset *kubernetes.Clientset
	DB        *gorm.DB
	Rdb     redis.UniversalClient
	Cache     *expirable.LRU[string, any]
}

func MustNewServiceContext() *ServiceContext {

	return &ServiceContext{
		DB:    storage.MustNewDB(),
		Rdb: storage.MustNewRdb(),
		Cache: expirable.NewLRU[string, any](5_000, nil, 24*time.Hour),
	}
}

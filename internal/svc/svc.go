package svc

import (
	"github.com/cy77cc/k8s-manage/storage"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"k8s.io/client-go/kubernetes"
)

type ServiceContext struct {
	Clientset *kubernetes.Clientset
	DB        *gorm.DB
	Redis     redis.UniversalClient
	Cache     *storage.Cache[string, any]
}

func MustNewServiceContext() *ServiceContext {

	return &ServiceContext{
		DB:    storage.MustNewDB(),
		Redis: storage.MustNewRedisClient(),
		Cache: storage.NewCache[string, any](5_000),
	}
}

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
}

func MustNewServiceContext() *ServiceContext {

	return &ServiceContext{
		DB:    storage.MustNewDB(),
		Redis: storage.MustNewRedisClient(),
	}
}

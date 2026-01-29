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
	Clientset *kubernetes.Clientset       // K8s 客户端
	DB        *gorm.DB                    // GORM 数据库实例
	Rdb       redis.UniversalClient       // Redis 客户端
	Cache     *expirable.LRU[string, any] // 本地缓存 (LRU)
}

// MustNewServiceContext 创建服务上下文，如果失败则 panic
func MustNewServiceContext() *ServiceContext {

	return &ServiceContext{
		DB:    storage.MustNewDB(),
		Rdb: storage.MustNewRdb(),
		Cache: expirable.NewLRU[string, any](5_000, nil, 24*time.Hour),
	}
}

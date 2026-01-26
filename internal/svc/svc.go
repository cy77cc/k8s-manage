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
	// configK8s, err := rest.InClusterConfig()
	// if err != nil {
	// 	log.Fatal("init clientset err", err)
	// }

	// var clientset *kubernetes.Clientset
	// if configK8s != nil {
	// 	clientset, err = kubernetes.NewForConfig(configK8s)
	// 	if err != nil {
	// 		log.Fatal(err.Error())
	// 	}
	// }

	return &ServiceContext{
		// Clientset: clientset,
		DB:    storage.MustNewDB(),
		Redis: storage.MustNewRedisClient(),
	}
}

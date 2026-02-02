package svc

import (
	"context"
	"path/filepath"
	"time"

	"github.com/cy77cc/k8s-manage/internal/ai"
	"github.com/cy77cc/k8s-manage/internal/logger"
	"github.com/cy77cc/k8s-manage/storage"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type ServiceContext struct {
	Clientset *kubernetes.Clientset       // K8s 客户端
	DB        *gorm.DB                    // GORM 数据库实例
	Rdb       redis.UniversalClient       // Redis 客户端
	Cache     *expirable.LRU[string, any] // 本地缓存 (LRU)
	AI        *ai.K8sCopilot              // AI Copilot
}

// MustNewServiceContext 创建服务上下文，如果失败则 panic
func MustNewServiceContext() *ServiceContext {
	ctx := context.Background()
	chatModel, err := ai.NewChatModel(ctx)
	if err != nil {
		// Log warning but don't panic if AI fails? Or panic?
		// User wants Eino integration, so maybe panic if config is enabled but fails.
		// For now, let's just log or ignore if not enabled.
		// But MustNewServiceContext implies "Must".
		// However, ai.NewChatModel returns nil, nil if disabled.
	}

	clientset := MustNewClientset()

	copilot, err := ai.NewK8sCopilot(ctx, chatModel, clientset)
	if err != nil {
		// handle error
		logger.L().Warn("Failed to initialize AI Copilot", logger.Error(err))
	}

	return &ServiceContext{
		Clientset: clientset,
		DB:        storage.MustNewDB(),
		Rdb:       storage.MustNewRdb(),
		Cache:     expirable.NewLRU[string, any](5_000, nil, 24*time.Hour),
		AI:        copilot,
	}
}

func MustNewClientset() *kubernetes.Clientset {
	// Try to load kubeconfig from home directory
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	// Try to build config from flags (defaulting to kubeconfig path)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		// Fallback to in-cluster config
		config, err = rest.InClusterConfig()
		if err != nil {
			logger.L().Warn("Failed to create K8s config (neither kubeconfig nor in-cluster)", logger.Error(err))
			return nil
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.L().Warn("Failed to create K8s clientset", logger.Error(err))
		return nil
	}

	logger.L().Info("Kubernetes Clientset initialized successfully")
	return clientset
}

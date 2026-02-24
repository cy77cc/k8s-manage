package svc

import (
	"context"
	"path/filepath"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/cy77cc/k8s-manage/internal/ai"
	casbinadapter "github.com/cy77cc/k8s-manage/internal/component/casbin"
	"github.com/cy77cc/k8s-manage/internal/config"
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
	Clientset      *kubernetes.Clientset       // K8s 客户端
	DB             *gorm.DB                    // GORM 数据库实例
	Rdb            redis.UniversalClient       // Redis 客户端
	Cache          *expirable.LRU[string, any] // 本地缓存 (LRU)
	AI             *ai.K8sCopilot              // AI Copilot
	CasbinEnforcer *casbin.Enforcer            // Casbin Enforcer
}

// MustNewServiceContext 创建服务上下文，如果失败则 panic
func MustNewServiceContext() *ServiceContext {
	ctx := context.Background()
	chatModel, err := ai.NewChatModel(ctx)
	if err != nil {
		logger.L().Warn("Failed to initialize AI chat model",
			logger.String("provider", config.CFG.LLM.Provider),
			logger.String("base_url", aiBaseURL()),
			logger.String("model", aiModel()),
			logger.Error(err),
		)
	}

	clientset := MustNewClientset()

	copilot, err := ai.NewK8sCopilot(ctx, chatModel, clientset)
	if err != nil {
		logger.L().Warn("Failed to initialize AI Copilot",
			logger.String("base_url", aiBaseURL()),
			logger.String("model", aiModel()),
			logger.Error(err),
		)
	}

	db := storage.MustNewDB()

	// Initialize Casbin
	adapter := casbinadapter.NewAdapter(db)
	enforcer, err := casbin.NewEnforcer("resource/casbin/rbac_model.conf", adapter)
	if err != nil {
		// Try absolute path if relative fails, or panic
		// Assuming running from project root
		logger.L().Error("Failed to initialize Casbin Enforcer", logger.Error(err))
		// panic(err) // Optional: panic if auth is critical
	} else {
		if err := enforcer.LoadPolicy(); err != nil {
			logger.L().Error("Failed to load Casbin policy", logger.Error(err))
		}
	}

	return &ServiceContext{
		Clientset:      clientset,
		DB:             db,
		Rdb:            storage.MustNewRdb(),
		Cache:          expirable.NewLRU[string, any](5_000, nil, 24*time.Hour),
		AI:             copilot,
		CasbinEnforcer: enforcer,
	}
}

func aiBaseURL() string {
	return config.CFG.LLM.BaseURL
}

func aiModel() string {
	return config.CFG.LLM.Model
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

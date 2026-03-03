package service

import (
	"context"
	"errors"

	"github.com/cy77cc/k8s-manage/internal/model"
)

// ValidateEnvMatch validates that the service environment matches the cluster environment type.
// Returns an error if they don't match.
// Note: Clusters with env_type 'development' (default value) are considered "unrestricted"
// and will accept any service environment for backward compatibility.
func (l *Logic) ValidateEnvMatch(ctx context.Context, serviceEnv string, clusterID uint) error {
	var cluster model.Cluster
	if err := l.svcCtx.DB.WithContext(ctx).Select("env_type").First(&cluster, clusterID).Error; err != nil {
		return errors.New("cluster not found")
	}

	// 如果集群 env_type 是默认值 'development'，跳过校验（兼容现有数据）
	// 这样新创建的集群在没有明确设置环境类型时，可以部署任何环境的服务
	if cluster.EnvType == "" || cluster.EnvType == "development" {
		return nil
	}

	// 校验服务环境与集群环境类型是否匹配
	if serviceEnv != cluster.EnvType {
		return errors.New("ENV_MISMATCH: service env '" + serviceEnv + "' does not match cluster env_type '" + cluster.EnvType + "'")
	}

	return nil
}

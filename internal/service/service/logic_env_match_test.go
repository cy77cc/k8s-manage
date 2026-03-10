package service

import (
	"context"
	"testing"

	"github.com/cy77cc/OpsPilot/internal/model"
	"github.com/cy77cc/OpsPilot/internal/testutil"
)

func TestValidateEnvMatch_Match(t *testing.T) {
	suite := testutil.NewIntegrationSuite(t)
	t.Cleanup(suite.Cleanup)

	// 创建集群，设置 env_type
	cluster := suite.SeedCluster(func(c *model.Cluster) {
		c.EnvType = "production"
	})

	// 创建 Logic
	logic := NewLogic(suite.SvcCtx)

	// 校验匹配的环境
	err := logic.ValidateEnvMatch(context.Background(), "production", cluster.ID)
	if err != nil {
		t.Fatalf("expected no error for matching env, got: %v", err)
	}
}

func TestValidateEnvMatch_Mismatch(t *testing.T) {
	suite := testutil.NewIntegrationSuite(t)
	t.Cleanup(suite.Cleanup)

	// 创建集群，设置 env_type
	cluster := suite.SeedCluster(func(c *model.Cluster) {
		c.EnvType = "production"
	})

	// 创建 Logic
	logic := NewLogic(suite.SvcCtx)

	// 校验不匹配的环境
	err := logic.ValidateEnvMatch(context.Background(), "staging", cluster.ID)
	if err == nil {
		t.Fatal("expected error for mismatching env, got nil")
	}
	if err.Error() != "ENV_MISMATCH: service env 'staging' does not match cluster env_type 'production'" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestValidateEnvMatch_DefaultDevelopment(t *testing.T) {
	suite := testutil.NewIntegrationSuite(t)
	t.Cleanup(suite.Cleanup)

	// 创建集群，使用默认值 development
	cluster := suite.SeedCluster(func(c *model.Cluster) {
		c.EnvType = "development" // 默认值
	})

	// 创建 Logic
	logic := NewLogic(suite.SvcCtx)

	// development 集群可以部署任何环境的服务（向后兼容）
	err := logic.ValidateEnvMatch(context.Background(), "staging", cluster.ID)
	if err != nil {
		t.Fatalf("expected no error when cluster env_type is default 'development', got: %v", err)
	}

	err = logic.ValidateEnvMatch(context.Background(), "production", cluster.ID)
	if err != nil {
		t.Fatalf("expected no error when cluster env_type is default 'development', got: %v", err)
	}
}

func TestValidateEnvMatch_ClusterNotFound(t *testing.T) {
	suite := testutil.NewIntegrationSuite(t)
	t.Cleanup(suite.Cleanup)

	// 创建 Logic
	logic := NewLogic(suite.SvcCtx)

	// 使用不存在的集群 ID
	err := logic.ValidateEnvMatch(context.Background(), "staging", 99999)
	if err == nil {
		t.Fatal("expected error for non-existent cluster, got nil")
	}
	if err.Error() != "cluster not found" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestValidateEnvMatch_StagingCluster(t *testing.T) {
	suite := testutil.NewIntegrationSuite(t)
	t.Cleanup(suite.Cleanup)

	// 创建 staging 环境集群
	cluster := suite.SeedCluster(func(c *model.Cluster) {
		c.EnvType = "staging"
	})

	// 创建 Logic
	logic := NewLogic(suite.SvcCtx)

	// staging 环境服务可以部署
	err := logic.ValidateEnvMatch(context.Background(), "staging", cluster.ID)
	if err != nil {
		t.Fatalf("expected no error for staging env, got: %v", err)
	}

	// production 环境服务不能部署到 staging 集群
	err = logic.ValidateEnvMatch(context.Background(), "production", cluster.ID)
	if err == nil {
		t.Fatal("expected error when deploying production service to staging cluster")
	}
}

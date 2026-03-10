package cluster

import (
	"context"
	"testing"
	"time"

	"github.com/cy77cc/OpsPilot/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupClusterRepoDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:cluster_repo_test?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.Cluster{}, &model.ClusterNode{}, &model.ClusterCredential{}, &model.ClusterBootstrapProfile{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestClusterRepositoryReadModels(t *testing.T) {
	db := setupClusterRepoDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	now := time.Now().UTC()

	c1 := model.Cluster{Name: "c1", Status: "active", Type: "kubernetes", Source: "platform_managed", Version: "v1.30.0", K8sVersion: "v1.30.0", CreatedAt: now, UpdatedAt: now}
	c2 := model.Cluster{Name: "c2", Status: "inactive", Type: "kubernetes", Source: "external_managed", Version: "v1.29.0", K8sVersion: "v1.29.0", CreatedAt: now, UpdatedAt: now}
	if err := db.Create(&c1).Error; err != nil {
		t.Fatalf("seed c1: %v", err)
	}
	if err := db.Create(&c2).Error; err != nil {
		t.Fatalf("seed c2: %v", err)
	}
	if err := db.Create(&model.ClusterNode{ClusterID: c1.ID, Name: "n1", IP: "10.0.0.1", Role: "worker", Status: "ready"}).Error; err != nil {
		t.Fatalf("seed node: %v", err)
	}

	rows, err := repo.ListClusters(ctx, "active", "")
	if err != nil {
		t.Fatalf("list clusters: %v", err)
	}
	if len(rows) != 1 || rows[0].ID != c1.ID || rows[0].NodeCount != 1 {
		t.Fatalf("unexpected list rows: %+v", rows)
	}

	detail, err := repo.GetClusterDetail(ctx, c1.ID)
	if err != nil {
		t.Fatalf("get detail: %v", err)
	}
	if detail.ID != c1.ID || detail.NodeCount != 1 {
		t.Fatalf("unexpected detail: %+v", detail)
	}

	nodes, err := repo.ListClusterNodes(ctx, c1.ID)
	if err != nil {
		t.Fatalf("list nodes: %v", err)
	}
	if len(nodes) != 1 || nodes[0].Name != "n1" {
		t.Fatalf("unexpected nodes: %+v", nodes)
	}
}

func TestClusterRepositoryDeleteClusterWithRelations(t *testing.T) {
	db := setupClusterRepoDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	cluster := model.Cluster{Name: "c-del", Status: "active", Type: "kubernetes", Source: "platform_managed"}
	if err := db.Create(&cluster).Error; err != nil {
		t.Fatalf("seed cluster: %v", err)
	}
	if err := db.Create(&model.ClusterNode{ClusterID: cluster.ID, Name: "n-del", IP: "10.0.0.9", Role: "worker", Status: "ready"}).Error; err != nil {
		t.Fatalf("seed node: %v", err)
	}
	if err := db.Create(&model.ClusterCredential{Name: "cred", ClusterID: cluster.ID, Source: "external_managed", RuntimeType: "k8s"}).Error; err != nil {
		t.Fatalf("seed cred: %v", err)
	}

	if err := repo.DeleteClusterWithRelations(ctx, cluster.ID); err != nil {
		t.Fatalf("delete cluster with relations: %v", err)
	}

	var cc int64
	db.Model(&model.Cluster{}).Where("id = ?", cluster.ID).Count(&cc)
	var nc int64
	db.Model(&model.ClusterNode{}).Where("cluster_id = ?", cluster.ID).Count(&nc)
	var rc int64
	db.Model(&model.ClusterCredential{}).Where("cluster_id = ?", cluster.ID).Count(&rc)
	if cc != 0 || nc != 0 || rc != 0 {
		t.Fatalf("expected cascade delete, cluster=%d nodes=%d creds=%d", cc, nc, rc)
	}
}

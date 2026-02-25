package cmdb

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestLogic(t *testing.T) *Logic {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:cmdbtest?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	err = db.AutoMigrate(
		&model.CMDBCI{}, &model.CMDBRelation{}, &model.CMDBSyncJob{}, &model.CMDBSyncRecord{}, &model.CMDBAudit{},
		&model.Node{}, &model.Cluster{}, &model.Service{}, &model.DeploymentTarget{}, &model.ServiceDeployTarget{},
	)
	if err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return NewLogic(&svc.ServiceContext{DB: db})
}

func TestCICRUDAndRelation(t *testing.T) {
	logic := newTestLogic(t)
	ctx := context.Background()

	ci, err := logic.createCI(ctx, 1, model.CMDBCI{CIType: "service", Name: "svc-a", ExternalID: "1", Source: "manual"})
	if err != nil {
		t.Fatalf("create ci: %v", err)
	}
	if ci.ID == 0 {
		t.Fatalf("expected ci id")
	}

	_, total, err := logic.listCIs(ctx, ciFilter{Page: 1, PageSize: 10})
	if err != nil || total != 1 {
		t.Fatalf("list ci failed total=%d err=%v", total, err)
	}

	updated, err := logic.updateCI(ctx, 2, ci.ID, map[string]any{"status": "inactive"})
	if err != nil {
		t.Fatalf("update ci: %v", err)
	}
	if updated.Status != "inactive" {
		t.Fatalf("expected status inactive")
	}

	ci2, err := logic.createCI(ctx, 1, model.CMDBCI{CIType: "cluster", Name: "cluster-a", ExternalID: "1", Source: "manual"})
	if err != nil {
		t.Fatalf("create ci2: %v", err)
	}
	rel, err := logic.createRelation(ctx, 1, model.CMDBRelation{FromCIID: ci.ID, ToCIID: ci2.ID, RelationType: "runs_on"})
	if err != nil {
		t.Fatalf("create relation: %v", err)
	}
	if rel.ID == 0 {
		t.Fatalf("expected relation id")
	}

	graph, err := logic.topology(ctx, 0, 0)
	if err != nil {
		t.Fatalf("topology: %v", err)
	}
	nodes, _ := graph["nodes"].([]map[string]any)
	edges, _ := graph["edges"].([]map[string]any)
	if len(nodes) == 0 || len(edges) == 0 {
		t.Fatalf("expected non-empty graph")
	}
}

func TestSyncAndAudit(t *testing.T) {
	logic := newTestLogic(t)
	ctx := context.Background()
	now := time.Now()

	err := logic.svcCtx.DB.Create(&model.Node{Name: "host-1", IP: "10.0.0.1", Status: "online", CreatedAt: now, UpdatedAt: now}).Error
	if err != nil {
		t.Fatalf("seed host: %v", err)
	}
	err = logic.svcCtx.DB.Create(&model.Cluster{Name: "cluster-1", Status: "ready", Type: "kubernetes", Endpoint: "https://x", AuthMethod: "token", CreatedAt: now, UpdatedAt: now}).Error
	if err != nil {
		t.Fatalf("seed cluster: %v", err)
	}
	err = logic.svcCtx.DB.Create(&model.Service{Name: "svc-1", Type: "stateless", Image: "nginx:latest", ProjectID: 1, RuntimeType: "k8s", Status: "running", CreatedAt: now, UpdatedAt: now}).Error
	if err != nil {
		t.Fatalf("seed service: %v", err)
	}

	job, err := logic.runSync(ctx, 1, "all")
	if err != nil {
		t.Fatalf("run sync: %v", err)
	}
	if job.Status != "succeeded" {
		t.Fatalf("unexpected sync status: %s", job.Status)
	}
	var summary map[string]int
	_ = json.Unmarshal([]byte(job.SummaryJSON), &summary)
	if summary["created"] == 0 {
		t.Fatalf("expected created > 0")
	}

	logic.writeAudit(ctx, model.CMDBAudit{CIID: 1, Action: "ci.update", ActorID: 1, Detail: "test"})
	audits, err := logic.listAudits(ctx, 1)
	if err != nil {
		t.Fatalf("list audits: %v", err)
	}
	if len(audits) == 0 {
		t.Fatalf("expected audits")
	}
}

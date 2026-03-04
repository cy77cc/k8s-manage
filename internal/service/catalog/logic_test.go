package catalog

import (
	"context"
	"testing"

	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestLogic(t *testing.T) *Logic {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:catalog_logic_test?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.ServiceCategory{}, &model.ServiceTemplate{}, &model.Service{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return NewLogic(&svc.ServiceContext{DB: db})
}

func TestRenderTemplateWithSchema(t *testing.T) {
	logic := &Logic{}
	content := "user={{ username }}\nport={{ port|default:3306 }}"
	schema := `[{"name":"username","type":"string","required":true},{"name":"port","type":"number","required":false}]`

	resolved, unresolved, err := logic.renderTemplateWithSchema(content, schema, map[string]any{"username": "admin"})
	if err != nil {
		t.Fatalf("renderTemplateWithSchema returned error: %v", err)
	}
	if resolved != "user=admin\nport=3306" {
		t.Fatalf("unexpected rendered yaml: %s", resolved)
	}
	if len(unresolved) != 0 {
		t.Fatalf("expected no unresolved vars, got: %v", unresolved)
	}

	if _, _, err := logic.renderTemplateWithSchema(content, schema, map[string]any{"username": "admin", "port": "NaN-Invalid"}); err == nil {
		t.Fatalf("expected number validation error")
	}
}

func TestReviewTransitions(t *testing.T) {
	logic := newTestLogic(t)
	ctx := context.Background()

	cat := model.ServiceCategory{Name: "database", DisplayName: "数据库"}
	if err := logic.svcCtx.DB.Create(&cat).Error; err != nil {
		t.Fatalf("create category: %v", err)
	}

	tpl, err := logic.CreateTemplate(ctx, 100, TemplateCreateRequest{
		Name:            "mysql-test",
		DisplayName:     "MySQL Test",
		CategoryID:      cat.ID,
		K8sTemplate:     "kind: Deployment\nmetadata:\n  name: {{ service_name }}",
		VariablesSchema: []CatalogVariableSchema{{Name: "service_name", Type: "string", Required: true}},
	})
	if err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}

	submitted, err := logic.SubmitForReview(ctx, tpl.ID, 100, false)
	if err != nil {
		t.Fatalf("SubmitForReview: %v", err)
	}
	if submitted.Status != model.TemplateStatusPendingReview {
		t.Fatalf("expected pending_review, got %s", submitted.Status)
	}

	published, err := logic.PublishTemplate(ctx, tpl.ID)
	if err != nil {
		t.Fatalf("PublishTemplate: %v", err)
	}
	if published.Status != model.TemplateStatusPublished {
		t.Fatalf("expected published, got %s", published.Status)
	}

	if _, err := logic.RejectTemplate(ctx, tpl.ID, "invalid"); err == nil {
		t.Fatalf("reject should fail from published state")
	}
}

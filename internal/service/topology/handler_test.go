package topology

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTopologyGraphQueryAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file:topologyapi?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.User{},
		&model.CMDBCI{},
		&model.CMDBRelation{},
		&model.TopologyAccessAudit{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	if err := db.Exec("INSERT INTO users (id, username, password_hash, email, phone, status) VALUES (1, 'admin', 'x', 'admin@example.com', '', 1)").Error; err != nil {
		t.Fatalf("seed admin: %v", err)
	}
	ci1 := model.CMDBCI{CIUID: "svc-1", CIType: "service", Name: "svc-demo", ProjectID: 11, Status: "active"}
	ci2 := model.CMDBCI{CIUID: "cluster-1", CIType: "cluster", Name: "cluster-demo", ProjectID: 11, Status: "active"}
	if err := db.Create(&ci1).Error; err != nil {
		t.Fatalf("seed ci1: %v", err)
	}
	if err := db.Create(&ci2).Error; err != nil {
		t.Fatalf("seed ci2: %v", err)
	}
	if err := db.Create(&model.CMDBRelation{FromCIID: ci1.ID, ToCIID: ci2.ID, RelationType: "deploys_to"}).Error; err != nil {
		t.Fatalf("seed relation: %v", err)
	}

	h := NewHandler(&svc.ServiceContext{DB: db})
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("uid", uint(1))
		c.Next()
	})
	g := r.Group("/api/v1/topology")
	g.GET("/graph", h.Graph)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/topology/graph?project_id=11&resource_type=service", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("graph status=%d body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		Data struct {
			Nodes []any `json:"nodes"`
			Edges []any `json:"edges"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode graph: %v", err)
	}
	if len(resp.Data.Nodes) != 1 {
		t.Fatalf("expected 1 service node, got %d", len(resp.Data.Nodes))
	}

	var auditCount int64
	if err := db.Model(&model.TopologyAccessAudit{}).Count(&auditCount).Error; err != nil {
		t.Fatalf("count topology audit: %v", err)
	}
	if auditCount < 1 {
		t.Fatalf("expected topology access audit")
	}
}

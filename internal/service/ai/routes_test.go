package ai

import (
	"testing"

	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
)

func TestRegisterAIHandlers_RegistersCanonicalEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	v1 := r.Group("/api/v1")
	RegisterAIHandlers(v1, &svc.ServiceContext{})

	routes := map[string]bool{}
	for _, item := range r.Routes() {
		routes[item.Method+" "+item.Path] = true
	}

	if !routes["POST /api/v1/ai/chat/respond"] {
		t.Fatalf("expected chat respond route")
	}
	if !routes["GET /api/v1/ai/tools"] {
		t.Fatalf("expected canonical tools route")
	}
	if !routes["POST /api/v1/ai/approval/respond"] {
		t.Fatalf("expected approval respond route")
	}
	if !routes["GET /api/v1/ai/capabilities"] {
		t.Fatalf("expected capabilities compatibility route")
	}
	if !routes["POST /api/v1/ai/chat"] {
		t.Fatalf("expected chat route")
	}
}

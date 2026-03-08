package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cy77cc/k8s-manage/internal/config"
	"github.com/gin-gonic/gin"
)

func TestStaticRouteRegistrationDevelopmentModeSkipsFrontendRoutes(t *testing.T) {
	prevEnv := config.CFG.App.Env
	config.CFG.App.Env = "development"
	t.Cleanup(func() {
		config.CFG.App.Env = prevEnv
	})

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	registerWebStaticRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for / in development mode, got %d", rec.Code)
	}

	apiReq := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	apiRec := httptest.NewRecorder()
	r.ServeHTTP(apiRec, apiReq)

	if apiRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for /api/health in development mode, got %d", apiRec.Code)
	}
}

func TestStaticRouteRegistrationProductionModeServesFrontendFallback(t *testing.T) {
	prevEnv := config.CFG.App.Env
	config.CFG.App.Env = "production"
	t.Cleanup(func() {
		config.CFG.App.Env = prevEnv
	})

	gin.SetMode(gin.TestMode)
	r := gin.New()

	registerWebStaticRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for / in production mode, got %d", rec.Code)
	}
}

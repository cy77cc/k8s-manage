package ai

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	aiplatform "github.com/cy77cc/k8s-manage/internal/ai"
	aitools "github.com/cy77cc/k8s-manage/internal/ai/tools"
	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newPerformanceTestHandler(tb testing.TB) *handler {
	tb.Helper()
	dsn := fmt.Sprintf("file:%s-%d?mode=memory&cache=shared&_busy_timeout=5000", strings.ReplaceAll(tb.Name(), "/", "_"), time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		tb.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		tb.Fatalf("sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	if err := db.AutoMigrate(
		&model.User{},
		&model.Role{},
		&model.Permission{},
		&model.UserRole{},
		&model.RolePermission{},
		&model.Node{},
		&model.Service{},
		&model.CICDRelease{},
		&model.AlertEvent{},
		&model.CMDBRelation{},
		&model.CMDBCI{},
		&model.CICDAuditEvent{},
		&model.AIChatSession{},
		&model.AIChatMessage{},
		&model.Cluster{},
	); err != nil {
		tb.Fatalf("auto migrate performance tables: %v", err)
	}
	h := newHandler(&svc.ServiceContext{DB: db})
	runner, err := aiplatform.NewPlatformRunner(
		context.Background(),
		&fakeE2EToolCallingModel{},
		aitools.PlatformDeps{DB: h.svcCtx.DB},
		&aiplatform.RunnerConfig{EnableStreaming: true},
	)
	if err != nil {
		tb.Fatalf("new platform runner: %v", err)
	}
	h.svcCtx.AI = runner
	if testT, ok := tb.(*testing.T); ok {
		seedPermissionTestData(testT, h)
	} else {
		seedPerformancePermissionData(tb, h)
	}
	return h
}

func seedPerformancePermissionData(tb testing.TB, h *handler) {
	tb.Helper()
	db := h.svcCtx.DB
	_ = db.Exec("DELETE FROM role_permissions").Error
	_ = db.Exec("DELETE FROM user_roles").Error
	_ = db.Exec("DELETE FROM permissions").Error
	_ = db.Exec("DELETE FROM roles").Error
	_ = db.Exec("DELETE FROM users").Error
	_ = db.Exec("DELETE FROM services").Error
	_ = db.Exec("DELETE FROM nodes").Error

	if err := db.Create(&model.User{ID: 1, Username: "admin001", PasswordHash: "x", Email: "admin001@example.com"}).Error; err != nil {
		tb.Fatalf("create user: %v", err)
	}
	if err := db.Create(&model.Role{ID: 1, Code: "admin", Name: "admin"}).Error; err != nil {
		tb.Fatalf("create role: %v", err)
	}
	if err := db.Create(&model.UserRole{UserID: 1, RoleID: 1}).Error; err != nil {
		tb.Fatalf("create user role: %v", err)
	}
	if err := db.Create(&model.Service{
		ID:          10,
		Name:        "svc",
		Type:        "stateless",
		Image:       "nginx:latest",
		ProjectID:   1,
		OwnerUserID: 1,
		Status:      "running",
	}).Error; err != nil {
		tb.Fatalf("create service: %v", err)
	}
}

func runSimpleChatRequest(tb testing.TB, h *handler, sessionID string) time.Duration {
	tb.Helper()
	c, w := newJSONContext(http.MethodPost, "/api/v1/ai/chat", map[string]any{
		"sessionId": sessionID,
		"message":   "帮我总结当前会话状态",
		"context":   map[string]any{"scene": "hosts"},
	}, 1)
	start := time.Now()
	h.chat(c)
	elapsed := time.Since(start)
	body := w.Body.String()
	if !strings.Contains(body, "event: done") {
		tb.Fatalf("expected done event, got %s", body)
	}
	return elapsed
}

func TestSimpleChatResponseCompletesWithinOneSecond(t *testing.T) {
	h := newE2ETestHandler(t)
	elapsed := runSimpleChatRequest(t, h, "sess-perf-threshold")
	if elapsed >= time.Second {
		t.Fatalf("expected simple chat response under 1s, got %s", elapsed)
	}
}

func BenchmarkChatSimpleResponse(b *testing.B) {
	h := newPerformanceTestHandler(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runSimpleChatRequest(b, h, fmt.Sprintf("sess-bench-%d", i))
	}
}

func BenchmarkSessionStoreConversationMemory(b *testing.B) {
	h := newPerformanceTestHandler(b)
	for i := 0; i < 24; i++ {
		if _, err := h.sessions.AppendMessage(1, "hosts", "sess-memory", map[string]any{
			"id":        fmt.Sprintf("msg-%d", i),
			"role":      "assistant",
			"content":   strings.Repeat("performance payload ", 8),
			"timestamp": time.Now(),
		}); err != nil {
			b.Fatalf("seed session message: %v", err)
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		session, ok := h.sessions.GetSession(1, "sess-memory")
		if !ok || session == nil {
			b.Fatalf("expected session to load")
		}
		if len(session.Messages) == 0 {
			b.Fatalf("expected seeded messages")
		}
	}
}

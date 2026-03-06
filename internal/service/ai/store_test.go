package ai

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/cy77cc/k8s-manage/internal/ai/tools"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newRuntimeStoreForTest(t *testing.T) (*runtimeStore, *gorm.DB) {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	return &runtimeStore{
		db:                db,
		approvals:         map[string]*approvalTicket{},
		executions:        map[string]*executionRecord{},
		recommendations:   map[string][]recommendationRecord{},
		toolParams:        map[string]map[string]any{},
		referencedContext: map[string]map[string]any{},
	}, db
}

func TestRuntimeStoreRecommendationsAndContext(t *testing.T) {
	store, _ := newRuntimeStoreForTest(t)
	now := time.Now()
	store.setRecommendations(7, " hosts ", []recommendationRecord{
		{ID: "old", UserID: 7, Scene: "hosts", Title: "old", CreatedAt: now.Add(-time.Minute)},
		{ID: "new", UserID: 7, Scene: "hosts", Title: "new", CreatedAt: now},
	})

	got := store.getRecommendations(7, "hosts", 1)
	if len(got) != 1 || got[0].ID != "new" {
		t.Fatalf("expected newest recommendation first, got %#v", got)
	}

	store.rememberContext(7, "", map[string]any{
		"target": "pod-a",
		"empty":  "",
		"nil":    nil,
	})
	remembered := store.getRememberedContext(7, "global")
	if remembered["target"] != "pod-a" {
		t.Fatalf("expected remembered target, got %#v", remembered)
	}
	if _, ok := remembered["empty"]; ok {
		t.Fatalf("expected empty values to be skipped: %#v", remembered)
	}
}

func TestRuntimeStoreToolMemory(t *testing.T) {
	store, _ := newRuntimeStoreForTest(t)

	acc := &toolMemoryAccessor{store: store, uid: 3, scene: "ops"}
	acc.SetLastToolParams("host_exec", map[string]any{"target": "node-1"})
	params := acc.GetLastToolParams("host_exec")
	if params["target"] != "node-1" {
		t.Fatalf("unexpected tool params: %#v", params)
	}
	params["target"] = "node-2"
	if acc.GetLastToolParams("host_exec")["target"] != "node-1" {
		t.Fatalf("expected tool params copy isolation")
	}
}

func TestRuntimeStoreApprovalsAndExecutions(t *testing.T) {
	store, _ := newRuntimeStoreForTest(t)

	approval := store.newApproval(9, approvalTicket{
		Tool:   "service_deploy",
		Risk:   tools.ToolRiskHigh,
		Mode:   tools.ToolModeMutating,
		Params: map[string]any{"service_id": 42},
	})
	if approval.ID == "" || approval.Status != "pending" || approval.RequestUID != 9 {
		t.Fatalf("unexpected approval: %#v", approval)
	}
	loadedApproval, ok := store.getApproval(approval.ID)
	if !ok || loadedApproval == nil {
		t.Fatalf("expected approval lookup")
	}
	updatedApproval, ok := store.setApprovalStatus(approval.ID, "approved", 12)
	if !ok || updatedApproval.Status != "approved" || updatedApproval.ReviewUID != 12 {
		t.Fatalf("unexpected approval update: %#v", updatedApproval)
	}

	finishedAt := time.Now()
	store.saveExecution(&executionRecord{
		ID:         "exec-1",
		Tool:       "host_exec",
		Status:     "succeeded",
		RequestUID: 9,
		FinishedAt: &finishedAt,
	})
	execution, ok := store.getExecution("exec-1")
	if !ok || execution == nil || execution.Tool != "host_exec" {
		t.Fatalf("unexpected execution: %#v", execution)
	}
}

func TestSessionStoreDBUnavailablePath(t *testing.T) {
	mini := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	sessionStore := NewSessionStore(nil, client)
	if sessionStore.dbEnabled() {
		t.Fatalf("expected db disabled without db")
	}
}

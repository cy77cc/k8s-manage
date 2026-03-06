package ai

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newSessionStoreForTest(t *testing.T) (*SessionStore, *gorm.DB) {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.AIChatSession{}, &model.AIChatMessage{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	miniRedis := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: miniRedis.Addr()})
	return NewSessionStore(db, client), db
}

func TestSessionStoreGetSessionUsesRedisCache(t *testing.T) {
	store, db := newSessionStoreForTest(t)

	if _, err := store.AppendMessage(1, "global", "sess-1", map[string]any{
		"id":        "u-1",
		"role":      "user",
		"content":   "first",
		"timestamp": time.Now(),
	}); err != nil {
		t.Fatalf("append message: %v", err)
	}

	sess, ok := store.GetSession(1, "sess-1")
	if !ok || sess == nil {
		t.Fatalf("expected session")
	}

	if err := db.Model(&model.AIChatSession{}).Where("id = ?", "sess-1").Update("title", "mutated in db").Error; err != nil {
		t.Fatalf("mutate db title: %v", err)
	}

	cached, ok := store.GetSession(1, "sess-1")
	if !ok || cached == nil {
		t.Fatalf("expected cached session")
	}
	if cached.Title != defaultAISessionTitle {
		t.Fatalf("expected cached title %q, got %q", defaultAISessionTitle, cached.Title)
	}
}

func TestSessionStoreAppendMessageInvalidatesSessionCache(t *testing.T) {
	store, _ := newSessionStoreForTest(t)

	if _, err := store.AppendMessage(1, "global", "sess-2", map[string]any{
		"id":        "u-1",
		"role":      "user",
		"content":   "first",
		"timestamp": time.Now(),
	}); err != nil {
		t.Fatalf("append first message: %v", err)
	}
	if _, ok := store.GetSession(1, "sess-2"); !ok {
		t.Fatalf("expected session after first append")
	}

	if _, err := store.AppendMessage(1, "global", "sess-2", map[string]any{
		"id":        "a-1",
		"role":      "assistant",
		"content":   "second",
		"timestamp": time.Now(),
	}); err != nil {
		t.Fatalf("append second message: %v", err)
	}

	updated, ok := store.GetSession(1, "sess-2")
	if !ok || updated == nil {
		t.Fatalf("expected updated session")
	}
	if len(updated.Messages) != 2 {
		t.Fatalf("expected 2 messages after cache invalidation, got %d", len(updated.Messages))
	}
}

func TestSessionStoreLoadsPersistedSessionFromDBAcrossInstances(t *testing.T) {
	store, db := newSessionStoreForTest(t)

	if _, err := store.AppendMessage(1, "global", "sess-persist", map[string]any{
		"id":        "u-1",
		"role":      "user",
		"content":   "persist me",
		"timestamp": time.Now(),
	}); err != nil {
		t.Fatalf("append message: %v", err)
	}

	reloaded := NewSessionStore(db, nil)
	session, ok := reloaded.GetSession(1, "sess-persist")
	if !ok || session == nil {
		t.Fatalf("expected reloaded session from db")
	}
	if len(session.Messages) != 1 {
		t.Fatalf("expected 1 persisted message, got %d", len(session.Messages))
	}
	if session.Messages[0]["content"] != "persist me" {
		t.Fatalf("unexpected persisted content: %#v", session.Messages[0]["content"])
	}
}

func TestSessionStoreListUpdateAndDeleteSession(t *testing.T) {
	store, db := newSessionStoreForTest(t)

	appendMessage := func(id, scene, msgID, content string) {
		t.Helper()
		if _, err := store.AppendMessage(1, scene, id, map[string]any{
			"id":        msgID,
			"role":      "user",
			"content":   content,
			"timestamp": time.Now(),
		}); err != nil {
			t.Fatalf("append message: %v", err)
		}
	}

	appendMessage("sess-a", "ops", "msg-a", "first")
	appendMessage("sess-b", "ops", "msg-b", "second")

	list := store.ListSessions(1, "ops")
	if len(list) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(list))
	}

	if err := db.Where("id = ?", "sess-a").Delete(&model.AIChatSession{}).Error; err != nil {
		t.Fatalf("delete session directly in db: %v", err)
	}
	cachedList := store.ListSessions(1, "ops")
	if len(cachedList) != 2 {
		t.Fatalf("expected cached session list, got %d", len(cachedList))
	}

	updated, err := store.UpdateSessionTitle(1, "sess-b", " renamed title ")
	if err != nil {
		t.Fatalf("update title: %v", err)
	}
	if updated.Title != "renamed title" {
		t.Fatalf("unexpected updated title: %q", updated.Title)
	}

	current, ok := store.CurrentSession(1, "ops")
	if !ok || current == nil || current.ID != "sess-b" {
		t.Fatalf("expected latest session as current, got %#v", current)
	}

	store.DeleteSession(1, "sess-b")
	if _, ok := store.GetSession(1, "sess-b"); ok {
		t.Fatalf("expected deleted session to be unavailable")
	}
}

func TestSessionStoreCurrentSessionIDAndCacheHelpers(t *testing.T) {
	store, _ := newSessionStoreForTest(t)

	if got := store.getOrCreateCurrentSessionID(1, "global"); got != "" {
		t.Fatalf("expected no current session id, got %q", got)
	}

	if _, err := store.AppendMessage(1, "global", "sess-current", map[string]any{
		"id":        "msg-current",
		"role":      "user",
		"content":   "hello",
		"timestamp": time.Now(),
	}); err != nil {
		t.Fatalf("append message: %v", err)
	}

	if got := store.getOrCreateCurrentSessionID(1, "global"); got != "sess-current" {
		t.Fatalf("unexpected current session id: %q", got)
	}

	key := store.sessionListCacheKey(1, "global")
	sessions := []*aiSession{{ID: "cached", Scene: "global", Title: "cached"}}
	store.cacheSessionList(key, sessions)
	cached, ok := store.loadCachedSessionList(key)
	if !ok || len(cached) != 1 || cached[0].ID != "cached" {
		t.Fatalf("unexpected cached session list: %#v", cached)
	}

	store.invalidateListCaches(1, "global")
	if _, ok := store.loadCachedSessionList(key); ok {
		t.Fatalf("expected invalidated cached session list")
	}
}

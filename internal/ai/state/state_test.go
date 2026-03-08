package state

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/cloudwego/eino/schema"
	"github.com/redis/go-redis/v9"
)

func TestSessionStateSaveLoadAndAppend(t *testing.T) {
	t.Parallel()

	client := newRedisClient(t)
	store := NewSessionState(client, "test:session:")

	if err := store.Save(context.Background(), SessionSnapshot{
		SessionID: "s-1",
		Context:   map[string]any{"tenant": "demo"},
	}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if err := store.AppendMessage(context.Background(), "s-1", schema.UserMessage("hello")); err != nil {
		t.Fatalf("AppendMessage() error = %v", err)
	}

	snapshot, err := store.Load(context.Background(), "s-1")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if snapshot == nil || len(snapshot.Messages) != 1 {
		t.Fatalf("Load() snapshot = %+v, want 1 message", snapshot)
	}
}

func TestCheckpointStoreGetSet(t *testing.T) {
	t.Parallel()

	client := newRedisClient(t)
	store := NewCheckpointStore(client, "test:checkpoint:")

	if err := store.Set(context.Background(), "cp-1", []byte("payload")); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	got, ok, err := store.Get(context.Background(), "cp-1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !ok || string(got) != "payload" {
		t.Fatalf("Get() = %q, %v; want payload, true", string(got), ok)
	}
}

func newRedisClient(t *testing.T) redis.UniversalClient {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run() error = %v", err)
	}
	t.Cleanup(mr.Close)

	return redis.NewClient(&redis.Options{Addr: mr.Addr()})
}

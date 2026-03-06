package ai

import (
	"context"
	"testing"
	"time"

	aitools "github.com/cy77cc/k8s-manage/internal/ai/tools"
)

func TestAgentFirstResponseCompletesWithinTwoSeconds(t *testing.T) {
	agent, err := NewHybridAgent(
		context.Background(),
		&fakeToolCallingModel{},
		&fakeClassifierModel{reply: "agentic"},
		aitools.PlatformDeps{},
		nil,
	)
	if err != nil {
		t.Fatalf("new hybrid agent failed: %v", err)
	}

	start := time.Now()
	iter := agent.Query(context.Background(), "sess-agent-perf", "查看 pod 日志")

	first, ok := iter.Next()
	if !ok {
		t.Fatalf("expected first agent result")
	}
	if first == nil {
		t.Fatalf("expected non-nil first agent result")
	}

	elapsed := time.Since(start)
	if elapsed >= 2*time.Second {
		t.Fatalf("expected agent first response under 2s, got %s", elapsed)
	}
}

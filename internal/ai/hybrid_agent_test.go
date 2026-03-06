package ai

import (
	"context"
	"testing"

	aitools "github.com/cy77cc/k8s-manage/internal/ai/tools"
)

func TestNewHybridAgent(t *testing.T) {
	agent, err := NewHybridAgent(context.Background(), &fakeToolCallingModel{}, &fakeClassifierModel{reply: "simple"}, aitools.PlatformDeps{}, nil)
	if err != nil {
		t.Fatalf("new hybrid agent failed: %v", err)
	}
	if agent == nil {
		t.Fatalf("expected non-nil hybrid agent")
	}
}

func TestHybridAgentQueryRoutesToSimpleChat(t *testing.T) {
	agent, err := NewHybridAgent(context.Background(), &fakeToolCallingModel{}, &fakeClassifierModel{reply: "simple"}, aitools.PlatformDeps{}, nil)
	if err != nil {
		t.Fatalf("new hybrid agent failed: %v", err)
	}

	results := collectAgentResults(agent.Query(context.Background(), "sess-1", "什么是 Pod"))
	if len(results) == 0 {
		t.Fatalf("expected results")
	}
	last := results[len(results)-1]
	if last.Type != "text" {
		t.Fatalf("expected text result, got %s", last.Type)
	}
}

func TestHybridAgentQueryRoutesToAgenticMode(t *testing.T) {
	agent := &HybridAgent{
		classifier:  NewIntentClassifier(&fakeClassifierModel{reply: "agentic"}),
		simpleChat:  NewSimpleChatMode(&fakeToolCallingModel{}),
		agenticMode: &AgenticMode{},
	}

	results := collectAgentResults(agent.Query(context.Background(), "sess-1", "查看 pod 日志"))
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Type != "error" {
		t.Fatalf("expected agentic path error result, got %s", results[0].Type)
	}
}

func TestNewHybridAgentFallsBackToChatModelForClassifier(t *testing.T) {
	agent, err := NewHybridAgent(context.Background(), &fakeToolCallingModel{}, nil, aitools.PlatformDeps{}, nil)
	if err != nil {
		t.Fatalf("new hybrid agent failed: %v", err)
	}
	if agent == nil || agent.classifier == nil {
		t.Fatalf("expected classifier to be initialized")
	}
}

func TestHybridAgentResumeWithoutAgenticModeReturnsError(t *testing.T) {
	agent := &HybridAgent{}
	if _, err := agent.Resume(context.Background(), "sess-1", "ask-1", true); err == nil {
		t.Fatalf("expected error")
	}
}

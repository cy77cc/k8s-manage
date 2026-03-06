package ai

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/adk"
)

func collectAgentResults(iter *adk.AsyncIterator[*AgentResult]) []*AgentResult {
	out := make([]*AgentResult, 0)
	for {
		item, ok := iter.Next()
		if !ok {
			break
		}
		if item != nil {
			out = append(out, item)
		}
	}
	return out
}

func TestSimpleChatModeExecuteReturnsTextResult(t *testing.T) {
	mode := NewSimpleChatMode(&fakeToolCallingModel{})
	iter, gen := adk.NewAsyncIteratorPair[*AgentResult]()

	go func() {
		defer gen.Close()
		mode.Execute(context.Background(), "什么是 Pod", gen)
	}()

	results := collectAgentResults(iter)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Type != "text" {
		t.Fatalf("expected text result, got %s", results[0].Type)
	}
	if results[0].Content == "" {
		t.Fatalf("expected non-empty content")
	}
}

func TestSimpleChatModeExecuteNilModelReturnsError(t *testing.T) {
	mode := NewSimpleChatMode(nil)
	iter, gen := adk.NewAsyncIteratorPair[*AgentResult]()

	go func() {
		defer gen.Close()
		mode.Execute(context.Background(), "你好", gen)
	}()

	results := collectAgentResults(iter)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Type != "error" {
		t.Fatalf("expected error result, got %s", results[0].Type)
	}
}

func TestSimpleChatModeExecuteModelErrorReturnsError(t *testing.T) {
	mode := NewSimpleChatMode(&fakeToolCallingModel{})
	iter, gen := adk.NewAsyncIteratorPair[*AgentResult]()

	go func() {
		defer gen.Close()
		mode.Execute(context.Background(), "trigger error", gen)
	}()

	results := collectAgentResults(iter)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Type != "error" {
		t.Fatalf("expected error result, got %s", results[0].Type)
	}
}

func TestSimpleChatModeExecuteEmptyContentFallsBackToDefault(t *testing.T) {
	mode := NewSimpleChatMode(&fakeClassifierModel{reply: "   "})
	iter, gen := adk.NewAsyncIteratorPair[*AgentResult]()

	go func() {
		defer gen.Close()
		mode.Execute(context.Background(), "你好", gen)
	}()

	results := collectAgentResults(iter)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Type != "text" {
		t.Fatalf("expected text result, got %s", results[0].Type)
	}
	if results[0].Content != "无输出。" {
		t.Fatalf("unexpected fallback content: %q", results[0].Content)
	}
}

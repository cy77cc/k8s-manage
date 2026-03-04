package experts

import (
	"testing"

	"github.com/cloudwego/eino/schema"
)

func TestBuildExpertMessagesIncludesHistory(t *testing.T) {
	exec := NewExpertExecutor(&fakeRegistry{})
	history := []*schema.Message{
		schema.SystemMessage("sys"),
		schema.UserMessage("u1"),
		schema.AssistantMessage("a1", nil),
	}
	step := &ExecutionStep{ExpertName: "service_expert", Task: "诊断"}
	messages := exec.buildExpertMessages(history, step, "服务异常", nil)
	if len(messages) != len(history)+1 {
		t.Fatalf("expected history + current message, got %d", len(messages))
	}
	if messages[len(messages)-1].Role != schema.User {
		t.Fatalf("expected last message to be user")
	}
	if messages[len(messages)-1].Content == "" {
		t.Fatalf("expected composed task message")
	}
}

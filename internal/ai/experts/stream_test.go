package experts

import (
	"context"
	"io"
	"strings"
	"testing"
)

func TestExpertExecutorStreamStepFallback(t *testing.T) {
	reg := &fakeRegistry{
		experts: map[string]*Expert{
			"service_expert": {Name: "service_expert"},
		},
	}
	exec := NewExpertExecutor(reg)
	stream, err := exec.StreamStep(context.Background(), &ExecutionStep{
		ExpertName: "service_expert",
		Task:       "analyze service",
	}, &ExecuteRequest{Message: "服务不可用"})
	if err != nil {
		t.Fatalf("stream step: %v", err)
	}
	defer stream.Close()

	msg, recvErr := stream.Recv()
	if recvErr != nil {
		t.Fatalf("recv stream: %v", recvErr)
	}
	if msg == nil || !strings.Contains(msg.Content, "专家模型未初始化") {
		t.Fatalf("unexpected stream content: %#v", msg)
	}
	_, recvErr = stream.Recv()
	if recvErr == nil || recvErr != io.EOF {
		t.Fatalf("expected EOF, got: %v", recvErr)
	}
}

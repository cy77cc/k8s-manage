package graph

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/cy77cc/k8s-manage/internal/ai/experts"
)

type fakeStreamPrimaryRunner struct{}

func (fakeStreamPrimaryRunner) RunPrimaryStream(_ context.Context, _ *experts.ExecuteRequest) (*schema.StreamReader[*schema.Message], error) {
	sr, sw := schema.Pipe[*schema.Message](4)
	sw.Send(schema.AssistantMessage("primary stream output", nil), nil)
	sw.Close()
	return sr, nil
}

type fakeStreamHelperRunner struct{}

func (fakeStreamHelperRunner) RunHelperStream(_ context.Context, _ *experts.ExecuteRequest, helper experts.HelperRequest) (*schema.StreamReader[*schema.Message], error) {
	sr, sw := schema.Pipe[*schema.Message](4)
	sw.Send(schema.AssistantMessage("helper "+helper.ExpertName, nil), nil)
	sw.Close()
	return sr, nil
}

func TestBuilderBuildStreamAndCompile(t *testing.T) {
	b := NewBuilderWithStreamRunners(fakeStreamPrimaryRunner{}, fakeStreamHelperRunner{})
	g, err := b.BuildStream(context.Background())
	if err != nil {
		t.Fatalf("build stream graph: %v", err)
	}
	if g == nil {
		t.Fatalf("stream graph is nil")
	}
	r, err := g.Compile(context.Background())
	if err != nil {
		t.Fatalf("compile stream graph: %v", err)
	}
	stream, err := r.Invoke(context.Background(), &GraphInput{
		Message:  "diag issue",
		Request:  &experts.ExecuteRequest{Message: "diag issue"},
		Strategy: experts.StrategyParallel,
		HelperRequests: []experts.HelperRequest{
			{ExpertName: "k8s", Task: "check pods"},
		},
	})
	if err != nil {
		t.Fatalf("invoke stream graph: %v", err)
	}
	if stream == nil {
		t.Fatalf("stream output is nil")
	}
	defer stream.Close()

	var out strings.Builder
	for {
		msg, recvErr := stream.Recv()
		if recvErr == io.EOF {
			break
		}
		if recvErr != nil {
			t.Fatalf("recv stream: %v", recvErr)
		}
		if msg != nil {
			out.WriteString(msg.Content)
		}
	}
	got := out.String()
	if !strings.Contains(got, "primary stream output") {
		t.Fatalf("expected primary output in stream, got: %q", got)
	}
	if !strings.Contains(got, "helper k8s") {
		t.Fatalf("expected helper output in stream, got: %q", got)
	}
}

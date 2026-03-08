package approval

import (
	"context"
	"testing"

	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/cy77cc/k8s-manage/internal/model"
)

func TestTaskGeneratorFallback(t *testing.T) {
	t.Parallel()

	gen := NewTaskGenerator(nil)
	detail, err := gen.Generate(context.Background(), &model.AIApprovalTask{
		ToolName:           "service_restart",
		TargetResourceName: "payment-api",
		RiskLevel:          "medium",
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if detail.Summary == "" || len(detail.Steps) == 0 {
		t.Fatalf("Generate() = %+v", detail)
	}
}

func TestTaskGeneratorLLM(t *testing.T) {
	t.Parallel()

	gen := NewTaskGenerator(staticModel{content: `{"summary":"restart payment-api","steps":[{"title":"review"}],"risk_assessment":{"level":"medium","summary":"safe","items":["window"]},"rollback_plan":"undo"}`})
	detail, err := gen.Generate(context.Background(), &model.AIApprovalTask{ToolName: "service_restart"})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if detail.Summary != "restart payment-api" {
		t.Fatalf("Generate().Summary = %q", detail.Summary)
	}
}

type staticModel struct{ content string }

var _ einomodel.ToolCallingChatModel = staticModel{}

func (s staticModel) Generate(_ context.Context, _ []*schema.Message, _ ...einomodel.Option) (*schema.Message, error) {
	return schema.AssistantMessage(s.content, nil), nil
}

func (s staticModel) Stream(_ context.Context, _ []*schema.Message, _ ...einomodel.Option) (*schema.StreamReader[*schema.Message], error) {
	return schema.StreamReaderFromArray([]*schema.Message{schema.AssistantMessage(s.content, nil)}), nil
}

func (s staticModel) WithTools(_ []*schema.ToolInfo) (einomodel.ToolCallingChatModel, error) {
	return s, nil
}

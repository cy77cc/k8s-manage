package router

import (
	"context"
	"testing"

	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/cy77cc/k8s-manage/internal/ai/tools"
)

func TestIntentClassifier_ClassifyByKeyword(t *testing.T) {
	t.Parallel()

	classifier := NewIntentClassifier(nil, nil)

	cases := []struct {
		name  string
		input string
		want  tools.ToolDomain
	}{
		{name: "infrastructure", input: "check k8s pod logs in cluster prod", want: tools.DomainInfrastructure},
		{name: "service", input: "deploy service payment-api to staging", want: tools.DomainService},
		{name: "cicd", input: "trigger pipeline for backend release", want: tools.DomainCICD},
		{name: "monitor", input: "show alert metrics for cpu saturation", want: tools.DomainMonitor},
		{name: "config", input: "compare config for app gateway", want: tools.DomainConfig},
		{name: "general fallback", input: "explain what happened yesterday", want: tools.DomainGeneral},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := classifier.Classify(context.Background(), tc.input)
			if err != nil {
				t.Fatalf("Classify() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("Classify() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestIntentClassifier_UsesModelWhenAvailable(t *testing.T) {
	t.Parallel()

	classifier := NewIntentClassifier(staticModel{content: "monitor"}, nil)

	got, err := classifier.Classify(context.Background(), "restart the payment app")
	if err != nil {
		t.Fatalf("Classify() error = %v", err)
	}
	if got != tools.DomainMonitor {
		t.Fatalf("Classify() = %q, want %q", got, tools.DomainMonitor)
	}
}

func TestIntentRouter_Route(t *testing.T) {
	t.Parallel()

	router, err := NewIntentRouter(context.Background(), NewIntentClassifier(nil, nil))
	if err != nil {
		t.Fatalf("NewIntentRouter() error = %v", err)
	}

	got, err := router.Route(context.Background(), "deploy the gateway service")
	if err != nil {
		t.Fatalf("Route() error = %v", err)
	}
	if got != tools.DomainService {
		t.Fatalf("Route() = %q, want %q", got, tools.DomainService)
	}
}

type staticModel struct {
	content string
}

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

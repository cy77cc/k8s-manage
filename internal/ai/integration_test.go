package ai

import (
	"context"
	"encoding/json"
	"testing"

	einomodel "github.com/cloudwego/eino/components/model"
	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	aigraph "github.com/cy77cc/k8s-manage/internal/ai/graph"
	airouter "github.com/cy77cc/k8s-manage/internal/ai/router"
	aitools "github.com/cy77cc/k8s-manage/internal/ai/tools"
)

func TestRouterGraphExecutionWorkflow(t *testing.T) {
	classifier := airouter.NewIntentClassifier(nil, nil)
	router, err := airouter.NewIntentRouter(context.Background(), classifier)
	if err != nil {
		t.Fatalf("new intent router: %v", err)
	}
	domain, err := router.Route(context.Background(), "restart service api")
	if err != nil {
		t.Fatalf("route: %v", err)
	}
	if domain != aitools.DomainService {
		t.Fatalf("expected service domain, got %s", domain)
	}

	graph, err := aigraph.NewActionGraph(context.Background(), aigraph.ActionGraphConfig{
		ChatModel: integrationChatModel{
			msg: schema.AssistantMessage("", []schema.ToolCall{{
				ID:   "call-1",
				Type: "function",
				Function: schema.FunctionCall{
					Name:      "service_echo",
					Arguments: `{"message":"done"}`,
				},
			}}),
		},
		Tools: []aicore.RegisteredTool{{
			Meta: aicore.ToolMeta{
				Name:        "service_echo",
				Description: "Echo tool",
				Domain:      aitools.DomainService,
			},
			Tool: integrationTool{name: "service_echo"},
		}},
	})
	if err != nil {
		t.Fatalf("new action graph: %v", err)
	}

	out, err := graph.Invoke(context.Background(), aigraph.ActionInput{Message: "restart service api"})
	if err != nil {
		t.Fatalf("invoke graph: %v", err)
	}
	if len(out.ToolCalls) != 1 || out.ToolCalls[0].Content != `{"echo":"done"}` {
		t.Fatalf("unexpected graph output: %+v", out)
	}
}

type integrationChatModel struct {
	msg *schema.Message
}

func (m integrationChatModel) Generate(_ context.Context, _ []*schema.Message, _ ...einomodel.Option) (*schema.Message, error) {
	return m.msg, nil
}

func (m integrationChatModel) Stream(_ context.Context, _ []*schema.Message, _ ...einomodel.Option) (*schema.StreamReader[*schema.Message], error) {
	return schema.StreamReaderFromArray([]*schema.Message{m.msg}), nil
}

func (m integrationChatModel) WithTools(_ []*schema.ToolInfo) (einomodel.ToolCallingChatModel, error) {
	return m, nil
}

type integrationTool struct {
	name string
}

func (t integrationTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: t.name,
		Desc: "integration tool",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"message": {Type: "string", Required: true},
		}),
	}, nil
}

func (t integrationTool) InvokableRun(_ context.Context, argumentsInJSON string, _ ...einotool.Option) (string, error) {
	var payload map[string]any
	if err := json.Unmarshal([]byte(argumentsInJSON), &payload); err != nil {
		return "", err
	}
	return `{"echo":"` + payload["message"].(string) + `"}`, nil
}

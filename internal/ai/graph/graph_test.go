package graph

import (
	"context"
	"encoding/json"
	"testing"

	einomodel "github.com/cloudwego/eino/components/model"
	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	airag "github.com/cy77cc/k8s-manage/internal/ai/rag"
	"github.com/cy77cc/k8s-manage/internal/ai/tools"
)

func TestActionGraphInvoke_NoToolCall(t *testing.T) {
	t.Parallel()

	g, err := NewActionGraph(context.Background(), ActionGraphConfig{
		ChatModel: staticChatModel{
			msg: schema.AssistantMessage("summarized response", nil),
		},
	})
	if err != nil {
		t.Fatalf("NewActionGraph() error = %v", err)
	}

	out, err := g.Invoke(context.Background(), ActionInput{Message: "hello token world"})
	if err != nil {
		t.Fatalf("Invoke() error = %v", err)
	}
	if out.Response != "summarized response" {
		t.Fatalf("Invoke().Response = %q, want %q", out.Response, "summarized response")
	}
	if len(out.ToolCalls) != 0 {
		t.Fatalf("Invoke().ToolCalls len = %d, want 0", len(out.ToolCalls))
	}
}

func TestActionGraphInvoke_ToolExecution(t *testing.T) {
	t.Parallel()

	g, err := NewActionGraph(context.Background(), ActionGraphConfig{
		ChatModel: staticChatModel{
			msg: schema.AssistantMessage("", []schema.ToolCall{{
				ID:   "call-1",
				Type: "function",
				Function: schema.FunctionCall{
					Name:      "service_echo",
					Arguments: `{"message":"ok"}`,
				},
			}}),
		},
		Tools: []tools.RegisteredTool{{
			Meta: tools.ToolMeta{
				Name:        "service_echo",
				Description: "Echoes a message.",
				Required:    []string{"message"},
				Domain:      tools.DomainService,
			},
			Tool: fakeTool{name: "service_echo"},
		}},
	})
	if err != nil {
		t.Fatalf("NewActionGraph() error = %v", err)
	}

	out, err := g.Invoke(context.Background(), ActionInput{Message: "deploy service"})
	if err != nil {
		t.Fatalf("Invoke() error = %v", err)
	}
	if len(out.ToolCalls) != 1 {
		t.Fatalf("Invoke().ToolCalls len = %d, want 1", len(out.ToolCalls))
	}
	if out.ToolCalls[0].Content != `{"echo":"ok"}` {
		t.Fatalf("Invoke().ToolCalls[0].Content = %q", out.ToolCalls[0].Content)
	}
}

func TestActionGraphInvoke_ValidationFailure(t *testing.T) {
	t.Parallel()

	g, err := NewActionGraph(context.Background(), ActionGraphConfig{
		ChatModel: staticChatModel{
			msg: schema.AssistantMessage("", []schema.ToolCall{{
				ID:   "call-1",
				Type: "function",
				Function: schema.FunctionCall{
					Name:      "k8s_apply_manifest",
					Arguments: `{"manifest":{"kind":"Deployment","metadata":{}}}`,
				},
			}}),
		},
		Tools: []tools.RegisteredTool{{
			Meta: tools.ToolMeta{
				Name:        "k8s_apply_manifest",
				Description: "Apply a manifest.",
				Domain:      tools.DomainInfrastructure,
			},
			Tool: fakeTool{name: "k8s_apply_manifest"},
		}},
	})
	if err != nil {
		t.Fatalf("NewActionGraph() error = %v", err)
	}

	if _, err := g.Invoke(context.Background(), ActionInput{Message: "apply manifest"}); err == nil {
		t.Fatal("Invoke() error = nil, want validation failure")
	}
}

func TestActionGraphInvoke_AugmentsPromptWithKnowledge(t *testing.T) {
	t.Parallel()

	indexer := airag.NewMilvusIndexer(nil)
	if _, err := indexer.AddUserKnowledge(context.Background(), "team-a", "restart service", "Use the service_restart tool"); err != nil {
		t.Fatalf("seed knowledge: %v", err)
	}

	model := &captureChatModel{msg: schema.AssistantMessage("ok", nil)}
	g, err := NewActionGraph(context.Background(), ActionGraphConfig{
		ChatModel: model,
		Retriever: airag.NewNamespaceRetriever(indexer),
	})
	if err != nil {
		t.Fatalf("NewActionGraph() error = %v", err)
	}

	if _, err := g.Invoke(context.Background(), ActionInput{
		Message: "restart service",
		Context: map[string]any{"namespace": "team-a"},
	}); err != nil {
		t.Fatalf("Invoke() error = %v", err)
	}
	if len(model.messages) == 0 || model.messages[len(model.messages)-1].Role != schema.User {
		t.Fatalf("expected captured user message, got %#v", model.messages)
	}
	if got := model.messages[len(model.messages)-1].Content; got == "restart service" {
		t.Fatalf("expected augmented prompt, got %q", got)
	}
}

type staticChatModel struct {
	msg *schema.Message
}

var _ einomodel.ToolCallingChatModel = staticChatModel{}

func (s staticChatModel) Generate(_ context.Context, _ []*schema.Message, _ ...einomodel.Option) (*schema.Message, error) {
	return s.msg, nil
}

func (s staticChatModel) Stream(_ context.Context, _ []*schema.Message, _ ...einomodel.Option) (*schema.StreamReader[*schema.Message], error) {
	return schema.StreamReaderFromArray([]*schema.Message{s.msg}), nil
}

func (s staticChatModel) WithTools(_ []*schema.ToolInfo) (einomodel.ToolCallingChatModel, error) {
	return s, nil
}

type captureChatModel struct {
	msg      *schema.Message
	messages []*schema.Message
}

func (c *captureChatModel) Generate(_ context.Context, messages []*schema.Message, _ ...einomodel.Option) (*schema.Message, error) {
	c.messages = append([]*schema.Message(nil), messages...)
	return c.msg, nil
}

func (c *captureChatModel) Stream(_ context.Context, _ []*schema.Message, _ ...einomodel.Option) (*schema.StreamReader[*schema.Message], error) {
	return schema.StreamReaderFromArray([]*schema.Message{c.msg}), nil
}

func (c *captureChatModel) WithTools(_ []*schema.ToolInfo) (einomodel.ToolCallingChatModel, error) {
	return c, nil
}

type fakeTool struct {
	name string
}

var _ einotool.InvokableTool = fakeTool{}

func (f fakeTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: f.name,
		Desc: "fake tool",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"message": {
				Type:     "string",
				Desc:     "message",
				Required: false,
			},
		}),
	}, nil
}

func (f fakeTool) InvokableRun(_ context.Context, argumentsInJSON string, _ ...einotool.Option) (string, error) {
	var payload map[string]any
	if err := json.Unmarshal([]byte(argumentsInJSON), &payload); err != nil {
		return "", err
	}
	return `{"echo":"` + payload["message"].(string) + `"}`, nil
}

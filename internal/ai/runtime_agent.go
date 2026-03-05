package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/cy77cc/k8s-manage/internal/ai/tools"
	"github.com/cy77cc/k8s-manage/internal/config"
	"github.com/cy77cc/k8s-manage/internal/rag"
)

type PlatformAgent struct {
	ADKAgent     adk.Agent
	Model        model.ToolCallingChatModel
	tools        map[string]tool.InvokableTool
	metas        map[string]tools.ToolMeta
	mcp          *tools.MCPClientManager
	ragRetriever ragPromptRetriever
}

type ragPromptRetriever interface {
	Retrieve(ctx context.Context, query string, topK int) (*rag.RAGContext, error)
	BuildAugmentedPrompt(query string, context *rag.RAGContext) string
}

func NewPlatformAgent(ctx context.Context, chatModel model.ToolCallingChatModel, deps tools.PlatformDeps) (*PlatformAgent, error) {
	if chatModel == nil {
		return nil, nil
	}

	mcpManager, err := tools.NewMCPClientManager(ctx, tools.MCPConfigFromEnv())
	if err != nil {
		return nil, err
	}
	registered, err := tools.BuildRegisteredToolsWithMCP(deps, mcpManager)
	if err != nil {
		return nil, err
	}

	baseTools := make([]tool.BaseTool, 0, len(registered))
	toolMap := make(map[string]tool.InvokableTool, len(registered))
	metaMap := make(map[string]tools.ToolMeta, len(registered))
	for _, item := range registered {
		baseTools = append(baseTools, tools.WrapRegisteredTool(item))
		toolMap[item.Meta.Name] = item.Tool
		metaMap[item.Meta.Name] = item.Meta
	}

	adkAgent, err := newADKPlanExecuteAgent(ctx, chatModel, baseTools)
	if err != nil {
		return nil, err
	}

	var ragRetriever ragPromptRetriever
	if deps.DB != nil && config.CFG.Milvus.Enable {
		milvusClient := rag.NewMilvusClient(config.CFG.Milvus)
		if err := milvusClient.EnsureCollections(ctx); err != nil {
			return nil, fmt.Errorf("ensure milvus collections: %w", err)
		}
		ragRetriever = rag.NewRAGRetriever(milvusClient, rag.NewEmbedder(config.CFG.Embedder))
	}

	return &PlatformAgent{
		ADKAgent:     adkAgent,
		Model:        chatModel,
		tools:        toolMap,
		metas:        metaMap,
		mcp:          mcpManager,
		ragRetriever: ragRetriever,
	}, nil
}

func (p *PlatformAgent) ToolMetas() []tools.ToolMeta {
	if p == nil {
		return nil
	}
	out := make([]tools.ToolMeta, 0, len(p.metas))
	for _, m := range p.metas {
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (p *PlatformAgent) Stream(ctx context.Context, messages []*schema.Message) (*schema.StreamReader[*schema.Message], error) {
	if p == nil {
		return nil, fmt.Errorf("agent not initialized")
	}
	if p.Model == nil {
		return nil, fmt.Errorf("chat model not initialized")
	}
	messages = p.injectRAGIntoMessages(ctx, messages)
	return p.Model.Stream(ctx, messages)
}

func (p *PlatformAgent) Generate(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	if p == nil {
		return nil, fmt.Errorf("agent not initialized")
	}
	if p.Model == nil {
		return nil, fmt.Errorf("chat model not initialized")
	}
	messages = p.injectRAGIntoMessages(ctx, messages)
	return p.Model.Generate(ctx, messages)
}

func (p *PlatformAgent) RunTool(ctx context.Context, toolName string, params map[string]any) (tools.ToolResult, error) {
	if p == nil {
		return tools.ToolResult{OK: false, Error: "agent not initialized", Source: "platform"}, fmt.Errorf("agent not initialized")
	}
	normalizedName := tools.NormalizeToolName(toolName)
	t, ok := p.tools[normalizedName]
	if !ok {
		return tools.ToolResult{OK: false, Error: "tool not found", Source: "platform"}, fmt.Errorf("tool not found")
	}
	raw, err := json.Marshal(params)
	if err != nil {
		return tools.ToolResult{OK: false, Error: err.Error(), Source: "platform"}, err
	}
	out, err := t.InvokableRun(ctx, string(raw))
	if err != nil {
		return tools.ToolResult{OK: false, Error: err.Error(), Source: "platform"}, nil
	}
	var result tools.ToolResult
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		return tools.ToolResult{OK: true, Data: out, Source: "platform"}, nil
	}
	return result, nil
}

func (p *PlatformAgent) injectRAGIntoMessages(ctx context.Context, messages []*schema.Message) []*schema.Message {
	if p == nil || p.ragRetriever == nil || len(messages) == 0 {
		return messages
	}
	userIdx := -1
	userContent := ""
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		if msg != nil && msg.Role == schema.User {
			userIdx = i
			userContent = strings.TrimSpace(msg.Content)
			break
		}
	}
	if userIdx < 0 || userContent == "" {
		return messages
	}
	contextData, err := p.ragRetriever.Retrieve(ctx, userContent, 6)
	if err != nil {
		return messages
	}
	augmented := p.ragRetriever.BuildAugmentedPrompt(userContent, contextData)
	if strings.TrimSpace(augmented) == "" || augmented == userContent {
		return messages
	}
	out := append([]*schema.Message(nil), messages...)
	copied := *out[userIdx]
	copied.Content = augmented
	out[userIdx] = &copied
	return out
}

func (p *PlatformAgent) Close() error {
	if p == nil || p.mcp == nil {
		return nil
	}
	return p.mcp.Close()
}

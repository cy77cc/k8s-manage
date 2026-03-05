package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"github.com/cy77cc/k8s-manage/internal/ai/experts"
	aigraph "github.com/cy77cc/k8s-manage/internal/ai/graph"
	askills "github.com/cy77cc/k8s-manage/internal/ai/skills"
	"github.com/cy77cc/k8s-manage/internal/ai/tools"
	"github.com/cy77cc/k8s-manage/internal/config"
	"github.com/cy77cc/k8s-manage/internal/rag"
)

const toolCallGuide = `Tool calling rules:
1) NEVER call a tool with empty {} arguments when the tool has required fields.
2) Prefer using runtime context (scene/page/host_id/cluster_id/service_id/namespace) as arguments.
3) If any required field is missing, ask for it or choose a safe readonly tool first.
4) For mutating tools, require approval token before execution.
5) For inventory/list/assets/清单/已添加资源 requests, call inventory tools first (host_list_inventory/cluster_list_inventory/service_list_inventory).
6) Return concise explanation after each tool result.`

type PlatformAgent struct {
	Runnable      *react.Agent
	Model         model.ToolCallingChatModel
	registry      experts.ExpertRegistry
	router        *experts.HybridRouter
	graphRunner   compose.Runnable[*aigraph.GraphInput, *aigraph.GraphOutput]
	streamRunner  compose.Runnable[*aigraph.GraphInput, *schema.StreamReader[*schema.Message]]
	agentOptions  []agent.AgentOption
	tools         map[string]tool.InvokableTool
	metas         map[string]tools.ToolMeta
	mcp           *tools.MCPClientManager
	ragRetriever  ragPromptRetriever
	skillRegistry *askills.Registry
	skillExecutor *askills.Executor
}

var scenePattern = regexp.MustCompile(`scene:([a-z0-9:_-]+)`)

type ragPromptRetriever interface {
	Retrieve(ctx context.Context, query string, topK int) (*rag.RAGContext, error)
	BuildAugmentedPrompt(query string, context *rag.RAGContext) string
}

func NewPlatformAgent(ctx context.Context, chatModel model.ToolCallingChatModel, deps tools.PlatformDeps) (*PlatformAgent, error) {
	if chatModel == nil {
		return nil, nil
	}

	localTools, err := tools.BuildLocalTools(deps)
	if err != nil {
		return nil, err
	}
	mcpManager, err := tools.NewMCPClientManager(ctx, tools.MCPConfigFromEnv())
	if err != nil {
		return nil, err
	}
	mcpTools, err := tools.BuildMCPProxyTools(mcpManager)
	if err != nil {
		return nil, err
	}
	registered := append(localTools, mcpTools...)
	baseTools := make([]tool.BaseTool, 0, len(registered))
	toolMap := make(map[string]tool.InvokableTool, len(registered))
	metaMap := make(map[string]tools.ToolMeta, len(registered))
	for _, item := range registered {
		baseTools = append(baseTools, item.Tool)
		toolMap[item.Meta.Name] = item.Tool
		metaMap[item.Meta.Name] = item.Meta
	}

	agentOptions, err := react.WithTools(ctx, baseTools...)
	if err != nil {
		return nil, err
	}
	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: chatModel,
		MaxStep:          20,
		MessageModifier:  react.NewPersonaModifier("You are Platform Ops Agent. Use tools safely and with complete parameters.\n" + toolCallGuide),
	})
	if err != nil {
		return nil, err
	}

	registry, err := experts.NewExpertRegistry(ctx, "configs/experts.yaml", toolMap, chatModel)
	if err != nil {
		return nil, err
	}
	router, err := experts.NewHybridRouter(registry, "configs/scene_mappings.yaml")
	if err != nil {
		return nil, err
	}
	graphBuilder := aigraph.NewBuilderWithRunners(
		aigraph.NewRegistryPrimaryRunner(registry),
		aigraph.NewRegistryHelperRunner(registry),
	)
	graphDef, err := graphBuilder.Build(ctx)
	if err != nil {
		return nil, err
	}
	graphRunner, err := graphDef.Compile(ctx)
	if err != nil {
		return nil, err
	}
	streamBuilder := aigraph.NewBuilderWithStreamRunners(
		aigraph.NewRegistryStreamPrimaryRunner(registry),
		aigraph.NewRegistryStreamHelperRunner(registry),
	)
	streamDef, err := streamBuilder.BuildStream(ctx)
	if err != nil {
		return nil, err
	}
	streamRunner, err := streamDef.Compile(ctx)
	if err != nil {
		return nil, err
	}
	var ragRetriever ragPromptRetriever
	if deps.DB != nil && config.CFG.Milvus.Enable {
		milvusClient := rag.NewMilvusClient(config.CFG.Milvus)
		if err := milvusClient.EnsureCollections(ctx); err != nil {
			log.Printf("warn: disable rag retriever because ensure milvus collections failed: %v", err)
		} else {
			ragRetriever = rag.NewRAGRetriever(milvusClient, rag.NewEmbedder(config.CFG.Embedder))
		}
	}
	var skillRegistry *askills.Registry
	if reg, err := askills.NewRegistry(askills.DefaultSkillsConfigPath); err == nil {
		skillRegistry = reg
	}
	skillExecutor := askills.NewExecutor(func(ctx context.Context, toolName string, params map[string]any) (tools.ToolResult, error) {
		normalized := tools.NormalizeToolName(toolName)
		t, ok := toolMap[normalized]
		if !ok {
			return tools.ToolResult{OK: false, Error: "tool not found", Source: "platform"}, fmt.Errorf("tool not found: %s", normalized)
		}
		raw, err := json.Marshal(params)
		if err != nil {
			return tools.ToolResult{OK: false, Error: err.Error(), Source: "platform"}, err
		}
		out, err := t.InvokableRun(ctx, string(raw))
		if err != nil {
			return tools.ToolResult{OK: false, Error: err.Error(), Source: "platform"}, err
		}
		var result tools.ToolResult
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			return tools.ToolResult{OK: true, Data: out, Source: "platform"}, nil
		}
		return result, nil
	}, nil, nil)

	return &PlatformAgent{
		Runnable:      agent,
		Model:         chatModel,
		registry:      registry,
		router:        router,
		graphRunner:   graphRunner,
		streamRunner:  streamRunner,
		agentOptions:  agentOptions,
		tools:         toolMap,
		metas:         metaMap,
		mcp:           mcpManager,
		ragRetriever:  ragRetriever,
		skillRegistry: skillRegistry,
		skillExecutor: skillExecutor,
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
	if stream, handled := p.executeSkillStream(ctx, messages); handled {
		return stream, nil
	}
	messages = p.injectRAGIntoMessages(ctx, messages)
	req := p.buildExecuteRequest(ctx, messages)
	if req == nil || req.Decision == nil {
		return p.Runnable.Stream(ctx, messages, p.agentOptions...)
	}
	if req.Decision.Strategy == experts.StrategySingle {
		if exp, ok := p.registry.GetExpert(req.Decision.PrimaryExpert); ok && exp != nil && exp.Agent != nil {
			return exp.Agent.Stream(ctx, messages, exp.AgentOptions...)
		}
		return p.Runnable.Stream(ctx, messages, p.agentOptions...)
	}
	if p.streamRunner != nil {
		stream, err := p.streamRunner.Invoke(ctx, p.buildGraphInput(req))
		if err == nil && stream != nil {
			return stream, nil
		}
	}
	return p.Runnable.Stream(ctx, messages, p.agentOptions...)
}

func (p *PlatformAgent) Generate(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	if p == nil {
		return nil, fmt.Errorf("agent not initialized")
	}
	messages = p.injectRAGIntoMessages(ctx, messages)
	req := p.buildExecuteRequest(ctx, messages)
	if req == nil || req.Decision == nil {
		return p.Runnable.Generate(ctx, messages, p.agentOptions...)
	}
	if p.graphRunner != nil {
		out, err := p.graphRunner.Invoke(ctx, p.buildGraphInput(req))
		if err == nil && out != nil {
			return schema.AssistantMessage(out.Response, nil), nil
		}
	}
	return p.Runnable.Generate(ctx, messages, p.agentOptions...)
}

func sceneFromMessages(content string) string {
	if strings.TrimSpace(content) == "" {
		return ""
	}
	matches := scenePattern.FindStringSubmatch(strings.ToLower(content))
	if len(matches) < 2 {
		return ""
	}
	return matches[1]
}

func (p *PlatformAgent) buildExecuteRequest(ctx context.Context, messages []*schema.Message) *experts.ExecuteRequest {
	if p == nil || p.router == nil {
		return nil
	}
	content := ""
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i] != nil && messages[i].Role == schema.User {
			content = strings.TrimSpace(messages[i].Content)
			break
		}
	}
	scene := sceneFromMessages(content)
	decision := p.router.Route(ctx, &experts.RouteRequest{
		Message: content,
		Scene:   scene,
		History: messages,
		RuntimeContext: map[string]any{
			"scene": scene,
		},
	})
	return &experts.ExecuteRequest{
		Message:  content,
		Decision: decision,
		RuntimeContext: map[string]any{
			"scene": scene,
		},
		History:      messages,
		EventEmitter: experts.ProgressEmitterFromContext(ctx),
	}
}

func (p *PlatformAgent) buildGraphInput(req *experts.ExecuteRequest) *aigraph.GraphInput {
	if req == nil || req.Decision == nil {
		return &aigraph.GraphInput{}
	}
	input := &aigraph.GraphInput{
		Message:  req.Message,
		Request:  req,
		Strategy: req.Decision.Strategy,
	}
	switch req.Decision.Strategy {
	case experts.StrategyParallel, experts.StrategySequential:
		input.HelperRequests = make([]experts.HelperRequest, 0, len(req.Decision.OptionalHelpers))
		for _, helper := range req.Decision.OptionalHelpers {
			name := strings.TrimSpace(helper)
			if name == "" {
				continue
			}
			input.HelperRequests = append(input.HelperRequests, experts.HelperRequest{
				ExpertName: name,
				Task:       fmt.Sprintf("协助分析用户请求：%s", req.Message),
			})
		}
	}
	return input
}

func (p *PlatformAgent) RunTool(ctx context.Context, toolName string, params map[string]any) (tools.ToolResult, error) {
	if p == nil {
		return tools.ToolResult{
				OK:     false,
				Error:  "agent not initialized",
				Source: "platform",
			},
			fmt.Errorf("agent not initialized")
	}
	normalizedName := tools.NormalizeToolName(toolName)
	t, ok := p.tools[normalizedName]
	if !ok {
		return tools.ToolResult{
				OK:     false,
				Error:  "tool not found",
				Source: "platform",
			},
			fmt.Errorf("tool not found")
	}
	raw, err := json.Marshal(params)
	if err != nil {
		return tools.ToolResult{
				OK:     false,
				Error:  err.Error(),
				Source: "platform",
			},
			err
	}
	out, err := t.InvokableRun(ctx, string(raw))
	if err != nil {
		return tools.ToolResult{
				OK:     false,
				Error:  err.Error(),
				Source: "platform",
			},
			nil
	}
	var result tools.ToolResult
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		return tools.ToolResult{
				OK:     true,
				Data:   out,
				Source: "platform",
			},
			nil
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
	augmented := strings.TrimSpace(p.ragRetriever.BuildAugmentedPrompt(userContent, contextData))
	if augmented == "" || augmented == userContent {
		return messages
	}
	copied := append([]*schema.Message(nil), messages...)
	copied[userIdx] = schema.UserMessage(augmented)
	return copied
}

func (p *PlatformAgent) executeSkillStream(ctx context.Context, messages []*schema.Message) (*schema.StreamReader[*schema.Message], bool) {
	if p == nil || p.skillRegistry == nil || p.skillExecutor == nil {
		return nil, false
	}
	content := latestUserMessage(messages)
	if content == "" {
		return nil, false
	}
	skill, score := p.skillRegistry.MatchSkill(content)
	if skill == nil || score <= 0 {
		return nil, false
	}
	result, err := p.skillExecutor.ExecuteFromMessage(ctx, *skill, content, nil)
	if err != nil {
		msg := fmt.Sprintf("技能 `%s` 执行失败: %v", skill.Name, err)
		if strings.Contains(strings.ToLower(err.Error()), "missing required parameter") {
			msg = fmt.Sprintf("技能 `%s` 缺少必填参数，请补充后重试。错误: %v", skill.Name, err)
		}
		stream := schema.StreamReaderFromArray([]*schema.Message{schema.AssistantMessage(msg, nil)})
		return stream, true
	}
	output := formatSkillExecutionResult(skill.Name, result)
	stream := schema.StreamReaderFromArray([]*schema.Message{schema.AssistantMessage(output, nil)})
	return stream, true
}

func latestUserMessage(messages []*schema.Message) string {
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		if msg != nil && msg.Role == schema.User {
			return strings.TrimSpace(msg.Content)
		}
	}
	return ""
}

func formatSkillExecutionResult(skillName string, result *askills.ExecutionResult) string {
	if result == nil {
		return fmt.Sprintf("技能 `%s` 执行完成。", skillName)
	}
	lines := []string{fmt.Sprintf("技能 `%s` 执行完成。", skillName)}
	for step, data := range result.StepResults {
		lines = append(lines, fmt.Sprintf("- 步骤 %s: %v", step, data))
	}
	return strings.Join(lines, "\n")
}

func (p *PlatformAgent) Close() error {
	if p == nil || p.mcp == nil {
		return nil
	}
	return p.mcp.Close()
}

package graph

import (
	"context"
	"fmt"
	"strings"

	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	airag "github.com/cy77cc/k8s-manage/internal/ai/rag"
	"github.com/cy77cc/k8s-manage/internal/ai/tools"
	"github.com/cy77cc/k8s-manage/internal/ai/tools/core"
)

const (
	nodeSanitize   = "sanitize"
	nodeReasoning  = "reasoning"
	nodeValidation = "validation"
	nodeExecution  = "execution"
)

type ActionGraph struct {
	workflow  *compose.Workflow[actionEnvelope, actionEnvelope]
	runnable  compose.Runnable[actionEnvelope, actionEnvelope]
	chatModel einomodel.ToolCallingChatModel
	toolsNode *compose.ToolsNode
	registry  *tools.Registry
	validator Validator
	retriever airag.Retriever
}

type ActionGraphConfig struct {
	ChatModel       einomodel.ToolCallingChatModel
	Tools           []core.RegisteredTool
	Validator       Validator
	CheckPointStore compose.CheckPointStore
	Retriever       airag.Retriever
}

type actionEnvelope struct {
	Input        ActionInput
	State        *GraphState
	Assistant    *schema.Message
	ToolMessages []*schema.Message
	Output       ActionOutput
}

func NewActionGraph(ctx context.Context, cfg ActionGraphConfig) (*ActionGraph, error) {
	registry := tools.NewRegistry(cfg.Tools)
	validator := cfg.Validator
	if validator == nil {
		validator = NewOpenAPIValidator(registry)
	}

	toolNode, err := newToolsNode(ctx, cfg.Tools)
	if err != nil {
		return nil, fmt.Errorf("build tools node: %w", err)
	}

	ag := &ActionGraph{
		workflow:  compose.NewWorkflow[actionEnvelope, actionEnvelope](),
		chatModel: cfg.ChatModel,
		toolsNode: toolNode,
		registry:  registry,
		validator: validator,
		retriever: cfg.Retriever,
	}

	ag.workflow.AddLambdaNode(nodeSanitize, compose.InvokableLambda(ag.sanitizeNode)).AddInput(compose.START)
	ag.workflow.AddLambdaNode(nodeReasoning, compose.InvokableLambda(ag.reasoningNode)).AddInput(nodeSanitize)
	ag.workflow.AddLambdaNode(nodeValidation, compose.InvokableLambda(ag.validationNode)).AddInput(nodeReasoning)
	ag.workflow.AddLambdaNode(nodeExecution, compose.InvokableLambda(ag.executionNode)).AddInput(nodeValidation)
	ag.workflow.End().AddInput(nodeExecution)

	compileOpts := make([]compose.GraphCompileOption, 0, 1)
	if cfg.CheckPointStore != nil {
		compileOpts = append(compileOpts, compose.WithCheckPointStore(cfg.CheckPointStore))
	}
	runnable, err := ag.workflow.Compile(ctx, compileOpts...)
	if err != nil {
		return nil, fmt.Errorf("compile workflow: %w", err)
	}
	ag.runnable = runnable

	return ag, nil
}

func (g *ActionGraph) Invoke(ctx context.Context, input ActionInput) (ActionOutput, error) {
	if g == nil || g.runnable == nil {
		return ActionOutput{}, fmt.Errorf("action graph is not initialized")
	}

	env, err := g.runnable.Invoke(ctx, actionEnvelope{
		Input: input,
		State: NewGraphState(),
	})
	if err != nil {
		return ActionOutput{}, err
	}
	return env.Output, nil
}

func (g *ActionGraph) Workflow() *compose.Workflow[actionEnvelope, actionEnvelope] {
	if g == nil {
		return nil
	}
	return g.workflow
}

func (g *ActionGraph) sanitizeNode(_ context.Context, env actionEnvelope) (actionEnvelope, error) {
	env.State = ensureState(env.State)
	env.Input.Message = sanitizeText(env.Input.Message)
	env.State.AddMessage(schema.UserMessage(env.Input.Message))
	return env, nil
}

func (g *ActionGraph) reasoningNode(ctx context.Context, env actionEnvelope) (actionEnvelope, error) {
	env.State = ensureState(env.State)
	if g.chatModel == nil {
		env.Assistant = schema.AssistantMessage(env.Input.Message, nil)
		env.State.AddMessage(env.Assistant)
		env.Output.Response = env.Assistant.Content
		return env, nil
	}

	modelToUse := g.chatModel
	messages := env.State.Messages
	if g.retriever != nil {
		namespace := knowledgeNamespace(env.Input)
		if entries, err := g.retriever.Retrieve(ctx, namespace, env.Input.Message, 4); err == nil && len(entries) > 0 {
			messages = append([]*schema.Message(nil), env.State.Messages...)
			if len(messages) > 0 {
				messages[len(messages)-1] = schema.UserMessage(airag.BuildAugmentedPrompt(env.Input.Message, entries))
			}
		}
	}
	toolInfos, err := g.toolInfos(ctx)
	if err != nil {
		return env, err
	}
	if len(toolInfos) > 0 {
		bound, err := modelToUse.WithTools(toolInfos)
		if err != nil {
			return env, fmt.Errorf("bind tools: %w", err)
		}
		modelToUse = bound
	}

	msg, err := modelToUse.Generate(ctx, messages)
	if err != nil {
		return env, fmt.Errorf("generate reasoning response: %w", err)
	}
	env.Assistant = msg
	env.State.AddMessage(msg)
	env.State.SetPendingToolCalls(msg.ToolCalls)
	env.Output.Response = msg.Content
	return env, nil
}

func (g *ActionGraph) validationNode(ctx context.Context, env actionEnvelope) (actionEnvelope, error) {
	env.State = ensureState(env.State)
	if env.Assistant == nil || len(env.Assistant.ToolCalls) == 0 {
		return env, nil
	}
	for _, validationErr := range g.validator.Validate(ctx, env.Assistant.ToolCalls) {
		env.State.AddValidationError(validationErr)
	}
	if env.State.HasValidationErrors() {
		msgs := make([]string, 0, len(env.State.ValidationErrors))
		for _, validationErr := range env.State.ValidationErrors {
			msgs = append(msgs, validationErr.Message)
		}
		return env, fmt.Errorf("validation failed: %s", strings.Join(msgs, "; "))
	}
	return env, nil
}

func (g *ActionGraph) executionNode(ctx context.Context, env actionEnvelope) (actionEnvelope, error) {
	env.State = ensureState(env.State)
	if env.Assistant == nil || len(env.Assistant.ToolCalls) == 0 || g.toolsNode == nil {
		return env, nil
	}

	msgs, err := g.toolsNode.Invoke(ctx, env.Assistant)
	if err != nil {
		return env, fmt.Errorf("execute tools: %w", err)
	}
	env.ToolMessages = append(env.ToolMessages, msgs...)
	results := make([]ToolCallResult, 0, len(msgs))
	for _, msg := range msgs {
		if msg == nil {
			continue
		}
		env.State.AddMessage(msg)
		env.State.AddToolResult(ToolResult{
			CallID:   msg.ToolCallID,
			ToolName: msg.ToolName,
			Content:  msg.Content,
		})
		results = append(results, ToolCallResult{
			ID:      msg.ToolCallID,
			Name:    msg.ToolName,
			Content: msg.Content,
		})
		if strings.TrimSpace(env.Output.Response) == "" {
			env.Output.Response = msg.Content
		}
	}
	env.Output.ToolCalls = results
	return env, nil
}

func (g *ActionGraph) toolInfos(ctx context.Context) ([]*schema.ToolInfo, error) {
	if g.toolsNode == nil || g.registry == nil {
		return nil, nil
	}
	allDomains := []tools.ToolDomain{
		tools.DomainInfrastructure,
		tools.DomainService,
		tools.DomainCICD,
		tools.DomainMonitor,
		tools.DomainConfig,
		tools.DomainGeneral,
		tools.DomainUser,
	}
	seen := make(map[string]struct{})
	out := make([]*schema.ToolInfo, 0)
	for _, domain := range allDomains {
		for _, registered := range g.registry.ByDomain(domain) {
			info, err := registered.Tool.Info(ctx)
			if err != nil {
				return nil, err
			}
			if _, ok := seen[info.Name]; ok {
				continue
			}
			seen[info.Name] = struct{}{}
			out = append(out, info)
		}
	}
	return out, nil
}

func newToolsNode(ctx context.Context, registered []core.RegisteredTool) (*compose.ToolsNode, error) {
	if len(registered) == 0 {
		return nil, nil
	}
	toolList := make([]tool.BaseTool, 0, len(registered))
	for _, item := range registered {
		toolList = append(toolList, item.Tool)
	}
	return compose.NewToolNode(ctx, &compose.ToolsNodeConfig{Tools: toolList})
}

func sanitizeText(input string) string {
	replacer := strings.NewReplacer("token", "[redacted]", "password", "[redacted]", "secret", "[redacted]")
	return strings.TrimSpace(replacer.Replace(input))
}

func ensureState(state *GraphState) *GraphState {
	if state != nil {
		return state
	}
	return NewGraphState()
}

func knowledgeNamespace(input ActionInput) string {
	if input.Context == nil {
		return "global"
	}
	if ns := strings.TrimSpace(fmt.Sprintf("%v", input.Context["namespace"])); ns != "" {
		return ns
	}
	if ns := strings.TrimSpace(fmt.Sprintf("%v", input.Context["tenant"])); ns != "" {
		return ns
	}
	return "global"
}

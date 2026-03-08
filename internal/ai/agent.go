package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	aigraph "github.com/cy77cc/k8s-manage/internal/ai/graph"
	airag "github.com/cy77cc/k8s-manage/internal/ai/rag"
	airouter "github.com/cy77cc/k8s-manage/internal/ai/router"
	aistate "github.com/cy77cc/k8s-manage/internal/ai/state"
	"github.com/cy77cc/k8s-manage/internal/ai/tools"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RunnerConfig struct {
	EnableStreaming bool
	RedisClient     redis.UniversalClient
}

type AIAgent struct {
	model        einomodel.ToolCallingChatModel
	registered   []tools.RegisteredTool
	registry     *tools.Registry
	router       *airouter.IntentRouter
	graphs       map[tools.ToolDomain]*aigraph.ActionGraph
	sessionState *aistate.SessionState
	knowledge    *airag.MilvusIndexer
	retriever    airag.Retriever
	feedback     *airag.SessionFeedbackCollector
}

func NewAIAgent(ctx context.Context, model einomodel.ToolCallingChatModel, deps tools.PlatformDeps, cfg *RunnerConfig) (*AIAgent, error) {
	registered, err := tools.BuildRegisteredTools(deps)
	if err != nil {
		return nil, err
	}
	agent := &AIAgent{
		model:      model,
		registered: registered,
		registry:   tools.NewRegistry(registered),
		graphs:     make(map[tools.ToolDomain]*aigraph.ActionGraph),
	}
	if cfg != nil && cfg.RedisClient != nil {
		agent.sessionState = aistate.NewSessionState(cfg.RedisClient, "")
	}
	agent.router, err = airouter.NewIntentRouter(ctx, airouter.NewIntentClassifier(model, nil))
	if err != nil {
		return nil, err
	}
	var checkpointStore *aistate.CheckpointStore
	knowledgeIndexer := airag.NewMilvusIndexer(nil)
	knowledgeRetriever := airag.NewNamespaceRetriever(knowledgeIndexer)
	agent.knowledge = knowledgeIndexer
	agent.retriever = knowledgeRetriever
	if cfg != nil && cfg.RedisClient != nil {
		checkpointStore = aistate.NewCheckpointStore(cfg.RedisClient, "")
		agent.feedback = airag.NewFeedbackCollector(knowledgeIndexer, airag.NewSessionQAExtractor(agent.sessionState))
	}
	for _, domain := range []tools.ToolDomain{
		tools.DomainGeneral,
		tools.DomainInfrastructure,
		tools.DomainService,
		tools.DomainCICD,
		tools.DomainMonitor,
		tools.DomainConfig,
		tools.DomainUser,
	} {
		graph, err := aigraph.NewActionGraph(ctx, aigraph.ActionGraphConfig{
			ChatModel:       model,
			Tools:           agent.toolsForDomain(domain),
			CheckPointStore: checkpointStore,
			Retriever:       knowledgeRetriever,
		})
		if err != nil {
			return nil, err
		}
		agent.graphs[domain] = graph
	}
	return agent, nil
}

func NewPlatformRunner(ctx context.Context, model einomodel.ToolCallingChatModel, deps tools.PlatformDeps, cfg *RunnerConfig) (*AIAgent, error) {
	return NewAIAgent(ctx, model, deps, cfg)
}

func (a *AIAgent) ToolMetas() []tools.ToolMeta {
	out := make([]tools.ToolMeta, 0, len(a.registered))
	for _, item := range a.registered {
		out = append(out, item.Meta)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (a *AIAgent) Generate(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	if a.model != nil {
		return a.model.Generate(ctx, messages)
	}
	last := ""
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i] != nil && messages[i].Role == schema.User {
			last = messages[i].Content
			break
		}
	}
	return schema.AssistantMessage(last, nil), nil
}

func (a *AIAgent) RunTool(ctx context.Context, toolName string, params map[string]any) (tools.ToolResult, error) {
	item, ok := a.registry.Get(toolName)
	if !ok {
		return tools.ToolResult{}, ErrToolNotFound
	}
	raw, err := json.Marshal(params)
	if err != nil {
		return tools.ToolResult{}, err
	}
	content, err := item.Tool.InvokableRun(ctx, string(raw))
	if err != nil {
		return tools.ToolResult{}, err
	}
	var result tools.ToolResult
	if json.Unmarshal([]byte(content), &result) == nil && (result.OK || result.Error != "" || result.Source != "") {
		return result, nil
	}
	return tools.ToolResult{OK: true, Data: map[string]any{"content": content}, Source: "tool"}, nil
}

func (a *AIAgent) Query(ctx context.Context, sessionID, message string) (*aigraph.ActionOutput, error) {
	if strings.TrimSpace(sessionID) == "" {
		sessionID = "sess-" + uuid.NewString()
	}
	domain, err := a.router.Route(ctx, message)
	if err != nil {
		domain = tools.DomainGeneral
	}
	graph := a.graphs[domain]
	if graph == nil {
		graph = a.graphs[tools.DomainGeneral]
	}
	out, err := graph.Invoke(ctx, aigraph.ActionInput{SessionID: sessionID, Message: message})
	if err != nil {
		return nil, err
	}
	if a.sessionState != nil {
		_ = a.sessionState.AppendMessage(ctx, sessionID, schema.UserMessage(message))
		_ = a.sessionState.AppendMessage(ctx, sessionID, schema.AssistantMessage(out.Response, nil))
	}
	return &out, nil
}

func (a *AIAgent) Resume(_ context.Context, sessionID string, response map[string]any) (map[string]any, error) {
	return map[string]any{
		"checkpoint_id": sessionID,
		"resumed":       true,
		"response":      response,
	}, nil
}

func (a *AIAgent) FindMeta(name string) (tools.ToolMeta, bool) {
	item, ok := a.registry.Get(name)
	if !ok {
		return tools.ToolMeta{}, false
	}
	return item.Meta, true
}

func (a *AIAgent) AddKnowledge(ctx context.Context, namespace, question, answer string) (*airag.KnowledgeEntry, error) {
	if a == nil || a.knowledge == nil {
		return nil, ErrToolNotFound
	}
	entry, err := a.knowledge.AddUserKnowledge(ctx, namespace, question, answer)
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (a *AIAgent) CollectFeedback(ctx context.Context, sessionID, namespace string, feedback airag.Feedback, question, answer string) (*airag.KnowledgeEntry, error) {
	if a == nil || a.knowledge == nil {
		return nil, ErrToolNotFound
	}
	if strings.TrimSpace(question) != "" || strings.TrimSpace(answer) != "" {
		entry := airag.KnowledgeEntry{
			ID:        "feedback-" + uuid.NewString(),
			Source:    airag.SourceFeedback,
			Namespace: strings.TrimSpace(namespace),
			Question:  strings.TrimSpace(question),
			Answer:    strings.TrimSpace(answer),
		}
		if err := a.knowledge.Index(ctx, []airag.KnowledgeEntry{entry}); err != nil {
			return nil, err
		}
		return &entry, nil
	}
	if a.feedback == nil {
		return nil, fmt.Errorf("feedback collector is not initialized")
	}
	return a.feedback.Collect(ctx, sessionID, namespace, feedback)
}

func (a *AIAgent) toolsForDomain(domain tools.ToolDomain) []tools.RegisteredTool {
	if domain == tools.DomainGeneral {
		return append([]tools.RegisteredTool(nil), a.registered...)
	}
	items := a.registry.ByDomain(domain)
	if len(items) == 0 {
		return append([]tools.RegisteredTool(nil), a.registered...)
	}
	return items
}

package ai

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cy77cc/k8s-manage/internal/ai/agent"
	"github.com/cy77cc/k8s-manage/internal/ai/classifier"
	"github.com/cy77cc/k8s-manage/internal/ai/modes"
	aitools "github.com/cy77cc/k8s-manage/internal/ai/tools"
	"github.com/cy77cc/k8s-manage/internal/ai/types"
	"github.com/cy77cc/k8s-manage/internal/config"
)

type HybridAgent struct {
	classifier      *classifier.IntentClassifier
	simpleChat      *modes.SimpleChatMode
	agenticMode     *modes.AgenticMode
	multiDomainMode *modes.MultiDomainMode
	useMultiDomain  bool
}

func NewHybridAgent(ctx context.Context, chatModel model.ToolCallingChatModel, classifierModel model.ToolCallingChatModel, deps aitools.PlatformDeps, cfg *agent.RunnerConfig) (*HybridAgent, error) {
	classifierBackend := classifierModel
	if classifierBackend == nil {
		classifierBackend = chatModel
	}
	agenticMode, err := modes.NewAgenticMode(ctx, chatModel, deps, cfg)
	if err != nil {
		return nil, err
	}
	useMultiDomain := config.CFG.AI.UseMultiDomainArch
	if cfg != nil && cfg.UseMultiDomainArch {
		useMultiDomain = true
	}
	return &HybridAgent{
		classifier:      classifier.NewIntentClassifier(classifierBackend),
		simpleChat:      modes.NewSimpleChatMode(chatModel),
		agenticMode:     agenticMode,
		multiDomainMode: modes.NewMultiDomainMode(chatModel, deps),
		useMultiDomain:  useMultiDomain,
	}, nil
}

func (a *HybridAgent) Query(ctx context.Context, sessionID, message string) *adk.AsyncIterator[*types.AgentResult] {
	iter, gen := adk.NewAsyncIteratorPair[*types.AgentResult]()

	go func() {
		defer gen.Close()

		intent, err := a.classifier.Classify(ctx, message)
		if err != nil {
			gen.Send(&types.AgentResult{Type: "error", Content: err.Error()})
			return
		}

		switch intent {
		case classifier.IntentAgentic:
			if a.useMultiDomain && a.multiDomainMode != nil {
				a.queryMultiDomain(ctx, sessionID, message, gen)
				return
			}
			a.agenticMode.Execute(ctx, sessionID, message, gen)
		case classifier.IntentSimple:
			fallthrough
		default:
			a.simpleChat.Execute(ctx, message, gen)
		}
	}()

	return iter
}

func (a *HybridAgent) queryMultiDomain(ctx context.Context, sessionID, message string, gen *adk.AsyncGenerator[*types.AgentResult]) {
	if a == nil || a.multiDomainMode == nil {
		gen.Send(&types.AgentResult{Type: "error", Content: "multi-domain mode not initialized"})
		return
	}
	a.multiDomainMode.Execute(ctx, sessionID, message, gen)
}

func (a *HybridAgent) Resume(ctx context.Context, sessionID, askID string, response any) (*types.AgentResult, error) {
	if a == nil || a.agenticMode == nil {
		return nil, fmt.Errorf("agentic mode not initialized")
	}
	return a.agenticMode.Resume(ctx, sessionID, askID, response)
}

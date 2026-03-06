package ai

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	aitools "github.com/cy77cc/k8s-manage/internal/ai/tools"
)

type HybridAgent struct {
	classifier  *IntentClassifier
	simpleChat  *SimpleChatMode
	agenticMode *AgenticMode
}

func NewHybridAgent(ctx context.Context, chatModel model.ToolCallingChatModel, classifierModel model.ToolCallingChatModel, deps aitools.PlatformDeps, cfg *RunnerConfig) (*HybridAgent, error) {
	classifierBackend := classifierModel
	if classifierBackend == nil {
		classifierBackend = chatModel
	}
	agenticMode, err := NewAgenticMode(ctx, chatModel, deps, cfg)
	if err != nil {
		return nil, err
	}
	return &HybridAgent{
		classifier:  NewIntentClassifier(classifierBackend),
		simpleChat:  NewSimpleChatMode(chatModel),
		agenticMode: agenticMode,
	}, nil
}

func (a *HybridAgent) Query(ctx context.Context, sessionID, message string) *adk.AsyncIterator[*AgentResult] {
	iter, gen := adk.NewAsyncIteratorPair[*AgentResult]()

	go func() {
		defer gen.Close()

		intent, err := a.classifier.Classify(ctx, message)
		if err != nil {
			gen.Send(&AgentResult{Type: "error", Content: err.Error()})
			return
		}

		switch intent {
		case IntentAgentic:
			a.agenticMode.Execute(ctx, sessionID, message, gen)
		case IntentSimple:
			fallthrough
		default:
			a.simpleChat.Execute(ctx, message, gen)
		}
	}()

	return iter
}

func (a *HybridAgent) Resume(ctx context.Context, sessionID, askID string, response any) (*AgentResult, error) {
	if a == nil || a.agenticMode == nil {
		return nil, fmt.Errorf("agentic mode not initialized")
	}
	return a.agenticMode.Resume(ctx, sessionID, askID, response)
}

// Package summarizer 实现 AI 编排的总结阶段。
//
// 本文件提供 ADK 集成，构建总结器 Agent。
package summarizer

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/adk"
	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// NewWithADK 使用 ADK 创建总结器实例。
func NewWithADK(ctx context.Context, model einomodel.BaseChatModel) (*Summarizer, error) {
	if model == nil {
		return nil, fmt.Errorf("summarizer model is required")
	}
	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:          "summarizer-stage",
		Description:   "Stream the final user-facing answer from executor outputs.",
		Instruction:   SystemPrompt(),
		Model:         model,
		MaxIterations: 2,
	})
	if err != nil {
		return nil, err
	}
	return &Summarizer{
		runner: adk.NewRunner(ctx, adk.RunnerConfig{Agent: agent}),
	}, nil
}

func runADKSummarizer(ctx context.Context, runner *adk.Runner, input string, onDelta func(string)) (string, error) {
	if runner == nil {
		return "", fmt.Errorf("summarizer ADK runner is not configured")
	}
	iter := runner.Query(ctx, input)
	var streamed string
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event == nil {
			continue
		}
		if event.Err != nil {
			return "", event.Err
		}
		msg, _, err := adk.GetMessage(event)
		if err != nil || msg == nil {
			continue
		}
		if msg.Role == schema.Assistant {
			streamed = emitSummaryDelta(streamed, msg.Content, onDelta)
		}
	}
	if strings.TrimSpace(streamed) == "" {
		return "", fmt.Errorf("summarizer stage produced empty output")
	}
	return strings.TrimSpace(streamed), nil
}

func emitSummaryDelta(previous, current string, onDelta func(string)) string {
	if onDelta == nil {
		return strings.TrimSpace(current)
	}
	current = strings.TrimSpace(current)
	previous = strings.TrimSpace(previous)
	if current == "" || current == previous {
		return current
	}
	if previous != "" && strings.HasPrefix(current, previous) {
		if delta := strings.TrimSpace(current[len(previous):]); delta != "" {
			onDelta(delta)
		}
		return current
	}
	onDelta(current)
	return current
}

package adkstage

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/adk"
	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type BaseModelFactory func(context.Context) (einomodel.BaseChatModel, error)

type Config struct {
	Name         string
	Description  string
	Instruction  string
	ModelFactory BaseModelFactory
	Handlers     []adk.ChatModelAgentMiddleware
	Middlewares  []adk.AgentMiddleware
	ToolsConfig  *adk.ToolsConfig
}

type Runner struct {
	cfg Config
}

func New(cfg Config) *Runner {
	return &Runner{cfg: cfg}
}

func (r *Runner) Run(ctx context.Context, input string) (string, error) {
	if r == nil {
		return "", fmt.Errorf("stage runner is nil")
	}
	if strings.TrimSpace(r.cfg.Name) == "" {
		return "", fmt.Errorf("stage name is required")
	}
	if r.cfg.ModelFactory == nil {
		return "", fmt.Errorf("stage model factory is required")
	}
	model, err := r.cfg.ModelFactory(ctx)
	if err != nil {
		return "", err
	}

	toolsConfig := adk.ToolsConfig{}
	if r.cfg.ToolsConfig != nil {
		toolsConfig = *r.cfg.ToolsConfig
	}
	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:          r.cfg.Name,
		Description:   firstNonEmpty(r.cfg.Description, r.cfg.Name),
		Instruction:   strings.TrimSpace(r.cfg.Instruction),
		Model:         model,
		ToolsConfig:   toolsConfig,
		MaxIterations: 1,
		Handlers:      append([]adk.ChatModelAgentMiddleware(nil), r.cfg.Handlers...),
		Middlewares:   append([]adk.AgentMiddleware(nil), r.cfg.Middlewares...),
	})
	if err != nil {
		return "", err
	}

	runner := adk.NewRunner(ctx, adk.RunnerConfig{Agent: agent})
	iter := runner.Query(ctx, input)

	var last string
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
			last = strings.TrimSpace(msg.Content)
		}
	}
	if last == "" {
		return "", fmt.Errorf("stage %s produced empty output", r.cfg.Name)
	}
	return last, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

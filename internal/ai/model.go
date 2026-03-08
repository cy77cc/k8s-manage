package ai

import (
	"context"
	"fmt"
	"strings"
	"time"

	ollamamodel "github.com/cloudwego/eino-ext/components/model/ollama"
	qwenmodel "github.com/cloudwego/eino-ext/components/model/qwen"
	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/cy77cc/k8s-manage/internal/config"
)

func NewToolCallingChatModel(ctx context.Context) (einomodel.ToolCallingChatModel, error) {
	if !config.CFG.LLM.Enable {
		return nil, fmt.Errorf("llm disabled")
	}
	switch strings.TrimSpace(strings.ToLower(config.CFG.LLM.Provider)) {
	case "ollama":
		return ollamamodel.NewChatModel(ctx, &ollamamodel.ChatModelConfig{
			BaseURL: config.CFG.LLM.BaseURL,
			Model:   config.CFG.LLM.Model,
			Timeout: 30 * time.Second,
		})
	case "qwen":
		thinking := true
		temp := float32(config.CFG.LLM.Temperature)
		return qwenmodel.NewChatModel(ctx, &qwenmodel.ChatModelConfig{
			APIKey:      config.CFG.LLM.APIKey,
			BaseURL:     config.CFG.LLM.BaseURL,
			Model:       config.CFG.LLM.Model,
			Temperature: &temp,
			Timeout:     30 * time.Second,
			EnableThinking: &thinking,
		})
	default:
		return nil, fmt.Errorf("unsupported llm provider %q", config.CFG.LLM.Provider)
	}
}

func CheckModelHealth(ctx context.Context, model einomodel.ToolCallingChatModel) error {
	if model == nil {
		return fmt.Errorf("chat model not initialized")
	}
	_, err := model.Generate(ctx, []*schema.Message{schema.UserMessage("ping")})
	return err
}

package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino-ext/components/model/qwen"
	"github.com/cloudwego/eino/components/model"
	"github.com/cy77cc/k8s-manage/internal/config"
)

func NewChatModel(ctx context.Context) (chatModel model.ToolCallingChatModel, err error) {
	if !config.CFG.LLM.Enable {
		return nil, nil
	}
	provider := strings.ToLower(strings.TrimSpace(config.CFG.LLM.Provider))

	switch provider {
	case "qwen":
		qwen.NewChatModel(ctx, &qwen.ChatModelConfig{
			BaseURL: config.CFG.LLM.BaseURL,
			Model:   config.CFG.LLM.Model,
			APIKey:  config.CFG.LLM.APIKey,
		})
	case "ollama":
		chatModel, err = ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
			BaseURL: config.CFG.LLM.BaseURL,
			Model:   config.CFG.LLM.Model,
			Options: &ollama.Options{
				Temperature: float32(config.CFG.LLM.Temperature),
				NumPredict:  1024,
			},
		})
	default:
		return nil, fmt.Errorf("unsupported llm provider %q", config.CFG.LLM.Provider)
	}

	if err != nil {
		return nil, err
	}
	return chatModel, nil
}

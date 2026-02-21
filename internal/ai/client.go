package ai

import (
	"context"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino/components/model"
	"github.com/cy77cc/k8s-manage/internal/config"
)

func NewChatModel(ctx context.Context) (model.ToolCallingChatModel, error) {
	if !config.CFG.LLM.Enable {
		return nil, nil
	}

	// Create OpenAI ChatModel
	// Note: API Key and BaseURL are required.
	// temp := float32(config.CFG.LLM.Temperature)
	chatModel, err := ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
		BaseURL:     config.CFG.LLM.BaseURL,
		Model:       config.CFG.LLM.Model,
	})
	if err != nil {
		return nil, err
	}
	return chatModel, nil
}

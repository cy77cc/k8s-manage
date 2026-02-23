package ai

import (
	"context"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino/components/model"
	"github.com/cy77cc/k8s-manage/internal/config"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func NewChatModel(ctx context.Context) (model.ToolCallingChatModel, error) {
	if !config.CFG.LLM.Enable {
		return nil, nil
	}

	// Create OpenAI ChatModel
	// Note: API Key and BaseURL are required.
	// temp := float32(config.CFG.LLM.Temperature)
	chatModel, err := ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
		BaseURL: config.CFG.LLM.BaseURL,
		Model:   config.CFG.LLM.Model,
	})
	if err != nil {
		return nil, err
	}
	return chatModel, nil
}

func NewMCPClient(ctx context.Context) (*client.Client, error) {
	cli, err := client.NewSSEMCPClient("http://localhost:12345/sse")
	if err != nil {
		return nil, err
	}
	err = cli.Start(ctx)
	if err != nil {
		return nil, err
	}

	cli.Initialize(ctx, mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			ClientInfo: mcp.Implementation{
				Name:    "mcp-tools",
				Version: "0.1.0",
			},
		},
	})
	return cli, nil
}

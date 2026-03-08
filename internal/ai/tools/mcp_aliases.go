package tools

import (
	"context"

	mcpimpl "github.com/cy77cc/k8s-manage/internal/ai/tools/impl/mcp"
)

type MCPConfig = mcpimpl.MCPConfig
type MCPToolInfo = mcpimpl.MCPToolInfo
type MCPClientManager = mcpimpl.MCPClientManager

func MCPConfigFromEnv() MCPConfig {
	return mcpimpl.MCPConfigFromEnv()
}

func NewMCPClientManager(ctx context.Context, cfg MCPConfig) (*MCPClientManager, error) {
	return mcpimpl.NewMCPClientManager(ctx, cfg)
}

func BuildMCPProxyTools(manager *MCPClientManager) ([]RegisteredTool, error) {
	return mcpimpl.BuildMCPProxyTools(manager)
}

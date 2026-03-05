package tools

import (
	"context"
	"encoding/json"

	"github.com/cloudwego/eino/components/tool"
)

// BuildAllTools returns all tools with risk-based wrappers applied.
func BuildAllTools(ctx context.Context, deps PlatformDeps) ([]tool.BaseTool, error) {
	registered, err := BuildRegisteredTools(deps)
	if err != nil {
		return nil, err
	}
	_ = ctx
	all := make([]tool.BaseTool, 0, len(registered))
	for _, item := range registered {
		all = append(all, WrapRegisteredTool(item))
	}
	return all, nil
}

// BuildRegisteredTools aggregates local tools.
func BuildRegisteredTools(deps PlatformDeps) ([]RegisteredTool, error) {
	return BuildLocalTools(deps)
}

// BuildRegisteredToolsWithMCP aggregates local tools and MCP proxy tools.
func BuildRegisteredToolsWithMCP(deps PlatformDeps, manager *MCPClientManager) ([]RegisteredTool, error) {
	localTools, err := BuildLocalTools(deps)
	if err != nil {
		return nil, err
	}
	mcpTools, err := BuildMCPProxyTools(manager)
	if err != nil {
		return nil, err
	}
	return append(localTools, mcpTools...), nil
}

func WrapRegisteredTool(item RegisteredTool) tool.BaseTool {
	switch item.Meta.Risk {
	case ToolRiskHigh:
		return NewApprovableTool(item.Tool, ToolRiskHigh, buildDefaultPreview(item.Meta.Name))
	case ToolRiskMedium:
		return NewReviewableTool(item.Tool)
	default:
		return item.Tool
	}
}

func buildDefaultPreview(toolName string) ApprovalPreviewFn {
	return func(_ context.Context, args string) (map[string]any, error) {
		preview := map[string]any{"tool": toolName}
		if args == "" {
			preview["arguments"] = map[string]any{}
			return preview, nil
		}
		var parsed map[string]any
		if err := json.Unmarshal([]byte(args), &parsed); err != nil {
			preview["arguments_raw"] = args
			preview["parse_error"] = err.Error()
			return preview, nil
		}
		preview["arguments"] = parsed
		return preview, nil
	}
}

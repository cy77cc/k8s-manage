package ai

import (
	"context"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/mark3labs/mcp-go/mcp"
)

func BuildMCPProxyTools(manager *MCPClientManager) ([]RegisteredTool, error) {
	if manager == nil {
		return nil, nil
	}
	list := manager.ListTools()
	out := make([]RegisteredTool, 0, len(list))
	for _, t := range list {
		meta := ToolMeta{
			Name:        "mcp.default." + t.Name,
			Description: t.Description,
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskMedium,
			Provider:    "mcp",
			Permission:  "ai:tool:read",
		}
		toolName := t.Name
		invoke := func(ctx context.Context, input map[string]any) (ToolResult, error) {
			start := time.Now()
			EmitToolEvent(ctx, "tool_call", map[string]any{"tool": meta.Name, "params": input})
			if err := CheckToolPolicy(ctx, meta, input); err != nil {
				return ToolResult{OK: false, Error: err.Error(), Source: "mcp", LatencyMS: time.Since(start).Milliseconds()}, err
			}
			res, err := manager.CallTool(ctx, toolName, input)
			if err != nil {
				result := ToolResult{OK: false, Error: err.Error(), Source: "mcp", LatencyMS: time.Since(start).Milliseconds()}
				EmitToolEvent(ctx, "tool_result", map[string]any{"tool": meta.Name, "result": result})
				return result, nil
			}
			texts := make([]string, 0, len(res.Content))
			for _, c := range res.Content {
				if tc, ok := mcp.AsTextContent(c); ok {
					texts = append(texts, tc.Text)
				}
			}
			result := ToolResult{OK: !res.IsError, Data: map[string]any{"text": strings.Join(texts, "\n"), "raw": res}, Source: "mcp", LatencyMS: time.Since(start).Milliseconds()}
			if res.IsError {
				result.Error = strings.Join(texts, "\n")
			}
			EmitToolEvent(ctx, "tool_result", map[string]any{"tool": meta.Name, "result": result})
			return result, nil
		}
		tl, err := utils.InferTool[map[string]any, ToolResult](meta.Name, meta.Description, invoke)
		if err != nil {
			return nil, err
		}
		out = append(out, RegisteredTool{Meta: meta, Tool: tl})
	}
	return out, nil
}

package tools

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
			Schema:      t.Schema,
			Required:    t.Required,
		}
		toolName := t.Name
		invoke := func(ctx context.Context, input map[string]any) (ToolResult, error) {
			start := time.Now()
			callID := nextToolCallID()
			EmitToolEvent(ctx, "tool_call", map[string]any{"tool": meta.Name, "call_id": callID, "params": input})

			if err := CheckToolPolicy(ctx, meta, input); err != nil {
				res := ToolResult{OK: false, ErrorCode: "policy_denied", Error: err.Error(), Source: "mcp", LatencyMS: time.Since(start).Milliseconds()}
				EmitToolEvent(ctx, "tool_result", map[string]any{"tool": meta.Name, "call_id": callID, "result": res})
				return res, err
			}

			callCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
			defer cancel()
			res, err := manager.CallTool(callCtx, toolName, input)
			if err != nil {
				code := "tool_error"
				if callCtx.Err() == context.DeadlineExceeded {
					code = "tool_timeout"
				}
				result := ToolResult{OK: false, ErrorCode: code, Error: err.Error(), Source: "mcp", LatencyMS: time.Since(start).Milliseconds()}
				EmitToolEvent(ctx, "tool_result", map[string]any{"tool": meta.Name, "call_id": callID, "result": result})
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
				result.ErrorCode = "tool_error"
				result.Error = strings.Join(texts, "\n")
			}
			EmitToolEvent(ctx, "tool_result", map[string]any{"tool": meta.Name, "call_id": callID, "result": result})
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

package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestBuildMCPProxyToolsUsesManagerPrefix(t *testing.T) {
	manager := &MCPClientManager{
		prefix: "mcp_default",
		tools: []MCPToolInfo{
			{
				Name:        "mcp_default_search_docs",
				RemoteName:  "search.docs",
				Description: "search docs",
				Schema:      map[string]any{"type": "object"},
				Required:    []string{"query"},
			},
		},
		remoteIndex: map[string]string{
			"mcp_default_search_docs": "search.docs",
		},
	}

	tools, err := BuildMCPProxyTools(manager)
	if err != nil {
		t.Fatalf("BuildMCPProxyTools failed: %v", err)
	}
	if len(tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(tools))
	}
	if tools[0].Meta.Name != "mcp_default_search_docs" {
		t.Fatalf("unexpected tool name: %s", tools[0].Meta.Name)
	}
	if tools[0].Meta.Provider != "mcp" {
		t.Fatalf("unexpected provider: %s", tools[0].Meta.Provider)
	}
}

func TestBuildMCPProxyToolsInvokesRemoteToolByMappedName(t *testing.T) {
	var gotTool string
	var gotArgs map[string]any
	manager := &MCPClientManager{
		prefix: "mcp_default",
		tools: []MCPToolInfo{
			{
				Name:        "mcp_default_search_docs",
				RemoteName:  "search.docs",
				Description: "search docs",
				Schema:      map[string]any{"type": "object"},
				Required:    []string{"query"},
			},
		},
		remoteIndex: map[string]string{
			"mcp_default_search_docs": "search.docs",
		},
		callToolFn: func(_ context.Context, toolName string, args map[string]any) (*mcp.CallToolResult, error) {
			gotTool = toolName
			gotArgs = args
			return mcp.NewToolResultText("matched docs"), nil
		},
	}

	items, err := BuildMCPProxyTools(manager)
	if err != nil {
		t.Fatalf("BuildMCPProxyTools failed: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(items))
	}

	out, err := items[0].Tool.InvokableRun(context.Background(), `{"query":"k8s"}`)
	if err != nil {
		t.Fatalf("InvokableRun failed: %v", err)
	}
	if gotTool != "search.docs" {
		t.Fatalf("expected proxy to call remote tool name, got %s", gotTool)
	}
	if gotArgs["query"] != "k8s" {
		t.Fatalf("unexpected args: %#v", gotArgs)
	}

	var result ToolResult
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if !result.OK {
		t.Fatalf("expected ok result, got %#v", result)
	}
	data, ok := result.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected map data, got %T", result.Data)
	}
	if data["text"] != "matched docs" {
		t.Fatalf("unexpected text payload: %#v", data["text"])
	}
}

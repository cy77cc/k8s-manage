package tools

import "testing"

func TestMCPClientManagerPrefixAndMapping(t *testing.T) {
	manager := &MCPClientManager{
		prefix:      sanitizeMCPPrefix("mcp.default"),
		remoteIndex: map[string]string{},
	}

	toolName := manager.ToolNameForRemote("search.docs")
	if toolName != "mcp_default_search_docs" {
		t.Fatalf("unexpected tool name: %s", toolName)
	}

	manager.remoteIndex[toolName] = "search.docs"
	if got := manager.RemoteNameForTool(toolName); got != "search.docs" {
		t.Fatalf("unexpected remote name: %s", got)
	}
}

package tools

import (
	"context"
	"encoding/json"
	"testing"
)

func TestBuildCategorySelectors(t *testing.T) {
	all, err := BuildLocalTools(PlatformDeps{})
	if err != nil {
		t.Fatalf("BuildLocalTools failed: %v", err)
	}

	if len(buildOpsTools(all)) == 0 {
		t.Fatalf("expected ops tools")
	}
	if len(buildK8sTools(all)) == 0 {
		t.Fatalf("expected k8s tools")
	}
	if len(buildServiceTools(all)) == 0 {
		t.Fatalf("expected service tools")
	}
	if len(buildDeploymentTools(all)) == 0 {
		t.Fatalf("expected deployment tools")
	}
	if len(buildCICDTools(all)) == 0 {
		t.Fatalf("expected cicd tools")
	}
	if len(buildMonitorTools(all)) == 0 {
		t.Fatalf("expected monitor tools")
	}
	if len(buildConfigTools(all)) == 0 {
		t.Fatalf("expected config tools")
	}
	if len(buildGovernanceTools(all)) == 0 {
		t.Fatalf("expected governance tools")
	}
	if len(buildInventoryTools(all)) == 0 {
		t.Fatalf("expected inventory tools")
	}
}

func TestBuildRegisteredToolsVariants(t *testing.T) {
	registered, err := BuildRegisteredTools(PlatformDeps{})
	if err != nil {
		t.Fatalf("BuildRegisteredTools failed: %v", err)
	}
	if len(registered) == 0 {
		t.Fatalf("expected registered tools")
	}

	allTools, err := BuildAllTools(context.Background(), PlatformDeps{})
	if err != nil {
		t.Fatalf("BuildAllTools failed: %v", err)
	}
	if len(allTools) != len(registered) {
		t.Fatalf("expected wrapped tool count to match registered count, got %d vs %d", len(allTools), len(registered))
	}

	manager := &MCPClientManager{
		prefix:      "mcp_default",
		tools:       []MCPToolInfo{{Name: "mcp_default_search_docs", RemoteName: "search.docs", Description: "Search docs"}},
		remoteIndex: map[string]string{"mcp_default_search_docs": "search.docs"},
	}
	withMCP, err := BuildRegisteredToolsWithMCP(PlatformDeps{}, manager)
	if err != nil {
		t.Fatalf("BuildRegisteredToolsWithMCP failed: %v", err)
	}
	if len(withMCP) <= len(registered) {
		t.Fatalf("expected MCP tools to extend registered tools")
	}
}

func TestWrapRegisteredToolAndDefaultPreview(t *testing.T) {
	registered, err := BuildRegisteredTools(PlatformDeps{})
	if err != nil {
		t.Fatalf("BuildRegisteredTools failed: %v", err)
	}
	if len(registered) == 0 {
		t.Fatalf("expected registered tools")
	}
	readonly := registered[0]
	readonly.Meta.Risk = ToolRiskLow
	if wrapped := WrapRegisteredTool(readonly); wrapped != readonly.Tool {
		t.Fatalf("expected low risk tool to remain unwrapped")
	}

	highRisk := RegisteredTool{
		Meta: ToolMeta{Name: "host_batch", Risk: ToolRiskHigh},
		Tool: readonly.Tool,
	}
	if _, ok := WrapRegisteredTool(highRisk).(*ApprovableTool); !ok {
		t.Fatalf("expected high risk tool to use approvable wrapper")
	}

	mediumRisk := RegisteredTool{
		Meta: ToolMeta{Name: "service_deploy", Risk: ToolRiskMedium},
		Tool: readonly.Tool,
	}
	if _, ok := WrapRegisteredTool(mediumRisk).(*ReviewableTool); !ok {
		t.Fatalf("expected medium risk tool to use review wrapper")
	}

	preview := buildDefaultPreview("host_batch")
	noArgs, err := preview(context.Background(), "")
	if err != nil {
		t.Fatalf("preview without args: %v", err)
	}
	if noArgs["tool"] != "host_batch" {
		t.Fatalf("unexpected preview tool name: %#v", noArgs)
	}

	withArgs, err := preview(context.Background(), `{"target":"localhost"}`)
	if err != nil {
		t.Fatalf("preview with json args: %v", err)
	}
	args, ok := withArgs["arguments"].(map[string]any)
	if !ok || args["target"] != "localhost" {
		t.Fatalf("unexpected parsed preview args: %#v", withArgs)
	}

	rawArgs, err := preview(context.Background(), "{not-json}")
	if err != nil {
		t.Fatalf("preview with raw args: %v", err)
	}
	if rawArgs["arguments_raw"] != "{not-json}" {
		t.Fatalf("expected raw args fallback: %#v", rawArgs)
	}
	if _, ok := rawArgs["parse_error"].(string); !ok {
		t.Fatalf("expected parse error string: %#v", rawArgs)
	}

	encoded, err := json.Marshal(withArgs)
	if err != nil || len(encoded) == 0 {
		t.Fatalf("expected preview payload to marshal: %v", err)
	}
}

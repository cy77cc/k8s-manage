package ai

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	aitools "github.com/cy77cc/k8s-manage/internal/ai/tools"
)

func TestNewToolRegistryBuildsDefinitions(t *testing.T) {
	registry := NewToolRegistry([]aitools.RegisteredTool{
		{
			Meta: aitools.ToolMeta{
				Name:        "k8s_list_resources",
				Description: "list k8s resources",
				Risk:        aitools.ToolRiskLow,
				Mode:        aitools.ToolModeReadonly,
				Schema:      map[string]any{"type": "object"},
				Required:    []string{"resource"},
			},
			Tool: &fakeRegistryTool{name: "k8s_list_resources"},
		},
		{
			Meta: aitools.ToolMeta{
				Name:        "mcp_default_search",
				Description: "search via mcp",
				Provider:    "mcp",
				Risk:        aitools.ToolRiskMedium,
			},
			Tool: &fakeRegistryTool{name: "mcp_default_search"},
		},
	})

	if got := len(registry.List()); got != 2 {
		t.Fatalf("expected 2 tools, got %d", got)
	}

	k8sDef, ok := registry.Get("k8s.list_resources")
	if !ok {
		t.Fatal("expected normalized k8s tool")
	}
	if k8sDef.Category != CategoryK8s {
		t.Fatalf("expected k8s category, got %s", k8sDef.Category)
	}
	if k8sDef.Schema["type"] != "object" {
		t.Fatalf("expected schema to be copied")
	}

	mcpDef, ok := registry.Get("mcp.default.search")
	if !ok {
		t.Fatal("expected mcp tool")
	}
	if mcpDef.Category != CategoryMCP {
		t.Fatalf("expected mcp category, got %s", mcpDef.Category)
	}
}

func TestToolRegistryByCategory(t *testing.T) {
	registry := NewToolRegistry([]aitools.RegisteredTool{
		{Meta: aitools.ToolMeta{Name: "host_ssh_exec_readonly"}, Tool: &fakeRegistryTool{name: "host_ssh_exec_readonly"}},
		{Meta: aitools.ToolMeta{Name: "os_get_cpu_mem"}, Tool: &fakeRegistryTool{name: "os_get_cpu_mem"}},
		{Meta: aitools.ToolMeta{Name: "monitor_metric_query"}, Tool: &fakeRegistryTool{name: "monitor_metric_query"}},
	})

	hostTools := registry.ByCategory(CategoryHost)
	if got := len(hostTools); got != 2 {
		t.Fatalf("expected 2 host tools, got %d", got)
	}
	if hostTools[0].Name != "host_ssh_exec_readonly" || hostTools[1].Name != "os_get_cpu_mem" {
		t.Fatalf("unexpected host tools order: %#v", hostTools)
	}
}

func TestToolRegistryMapsExposeWrappedInputs(t *testing.T) {
	registry := NewToolRegistry([]aitools.RegisteredTool{
		{
			Meta: aitools.ToolMeta{
				Name:        "service_deploy_apply",
				Description: "deploy service",
				Risk:        aitools.ToolRiskHigh,
				Mode:        aitools.ToolModeMutating,
			},
			Tool: &fakeRegistryTool{name: "service_deploy_apply"},
		},
	})

	baseTools := registry.BaseTools()
	if len(baseTools) != 1 {
		t.Fatalf("expected one base tool, got %d", len(baseTools))
	}

	if _, ok := registry.ToolMap()["service_deploy_apply"]; !ok {
		t.Fatal("expected tool map to include service_deploy_apply")
	}
	meta, ok := registry.MetaMap()["service_deploy_apply"]
	if !ok {
		t.Fatal("expected meta map to include service_deploy_apply")
	}
	if meta.Risk != aitools.ToolRiskHigh {
		t.Fatalf("expected high risk meta, got %s", meta.Risk)
	}
}

type fakeRegistryTool struct {
	name string
}

func (f *fakeRegistryTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{Name: f.name}, nil
}

func (f *fakeRegistryTool) InvokableRun(_ context.Context, _ string, _ ...tool.Option) (string, error) {
	return `{"ok":true}`, nil
}

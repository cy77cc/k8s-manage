package experts

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cloudwego/eino/components/tool"
	toolutils "github.com/cloudwego/eino/components/tool/utils"
)

type noopInput struct {
	Value string `json:"value,omitempty"`
}

func mustTool(t *testing.T, name string) tool.InvokableTool {
	t.Helper()
	info, err := toolutils.GoStruct2ToolInfo[noopInput](name, "test tool")
	if err != nil {
		t.Fatalf("build tool info: %v", err)
	}
	return toolutils.NewTool(info, func(ctx context.Context, input noopInput) (string, error) {
		return "ok", nil
	})
}

func TestNewExpertRegistryLoadAndMatch(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "experts.yaml")
	raw := `version: "1.0"
experts:
  - name: host_expert
    display_name: "host"
    persona: "host persona"
    tool_patterns: ["host_*", "credential_*"]
    domains:
      - name: host_management
        weight: 0.9
    keywords: ["主机", "host"]
    capabilities: ["host diagnosis"]
    risk_level: medium
  - name: service_expert
    display_name: "service"
    persona: "service persona"
    tool_patterns: ["service_*"]
    domains:
      - name: service_management
        weight: 0.95
    keywords: ["服务", "service"]
    capabilities: ["service diagnosis"]
    risk_level: low
`
	if err := os.WriteFile(cfgPath, []byte(raw), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	tools := map[string]tool.InvokableTool{
		"host_list_inventory": mustTool(t, "host_list_inventory"),
		"credential_list":     mustTool(t, "credential_list"),
		"service_get_detail":  mustTool(t, "service_get_detail"),
	}
	reg, err := NewExpertRegistry(context.Background(), cfgPath, tools, nil)
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}

	host, ok := reg.GetExpert("host_expert")
	if !ok || host == nil {
		t.Fatalf("expected host expert")
	}
	if len(host.Tools) < 2 {
		t.Fatalf("expected host tools>=2, got=%d", len(host.Tools))
	}
	if _, ok := host.Tools["host_list_inventory"]; !ok {
		t.Fatalf("expected host_list_inventory in host tools")
	}
	if _, ok := host.Tools["credential_list"]; !ok {
		t.Fatalf("expected credential_list in host tools")
	}
	helperTool, ok := host.Tools["service_expert"]
	if !ok {
		t.Fatalf("expected helper expert tool service_expert in host tools")
	}
	toolInput, _ := json.Marshal(ExpertToolInput{Task: "请协助分析服务状态"})
	helperOut, err := helperTool.InvokableRun(context.Background(), string(toolInput))
	if err != nil {
		t.Fatalf("invoke helper expert tool: %v", err)
	}
	if helperOut == "" {
		t.Fatalf("expected helper output")
	}

	matches := reg.MatchByKeywords("请检查主机 CPU")
	if len(matches) == 0 || matches[0].Expert == nil || matches[0].Expert.Name != "host_expert" {
		t.Fatalf("expected host_expert keyword match")
	}

	domain := reg.MatchByDomain("service_management")
	if len(domain) == 0 || domain[0].Expert == nil || domain[0].Expert.Name != "service_expert" {
		t.Fatalf("expected service_expert domain match")
	}
}

func TestExpertNestedCallPathViaHelperTool(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "experts.yaml")
	raw := `version: "1.0"
experts:
  - name: primary_expert
    display_name: "primary"
    persona: "primary persona"
    tool_patterns: ["host_*"]
    keywords: ["primary"]
    capabilities: ["primary"]
    risk_level: low
  - name: helper_expert
    display_name: "helper"
    persona: "helper persona"
    tool_patterns: ["service_*"]
    keywords: ["helper"]
    capabilities: ["helper"]
    risk_level: low
`
	if err := os.WriteFile(cfgPath, []byte(raw), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	tools := map[string]tool.InvokableTool{
		"host_list_inventory": mustTool(t, "host_list_inventory"),
		"service_get_detail":  mustTool(t, "service_get_detail"),
	}
	reg, err := NewExpertRegistry(context.Background(), cfgPath, tools, nil)
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	primary, ok := reg.GetExpert("primary_expert")
	if !ok || primary == nil {
		t.Fatalf("expected primary_expert")
	}

	helperTool, ok := primary.Tools["helper_expert"]
	if !ok {
		t.Fatalf("expected helper_expert tool in primary tools")
	}
	payload, _ := json.Marshal(ExpertToolInput{Task: "检查服务依赖"})
	out, err := helperTool.InvokableRun(context.Background(), string(payload))
	if err != nil {
		t.Fatalf("invoke helper_expert tool: %v", err)
	}
	if !strings.Contains(out, "helper_expert") {
		t.Fatalf("unexpected helper tool output: %s", out)
	}
}

func TestNewExpertRegistryFallbackDefaults(t *testing.T) {
	reg, err := NewExpertRegistry(context.Background(), "/non/exist/experts.yaml", map[string]tool.InvokableTool{}, nil)
	if err != nil {
		t.Fatalf("new registry fallback: %v", err)
	}
	list := reg.ListExperts()
	if len(list) != 1 || list[0] == nil || list[0].Name != "general_expert" {
		t.Fatalf("expected default general_expert fallback")
	}
}

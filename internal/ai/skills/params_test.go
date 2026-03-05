package skills

import (
	"strings"
	"testing"
)

type fakeEnumProvider struct {
	data map[string][]string
}

func (f fakeEnumProvider) ListValues(source string) ([]string, error) {
	return f.data[source], nil
}

func TestValidateParams(t *testing.T) {
	defs := []SkillParameter{
		{Name: "service_id", Type: "int", Required: true},
		{Name: "env", Type: "string", Default: "staging"},
		{Name: "host_ids", Type: "array<int>"},
	}
	params, err := validateParams(defs, map[string]any{"service_id": "42", "host_ids": "1,2,3"})
	if err != nil {
		t.Fatalf("validate params: %v", err)
	}
	if params["service_id"].(int) != 42 {
		t.Fatalf("unexpected service_id: %#v", params["service_id"])
	}
	if params["env"].(string) != "staging" {
		t.Fatalf("expected default env")
	}
	if len(params["host_ids"].([]int)) != 3 {
		t.Fatalf("expected 3 host ids")
	}
}

func TestExtractParamsWithEnumSource(t *testing.T) {
	defs := []SkillParameter{
		{Name: "cluster", Type: "string", Required: true, EnumSource: "clusters"},
		{Name: "service_id", Type: "int", Required: true},
	}
	params, err := extractParams("deploy service_id=12 cluster=prod", defs, fakeEnumProvider{data: map[string][]string{"clusters": {"prod-cluster", "staging-cluster"}}})
	if err != nil {
		t.Fatalf("extract params: %v", err)
	}
	if params["cluster"].(string) != "prod-cluster" {
		t.Fatalf("expected enum resolved cluster, got %#v", params["cluster"])
	}
}

func TestRenderParamsTemplate(t *testing.T) {
	tpl := map[string]any{
		"service_id": "{{params.service_id}}",
		"env":        "{{params.env}}",
		"summary":    "service={{params.service_id}} status={{steps.preview.status}}",
	}
	rendered, err := renderParamsTemplate(tpl, map[string]any{"service_id": 100, "env": "prod"}, map[string]any{"preview": map[string]any{"status": "ok"}})
	if err != nil {
		t.Fatalf("render params template: %v", err)
	}
	if rendered["service_id"].(int) != 100 {
		t.Fatalf("expected integer substitution")
	}
	if !strings.Contains(rendered["summary"].(string), "status=ok") {
		t.Fatalf("unexpected summary template output: %s", rendered["summary"].(string))
	}
}

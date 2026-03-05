package experts

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestBuildExpertToolsExcludeSelf(t *testing.T) {
	expertsMap := map[string]*Expert{
		"primary": {Name: "primary", Persona: "primary persona"},
		"helper":  {Name: "helper", Persona: "helper persona"},
	}
	tools := BuildExpertTools(expertsMap, "primary")
	if _, ok := tools["primary"]; ok {
		t.Fatalf("self expert tool should be excluded")
	}
	if _, ok := tools["helper"]; !ok {
		t.Fatalf("helper expert tool should be present")
	}
}

func TestBuildExpertToolFallbackWhenAgentNil(t *testing.T) {
	tool, err := BuildExpertTool(&Expert{Name: "k8s_expert", Persona: "k8s"})
	if err != nil {
		t.Fatalf("build expert tool: %v", err)
	}
	in, _ := json.Marshal(ExpertToolInput{Task: "check pods"})
	out, err := tool.InvokableRun(context.Background(), string(in))
	if err != nil {
		t.Fatalf("invoke expert tool: %v", err)
	}
	if !strings.Contains(out, "k8s_expert") {
		t.Fatalf("unexpected output: %s", out)
	}
}

package aiv2

import (
	"context"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	expertspkg "github.com/cy77cc/OpsPilot/internal/ai/experts"
	expertspec "github.com/cy77cc/OpsPilot/internal/ai/experts/spec"
	"github.com/cy77cc/OpsPilot/internal/ai/tools/common"
)

type ToolRegistry struct {
	tools    []tool.BaseTool
	policies map[string]ToolPolicy
}

func NewToolRegistry(ctx context.Context, deps common.PlatformDeps, sessionID, turnID string, runtimeContext map[string]any) (*ToolRegistry, error) {
	registry := expertspkg.DefaultRegistry(deps)
	policies := make(map[string]ToolPolicy)
	items := make([]tool.BaseTool, 0, 32)
	seen := make(map[string]struct{})

	for _, exp := range registry.List() {
		for _, capability := range exp.Capabilities() {
			policies[strings.TrimSpace(capability.Name)] = ToolPolicy{
				Name:             strings.TrimSpace(capability.Name),
				Expert:           exp.Name(),
				Mode:             normalizeMode(capability.Mode),
				Risk:             normalizeRisk(capability.Risk),
				ApprovalRequired: normalizeMode(capability.Mode) == "mutating",
			}
		}
		for _, invokable := range exp.Tools(ctx) {
			if invokable == nil {
				continue
			}
			info, err := invokable.Info(ctx)
			if err != nil || info == nil || strings.TrimSpace(info.Name) == "" {
				continue
			}
			name := strings.TrimSpace(info.Name)
			if _, ok := seen[name]; ok {
				continue
			}
			policy := policies[name]
			if policy.Name == "" {
				policy = ToolPolicy{Name: name, Expert: exp.Name(), Mode: "readonly", Risk: "low"}
			}
			items = append(items, wrapApprovalTool(invokable, policy, sessionID, turnID, runtimeContext))
			seen[name] = struct{}{}
		}
	}

	return &ToolRegistry{tools: items, policies: policies}, nil
}

func (r *ToolRegistry) Tools() []tool.BaseTool {
	if r == nil {
		return nil
	}
	return append([]tool.BaseTool(nil), r.tools...)
}

func (r *ToolRegistry) Policy(name string) ToolPolicy {
	if r == nil {
		return ToolPolicy{}
	}
	return r.policies[strings.TrimSpace(name)]
}

func policyByCapabilities(exp expertspec.Expert) map[string]ToolPolicy {
	out := make(map[string]ToolPolicy)
	if exp == nil {
		return out
	}
	for _, capability := range exp.Capabilities() {
		name := strings.TrimSpace(capability.Name)
		if name == "" {
			continue
		}
		out[name] = ToolPolicy{
			Name:             name,
			Expert:           exp.Name(),
			Mode:             normalizeMode(capability.Mode),
			Risk:             normalizeRisk(capability.Risk),
			ApprovalRequired: normalizeMode(capability.Mode) == "mutating",
		}
	}
	return out
}

func normalizeMode(mode string) string {
	switch strings.TrimSpace(strings.ToLower(mode)) {
	case "mutating":
		return "mutating"
	default:
		return "readonly"
	}
}

func normalizeRisk(risk string) string {
	switch strings.TrimSpace(strings.ToLower(risk)) {
	case "high", "medium", "low":
		return strings.TrimSpace(strings.ToLower(risk))
	default:
		return "low"
	}
}


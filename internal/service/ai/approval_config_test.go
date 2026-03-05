package ai

import (
	"path/filepath"
	"testing"

	"github.com/cy77cc/k8s-manage/internal/ai/tools"
)

func TestLoadApprovalConfig(t *testing.T) {
	cfg, err := LoadApprovalConfig(filepath.Join("..", "..", "..", "configs", "approval_config.yaml"))
	if err != nil {
		t.Fatalf("load approval config: %v", err)
	}
	if cfg.Defaults.ApprovalTimeoutSeconds <= 0 {
		t.Fatalf("invalid defaults approval timeout")
	}
}

func TestApprovalConfigDecideWithToolOverride(t *testing.T) {
	cfg, err := LoadApprovalConfig(filepath.Join("..", "..", "..", "configs", "approval_config.yaml"))
	if err != nil {
		t.Fatalf("load approval config: %v", err)
	}
	out := cfg.Decide(tools.ToolMeta{
		Name: "service_deploy_apply",
		Risk: tools.ToolRiskLow,
	})
	if out.RiskLevel != "high" {
		t.Fatalf("expected high risk from tool override, got %s", out.RiskLevel)
	}
	if out.ApprovalTimeout <= 0 {
		t.Fatalf("approval timeout should be configured")
	}
}

func TestApprovalConfigValidateRejectsInvalidRisk(t *testing.T) {
	cfg := &ApprovalConfig{
		Defaults: ApprovalDefaultConfig{
			ConfirmationTimeoutSeconds: 300,
			ApprovalTimeoutSeconds:     1800,
			RiskLevel:                  "critical",
		},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected validation error for invalid risk level")
	}
}

package tools

import (
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

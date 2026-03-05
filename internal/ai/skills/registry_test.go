package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSkillsAndMatch(t *testing.T) {
	configPath := testSkillsConfigPath(t)
	cfg, err := LoadSkills(configPath)
	if err != nil {
		t.Fatalf("load skills config: %v", err)
	}
	if len(cfg.Skills) < 3 {
		t.Fatalf("expected at least 3 skills, got %d", len(cfg.Skills))
	}

	r, err := NewRegistry(configPath)
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	matched, score := r.MatchSkill("帮我部署服务到生产环境")
	if matched == nil {
		t.Fatalf("expected matched skill")
	}
	if matched.Name != "deploy_service" {
		t.Fatalf("expected deploy_service, got %s", matched.Name)
	}
	if score <= 0 {
		t.Fatalf("expected positive score")
	}
}

func TestSkillConfigValidate(t *testing.T) {
	cfg := &SkillConfig{
		Version: "1.0",
		Skills: []Skill{
			{
				Name:            "test",
				TriggerPatterns: []string{"test"},
				Parameters:      []SkillParameter{{Name: "id", Type: "int", Required: true}},
				Steps:           []SkillStep{{Name: "run", Type: "tool", Tool: "service.status"}},
			},
		},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("validate config: %v", err)
	}

	cfg.Skills[0].Steps[0].Type = "unknown"
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected invalid step type error")
	}
}

func TestRegistryReload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "skills.yaml")
	content := `version: "1.0"
skills:
  - name: quick_check
    trigger_patterns: ["检查"]
    parameters:
      - name: host_id
        type: int
        required: true
    steps:
      - name: query
        type: tool
        tool: host.status
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp config: %v", err)
	}
	r, err := NewRegistry(path)
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	if _, ok := r.Get("quick_check"); !ok {
		t.Fatalf("expected quick_check skill")
	}
}

func testSkillsConfigPath(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	return filepath.Clean(filepath.Join(wd, "..", "..", "..", "configs", "skills.yaml"))
}

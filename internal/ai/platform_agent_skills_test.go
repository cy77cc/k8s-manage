package ai

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cloudwego/eino/schema"
	askills "github.com/cy77cc/k8s-manage/internal/ai/skills"
	"github.com/cy77cc/k8s-manage/internal/ai/tools"
)

func TestExecuteSkillStream(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "skills.yaml")
	content := `version: "1.0"
skills:
  - name: deploy_service
    trigger_patterns: ["部署服务", "deploy"]
    parameters:
      - name: service_id
        type: int
        required: true
    steps:
      - name: run
        type: tool
        tool: deployment.release.apply
        params_template:
          service_id: "{{params.service_id}}"
`
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write skills config: %v", err)
	}
	reg, err := askills.NewRegistry(configPath)
	if err != nil {
		t.Fatalf("new skills registry: %v", err)
	}
	exec := askills.NewExecutor(func(_ context.Context, toolName string, params map[string]any) (tools.ToolResult, error) {
		if toolName != "deployment.release.apply" {
			t.Fatalf("unexpected tool name: %s", toolName)
		}
		if params["service_id"].(int) != 9 {
			t.Fatalf("unexpected service_id: %#v", params["service_id"])
		}
		return tools.ToolResult{OK: true, Data: "ok"}, nil
	}, nil, nil)

	agent := &PlatformAgent{skillRegistry: reg, skillExecutor: exec}
	stream, handled := agent.executeSkillStream(context.Background(), []*schema.Message{schema.UserMessage("请帮我deploy service_id=9")})
	if !handled {
		t.Fatalf("expected skill stream to be handled")
	}
	if stream == nil {
		t.Fatalf("expected non-nil stream")
	}
	defer stream.Close()
	msg, err := stream.Recv()
	if err != nil {
		t.Fatalf("recv stream message: %v", err)
	}
	if !strings.Contains(msg.Content, "技能 `deploy_service` 执行完成") {
		t.Fatalf("unexpected skill result output: %s", msg.Content)
	}
	_, err = stream.Recv()
	if err != io.EOF {
		t.Fatalf("expected EOF, got %v", err)
	}
}

func TestExecuteSkillStreamNoMatch(t *testing.T) {
	agent := &PlatformAgent{}
	stream, handled := agent.executeSkillStream(context.Background(), []*schema.Message{schema.UserMessage("hello")})
	if handled || stream != nil {
		t.Fatalf("expected no skill handling")
	}
}

package logic

import (
	"strings"
	"testing"

	v1 "github.com/cy77cc/k8s-manage/api/project/v1"
	"github.com/cy77cc/k8s-manage/internal/svc"
)

func TestServiceLogic_generateK8sYAML(t *testing.T) {
	l := &ServiceLogic{svcCtx: &svc.ServiceContext{}} // Mock context not needed for this test

	req := v1.CreateServiceReq{
		ProjectID:     1,
		Name:          "test-app",
		Type:          "stateless",
		Image:         "nginx:latest",
		Replicas:      3,
		ServicePort:   80,
		ContainerPort: 8080,
		NodePort:      30080,
		EnvVars: []v1.EnvVar{
			{Key: "ENV_KEY", Value: "ENV_VAL"},
		},
		Resources: &v1.ResourceReq{
			Limits:   map[string]string{"cpu": "500m", "memory": "512Mi"},
			Requests: map[string]string{"cpu": "250m", "memory": "256Mi"},
		},
	}

	yamlContent, err := l.generateK8sYAML(req)
	if err != nil {
		t.Fatalf("generateK8sYAML failed: %v", err)
	}

	if !strings.Contains(yamlContent, "kind: Deployment") {
		t.Error("YAML should contain Deployment")
	}
	if !strings.Contains(yamlContent, "image: nginx:latest") {
		t.Error("YAML should contain image")
	}
	if !strings.Contains(yamlContent, "kind: Service") {
		t.Error("YAML should contain Service")
	}
	if !strings.Contains(yamlContent, "nodePort: 30080") {
		t.Error("YAML should contain nodePort")
	}
	if !strings.Contains(yamlContent, "ENV_KEY") {
		t.Error("YAML should contain ENV_KEY")
	}
	if !strings.Contains(yamlContent, "500m") {
		t.Error("YAML should contain cpu limit")
	}

	t.Logf("Generated YAML:\n%s", yamlContent)
}

package service

import (
	"strings"
	"testing"
)

func TestRenderFromStandard_K8s(t *testing.T) {
	cfg := &StandardServiceConfig{
		Image:    "nginx:1.27",
		Replicas: 2,
		Ports: []PortConfig{{
			Name:          "http",
			Protocol:      "TCP",
			ContainerPort: 8080,
			ServicePort:   80,
		}},
		Envs: []EnvKV{{Key: "APP_ENV", Value: "staging"}},
	}

	resp, err := renderFromStandard("demo-svc", "stateless", "k8s", cfg)
	if err != nil {
		t.Fatalf("render k8s failed: %v", err)
	}
	if !strings.Contains(resp.RenderedYAML, "kind: Deployment") {
		t.Fatalf("expected deployment yaml, got: %s", resp.RenderedYAML)
	}
	if !strings.Contains(resp.RenderedYAML, "kind: Service") {
		t.Fatalf("expected service yaml, got: %s", resp.RenderedYAML)
	}
}

func TestRenderFromStandard_Compose(t *testing.T) {
	cfg := &StandardServiceConfig{
		Image:    "nginx:1.27",
		Replicas: 1,
		Ports: []PortConfig{{
			Name:          "http",
			Protocol:      "TCP",
			ContainerPort: 8080,
			ServicePort:   80,
		}},
	}

	resp, err := renderFromStandard("demo-svc", "stateless", "compose", cfg)
	if err != nil {
		t.Fatalf("render compose failed: %v", err)
	}
	if !strings.Contains(resp.RenderedYAML, "services:") {
		t.Fatalf("expected compose services, got: %s", resp.RenderedYAML)
	}
	if !strings.Contains(resp.RenderedYAML, "80:8080") {
		t.Fatalf("expected ports mapping, got: %s", resp.RenderedYAML)
	}
}

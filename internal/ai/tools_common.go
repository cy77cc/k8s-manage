package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	sshclient "github.com/cy77cc/k8s-manage/internal/client/ssh"
	"github.com/cy77cc/k8s-manage/internal/model"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func runWithPolicyAndEvent(ctx context.Context, meta ToolMeta, input map[string]any, do func() (any, string, error)) (ToolResult, error) {
	start := time.Now()
	EmitToolEvent(ctx, "tool_call", map[string]any{"tool": meta.Name, "params": input})
	if err := CheckToolPolicy(ctx, meta, input); err != nil {
		res := ToolResult{OK: false, Error: err.Error(), Source: meta.Provider, LatencyMS: time.Since(start).Milliseconds()}
		EmitToolEvent(ctx, "tool_result", map[string]any{"tool": meta.Name, "result": res})
		return res, err
	}
	data, source, err := do()
	if source == "" {
		source = meta.Provider
	}
	if err != nil {
		res := ToolResult{OK: false, Error: err.Error(), Source: source, LatencyMS: time.Since(start).Milliseconds()}
		EmitToolEvent(ctx, "tool_result", map[string]any{"tool": meta.Name, "result": res})
		return res, nil
	}
	res := ToolResult{OK: true, Data: data, Source: source, LatencyMS: time.Since(start).Milliseconds()}
	EmitToolEvent(ctx, "tool_result", map[string]any{"tool": meta.Name, "result": res})
	return res, nil
}

func resolveK8sClient(deps PlatformDeps, params map[string]any) (*kubernetes.Clientset, string, error) {
	clusterID := toInt(params["cluster_id"])
	if clusterID > 0 && deps.DB != nil {
		var cluster model.Cluster
		if err := deps.DB.First(&cluster, clusterID).Error; err == nil && strings.TrimSpace(cluster.KubeConfig) != "" {
			cfg, err := clientcmd.RESTConfigFromKubeConfig([]byte(cluster.KubeConfig))
			if err != nil {
				return nil, "cluster_kubeconfig", err
			}
			cli, err := kubernetes.NewForConfig(cfg)
			if err != nil {
				return nil, "cluster_kubeconfig", err
			}
			return cli, "cluster_kubeconfig", nil
		}
	}
	if deps.Clientset != nil {
		return deps.Clientset, "default_clientset", nil
	}
	return nil, "fallback", errors.New("k8s client unavailable")
}

func runLocalCommand(ctx context.Context, timeout time.Duration, name string, args ...string) (string, error) {
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(cctx, name, args...)
	out, err := cmd.CombinedOutput()
	if cctx.Err() == context.DeadlineExceeded {
		return strings.TrimSpace(string(out)), errors.New("command timeout")
	}
	return strings.TrimSpace(string(out)), err
}

func runOnTarget(ctx context.Context, deps PlatformDeps, target, localName string, localArgs []string, remoteCmd string) (string, string, error) {
	node, err := resolveNodeByTarget(deps, target)
	if err != nil {
		return "", "target_check", err
	}
	if node == nil {
		out, err := runLocalCommand(ctx, 6*time.Second, localName, localArgs...)
		return out, "local", err
	}
	cli, err := sshclient.NewSSHClient(node.SSHUser, node.SSHPassword, node.IP, node.Port, "")
	if err != nil {
		return "", "remote_ssh", err
	}
	defer cli.Close()
	out, err := sshclient.RunCommand(cli, remoteCmd)
	return out, "remote_ssh", err
}

func resolveNodeByTarget(deps PlatformDeps, target string) (*model.Node, error) {
	trimmed := strings.TrimSpace(target)
	if trimmed == "" || trimmed == "localhost" {
		return nil, nil
	}
	if deps.DB == nil {
		return nil, errors.New("db unavailable")
	}
	var node model.Node
	if id, err := strconv.ParseUint(trimmed, 10, 64); err == nil {
		if err := deps.DB.First(&node, id).Error; err == nil {
			return &node, nil
		}
	}
	if err := deps.DB.Where("ip = ? OR name = ? OR hostname = ?", trimmed, trimmed, trimmed).First(&node).Error; err != nil {
		return nil, errors.New("target not in host whitelist")
	}
	return &node, nil
}

func toInt(v any) int {
	switch x := v.(type) {
	case int:
		return x
	case int64:
		return int(x)
	case float64:
		return int(x)
	case uint64:
		return int(x)
	case json.Number:
		n, _ := strconv.Atoi(x.String())
		return n
	case string:
		n, _ := strconv.Atoi(strings.TrimSpace(x))
		return n
	default:
		return 0
	}
}

func toString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case json.Number:
		return x.String()
	default:
		return fmt.Sprintf("%v", x)
	}
}

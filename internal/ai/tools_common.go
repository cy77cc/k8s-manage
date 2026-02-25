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

func runWithPolicyAndEvent[T any](ctx context.Context, meta ToolMeta, input T, do func(T) (any, string, error)) (ToolResult, error) {
	start := time.Now()
	originalParams := structToMap(input)
	resolvedParams, resolution := resolveToolParams(ctx, meta, originalParams, "")
	runtimeInput, convErr := mapToInput[T](resolvedParams)
	if convErr != nil {
		res := ToolResult{OK: false, ErrorCode: "invalid_param", Error: convErr.Error(), Source: meta.Provider, LatencyMS: time.Since(start).Milliseconds()}
		EmitToolEvent(ctx, "tool_result", map[string]any{"tool": meta.Name, "call_id": nextToolCallID(), "result": res, "param_resolution": resolution, "retry": false})
		return res, nil
	}
	callID := nextToolCallID()
	EmitToolEvent(ctx, "tool_call", map[string]any{"tool": meta.Name, "call_id": callID, "params": resolvedParams, "param_resolution": resolution, "retry": false})
	if err := CheckToolPolicy(ctx, meta, resolvedParams); err != nil {
		res := ToolResult{OK: false, ErrorCode: "policy_denied", Error: err.Error(), Source: meta.Provider, LatencyMS: time.Since(start).Milliseconds()}
		EmitToolEvent(ctx, "tool_result", map[string]any{"tool": meta.Name, "call_id": callID, "result": res, "param_resolution": resolution, "retry": false})
		return res, err
	}
	data, source, err := runToolAttempt(ctx, meta, runtimeInput, do)
	retried := false
	if err != nil {
		if ie, ok := AsToolInputError(err); ok && ie.Code == "missing_param" {
			// First attempt still gets a terminal result event before retry.
			firstRes := ToolResult{
				OK:        false,
				ErrorCode: ie.Code,
				Error:     err.Error(),
				Source:    firstNonEmpty(source, meta.Provider),
				LatencyMS: time.Since(start).Milliseconds(),
			}
			EmitToolEvent(ctx, "tool_result", map[string]any{"tool": meta.Name, "call_id": callID, "result": firstRes, "param_resolution": resolution, "retry": true})
			retried = true
			resolvedRetry, resolutionRetry := resolveToolParams(ctx, meta, resolvedParams, ie.Field)
			if !equalJSONMap(resolvedRetry, resolvedParams) {
				resolvedParams = resolvedRetry
				runtimeInput, convErr = mapToInput[T](resolvedParams)
				if convErr == nil {
					callID = nextToolCallID()
					EmitToolEvent(ctx, "tool_call", map[string]any{"tool": meta.Name, "call_id": callID, "params": resolvedParams, "param_resolution": resolutionRetry, "retry": true})
					if err := CheckToolPolicy(ctx, meta, resolvedParams); err != nil {
						res := ToolResult{
							OK:        false,
							ErrorCode: "policy_denied",
							Error:     err.Error(),
							Source:    meta.Provider,
							LatencyMS: time.Since(start).Milliseconds(),
						}
						EmitToolEvent(ctx, "tool_result", map[string]any{"tool": meta.Name, "call_id": callID, "result": res, "param_resolution": resolutionRetry, "retry": true})
						return res, err
					}
					data, source, err = runToolAttempt(ctx, meta, runtimeInput, do)
					resolution = resolutionRetry
				}
			}
		}
	}
	if source == "" {
		source = meta.Provider
	}
	if err != nil {
		errCode := "tool_error"
		if ie, ok := AsToolInputError(err); ok {
			errCode = ie.Code
		}
		res := ToolResult{OK: false, ErrorCode: errCode, Error: err.Error(), Source: source, LatencyMS: time.Since(start).Milliseconds()}
		EmitToolEvent(ctx, "tool_result", map[string]any{"tool": meta.Name, "call_id": callID, "result": res, "param_resolution": resolution, "retry": retried})
		return res, nil
	}
	res := ToolResult{OK: true, Data: data, Source: source, LatencyMS: time.Since(start).Milliseconds()}
	EmitToolEvent(ctx, "tool_result", map[string]any{"tool": meta.Name, "call_id": callID, "result": res, "params": resolvedParams, "param_resolution": resolution, "retry": retried})
	if mem := ToolMemoryAccessorFromContext(ctx); mem != nil && res.OK {
		mem.SetLastToolParams(meta.Name, resolvedParams)
	}
	return res, nil
}

func runToolAttempt[T any](ctx context.Context, meta ToolMeta, input T, do func(T) (any, string, error)) (data any, source string, err error) {
	timeout := 8 * time.Second
	if meta.Mode == ToolModeMutating {
		timeout = 20 * time.Second
	}
	attemptCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	type attemptOut struct {
		data   any
		source string
		err    error
	}
	ch := make(chan attemptOut, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				ch <- attemptOut{err: &ToolInputError{Code: "tool_panic", Message: fmt.Sprintf("tool panic: %v", r)}}
			}
		}()
		d, s, e := do(input)
		ch <- attemptOut{data: d, source: s, err: e}
	}()

	select {
	case <-attemptCtx.Done():
		if errors.Is(attemptCtx.Err(), context.DeadlineExceeded) {
			return nil, meta.Provider, &ToolInputError{Code: "tool_timeout", Message: "tool execution timeout"}
		}
		return nil, meta.Provider, &ToolInputError{Code: "tool_canceled", Message: "tool execution canceled"}
	case out := <-ch:
		if out.source == "" {
			out.source = meta.Provider
		}
		if out.err == nil {
			return out.data, out.source, nil
		}
		return out.data, out.source, normalizeToolErr(out.err)
	}
}

func normalizeToolErr(err error) error {
	if err == nil {
		return nil
	}
	if _, ok := AsToolInputError(err); ok {
		return err
	}
	msg := strings.ToLower(strings.TrimSpace(err.Error()))
	switch {
	case strings.Contains(msg, "timeout"):
		return &ToolInputError{Code: "tool_timeout", Message: err.Error()}
	case strings.Contains(msg, "canceled"):
		return &ToolInputError{Code: "tool_canceled", Message: err.Error()}
	default:
		return err
	}
}

func firstNonEmpty(v ...string) string {
	for _, item := range v {
		if strings.TrimSpace(item) != "" {
			return item
		}
	}
	return ""
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

func structToMap(v any) map[string]any {
	raw, err := json.Marshal(v)
	if err != nil {
		return map[string]any{}
	}
	out := map[string]any{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return map[string]any{}
	}
	return out
}

func mapToInput[T any](input map[string]any) (T, error) {
	var out T
	raw, err := json.Marshal(input)
	if err != nil {
		return out, err
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return out, err
	}
	return out, nil
}

func equalJSONMap(a, b map[string]any) bool {
	ra, _ := json.Marshal(a)
	rb, _ := json.Marshal(b)
	return string(ra) == string(rb)
}

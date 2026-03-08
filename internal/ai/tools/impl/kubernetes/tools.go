package kubernetes

import (
	"context"
	"strings"

	. "github.com/cy77cc/k8s-manage/internal/ai/tools/core"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func K8sQuery(ctx context.Context, deps PlatformDeps, input K8sQueryInput) (ToolResult, error) {
	return RunWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:       "k8s_query",
			Mode:       ToolModeReadonly,
			Risk:       ToolRiskLow,
			Provider:   "local",
			Permission: "ai:tool:read",
		},
		input,
		func(in K8sQueryInput) (any, string, error) {
			if strings.TrimSpace(in.Resource) == "" {
				return nil, "validation", NewMissingParam("resource", "resource is required")
			}
			cli, source, err := ResolveK8sClient(deps, StructToMap(in))
			if err != nil {
				return nil, source, err
			}
			ns := strings.TrimSpace(in.Namespace)
			if ns == "" {
				ns = corev1.NamespaceAll
			}
			limit := in.Limit
			if limit <= 0 {
				limit = 50
			}
			name := strings.TrimSpace(in.Name)
			label := strings.TrimSpace(in.Label)
			resource := strings.ToLower(strings.TrimSpace(in.Resource))
			listOpts := metav1.ListOptions{}
			if label != "" {
				listOpts.LabelSelector = label
			}

			switch resource {
			case "pods":
				list, err := cli.CoreV1().Pods(ns).List(ctx, listOpts)
				if err != nil {
					return nil, source, err
				}
				out := make([]map[string]any, 0, len(list.Items))
				for _, p := range list.Items {
					if name != "" && p.Name != name {
						continue
					}
					out = append(out, map[string]any{
						"name":      p.Name,
						"namespace": p.Namespace,
						"phase":     p.Status.Phase,
						"node":      p.Spec.NodeName,
						"labels":    p.Labels,
					})
					if len(out) >= limit {
						break
					}
				}
				return out, source, nil
			case "services":
				list, err := cli.CoreV1().Services(ns).List(ctx, listOpts)
				if err != nil {
					return nil, source, err
				}
				out := make([]map[string]any, 0, len(list.Items))
				for _, s := range list.Items {
					if name != "" && s.Name != name {
						continue
					}
					out = append(out, map[string]any{
						"name":       s.Name,
						"namespace":  s.Namespace,
						"type":       s.Spec.Type,
						"cluster_ip": s.Spec.ClusterIP,
						"labels":     s.Labels,
					})
					if len(out) >= limit {
						break
					}
				}
				return out, source, nil
			case "deployments":
				list, err := cli.AppsV1().Deployments(ns).List(ctx, listOpts)
				if err != nil {
					return nil, source, err
				}
				out := make([]map[string]any, 0, len(list.Items))
				for _, d := range list.Items {
					if name != "" && d.Name != name {
						continue
					}
					out = append(out, map[string]any{
						"name":      d.Name,
						"namespace": d.Namespace,
						"ready":     d.Status.ReadyReplicas,
						"replicas":  d.Status.Replicas,
						"updated":   d.Status.UpdatedReplicas,
						"labels":    d.Labels,
					})
					if len(out) >= limit {
						break
					}
				}
				return out, source, nil
			case "nodes":
				list, err := cli.CoreV1().Nodes().List(ctx, listOpts)
				if err != nil {
					return nil, source, err
				}
				out := make([]map[string]any, 0, len(list.Items))
				for _, n := range list.Items {
					if name != "" && n.Name != name {
						continue
					}
					out = append(out, map[string]any{
						"name":   n.Name,
						"labels": n.Labels,
					})
					if len(out) >= limit {
						break
					}
				}
				return out, source, nil
			default:
				return nil, source, NewInvalidParam("resource", "unsupported resource")
			}
		})
}

func K8sListResources(ctx context.Context, deps PlatformDeps, input K8sListInput) (ToolResult, error) {
	return RunWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:       "k8s_list_resources",
			Mode:       ToolModeReadonly,
			Risk:       ToolRiskLow,
			Provider:   "local",
			Permission: "ai:tool:read",
		},
		input,
		func(in K8sListInput) (any, string, error) {
			if strings.TrimSpace(in.Resource) == "" {
				return nil, "validation", NewMissingParam("resource", "resource is required")
			}
			cli, source, err := ResolveK8sClient(deps, StructToMap(in))
			if err != nil {
				return nil, source, err
			}
			ns := strings.TrimSpace(in.Namespace)
			if ns == "" {
				ns = corev1.NamespaceAll
			}
			resource := strings.ToLower(strings.TrimSpace(in.Resource))
			limit := in.Limit
			if limit <= 0 {
				limit = 50
			}
			switch resource {
			case "pods":
				list, err := cli.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
				if err != nil {
					return nil, source, err
				}
				out := make([]map[string]any, 0, len(list.Items))
				for i, p := range list.Items {
					if i >= limit {
						break
					}
					out = append(out, map[string]any{"name": p.Name, "namespace": p.Namespace, "phase": p.Status.Phase})
				}
				return out, source, nil
			case "services":
				list, err := cli.CoreV1().Services(ns).List(ctx, metav1.ListOptions{})
				if err != nil {
					return nil, source, err
				}
				out := make([]map[string]any, 0, len(list.Items))
				for i, s := range list.Items {
					if i >= limit {
						break
					}
					out = append(out, map[string]any{"name": s.Name, "namespace": s.Namespace, "type": s.Spec.Type})
				}
				return out, source, nil
			case "deployments":
				list, err := cli.AppsV1().Deployments(ns).List(ctx, metav1.ListOptions{})
				if err != nil {
					return nil, source, err
				}
				out := make([]map[string]any, 0, len(list.Items))
				for i, d := range list.Items {
					if i >= limit {
						break
					}
					out = append(out, map[string]any{"name": d.Name, "namespace": d.Namespace, "ready": d.Status.ReadyReplicas, "replicas": d.Status.Replicas})
				}
				return out, source, nil
			case "nodes":
				list, err := cli.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
				if err != nil {
					return nil, source, err
				}
				out := make([]map[string]any, 0, len(list.Items))
				for i, n := range list.Items {
					if i >= limit {
						break
					}
					out = append(out, map[string]any{"name": n.Name, "labels": n.Labels})
				}
				return out, source, nil
			default:
				return nil, source, NewInvalidParam("resource", "unsupported resource")
			}
		})
}

func K8sEvents(ctx context.Context, deps PlatformDeps, input K8sEventsQueryInput) (ToolResult, error) {
	return RunWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:       "k8s_events",
			Mode:       ToolModeReadonly,
			Risk:       ToolRiskLow,
			Provider:   "local",
			Permission: "ai:tool:read",
		},
		input,
		func(in K8sEventsQueryInput) (any, string, error) {
			cli, source, err := ResolveK8sClient(deps, StructToMap(in))
			if err != nil {
				return nil, source, err
			}
			ns := strings.TrimSpace(in.Namespace)
			if ns == "" {
				ns = corev1.NamespaceAll
			}
			limit := in.Limit
			if limit <= 0 {
				limit = 50
			}
			kind := strings.TrimSpace(in.Kind)
			name := strings.TrimSpace(in.Name)
			list, err := cli.CoreV1().Events(ns).List(ctx, metav1.ListOptions{})
			if err != nil {
				return nil, source, err
			}
			out := make([]map[string]any, 0, len(list.Items))
			for _, e := range list.Items {
				if kind != "" && !strings.EqualFold(e.InvolvedObject.Kind, kind) {
					continue
				}
				if name != "" && e.InvolvedObject.Name != name {
					continue
				}
				out = append(out, map[string]any{
					"type":      e.Type,
					"reason":    e.Reason,
					"message":   e.Message,
					"namespace": e.Namespace,
					"kind":      e.InvolvedObject.Kind,
					"name":      e.InvolvedObject.Name,
				})
				if len(out) >= limit {
					break
				}
			}
			return out, source, nil
		})
}

func K8sGetEvents(ctx context.Context, deps PlatformDeps, input K8sEventsInput) (ToolResult, error) {
	return RunWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:       "k8s_get_events",
			Mode:       ToolModeReadonly,
			Risk:       ToolRiskLow,
			Provider:   "local",
			Permission: "ai:tool:read",
		},
		input,
		func(in K8sEventsInput) (any, string, error) {
			cli, source, err := ResolveK8sClient(deps, StructToMap(in))
			if err != nil {
				return nil, source, err
			}
			ns := strings.TrimSpace(in.Namespace)
			if ns == "" {
				ns = corev1.NamespaceAll
			}
			limit := in.Limit
			if limit <= 0 {
				limit = 50
			}
			list, err := cli.CoreV1().Events(ns).List(ctx, metav1.ListOptions{})
			if err != nil {
				return nil, source, err
			}
			out := make([]map[string]any, 0, len(list.Items))
			for i, e := range list.Items {
				if i >= limit {
					break
				}
				out = append(out, map[string]any{"type": e.Type, "reason": e.Reason, "message": e.Message})
			}
			return out, source, nil
		})
}

func K8sLogs(ctx context.Context, deps PlatformDeps, input K8sLogsInput) (ToolResult, error) {
	return RunWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:       "k8s_logs",
			Mode:       ToolModeReadonly,
			Risk:       ToolRiskMedium,
			Provider:   "local",
			Permission: "ai:tool:read",
		},
		input,
		func(in K8sLogsInput) (any, string, error) {
			cli, source, err := ResolveK8sClient(deps, StructToMap(in))
			if err != nil {
				return nil, source, err
			}
			ns := strings.TrimSpace(in.Namespace)
			if ns == "" {
				ns = "default"
			}
			pod := strings.TrimSpace(in.Pod)
			if pod == "" {
				return nil, source, NewMissingParam("pod", "pod is required")
			}
			tailLines := int64(in.TailLines)
			if tailLines <= 0 {
				tailLines = 200
			}
			opt := &corev1.PodLogOptions{Container: strings.TrimSpace(in.Container), TailLines: &tailLines}
			raw, err := cli.CoreV1().Pods(ns).GetLogs(pod, opt).DoRaw(ctx)
			if err != nil {
				return nil, source, err
			}
			return map[string]any{
				"namespace": ns,
				"pod":       pod,
				"container": strings.TrimSpace(in.Container),
				"logs":      string(raw),
			}, source, nil
		})
}

func K8sGetPodLogs(ctx context.Context, deps PlatformDeps, input K8sPodLogsInput) (ToolResult, error) {
	return RunWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:       "k8s_get_pod_logs",
			Mode:       ToolModeReadonly,
			Risk:       ToolRiskMedium,
			Provider:   "local",
			Permission: "ai:tool:read",
		},
		input,
		func(in K8sPodLogsInput) (any, string, error) {
			cli, source, err := ResolveK8sClient(deps, StructToMap(in))
			if err != nil {
				return nil, source, err
			}
			ns := strings.TrimSpace(in.Namespace)
			if ns == "" {
				ns = "default"
			}
			pod := strings.TrimSpace(in.Pod)
			if pod == "" {
				return nil, source, NewMissingParam("pod", "pod is required")
			}
			tailLines := int64(in.TailLines)
			if tailLines <= 0 {
				tailLines = 200
			}
			opt := &corev1.PodLogOptions{Container: strings.TrimSpace(in.Container), TailLines: &tailLines}
			raw, err := cli.CoreV1().Pods(ns).GetLogs(pod, opt).DoRaw(ctx)
			if err != nil {
				return nil, source, err
			}
			return map[string]any{"namespace": ns, "pod": pod, "logs": string(raw)}, source, nil
		})
}

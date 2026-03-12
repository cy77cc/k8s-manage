// Package observability 提供 AI 编排层的可观测性能力。
//
// 本文件提供 HTTP 接口来暴露可观测性指标。
package observability

import (
	"net/http"
	"strings"
)

// HTTPHandler 返回可观测性指标的 HTTP handler。
// 可用于暴露 /metrics/observability 等端点。
func HTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snapshot := Snapshot()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// 简单的 JSON 响应
		w.Write([]byte(`{
  "llm_calls": ` + marshalLLMMetrics(snapshot.LLMCalls) + `,
  "tool_calls": ` + marshalToolMetrics(snapshot.ToolCalls) + `,
  "agent_runs": ` + marshalAgentMetrics(snapshot.AgentRuns) + `,
  "total_tokens": ` + itoa(snapshot.TotalTokens) + `,
  "total_llm_calls": ` + itoa(snapshot.TotalLLMCalls) + `,
  "total_tool_calls": ` + itoa(snapshot.TotalToolCalls) + `,
  "total_errors": ` + itoa(snapshot.TotalErrors) + `,
  "avg_latency_ms": ` + itoa(snapshot.AvgLatencyMs) + `
}`))
	}
}

// marshalLLMMetrics 序列化 LLM 指标。
func marshalLLMMetrics(m map[string]*LLMMetrics) string {
	if len(m) == 0 {
		return "{}"
	}
	var result strings.Builder
	result.WriteString("{")
	first := true
	for name, metrics := range m {
		if !first {
			result.WriteString(",")
		}
		first = false
		result.WriteString(`"`)
		result.WriteString(name)
		result.WriteString(`":{`)
		result.WriteString(`"call_count":`)
		result.WriteString(itoa(metrics.CallCount))
		result.WriteString(`,"token_count":`)
		result.WriteString(itoa(metrics.TokenCount))
		result.WriteString(`,"error_count":`)
		result.WriteString(itoa(metrics.ErrorCount))
		result.WriteString(`,"avg_latency_ms":`)
		result.WriteString(itoa(avgLatency(metrics.CallCount, metrics.TotalLatencyMs)))
		result.WriteString("}")
	}
	result.WriteString("}")
	return result.String()
}

// marshalToolMetrics 序列化工具指标。
func marshalToolMetrics(m map[string]*ToolMetrics) string {
	if len(m) == 0 {
		return "{}"
	}
	var result strings.Builder
	result.WriteString("{")
	first := true
	for name, metrics := range m {
		if !first {
			result.WriteString(",")
		}
		first = false
		result.WriteString(`"`)
		result.WriteString(name)
		result.WriteString(`":{`)
		result.WriteString(`"call_count":`)
		result.WriteString(itoa(metrics.CallCount))
		result.WriteString(`,"error_count":`)
		result.WriteString(itoa(metrics.ErrorCount))
		result.WriteString(`,"avg_latency_ms":`)
		result.WriteString(itoa(avgLatency(metrics.CallCount, metrics.TotalLatencyMs)))
		result.WriteString("}")
	}
	result.WriteString("}")
	return result.String()
}

// marshalAgentMetrics 序列化 Agent 指标。
func marshalAgentMetrics(m map[string]*AgentMetrics) string {
	if len(m) == 0 {
		return "{}"
	}
	var result strings.Builder
	result.WriteString("{")
	first := true
	for name, metrics := range m {
		if !first {
			result.WriteString(",")
		}
		first = false
		result.WriteString(`"`)
		result.WriteString(name)
		result.WriteString(`":{`)
		result.WriteString(`"run_count":`)
		result.WriteString(itoa(metrics.RunCount))
		result.WriteString(`,"error_count":`)
		result.WriteString(itoa(metrics.ErrorCount))
		result.WriteString(`,"tool_call_count":`)
		result.WriteString(itoa(metrics.ToolCallCount))
		result.WriteString(`,"avg_latency_ms":`)
		result.WriteString(itoa(avgLatency(metrics.RunCount, metrics.TotalLatencyMs)))
		result.WriteString("}")
	}
	result.WriteString("}")
	return result.String()
}

// avgLatency 计算平均延迟。
func avgLatency(count int64, totalMs int64) int64 {
	if count == 0 {
		return 0
	}
	return totalMs / count
}

// itoa 将 int64 转为字符串。
func itoa(n int64) string {
	if n == 0 {
		return "0"
	}

	var negative bool
	if n < 0 {
		negative = true
		n = -n
	}

	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}

	if negative {
		digits = append([]byte{'-'}, digits...)
	}

	return string(digits)
}

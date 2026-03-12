// Package observability 提供 AI 编排层的可观测性能力。
//
// 本文件实现 Prometheus 指标暴露，用于实时仪表盘和告警。
package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Prometheus 指标定义
var (
	// LLM 调用计数
	llmCallsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_llm_calls_total",
			Help: "Total number of LLM calls.",
		},
		[]string{"model", "status"},
	)

	// LLM Token 消耗
	llmTokensTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_llm_tokens_total",
			Help: "Total tokens consumed by LLM calls.",
		},
		[]string{"model", "type"}, // type: input/output
	)

	// LLM 调用延迟
	llmLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ai_llm_latency_seconds",
			Help:    "LLM call latency in seconds.",
			Buckets: []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30, 60},
		},
		[]string{"model"},
	)

	// 工具调用计数
	toolCallsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_tool_calls_total",
			Help: "Total number of tool calls.",
		},
		[]string{"tool", "status"},
	)

	// 工具调用延迟
	toolLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ai_tool_latency_seconds",
			Help:    "Tool call latency in seconds.",
			Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"tool"},
	)

	// Agent 运行计数
	agentRunsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_agent_runs_total",
			Help: "Total number of agent runs.",
		},
		[]string{"agent", "status"},
	)

	// Agent 运行延迟
	agentLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ai_agent_latency_seconds",
			Help:    "Agent run latency in seconds.",
			Buckets: []float64{0.5, 1, 2.5, 5, 10, 30, 60, 120, 300},
		},
		[]string{"agent"},
	)

	// Agent 迭代次数
	agentIterations = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ai_agent_iterations",
			Help:    "Number of iterations per agent run.",
			Buckets: []float64{1, 2, 3, 5, 10, 20},
		},
		[]string{"agent"},
	)

	// 活跃会话数
	activeSessions = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "ai_active_sessions",
			Help: "Number of active AI sessions.",
		},
	)

	// 队列深度（等待处理的请求数）
	queueDepth = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "ai_queue_depth",
			Help: "Number of requests waiting to be processed.",
		},
	)

	// 错误计数
	errorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_errors_total",
			Help: "Total number of errors by type.",
		},
		[]string{"component", "error_type"},
	)
)

// RecordLLMMetric 记录 LLM 调用指标到 Prometheus。
func RecordLLMMetric(model string, inputTokens, outputTokens int, latencySeconds float64, err error) {
	status := "success"
	if err != nil {
		status = "error"
		errorsTotal.WithLabelValues("llm", "call_error").Inc()
	}

	llmCallsTotal.WithLabelValues(model, status).Inc()
	llmTokensTotal.WithLabelValues(model, "input").Add(float64(inputTokens))
	llmTokensTotal.WithLabelValues(model, "output").Add(float64(outputTokens))
	llmLatency.WithLabelValues(model).Observe(latencySeconds)
}

// RecordToolMetric 记录工具调用指标到 Prometheus。
func RecordToolMetric(tool string, latencySeconds float64, err error) {
	status := "success"
	if err != nil {
		status = "error"
		errorsTotal.WithLabelValues("tool", "call_error").Inc()
	}

	toolCallsTotal.WithLabelValues(tool, status).Inc()
	toolLatency.WithLabelValues(tool).Observe(latencySeconds)
}

// RecordAgentMetric 记录 Agent 运行指标到 Prometheus。
func RecordAgentMetric(agent string, latencySeconds float64, iterations int, err error) {
	status := "success"
	if err != nil {
		status = "error"
		errorsTotal.WithLabelValues("agent", "run_error").Inc()
	}

	agentRunsTotal.WithLabelValues(agent, status).Inc()
	agentLatency.WithLabelValues(agent).Observe(latencySeconds)
	agentIterations.WithLabelValues(agent).Observe(float64(iterations))
}

// SetActiveSessions 设置活跃会话数。
func SetActiveSessions(count int) {
	activeSessions.Set(float64(count))
}

// SetQueueDepth 设置队列深度。
func SetQueueDepth(count int) {
	queueDepth.Set(float64(count))
}

// IncQueueDepth 增加队列深度。
func IncQueueDepth() {
	queueDepth.Inc()
}

// DecQueueDepth 减少队列深度。
func DecQueueDepth() {
	queueDepth.Dec()
}

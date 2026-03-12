// Package observability 提供 AI 编排层的可观测性能力。
//
// 本文件实现指标收集器，用于统计 LLM 调用、Token 消耗、工具调用等指标。
package observability

import (
	"sync"
	"sync/atomic"
	"time"
)

// Metrics 指标收集器。
type Metrics struct {
	mu sync.RWMutex

	// LLM 指标
	llmCalls      map[string]*LLMMetrics // 按模型名称分组

	// 工具指标
	toolCalls      map[string]*ToolMetrics // 按工具名称分组

	// Agent 指标
	agentRuns      map[string]*AgentMetrics // 按 Agent 名称分组

	// 全局计数器
	totalTokens      int64
	totalLLMCalls    int64
	totalToolCalls   int64
	totalErrors      int64
	totalLatencyMs   int64
}

// LLMMetrics LLM 调用指标。
type LLMMetrics struct {
	Name          string
	CallCount     int64
	TokenCount    int64
	InputTokens   int64
	OutputTokens  int64
	ErrorCount    int64
	TotalLatencyMs int64
	LastCallAt    time.Time
}

// ToolMetrics 工具调用指标。
type ToolMetrics struct {
	Name           string
	CallCount      int64
	ErrorCount     int64
	TotalLatencyMs int64
	LastCallAt     time.Time
}

// AgentMetrics Agent 运行指标。
type AgentMetrics struct {
	Name            string
	RunCount        int64
	ErrorCount      int64
	TotalLatencyMs  int64
	ToolCallCount   int64
	IterationCount  int64
	LastRunAt       time.Time
}

// MetricsSnapshot 指标快照。
type MetricsSnapshot struct {
	LLMCalls     map[string]*LLMMetrics  `json:"llm_calls"`
	ToolCalls    map[string]*ToolMetrics `json:"tool_calls"`
	AgentRuns    map[string]*AgentMetrics `json:"agent_runs"`
	TotalTokens  int64                   `json:"total_tokens"`
	TotalLLMCalls int64                  `json:"total_llm_calls"`
	TotalToolCalls int64                 `json:"total_tool_calls"`
	TotalErrors  int64                   `json:"total_errors"`
	AvgLatencyMs int64                   `json:"avg_latency_ms"`
}

// NewMetrics 创建新的指标收集器。
func NewMetrics() *Metrics {
	return &Metrics{
		llmCalls:   make(map[string]*LLMMetrics),
		toolCalls:  make(map[string]*ToolMetrics),
		agentRuns:  make(map[string]*AgentMetrics),
	}
}

// RecordLLMCall 记录 LLM 调用。
func (m *Metrics) RecordLLMCall(model string, inputTokens, outputTokens int, latencyMs int64, err error) {
	if m == nil {
		return
	}

	atomic.AddInt64(&m.totalLLMCalls, 1)
	atomic.AddInt64(&m.totalTokens, int64(inputTokens+outputTokens))
	atomic.AddInt64(&m.totalLatencyMs, latencyMs)

	if err != nil {
		atomic.AddInt64(&m.totalErrors, 1)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	metrics, ok := m.llmCalls[model]
	if !ok {
		metrics = &LLMMetrics{Name: model}
		m.llmCalls[model] = metrics
	}

	metrics.CallCount++
	metrics.TokenCount += int64(inputTokens + outputTokens)
	metrics.InputTokens += int64(inputTokens)
	metrics.OutputTokens += int64(outputTokens)
	metrics.TotalLatencyMs += latencyMs
	metrics.LastCallAt = time.Now().UTC()

	if err != nil {
		metrics.ErrorCount++
	}
}

// RecordToolCall 记录工具调用。
func (m *Metrics) RecordToolCall(tool string, latencyMs int64, err error) {
	if m == nil {
		return
	}

	atomic.AddInt64(&m.totalToolCalls, 1)
	atomic.AddInt64(&m.totalLatencyMs, latencyMs)

	if err != nil {
		atomic.AddInt64(&m.totalErrors, 1)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	metrics, ok := m.toolCalls[tool]
	if !ok {
		metrics = &ToolMetrics{Name: tool}
		m.toolCalls[tool] = metrics
	}

	metrics.CallCount++
	metrics.TotalLatencyMs += latencyMs
	metrics.LastCallAt = time.Now().UTC()

	if err != nil {
		metrics.ErrorCount++
	}
}

// RecordAgentRun 记录 Agent 运行。
func (m *Metrics) RecordAgentRun(agent string, latencyMs int64, toolCalls, iterations int, err error) {
	if m == nil {
		return
	}

	atomic.AddInt64(&m.totalLatencyMs, latencyMs)

	if err != nil {
		atomic.AddInt64(&m.totalErrors, 1)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	metrics, ok := m.agentRuns[agent]
	if !ok {
		metrics = &AgentMetrics{Name: agent}
		m.agentRuns[agent] = metrics
	}

	metrics.RunCount++
	metrics.TotalLatencyMs += latencyMs
	metrics.ToolCallCount += int64(toolCalls)
	metrics.IterationCount += int64(iterations)
	metrics.LastRunAt = time.Now().UTC()

	if err != nil {
		metrics.ErrorCount++
	}
}

// Snapshot 返回当前指标的快照。
func (m *Metrics) Snapshot() MetricsSnapshot {
	if m == nil {
		return MetricsSnapshot{}
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	totalCalls := m.totalLLMCalls + m.totalToolCalls
	var avgLatency int64
	if totalCalls > 0 {
		avgLatency = m.totalLatencyMs / totalCalls
	}

	return MetricsSnapshot{
		LLMCalls:       m.llmCalls,
		ToolCalls:      m.toolCalls,
		AgentRuns:      m.agentRuns,
		TotalTokens:    m.totalTokens,
		TotalLLMCalls:  m.totalLLMCalls,
		TotalToolCalls: m.totalToolCalls,
		TotalErrors:    m.totalErrors,
		AvgLatencyMs:   avgLatency,
	}
}

// Reset 重置所有指标。
func (m *Metrics) Reset() {
	if m == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.llmCalls = make(map[string]*LLMMetrics)
	m.toolCalls = make(map[string]*ToolMetrics)
	m.agentRuns = make(map[string]*AgentMetrics)
	m.totalTokens = 0
	m.totalLLMCalls = 0
	m.totalToolCalls = 0
	m.totalErrors = 0
	m.totalLatencyMs = 0
}

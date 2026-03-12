// Package observability 提供 AI 编排层的可观测性能力。
//
// 使用示例:
//
//	package main
//
//	import (
//	    "github.com/cy77cc/OpsPilot/internal/ai/observability"
//	)
//
//	func main() {
//	    // 1. 在应用启动时初始化可观测性
//	    handler := observability.Setup(observability.Config{
//	        EnableTracing: true,
//	        EnableMetrics: true,
//	        EventHandler: func(name string, meta events.EventMeta, data map[string]any) {
//	            // 处理可观测性事件
//	            // 可以发送到日志、SSE、或监控系统
//	            log.Printf("[%s] %v", name, data)
//	        },
//	    })
//
//	    // 2. 在需要时获取指标快照
//	    snapshot := handler.Snapshot()
//	    fmt.Printf("Total LLM calls: %d\n", snapshot.TotalLLMCalls)
//	    fmt.Printf("Total tokens: %d\n", snapshot.TotalTokens)
//	    fmt.Printf("Average latency: %dms\n", snapshot.AvgLatencyMs)
//
//	    // 3. 或者在 Orchestrator 中获取
//	    orch := NewOrchestrator(...)
//	    obsSnapshot := orch.ObservabilitySnapshot()
//	}
package observability

// Package observability 提供 AI 编排层的可观测性能力。
//
// 本文件实现基于 Eino Callback 机制的可观测性处理器，
// 支持 LLM 调用、工具调用、Agent 运行的追踪和指标收集。
package observability

import (
	"context"
	"sync"
	"time"
)

// EventEmitter 事件发射器函数类型。
type EventEmitter func(name string, data map[string]any)

// Span 表示一次调用的追踪跨度。
type Span struct {
	ID        string        // 跨度 ID
	Name      string        // 跨度名称
	StartTime time.Time     // 开始时间
	EndTime   time.Time     // 结束时间
	Duration  time.Duration // 持续时间
	ParentID  string        // 父跨度 ID
	Metadata  map[string]any // 元数据
}

// Tracer 链路追踪器，管理调用跨度的生命周期。
type Tracer struct {
	mu    sync.RWMutex
	spans map[string]*Span // 活跃的跨度
}

// NewTracer 创建新的追踪器。
func NewTracer() *Tracer {
	return &Tracer{
		spans: make(map[string]*Span),
	}
}

// StartSpan 开始一个新的跨度。
func (t *Tracer) StartSpan(ctx context.Context, spanType, name string) *Span {
	span := &Span{
		ID:        generateSpanID(),
		Name:      spanType + ":" + name,
		StartTime: time.Now().UTC(),
		Metadata:  make(map[string]any),
	}

	// 尝试从上下文获取父跨度
	if parentID, ok := ctx.Value(spanIDKey{}).(string); ok && parentID != "" {
		span.ParentID = parentID
	}

	t.mu.Lock()
	t.spans[span.ID] = span
	t.mu.Unlock()

	return span
}

// End 结束跨度。
func (s *Span) End() {
	if s == nil {
		return
	}
	s.EndTime = time.Now().UTC()
	s.Duration = s.EndTime.Sub(s.StartTime)
}

// spanIDKey 上下文键类型。
type spanIDKey struct{}

// ContextWithSpanID 将跨度 ID 存入上下文。
func ContextWithSpanID(ctx context.Context, spanID string) context.Context {
	return context.WithValue(ctx, spanIDKey{}, spanID)
}

// SpanIDFromContext 从上下文获取跨度 ID。
func SpanIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(spanIDKey{}).(string); ok {
		return id
	}
	return ""
}

// generateSpanID 生成唯一的跨度 ID。
func generateSpanID() string {
	return time.Now().Format("20060102150405.000000")
}

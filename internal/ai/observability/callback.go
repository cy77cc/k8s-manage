// Package observability 提供 AI 编排层的可观测性能力。
//
// 本文件实现基于 Eino Callback 机制的可观测性处理器，
// 支持 LLM 调用、工具调用、Agent 运行的追踪和指标收集。
package observability

import (
	"context"
	"strings"
	"time"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	template "github.com/cloudwego/eino/utils/callbacks"

	"github.com/cy77cc/OpsPilot/internal/ai/events"
)

// EventHandler 事件处理器函数类型。
type EventHandler func(name string, meta events.EventMeta, data map[string]any)

// Handler 可观测性处理器，实现 Eino 回调接口。
type Handler struct {
	tracer       *Tracer
	metrics      *Metrics
	eventHandler EventHandler
}

// NewHandler 创建新的可观测性处理器。
func NewHandler(eventHandler EventHandler) *Handler {
	return &Handler{
		tracer:       NewTracer(),
		metrics:      NewMetrics(),
		eventHandler: eventHandler,
	}
}

// BuildCallbackHandler 构建 Eino 回调处理器。
func (h *Handler) BuildCallbackHandler() callbacks.Handler {
	return template.NewHandlerHelper().
		ChatModel(h.chatModelHandler()).
		Tool(h.toolHandler()).
		Agent(h.agentHandler()).
		Handler()
}

// Tracer 返回追踪器实例。
func (h *Handler) Tracer() *Tracer {
	return h.tracer
}

// Metrics 返回指标收集器实例。
func (h *Handler) Metrics() *Metrics {
	return h.metrics
}

// Snapshot 返回当前指标快照。
func (h *Handler) Snapshot() MetricsSnapshot {
	return h.metrics.Snapshot()
}

// =============================================================================
// ChatModel 回调处理器
// =============================================================================

func (h *Handler) chatModelHandler() *template.ModelCallbackHandler {
	return &template.ModelCallbackHandler{
		OnStart: func(ctx context.Context, info *callbacks.RunInfo, input *model.CallbackInput) context.Context {
			span := h.tracer.StartSpan(ctx, "llm", info.Name)
			ctx = ContextWithSpanID(ctx, span.ID)

			if h.eventHandler != nil {
				h.eventHandler("llm_start", events.EventMeta{
					Timestamp: time.Now().UTC(),
				}, map[string]any{
					"model":       info.Name,
					"component":   string(info.Component),
					"span_id":     span.ID,
					"messages":    len(input.Messages),
					"message_preview": truncateMessages(input.Messages, 200),
				})
			}

			return ctx
		},
		OnEnd: func(ctx context.Context, info *callbacks.RunInfo, output *model.CallbackOutput) context.Context {
			spanID := SpanIDFromContext(ctx)
			span := h.getSpan(spanID)
			if span != nil {
				span.End()
			}

			// 估算 token 数量
			content := ""
			if output.Message != nil {
				content = output.Message.Content
			}
			tokens := estimateTokens(content)
			latencyMs := int64(0)
			if span != nil {
				latencyMs = span.Duration.Milliseconds()
			}

			h.metrics.RecordLLMCall(info.Name, 0, tokens, latencyMs, nil)

			if h.eventHandler != nil {
				h.eventHandler("llm_end", events.EventMeta{
					Timestamp: time.Now().UTC(),
				}, map[string]any{
					"model":        info.Name,
					"span_id":      spanID,
					"tokens":       tokens,
					"duration_ms":  latencyMs,
					"content_len":  len(content),
					"response_preview": truncate(content, 500),
				})
			}

			return ctx
		},
		OnError: func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
			spanID := SpanIDFromContext(ctx)
			span := h.getSpan(spanID)
			if span != nil {
				span.End()
			}

			latencyMs := int64(0)
			if span != nil {
				latencyMs = span.Duration.Milliseconds()
			}

			h.metrics.RecordLLMCall(info.Name, 0, 0, latencyMs, err)

			if h.eventHandler != nil {
				h.eventHandler("llm_error", events.EventMeta{
					Timestamp: time.Now().UTC(),
				}, map[string]any{
					"model":       info.Name,
					"span_id":     spanID,
					"error":       err.Error(),
					"duration_ms": latencyMs,
				})
			}

			return ctx
		},
		OnEndWithStreamOutput: func(ctx context.Context, info *callbacks.RunInfo, output *schema.StreamReader[*model.CallbackOutput]) context.Context {
			spanID := SpanIDFromContext(ctx)

			// 流式输出时，在后台消费并追踪
			go func() {
				var content strings.Builder
				start := time.Now()

				for {
					chunk, err := output.Recv()
					if err != nil {
						break
					}
					if chunk != nil && chunk.Message != nil {
						content.WriteString(chunk.Message.Content)
					}
				}

				span := h.getSpan(spanID)
				if span != nil {
					span.End()
				}

				tokens := estimateTokens(content.String())
				latencyMs := time.Since(start).Milliseconds()

				h.metrics.RecordLLMCall(info.Name, 0, tokens, latencyMs, nil)

				if h.eventHandler != nil {
					h.eventHandler("llm_stream_end", events.EventMeta{
						Timestamp: time.Now().UTC(),
					}, map[string]any{
						"model":        info.Name,
						"span_id":      spanID,
						"tokens":       tokens,
						"duration_ms":  latencyMs,
						"content_len":  content.Len(),
					})
				}
			}()

			return ctx
		},
	}
}

// =============================================================================
// Tool 回调处理器
// =============================================================================

func (h *Handler) toolHandler() *template.ToolCallbackHandler {
	return &template.ToolCallbackHandler{
		OnStart: func(ctx context.Context, info *callbacks.RunInfo, input *tool.CallbackInput) context.Context {
			span := h.tracer.StartSpan(ctx, "tool", info.Name)
			ctx = ContextWithSpanID(ctx, span.ID)

			if h.eventHandler != nil {
				h.eventHandler("tool_start", events.EventMeta{
					Timestamp: time.Now().UTC(),
				}, map[string]any{
					"tool":        info.Name,
					"span_id":     span.ID,
					"arguments":   truncate(input.ArgumentsInJSON, 1000),
				})
			}

			return ctx
		},
		OnEnd: func(ctx context.Context, info *callbacks.RunInfo, output *tool.CallbackOutput) context.Context {
			spanID := SpanIDFromContext(ctx)
			span := h.getSpan(spanID)
			if span != nil {
				span.End()
			}

			latencyMs := int64(0)
			if span != nil {
				latencyMs = span.Duration.Milliseconds()
			}

			h.metrics.RecordToolCall(info.Name, latencyMs, nil)

			result := extractToolResult(output)

			if h.eventHandler != nil {
				h.eventHandler("tool_end", events.EventMeta{
					Timestamp: time.Now().UTC(),
				}, map[string]any{
					"tool":          info.Name,
					"span_id":       spanID,
					"duration_ms":   latencyMs,
					"result_preview": truncate(result, 500),
				})
			}

			return ctx
		},
		OnError: func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
			spanID := SpanIDFromContext(ctx)
			span := h.getSpan(spanID)
			if span != nil {
				span.End()
			}

			latencyMs := int64(0)
			if span != nil {
				latencyMs = span.Duration.Milliseconds()
			}

			h.metrics.RecordToolCall(info.Name, latencyMs, err)

			if h.eventHandler != nil {
				h.eventHandler("tool_error", events.EventMeta{
					Timestamp: time.Now().UTC(),
				}, map[string]any{
					"tool":        info.Name,
					"span_id":     spanID,
					"error":       err.Error(),
					"duration_ms": latencyMs,
				})
			}

			return ctx
		},
		OnEndWithStreamOutput: func(ctx context.Context, info *callbacks.RunInfo, output *schema.StreamReader[*tool.CallbackOutput]) context.Context {
			spanID := SpanIDFromContext(ctx)

			go func() {
				var result strings.Builder
				start := time.Now()

				for {
					chunk, err := output.Recv()
					if err != nil {
						break
					}
					if chunk != nil {
						if chunk.Response != "" {
							result.WriteString(chunk.Response)
						}
						if chunk.ToolOutput != nil {
							result.WriteString(extractToolResultFromParts(chunk.ToolOutput.Parts))
						}
					}
				}

				span := h.getSpan(spanID)
				if span != nil {
					span.End()
				}

				latencyMs := time.Since(start).Milliseconds()
				h.metrics.RecordToolCall(info.Name, latencyMs, nil)

				if h.eventHandler != nil {
					h.eventHandler("tool_stream_end", events.EventMeta{
						Timestamp: time.Now().UTC(),
					}, map[string]any{
						"tool":        info.Name,
						"span_id":     spanID,
						"duration_ms": latencyMs,
						"result_len":  result.Len(),
					})
				}
			}()

			return ctx
		},
	}
}

// =============================================================================
// Agent 回调处理器
// =============================================================================

func (h *Handler) agentHandler() *template.AgentCallbackHandler {
	return &template.AgentCallbackHandler{
		OnStart: func(ctx context.Context, info *callbacks.RunInfo, input *adk.AgentCallbackInput) context.Context {
			span := h.tracer.StartSpan(ctx, "agent", info.Name)
			ctx = ContextWithSpanID(ctx, span.ID)

			messageCount := 0
			if input != nil && input.Input != nil {
				messageCount = len(input.Input.Messages)
			}

			if h.eventHandler != nil {
				h.eventHandler("agent_start", events.EventMeta{
					Timestamp: time.Now().UTC(),
				}, map[string]any{
					"agent":       info.Name,
					"span_id":     span.ID,
					"messages":    messageCount,
				})
			}

			return ctx
		},
		OnEnd: func(ctx context.Context, info *callbacks.RunInfo, output *adk.AgentCallbackOutput) context.Context {
			spanID := SpanIDFromContext(ctx)
			span := h.getSpan(spanID)
			if span != nil {
				span.End()
			}

			latencyMs := int64(0)
			if span != nil {
				latencyMs = span.Duration.Milliseconds()
			}

			// 从跨度元数据获取迭代次数
			iterations := 0
			if span != nil {
				if iter, ok := span.Metadata["iterations"].(int); ok {
					iterations = iter
				}
			}

			h.metrics.RecordAgentRun(info.Name, latencyMs, 0, iterations, nil)

			if h.eventHandler != nil {
				h.eventHandler("agent_end", events.EventMeta{
					Timestamp: time.Now().UTC(),
				}, map[string]any{
					"agent":       info.Name,
					"span_id":     spanID,
					"duration_ms": latencyMs,
					"iterations":  iterations,
				})
			}

			return ctx
		},
	}
}

// =============================================================================
// 辅助方法
// =============================================================================

func (h *Handler) getSpan(spanID string) *Span {
	if h == nil || h.tracer == nil || spanID == "" {
		return nil
	}
	h.tracer.mu.RLock()
	defer h.tracer.mu.RUnlock()
	return h.tracer.spans[spanID]
}

// truncate 截断字符串。
func truncate(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if maxLen <= 0 || len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// truncateMessages 截断消息列表预览。
func truncateMessages(messages []*schema.Message, maxLen int) string {
	if len(messages) == 0 {
		return ""
	}

	var preview strings.Builder
	for i, msg := range messages {
		if i > 0 {
			preview.WriteString(" | ")
		}
		content := truncate(msg.Content, maxLen/len(messages))
		preview.WriteString(string(msg.Role) + ": " + content)
	}

	return truncate(preview.String(), maxLen)
}

// estimateTokens 估算 token 数量。
// 简单实现：中文约 2 字符/token，英文约 4 字符/token
func estimateTokens(content string) int {
	content = strings.TrimSpace(content)
	if content == "" {
		return 0
	}

	// 简单估算：平均 3 字符一个 token
	return len(content) / 3
}

// extractToolResult 从工具回调输出中提取结果文本。
func extractToolResult(output *tool.CallbackOutput) string {
	if output == nil {
		return ""
	}
	if output.Response != "" {
		return output.Response
	}
	if output.ToolOutput != nil {
		return extractToolResultFromParts(output.ToolOutput.Parts)
	}
	return ""
}

// extractToolResultFromParts 从工具输出部分中提取文本。
func extractToolResultFromParts(parts []schema.ToolOutputPart) string {
	if len(parts) == 0 {
		return ""
	}
	var texts []string
	for _, part := range parts {
		if part.Type == schema.ToolPartTypeText && part.Text != "" {
			texts = append(texts, part.Text)
		}
	}
	return strings.Join(texts, "\n")
}

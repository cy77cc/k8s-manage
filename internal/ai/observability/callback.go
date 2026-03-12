// Package observability 提供 AI 编排层的可观测性能力。
//
// 本文件实现基于 Eino Callback 机制的可观测性处理器，
// 支持 LLM 调用、工具调用、Agent 运行的追踪和指标收集。
// 同时将指标推送到 Prometheus，追踪数据存储到数据库。
package observability

import (
	"context"
	"strings"
	"time"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/callbacks"
	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	template "github.com/cloudwego/eino/utils/callbacks"

	"github.com/cy77cc/OpsPilot/internal/ai/events"
	dbmodel "github.com/cy77cc/OpsPilot/internal/model"
)

// EventHandler 事件处理器函数类型。
type EventHandler func(name string, meta events.EventMeta, data map[string]any)

// Handler 可观测性处理器，实现 Eino 回调接口。
type Handler struct {
	tracer       *Tracer
	traceStore   *TraceStore
	eventHandler EventHandler
}

// NewHandler 创建新的可观测性处理器。
func NewHandler(traceStore *TraceStore, eventHandler EventHandler) *Handler {
	return &Handler{
		tracer:       NewTracer(),
		traceStore:   traceStore,
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

// =============================================================================
// ChatModel 回调处理器
// =============================================================================

type llmContext struct {
	spanID    string
	startTime time.Time
	sessionID string
	traceID   string
}

func (h *Handler) chatModelHandler() *template.ModelCallbackHandler {
	return &template.ModelCallbackHandler{
		OnStart: func(ctx context.Context, info *callbacks.RunInfo, input *einomodel.CallbackInput) context.Context {
			span := h.tracer.StartSpan(ctx, "llm", info.Name)
			ctx = ContextWithSpanID(ctx, span.ID)

			var sessionID, traceID string
			for _, msg := range input.Messages {
				if strings.Contains(msg.Content, `"session_id"`) {
					sessionID = extractFieldFromJSON(msg.Content, "session_id")
					traceID = extractFieldFromJSON(msg.Content, "trace_id")
					break
				}
			}

			lctx := &llmContext{
				spanID:    span.ID,
				startTime: time.Now().UTC(),
				sessionID: sessionID,
				traceID:   traceID,
			}
			ctx = context.WithValue(ctx, llmCtxKey{}, lctx)

			if h.eventHandler != nil {
				h.eventHandler("llm_start", events.EventMeta{
					Timestamp: time.Now().UTC(),
				}, map[string]any{
					"model":           info.Name,
					"component":       string(info.Component),
					"span_id":         span.ID,
					"messages":        len(input.Messages),
					"message_preview": truncateMessages(input.Messages, 200),
				})
			}

			return ctx
		},
		OnEnd: func(ctx context.Context, info *callbacks.RunInfo, output *einomodel.CallbackOutput) context.Context {
			spanID := SpanIDFromContext(ctx)
			span := h.getSpan(spanID)
			if span != nil {
				span.End()
			}

			lctx, _ := ctx.Value(llmCtxKey{}).(*llmContext)
			content := ""
			if output.Message != nil {
				content = output.Message.Content
			}
			tokens := estimateTokens(content)
			var latencySeconds float64
			var latencyMs int64
			if lctx != nil {
				latencySeconds = time.Since(lctx.startTime).Seconds()
				latencyMs = int64(latencySeconds * 1000)
			}

			RecordLLMMetric(info.Name, 0, tokens, latencySeconds, nil)

			if h.traceStore != nil && lctx != nil {
				traceSpan := BuildSpan(
					spanID,
					dbmodel.SpanTypeLLM,
					info.Name,
					lctx.sessionID,
					lctx.traceID,
					"",
					lctx.startTime,
					latencyMs,
					dbmodel.SpanStatusSuccess,
					"",
					"",
					truncateString(content, 4000),
					int64(tokens),
					nil,
				)
				h.traceStore.SaveSpanAsync(traceSpan)
			}

			if h.eventHandler != nil {
				h.eventHandler("llm_end", events.EventMeta{
					Timestamp: time.Now().UTC(),
				}, map[string]any{
					"model":            info.Name,
					"span_id":          spanID,
					"tokens":           tokens,
					"duration_ms":      latencyMs,
					"content_len":      len(content),
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

			lctx, _ := ctx.Value(llmCtxKey{}).(*llmContext)
			var latencySeconds float64
			var latencyMs int64
			if lctx != nil {
				latencySeconds = time.Since(lctx.startTime).Seconds()
				latencyMs = int64(latencySeconds * 1000)
			}

			RecordLLMMetric(info.Name, 0, 0, latencySeconds, err)

			if h.traceStore != nil && lctx != nil {
				traceSpan := BuildSpan(
					spanID,
					dbmodel.SpanTypeLLM,
					info.Name,
					lctx.sessionID,
					lctx.traceID,
					"",
					lctx.startTime,
					latencyMs,
					dbmodel.SpanStatusError,
					err.Error(),
					"",
					"",
					0,
					nil,
				)
				h.traceStore.SaveSpanAsync(traceSpan)
			}

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
		OnEndWithStreamOutput: func(ctx context.Context, info *callbacks.RunInfo, output *schema.StreamReader[*einomodel.CallbackOutput]) context.Context {
			spanID := SpanIDFromContext(ctx)
			lctx, _ := ctx.Value(llmCtxKey{}).(*llmContext)

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
				latencySeconds := time.Since(start).Seconds()
				latencyMs := int64(latencySeconds * 1000)

				RecordLLMMetric(info.Name, 0, tokens, latencySeconds, nil)

				if h.traceStore != nil && lctx != nil {
					traceSpan := BuildSpan(
						spanID,
						dbmodel.SpanTypeLLM,
						info.Name,
						lctx.sessionID,
						lctx.traceID,
						"",
						lctx.startTime,
						latencyMs,
						dbmodel.SpanStatusSuccess,
						"",
						"",
						truncateString(content.String(), 4000),
						int64(tokens),
						map[string]any{"streaming": true},
					)
					h.traceStore.SaveSpanAsync(traceSpan)
				}

				if h.eventHandler != nil {
					h.eventHandler("llm_stream_end", events.EventMeta{
						Timestamp: time.Now().UTC(),
					}, map[string]any{
						"model":       info.Name,
						"span_id":     spanID,
						"tokens":      tokens,
						"duration_ms": latencyMs,
						"content_len": content.Len(),
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

type toolContext struct {
	spanID    string
	startTime time.Time
	sessionID string
	traceID   string
}

func (h *Handler) toolHandler() *template.ToolCallbackHandler {
	return &template.ToolCallbackHandler{
		OnStart: func(ctx context.Context, info *callbacks.RunInfo, input *tool.CallbackInput) context.Context {
			span := h.tracer.StartSpan(ctx, "tool", info.Name)
			ctx = ContextWithSpanID(ctx, span.ID)

			tctx := &toolContext{
				spanID:    span.ID,
				startTime: time.Now().UTC(),
			}
			ctx = context.WithValue(ctx, toolCtxKey{}, tctx)

			if h.eventHandler != nil {
				h.eventHandler("tool_start", events.EventMeta{
					Timestamp: time.Now().UTC(),
				}, map[string]any{
					"tool":      info.Name,
					"span_id":   span.ID,
					"arguments": truncate(input.ArgumentsInJSON, 1000),
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

			tctx, _ := ctx.Value(toolCtxKey{}).(*toolContext)
			var latencySeconds float64
			var latencyMs int64
			if tctx != nil {
				latencySeconds = time.Since(tctx.startTime).Seconds()
				latencyMs = int64(latencySeconds * 1000)
			}

			result := extractToolResult(output)

			RecordToolMetric(info.Name, latencySeconds, nil)

			if h.traceStore != nil && tctx != nil {
				traceSpan := BuildSpan(
					spanID,
					dbmodel.SpanTypeTool,
					info.Name,
					tctx.sessionID,
					tctx.traceID,
					"",
					tctx.startTime,
					latencyMs,
					dbmodel.SpanStatusSuccess,
					"",
					"",
					truncateString(result, 4000),
					0,
					nil,
				)
				h.traceStore.SaveSpanAsync(traceSpan)
			}

			if h.eventHandler != nil {
				h.eventHandler("tool_end", events.EventMeta{
					Timestamp: time.Now().UTC(),
				}, map[string]any{
					"tool":           info.Name,
					"span_id":        spanID,
					"duration_ms":    latencyMs,
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

			tctx, _ := ctx.Value(toolCtxKey{}).(*toolContext)
			var latencySeconds float64
			var latencyMs int64
			if tctx != nil {
				latencySeconds = time.Since(tctx.startTime).Seconds()
				latencyMs = int64(latencySeconds * 1000)
			}

			RecordToolMetric(info.Name, latencySeconds, err)

			if h.traceStore != nil && tctx != nil {
				traceSpan := BuildSpan(
					spanID,
					dbmodel.SpanTypeTool,
					info.Name,
					tctx.sessionID,
					tctx.traceID,
					"",
					tctx.startTime,
					latencyMs,
					dbmodel.SpanStatusError,
					err.Error(),
					"",
					"",
					0,
					nil,
				)
				h.traceStore.SaveSpanAsync(traceSpan)
			}

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
			tctx, _ := ctx.Value(toolCtxKey{}).(*toolContext)

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

				latencySeconds := time.Since(start).Seconds()
				latencyMs := int64(latencySeconds * 1000)

				RecordToolMetric(info.Name, latencySeconds, nil)

				if h.traceStore != nil && tctx != nil {
					traceSpan := BuildSpan(
						spanID,
						dbmodel.SpanTypeTool,
						info.Name,
						tctx.sessionID,
						tctx.traceID,
						"",
						tctx.startTime,
						latencyMs,
						dbmodel.SpanStatusSuccess,
						"",
						"",
						truncateString(result.String(), 4000),
						0,
						map[string]any{"streaming": true},
					)
					h.traceStore.SaveSpanAsync(traceSpan)
				}

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

type agentContext struct {
	spanID    string
	startTime time.Time
	sessionID string
	traceID   string
}

func (h *Handler) agentHandler() *template.AgentCallbackHandler {
	return &template.AgentCallbackHandler{
		OnStart: func(ctx context.Context, info *callbacks.RunInfo, input *adk.AgentCallbackInput) context.Context {
			span := h.tracer.StartSpan(ctx, "agent", info.Name)
			ctx = ContextWithSpanID(ctx, span.ID)

			messageCount := 0
			if input != nil && input.Input != nil {
				messageCount = len(input.Input.Messages)
			}

			actx := &agentContext{
				spanID:    span.ID,
				startTime: time.Now().UTC(),
			}
			ctx = context.WithValue(ctx, agentCtxKey{}, actx)

			if h.eventHandler != nil {
				h.eventHandler("agent_start", events.EventMeta{
					Timestamp: time.Now().UTC(),
				}, map[string]any{
					"agent":    info.Name,
					"span_id":  span.ID,
					"messages": messageCount,
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

			actx, _ := ctx.Value(agentCtxKey{}).(*agentContext)
			var latencySeconds float64
			var latencyMs int64
			iterations := 0
			if actx != nil {
				latencySeconds = time.Since(actx.startTime).Seconds()
				latencyMs = int64(latencySeconds * 1000)
			}

			if span != nil {
				if iter, ok := span.Metadata["iterations"].(int); ok {
					iterations = iter
				}
			}

			RecordAgentMetric(info.Name, latencySeconds, iterations, nil)

			if h.traceStore != nil && actx != nil {
				traceSpan := BuildSpan(
					spanID,
					dbmodel.SpanTypeAgent,
					info.Name,
					actx.sessionID,
					actx.traceID,
					"",
					actx.startTime,
					latencyMs,
					dbmodel.SpanStatusSuccess,
					"",
					"",
					"",
					0,
					map[string]any{"iterations": iterations},
				)
				h.traceStore.SaveSpanAsync(traceSpan)
			}

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
// 辅助类型和方法
// =============================================================================

type llmCtxKey struct{}
type toolCtxKey struct{}
type agentCtxKey struct{}

func (h *Handler) getSpan(spanID string) *Span {
	if h == nil || h.tracer == nil || spanID == "" {
		return nil
	}
	h.tracer.mu.RLock()
	defer h.tracer.mu.RUnlock()
	return h.tracer.spans[spanID]
}

func truncate(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if maxLen <= 0 || len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

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

func estimateTokens(content string) int {
	content = strings.TrimSpace(content)
	if content == "" {
		return 0
	}
	return len(content) / 3
}

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

func extractFieldFromJSON(jsonStr, field string) string {
	needle := `"` + field + `":"`
	start := strings.Index(jsonStr, needle)
	if start == -1 {
		return ""
	}
	start += len(needle)
	end := strings.Index(jsonStr[start:], `"`)
	if end == -1 {
		return ""
	}
	return jsonStr[start : start+end]
}

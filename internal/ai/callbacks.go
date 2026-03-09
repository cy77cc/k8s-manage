package ai

import (
	"context"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components"
	cbtool "github.com/cloudwego/eino/components/tool"
	"github.com/google/uuid"
)

// SSEEmitter 是 SSE 事件发射函数类型。
// 参数 event 是事件名称，payload 是事件负载。
// 返回 false 表示发射失败（如客户端已断开）。
type SSEEmitter func(event string, payload map[string]any) bool

// streamingCallbacks 捕获工具执行事件并发射 SSE 事件。
type streamingCallbacks struct {
	emit SSEEmitter
}

// NewStreamingCallbacks 创建一个回调处理器，用于发射工具执行的 SSE 事件。
//
// 参数:
//   - emit: SSE 事件发射函数。
//
// 返回:
//   - callbacks.Handler: 回调处理器。
func NewStreamingCallbacks(emit SSEEmitter) callbacks.Handler {
	return callbacks.NewHandlerBuilder().
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			if info == nil || emit == nil {
				return ctx
			}
			// Only handle tool component callbacks
			if info.Component != components.ComponentOfTool {
				return ctx
			}

			// Extract tool output
			toolOutput := cbtool.ConvCallbackOutput(output)
			if toolOutput == nil {
				return ctx
			}

			// Emit tool_result event
			toolName := info.Name
			if toolName == "" {
				toolName = "unknown"
			}

			emit("tool_result", map[string]any{
				"call_id":   "call-" + uuid.NewString(),
				"tool_name": toolName,
				"result":    toolOutput.Response,
				"status":    "success",
			})

			return ctx
		}).
		OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
			if info == nil || emit == nil {
				return ctx
			}
			if info.Component != components.ComponentOfTool {
				return ctx
			}

			toolName := info.Name
			if toolName == "" {
				toolName = "unknown"
			}

			emit("tool_result", map[string]any{
				"call_id":   "call-" + uuid.NewString(),
				"tool_name": toolName,
				"error":     err.Error(),
				"status":    "error",
			})

			return ctx
		}).
		Build()
}

// callbacksKey 是存储 SSE 发射器的上下文键。
type callbacksKey struct{}

// WithSSEEmitter 将 SSE 发射器存储到上下文中。
//
// 参数:
//   - ctx: 上下文。
//   - emit: SSE 发射函数。
//
// 返回:
//   - context.Context: 包含 SSE 发射器的上下文。
func WithSSEEmitter(ctx context.Context, emit SSEEmitter) context.Context {
	return context.WithValue(ctx, callbacksKey{}, emit)
}

// SSEEmitterFromContext 从上下文中获取 SSE 发射器。
//
// 参数:
//   - ctx: 上下文。
//
// 返回:
//   - SSEEmitter: SSE 发射函数，如果不存在则返回 nil。
func SSEEmitterFromContext(ctx context.Context) SSEEmitter {
	if emit, ok := ctx.Value(callbacksKey{}).(SSEEmitter); ok {
		return emit
	}
	return nil
}

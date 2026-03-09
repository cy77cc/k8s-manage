package ai

import (
	"context"
	"time"

	"github.com/cy77cc/k8s-manage/internal/service/ai/logic"
)

// ChatStreamRequest 封装聊天流请求的参数。
type ChatStreamRequest struct {
	// UserID 是发起请求的用户 ID。
	UserID uint64
	// SessionID 是会话 ID，可选。
	SessionID string
	// Message 是用户消息内容。
	Message string
	// Context 是运行时上下文，包含场景、命名空间等信息。
	Context map[string]any
}

// Orchestrator 编排 AI 对话流程，协调 Agent、会话存储、运行时存储和控制平面。
type Orchestrator struct {
	// ai 是 AI Agent 实例。
	ai *AIAgent
	// sessions 管理会话状态。
	sessions *logic.SessionStore
	// runtime 管理运行时上下文和执行记录。
	runtime *logic.RuntimeStore
	// control 提供权限检查和策略控制。
	control *ControlPlane
}

// NewOrchestrator 创建一个新的编排器实例。
//
// 参数:
//   - ai: AI Agent 实例。
//   - sessions: 会话存储。
//   - runtime: 运行时存储。
//   - control: 控制平面。
//
// 返回:
//   - *Orchestrator: 编排器实例。
func NewOrchestrator(ai *AIAgent, sessions *logic.SessionStore, runtime *logic.RuntimeStore, control *ControlPlane) *Orchestrator {
	return &Orchestrator{ai: ai, sessions: sessions, runtime: runtime, control: control}
}

// ChatStream 执行流式聊天。
// 处理会话管理、心跳、流式执行和中断恢复。
//
// 参数:
//   - ctx: 上下文。
//   - req: 聊天请求。
//   - emit: SSE 事件发射函数。
//
// 返回:
//   - error: 执行错误。
//
// SSE 事件类型:
//   - meta: 会话元信息。
//   - heartbeat: 心跳事件。
//   - delta/thinking_delta/tool_call/tool_result/approval_required: 来自 Agent。
//   - done: 完成事件。
//   - error: 错误事件。
func (o *Orchestrator) ChatStream(ctx context.Context, req ChatStreamRequest, emit func(event string, payload map[string]any) bool) error {
	scene := logic.NormalizeScene(logic.ToString(req.Context["scene"]))
	session := o.sessions.Ensure(req.UserID, scene)
	if req.SessionID != "" {
		session.ID = req.SessionID
		o.sessions.Put(session)
	}
	o.runtime.RememberContext(req.UserID, session.Scene, req.Context)

	// 发送 meta 事件
	emit("meta", map[string]any{"sessionId": session.ID})

	// 启动心跳
	stopHeartbeat := make(chan struct{})
	go heartbeatLoop(ctx, stopHeartbeat, emit)
	defer close(stopHeartbeat)

	// 流式执行
	result, err := o.ai.Stream(ctx, session.ID, req.Message, req.UserID, req.Context, emit)
	if err != nil {
		emit("error", map[string]any{"message": err.Error()})
		return err
	}

	// 检查是否中断
	if result.CheckpointID != "" {
		emit("done", map[string]any{
			"stream_state":  "interrupted",
			"checkpoint_id": result.CheckpointID,
		})
		return nil
	}

	emit("done", map[string]any{"stream_state": "ok"})
	return nil
}

// ResumePayload 从检查点恢复执行。
// 用于处理审批后的恢复流程。
//
// 参数:
//   - ctx: 上下文。
//   - checkpointID: 检查点 ID。
//   - targets: 恢复目标，通常包含审批结果。
//
// 返回:
//   - map[string]any: 恢复执行的结果。
//   - error: 恢复错误。
func (o *Orchestrator) ResumePayload(ctx context.Context, checkpointID string, targets map[string]any) (map[string]any, error) {
	// 提取审批结果
	approval := any(targets)
	if v, ok := targets["approval"]; ok {
		approval = v
	}

	// 执行恢复
	result, err := o.ai.Resume(ctx, checkpointID, approval, nil)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"checkpoint_id": checkpointID,
		"resumed":       true,
		"response":      result.Content,
		"tool_calls":    result.ToolCalls,
	}, nil
}

// heartbeatLoop 运行心跳循环，定期发送心跳事件。
// 用于保持 SSE 连接活跃，防止超时断开。
//
// 参数:
//   - ctx: 上下文，取消时停止循环。
//   - stop: 停止信号通道。
//   - emit: SSE 事件发射函数。
func heartbeatLoop(ctx context.Context, stop <-chan struct{}, emit func(event string, payload map[string]any) bool) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-stop:
			return
		case t := <-ticker.C:
			if emit != nil && !emit("heartbeat", map[string]any{"ts": t.UTC().Format(time.RFC3339Nano)}) {
				return
			}
		}
	}
}

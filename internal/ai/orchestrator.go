package ai

import (
	"context"

	"github.com/cy77cc/k8s-manage/internal/service/ai/logic"
)

type ChatStreamRequest struct {
	UserID    uint64
	SessionID string
	Message   string
	Context   map[string]any
}

type Orchestrator struct {
	ai       *AIAgent
	sessions *logic.SessionStore
	runtime  *logic.RuntimeStore
	control  *ControlPlane
}

func NewOrchestrator(ai *AIAgent, sessions *logic.SessionStore, runtime *logic.RuntimeStore, control *ControlPlane) *Orchestrator {
	return &Orchestrator{ai: ai, sessions: sessions, runtime: runtime, control: control}
}

func (o *Orchestrator) ChatStream(ctx context.Context, req ChatStreamRequest, emit func(event string, payload map[string]any) bool) error {
	scene := logic.NormalizeScene(logic.ToString(req.Context["scene"]))
	session := o.sessions.Ensure(req.UserID, scene)
	if req.SessionID != "" {
		session.ID = req.SessionID
		o.sessions.Put(session)
	}
	o.runtime.RememberContext(req.UserID, session.Scene, req.Context)
	emit("meta", map[string]any{"sessionId": session.ID})
	emit(EventPlanCreated, map[string]any{"plan_id": "plan-" + session.ID, "objective": req.Message})
	out, err := o.ai.Query(ctx, session.ID, req.Message)
	if err != nil {
		return err
	}
	emit("message", map[string]any{"role": "assistant", "content": out.Response})
	emit("done", map[string]any{"stream_state": "ok"})
	return nil
}

func (o *Orchestrator) ResumePayload(ctx context.Context, checkpointID string, targets map[string]any) (map[string]any, error) {
	return o.ai.Resume(ctx, checkpointID, targets)
}

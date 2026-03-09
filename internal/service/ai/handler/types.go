// Package handler provides AI service HTTP handlers
package handler

import (
	"context"

	"github.com/cloudwego/eino/schema"
	coreai "github.com/cy77cc/k8s-manage/internal/ai"
	"github.com/cy77cc/k8s-manage/internal/ai/tools/core"
	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/service/ai/logic"
	"github.com/cy77cc/k8s-manage/internal/svc"
)

// ChatRequest represents the request body for chat endpoint
type ChatRequest struct {
	SessionID string         `json:"sessionId"`
	Message   string         `json:"message" binding:"required"`
	Context   map[string]any `json:"context"`
}

type aiToolRunner interface {
	ToolMetas() []core.ToolMeta
	RunTool(ctx context.Context, toolName string, params map[string]any) (core.ToolResult, error)
	Generate(ctx context.Context, messages []*schema.Message) (*schema.Message, error)
}

type aiOrchestrator interface {
	ChatStream(ctx context.Context, req coreai.ChatStreamRequest, emit func(event string, payload map[string]any) bool) error
	ResumePayload(ctx context.Context, checkpointID string, targets map[string]any) (map[string]any, error)
}

type aiControlPlane interface {
	ToolPolicy(ctx context.Context, meta core.ToolMeta, params map[string]any) error
	HasPermission(uid uint64, code string) bool
	IsAdmin(uid uint64) bool
	FindMeta(name string) (core.ToolMeta, bool)
}

type aiGatewayRuntime interface {
	ListSessions(uid uint64, scene string) []*logic.AISession
	CurrentSession(uid uint64, scene string) (*logic.AISession, bool)
	GetSession(uid uint64, id string) (*logic.AISession, bool)
	BranchSession(uid uint64, sourceSessionID, anchorMessageID, title string) (*logic.AISession, error)
	DeleteSession(uid uint64, id string)
	UpdateSessionTitle(uid uint64, id, title string) (*logic.AISession, error)
	PreviewTool(uid uint64, tool string, params map[string]any) (map[string]any, error)
	ExecuteTool(ctx context.Context, uid uint64, tool string, params map[string]any, approvalToken string) (*logic.ExecutionRecord, error)
	GetExecution(id string) (*logic.ExecutionRecord, bool)
	CreateApproval(uid uint64, tool string, params map[string]any) (*logic.ApprovalTicket, error)
	ConfirmApproval(uid uint64, id string, approve bool) (*logic.ApprovalTicket, error)
	CreateApprovalTask(ctx context.Context, uid uint64, tool string, params map[string]any) (*model.AIApprovalTask, error)
	ListApprovalTasks(ctx context.Context, uid uint64, status string) ([]model.AIApprovalTask, error)
	GetApprovalTask(ctx context.Context, uid uint64, id string) (*model.AIApprovalTask, error)
	ReviewApprovalTask(ctx context.Context, uid uint64, id string, approve bool, reason string) (*model.AIApprovalTask, *logic.ExecutionRecord, error)
	ConfirmConfirmation(ctx context.Context, uid uint64, id string, approve bool) (*model.ConfirmationRequest, error)
}

// AIHandler handles AI service HTTP requests
type AIHandler struct {
	svcCtx       *svc.ServiceContext
	ai           aiToolRunner
	orchestrator aiOrchestrator
	control      aiControlPlane
	gateway      aiGatewayRuntime
	sessions     *logic.SessionStore
	runtime      *logic.RuntimeStore
}

// NewAIHandler creates a new AIHandler instance
func NewAIHandler(svcCtx *svc.ServiceContext) *AIHandler {
	sessions := logic.NewSessionStore(svcCtx.DB, svcCtx.Rdb)
	runtime := logic.NewRuntimeStore(svcCtx.DB)
	control := coreai.NewControlPlane(svcCtx.DB, runtime, svcCtx.AI)
	return &AIHandler{
		svcCtx:       svcCtx,
		ai:           svcCtx.AI,
		orchestrator: coreai.NewOrchestrator(svcCtx.AI, sessions, runtime, control),
		control:      control,
		gateway:      coreai.NewGatewayRuntime(svcCtx.DB, svcCtx.AI, control, sessions, runtime),
		sessions:     sessions,
		runtime:      runtime,
	}
}

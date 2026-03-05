package ai

import (
	"context"
	"sync"
	"time"

	askills "github.com/cy77cc/k8s-manage/internal/ai/skills"
	"github.com/cy77cc/k8s-manage/internal/ai/tools"
	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"gorm.io/gorm"
)

type chatRequest struct {
	SessionID string         `json:"sessionId"`
	Message   string         `json:"message" binding:"required"`
	Context   map[string]any `json:"context"`
}

const defaultAISessionTitle = "AI Session"

type aiSession struct {
	ID        string           `json:"id"`
	Scene     string           `json:"scene,omitempty"`
	Title     string           `json:"title"`
	Messages  []map[string]any `json:"messages"`
	CreatedAt time.Time        `json:"createdAt"`
	UpdatedAt time.Time        `json:"updatedAt"`
}

type recommendationRecord struct {
	ID             string    `json:"id"`
	UserID         uint64    `json:"userId"`
	Scene          string    `json:"scene"`
	Type           string    `json:"type"`
	Title          string    `json:"title"`
	Content        string    `json:"content"`
	FollowupPrompt string    `json:"followup_prompt,omitempty"`
	Reasoning      string    `json:"reasoning,omitempty"`
	Relevance      float64   `json:"relevance"`
	CreatedAt      time.Time `json:"createdAt"`
}

type approvalTicket struct {
	ID         string         `json:"id"`
	Tool       string         `json:"tool"`
	Params     map[string]any `json:"params"`
	Risk       tools.ToolRisk `json:"risk"`
	Mode       tools.ToolMode `json:"mode"`
	Status     string         `json:"status"`
	CreatedAt  time.Time      `json:"createdAt"`
	ExpiresAt  time.Time      `json:"expiresAt"`
	RequestUID uint64         `json:"requestUid"`
	ReviewUID  uint64         `json:"reviewUid,omitempty"`
	Meta       tools.ToolMeta `json:"-"`
}

type executionRecord struct {
	ID         string            `json:"id"`
	Tool       string            `json:"tool"`
	Params     map[string]any    `json:"params"`
	Mode       tools.ToolMode    `json:"mode"`
	Status     string            `json:"status"`
	Result     *tools.ToolResult `json:"result,omitempty"`
	ApprovalID string            `json:"approvalId,omitempty"`
	RequestUID uint64            `json:"requestUid"`
	CreatedAt  time.Time         `json:"createdAt"`
	FinishedAt *time.Time        `json:"finishedAt,omitempty"`
	Error      string            `json:"error,omitempty"`
}

type handler struct {
	svcCtx         *svc.ServiceContext
	store          *memoryStore
	skillRegistry  *askills.Registry
	skillExecutor  *askills.Executor
}

type memoryStore struct {
	mu                sync.RWMutex
	db                *gorm.DB
	approvals         map[string]*approvalTicket
	executions        map[string]*executionRecord
	recommendations   map[string][]recommendationRecord
	toolParams        map[string]map[string]any
	commandAliases    map[string]map[string]string
	commandTemplates  map[string]map[string]map[string]any
	referencedContext map[string]map[string]any
}

func newHandler(svcCtx *svc.ServiceContext) *handler {
	h := &handler{
		svcCtx: svcCtx,
		store: &memoryStore{
			db:                svcCtx.DB,
			approvals:         map[string]*approvalTicket{},
			executions:        map[string]*executionRecord{},
			recommendations:   map[string][]recommendationRecord{},
			toolParams:        map[string]map[string]any{},
			commandAliases:    map[string]map[string]string{},
			commandTemplates:  map[string]map[string]map[string]any{},
			referencedContext: map[string]map[string]any{},
		},
	}
	if registry, err := askills.NewRegistry(askills.DefaultSkillsConfigPath); err == nil {
		h.skillRegistry = registry
	}
	if svcCtx != nil && svcCtx.AI != nil {
		h.skillExecutor = askills.NewExecutor(func(ctx context.Context, toolName string, params map[string]any) (tools.ToolResult, error) {
			return svcCtx.AI.RunTool(ctx, toolName, params)
		}, nil, nil)
	}
	return h
}

func (s *memoryStore) dbEnabled() bool { return s != nil && s.db != nil }

func toSessionModel(uid uint64, scene string, in *aiSession) *model.AIChatSession {
	return &model.AIChatSession{
		ID:        in.ID,
		UserID:    uid,
		Scene:     scene,
		Title:     in.Title,
		CreatedAt: in.CreatedAt,
		UpdatedAt: in.UpdatedAt,
	}
}

package ai

import (
	"sync"
	"time"

	ai2 "github.com/cy77cc/k8s-manage/internal/ai"
	"github.com/cy77cc/k8s-manage/internal/svc"
)

type chatRequest struct {
	SessionID string         `json:"sessionId"`
	Message   string         `json:"message" binding:"required"`
	Context   map[string]any `json:"context"`
}

type aiSession struct {
	ID        string           `json:"id"`
	Title     string           `json:"title"`
	Messages  []map[string]any `json:"messages"`
	CreatedAt time.Time        `json:"createdAt"`
	UpdatedAt time.Time        `json:"updatedAt"`
}

type approvalTicket struct {
	ID         string         `json:"id"`
	Tool       string         `json:"tool"`
	Params     map[string]any `json:"params"`
	Risk       ai2.ToolRisk   `json:"risk"`
	Mode       ai2.ToolMode   `json:"mode"`
	Status     string         `json:"status"`
	CreatedAt  time.Time      `json:"createdAt"`
	ExpiresAt  time.Time      `json:"expiresAt"`
	RequestUID uint64         `json:"requestUid"`
	ReviewUID  uint64         `json:"reviewUid,omitempty"`
	Meta       ai2.ToolMeta   `json:"-"`
}

type executionRecord struct {
	ID         string          `json:"id"`
	Tool       string          `json:"tool"`
	Params     map[string]any  `json:"params"`
	Mode       ai2.ToolMode    `json:"mode"`
	Status     string          `json:"status"`
	Result     *ai2.ToolResult `json:"result,omitempty"`
	ApprovalID string          `json:"approvalId,omitempty"`
	RequestUID uint64          `json:"requestUid"`
	CreatedAt  time.Time       `json:"createdAt"`
	FinishedAt *time.Time      `json:"finishedAt,omitempty"`
	Error      string          `json:"error,omitempty"`
}

type handler struct {
	svcCtx *svc.ServiceContext
	store  *memoryStore
}

type memoryStore struct {
	mu         sync.RWMutex
	sessions   map[string]*aiSession
	approvals  map[string]*approvalTicket
	executions map[string]*executionRecord
}

func newHandler(svcCtx *svc.ServiceContext) *handler {
	return &handler{
		svcCtx: svcCtx,
		store: &memoryStore{
			sessions:   map[string]*aiSession{},
			approvals:  map[string]*approvalTicket{},
			executions: map[string]*executionRecord{},
		},
	}
}

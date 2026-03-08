package events

import (
	"sync"
	"time"
)

type ApprovalUpdate struct {
	ID             string         `json:"id"`
	ApprovalToken  string         `json:"approval_token,omitempty"`
	ToolName       string         `json:"tool_name,omitempty"`
	Status         string         `json:"status"`
	RequestUserID  uint64         `json:"request_user_id,omitempty"`
	ApproverUserID uint64         `json:"approver_user_id,omitempty"`
	Execution      map[string]any `json:"execution,omitempty"`
	UpdatedAt      time.Time      `json:"updated_at,omitempty"`
}

type ApprovalHub struct {
	mu          sync.RWMutex
	subscribers map[uint64]map[chan ApprovalUpdate]struct{}
}

func NewApprovalHub() *ApprovalHub {
	return &ApprovalHub{subscribers: make(map[uint64]map[chan ApprovalUpdate]struct{})}
}

func (h *ApprovalHub) Subscribe(userID uint64, buffer int) (<-chan ApprovalUpdate, func()) {
	ch := make(chan ApprovalUpdate, buffer)
	h.mu.Lock()
	if h.subscribers[userID] == nil {
		h.subscribers[userID] = make(map[chan ApprovalUpdate]struct{})
	}
	h.subscribers[userID][ch] = struct{}{}
	h.mu.Unlock()
	return ch, func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		if subs := h.subscribers[userID]; subs != nil {
			delete(subs, ch)
			if len(subs) == 0 {
				delete(h.subscribers, userID)
			}
		}
		close(ch)
	}
}

func (h *ApprovalHub) Publish(update any, userIDs ...uint64) {
	payload, ok := update.(ApprovalUpdate)
	if !ok {
		if src, ok := update.(map[string]any); ok {
			payload = approvalUpdateFromMap(src)
		} else {
			return
		}
	}

	seen := make(map[uint64]struct{}, len(userIDs))
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, userID := range userIDs {
		if userID == 0 {
			continue
		}
		if _, exists := seen[userID]; exists {
			continue
		}
		seen[userID] = struct{}{}
		for ch := range h.subscribers[userID] {
			select {
			case ch <- payload:
			default:
			}
		}
	}
}

func approvalUpdateFromMap(src map[string]any) ApprovalUpdate {
	out := ApprovalUpdate{
		ID:             asString(src["id"]),
		ApprovalToken:  asString(src["approval_token"]),
		ToolName:       asString(src["tool_name"]),
		Status:         asString(src["status"]),
		RequestUserID:  asUint64(src["request_user_id"]),
		ApproverUserID: asUint64(src["approver_user_id"]),
	}
	if execution, ok := src["execution"].(map[string]any); ok {
		out.Execution = execution
	}
	if updatedAt, ok := src["updated_at"].(time.Time); ok {
		out.UpdatedAt = updatedAt
	}
	return out
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}

func asUint64(v any) uint64 {
	switch x := v.(type) {
	case uint64:
		return x
	case uint:
		return uint64(x)
	case int:
		if x > 0 {
			return uint64(x)
		}
	}
	return 0
}

var defaultApprovalHub = NewApprovalHub()

func DefaultApprovalHub() *ApprovalHub {
	return defaultApprovalHub
}

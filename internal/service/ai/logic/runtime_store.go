package logic

import (
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"
)

type RuntimeStore struct {
	db              *gorm.DB
	mu              sync.RWMutex
	remembered      map[string]map[string]any
	recommendations map[string][]RecommendationRecord
	executions      map[string]*ExecutionRecord
	approvals       map[string]*ApprovalTicket
	lastToolParams  map[string]map[string]any
}

func NewRuntimeStore(db *gorm.DB) *RuntimeStore {
	return &RuntimeStore{
		db:              db,
		remembered:      make(map[string]map[string]any),
		recommendations: make(map[string][]RecommendationRecord),
		executions:      make(map[string]*ExecutionRecord),
		approvals:       make(map[string]*ApprovalTicket),
		lastToolParams:  make(map[string]map[string]any),
	}
}

func (r *RuntimeStore) GetRememberedContext(uid uint64, scene string) map[string]any {
	r.mu.RLock()
	defer r.mu.RUnlock()
	src := r.remembered[r.key(uid, scene)]
	out := make(map[string]any, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

func (r *RuntimeStore) RememberContext(uid uint64, scene string, ctx map[string]any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := r.key(uid, scene)
	dst := make(map[string]any, len(ctx))
	for k, v := range ctx {
		dst[k] = v
	}
	r.remembered[key] = dst
}

func (r *RuntimeStore) SetRecommendations(uid uint64, scene string, items []RecommendationRecord) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := r.key(uid, scene)
	copied := make([]RecommendationRecord, len(items))
	copy(copied, items)
	r.recommendations[key] = copied
}

func (r *RuntimeStore) SaveExecution(rec *ExecutionRecord) {
	if r == nil || rec == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	c := *rec
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now()
	}
	r.executions[c.ID] = &c
}

func (r *RuntimeStore) GetExecution(id string) (*ExecutionRecord, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rec, ok := r.executions[id]
	if !ok {
		return nil, false
	}
	c := *rec
	return &c, true
}

func (r *RuntimeStore) SaveApproval(ticket *ApprovalTicket) {
	if r == nil || ticket == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	c := *ticket
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now()
	}
	r.approvals[c.ID] = &c
}

func (r *RuntimeStore) CreateApprovalTask(ticket *ApprovalTicket) {
	r.SaveApproval(ticket)
}

func (r *RuntimeStore) GetApproval(id string) (*ApprovalTicket, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, ok := r.approvals[id]
	if !ok {
		return nil, false
	}
	c := *item
	return &c, true
}

func (r *RuntimeStore) GetApprovalTask(id string) (*ApprovalTicket, bool) {
	return r.GetApproval(id)
}

func (r *RuntimeStore) UpdateApproval(id, status string) (*ApprovalTicket, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	item, ok := r.approvals[id]
	if !ok {
		return nil, false
	}
	item.Status = status
	c := *item
	return &c, true
}

func (r *RuntimeStore) UpdateApprovalTask(id, status string) (*ApprovalTicket, bool) {
	return r.UpdateApproval(id, status)
}

func (r *RuntimeStore) DeleteApprovalTask(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.approvals, id)
}

type ToolMemoryAccessor struct {
	runtime *RuntimeStore
	key     string
}

func NewToolMemoryAccessor(runtime *RuntimeStore, uid uint64, scene string) *ToolMemoryAccessor {
	return &ToolMemoryAccessor{
		runtime: runtime,
		key:     runtime.key(uid, scene),
	}
}

func (a *ToolMemoryAccessor) GetLastToolParams(toolName string) map[string]any {
	if a == nil || a.runtime == nil {
		return nil
	}
	a.runtime.mu.RLock()
	defer a.runtime.mu.RUnlock()
	src := a.runtime.lastToolParams[a.key+":"+toolName]
	out := make(map[string]any, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

func (a *ToolMemoryAccessor) SetLastToolParams(toolName string, params map[string]any) {
	if a == nil || a.runtime == nil {
		return
	}
	a.runtime.mu.Lock()
	defer a.runtime.mu.Unlock()
	dst := make(map[string]any, len(params))
	for k, v := range params {
		dst[k] = v
	}
	a.runtime.lastToolParams[a.key+":"+toolName] = dst
}

func (r *RuntimeStore) key(uid uint64, scene string) string {
	return fmt.Sprintf("%d:%s", uid, NormalizeScene(scene))
}

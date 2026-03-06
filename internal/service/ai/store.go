package ai

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

func (s *runtimeStore) setRecommendations(userID uint64, scene string, recs []recommendationRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := recommendationKey(userID, scene)
	s.recommendations[key] = recs
}

func (s *runtimeStore) getRecommendations(userID uint64, scene string, limit int) []recommendationRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	key := recommendationKey(userID, scene)
	list := s.recommendations[key]
	if len(list) == 0 {
		return nil
	}
	cp := make([]recommendationRecord, 0, len(list))
	cp = append(cp, list...)
	sort.Slice(cp, func(i, j int) bool { return cp[i].CreatedAt.After(cp[j].CreatedAt) })
	if limit > 0 && len(cp) > limit {
		cp = cp[:limit]
	}
	return cp
}

func recommendationKey(userID uint64, scene string) string {
	return fmt.Sprintf("%d:%s", userID, normalizeScene(scene))
}

func toolParamKey(userID uint64, scene, tool string) string {
	return fmt.Sprintf("%d:%s:%s", userID, normalizeScene(scene), strings.TrimSpace(tool))
}

func referencedContextKey(userID uint64, scene string) string {
	return fmt.Sprintf("%d:%s", userID, normalizeScene(scene))
}

func normalizeScene(scene string) string {
	v := strings.TrimSpace(scene)
	if v == "" {
		return "global"
	}
	return v
}

func (s *runtimeStore) rememberContext(userID uint64, scene string, ctx map[string]any) {
	if s == nil || len(ctx) == 0 {
		return
	}
	key := referencedContextKey(userID, scene)
	s.mu.Lock()
	if s.referencedContext[key] == nil {
		s.referencedContext[key] = map[string]any{}
	}
	for k, v := range ctx {
		if strings.TrimSpace(k) == "" || v == nil {
			continue
		}
		if strings.TrimSpace(toString(v)) == "" {
			continue
		}
		s.referencedContext[key][k] = v
	}
	s.mu.Unlock()
}

func (s *runtimeStore) getRememberedContext(userID uint64, scene string) map[string]any {
	if s == nil {
		return map[string]any{}
	}
	key := referencedContextKey(userID, scene)
	s.mu.RLock()
	defer s.mu.RUnlock()
	raw := s.referencedContext[key]
	out := map[string]any{}
	for k, v := range raw {
		out[k] = v
	}
	return out
}

type toolMemoryAccessor struct {
	store *runtimeStore
	uid   uint64
	scene string
}

func (a *toolMemoryAccessor) GetLastToolParams(toolName string) map[string]any {
	if a == nil || a.store == nil {
		return nil
	}
	a.store.mu.RLock()
	defer a.store.mu.RUnlock()
	v := a.store.toolParams[toolParamKey(a.uid, a.scene, toolName)]
	if len(v) == 0 {
		return nil
	}
	out := map[string]any{}
	for k, val := range v {
		out[k] = val
	}
	return out
}

func (a *toolMemoryAccessor) SetLastToolParams(toolName string, params map[string]any) {
	if a == nil || a.store == nil || strings.TrimSpace(toolName) == "" || len(params) == 0 {
		return
	}
	key := toolParamKey(a.uid, a.scene, toolName)
	cp := map[string]any{}
	for k, v := range params {
		cp[k] = v
	}
	a.store.mu.Lock()
	a.store.toolParams[key] = cp
	a.store.mu.Unlock()
}

func (s *runtimeStore) newApproval(uid uint64, metaTool approvalTicket) *approvalTicket {
	s.mu.Lock()
	defer s.mu.Unlock()
	t := metaTool
	t.ID = fmt.Sprintf("apv-%d", time.Now().UnixNano())
	t.CreatedAt = time.Now()
	t.ExpiresAt = time.Now().Add(10 * time.Minute)
	t.Status = "pending"
	t.RequestUID = uid
	s.approvals[t.ID] = &t
	return &t
}

func (s *runtimeStore) getApproval(id string) (*approvalTicket, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.approvals[id]
	if !ok {
		return nil, false
	}
	cp := *t
	return &cp, true
}

func (s *runtimeStore) setApprovalStatus(id, status string, reviewUID uint64) (*approvalTicket, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.approvals[id]
	if !ok {
		return nil, false
	}
	t.Status = status
	t.ReviewUID = reviewUID
	cp := *t
	return &cp, true
}

func (s *runtimeStore) saveExecution(rec *executionRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.executions[rec.ID] = rec
}

func (s *runtimeStore) getExecution(id string) (*executionRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rec, ok := s.executions[id]
	if !ok {
		return nil, false
	}
	cp := *rec
	return &cp, true
}

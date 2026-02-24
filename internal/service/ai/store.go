package ai

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/cy77cc/k8s-manage/internal/model"
)

func (s *memoryStore) appendMessage(userID uint64, scene, sessionID string, message map[string]any) *aiSession {
	now := time.Now()
	scene = normalizeScene(scene)
	sid := strings.TrimSpace(sessionID)
	if sid == "" {
		sid = s.getOrCreateCurrentSessionID(userID, scene)
	}
	if sid == "" {
		sid = fmt.Sprintf("sess-%d", now.UnixNano())
	}

	if !s.dbEnabled() {
		return &aiSession{ID: sid, Scene: scene, Title: "AI Session", Messages: []map[string]any{message}, CreatedAt: now, UpdatedAt: now}
	}

	var sess model.AIChatSession
	err := s.db.Where("id = ? AND user_id = ?", sid, userID).First(&sess).Error
	if err != nil {
		sess = model.AIChatSession{ID: sid, UserID: userID, Scene: scene, Title: "AI Session", CreatedAt: now, UpdatedAt: now}
		_ = s.db.Create(&sess).Error
	} else {
		if sess.Scene == "" {
			sess.Scene = scene
		}
		sess.UpdatedAt = now
		_ = s.db.Save(&sess).Error
	}

	msgID := strings.TrimSpace(toString(message["id"]))
	if msgID == "" {
		msgID = fmt.Sprintf("msg-%d", time.Now().UnixNano())
	}
	role := strings.TrimSpace(toString(message["role"]))
	content := toString(message["content"])
	thinking := toString(message["thinking"])
	createdAt := now
	if t, ok := message["timestamp"].(time.Time); ok {
		createdAt = t
	}
	_ = s.db.Create(&model.AIChatMessage{
		ID:        msgID,
		SessionID: sid,
		Role:      role,
		Content:   content,
		Thinking:  thinking,
		CreatedAt: createdAt,
	}).Error

	return s.mustLoadSession(userID, sid)
}

func (s *memoryStore) getOrCreateCurrentSessionID(userID uint64, scene string) string {
	if !s.dbEnabled() {
		return ""
	}
	var sess model.AIChatSession
	if err := s.db.Where("user_id = ? AND scene = ?", userID, scene).Order("updated_at DESC").First(&sess).Error; err == nil {
		return sess.ID
	}
	return ""
}

func (s *memoryStore) currentSession(userID uint64, scene string) (*aiSession, bool) {
	if !s.dbEnabled() {
		return nil, false
	}
	var sess model.AIChatSession
	if err := s.db.Where("user_id = ? AND scene = ?", userID, normalizeScene(scene)).Order("updated_at DESC").First(&sess).Error; err != nil {
		return nil, false
	}
	loaded := s.mustLoadSession(userID, sess.ID)
	if loaded == nil {
		return nil, false
	}
	return loaded, true
}

func (s *memoryStore) listSessions(userID uint64, scene string) []*aiSession {
	if !s.dbEnabled() {
		return nil
	}
	q := s.db.Where("user_id = ?", userID)
	if trim := strings.TrimSpace(scene); trim != "" {
		q = q.Where("scene = ?", normalizeScene(trim))
	}
	var sessions []model.AIChatSession
	if err := q.Order("updated_at DESC").Find(&sessions).Error; err != nil {
		return nil
	}
	out := make([]*aiSession, 0, len(sessions))
	for i := range sessions {
		out = append(out, s.mustLoadSession(userID, sessions[i].ID))
	}
	return out
}

func (s *memoryStore) getSession(userID uint64, id string) (*aiSession, bool) {
	if !s.dbEnabled() {
		return nil, false
	}
	var sess model.AIChatSession
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&sess).Error; err != nil {
		return nil, false
	}
	loaded := s.mustLoadSession(userID, id)
	if loaded == nil {
		return nil, false
	}
	return loaded, true
}

func (s *memoryStore) deleteSession(userID uint64, id string) {
	if !s.dbEnabled() {
		return
	}
	_ = s.db.Where("session_id = ?", id).Delete(&model.AIChatMessage{}).Error
	_ = s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.AIChatSession{}).Error
}

func (s *memoryStore) mustLoadSession(userID uint64, id string) *aiSession {
	if !s.dbEnabled() {
		return nil
	}
	var sess model.AIChatSession
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&sess).Error; err != nil {
		return nil
	}
	var msgs []model.AIChatMessage
	_ = s.db.Where("session_id = ?", id).Order("created_at ASC").Find(&msgs).Error
	arr := make([]map[string]any, 0, len(msgs))
	for i := range msgs {
		m := map[string]any{
			"id":        msgs[i].ID,
			"role":      msgs[i].Role,
			"content":   msgs[i].Content,
			"timestamp": msgs[i].CreatedAt,
		}
		if strings.TrimSpace(msgs[i].Thinking) != "" {
			m["thinking"] = msgs[i].Thinking
		}
		arr = append(arr, m)
	}
	return &aiSession{
		ID:        sess.ID,
		Scene:     sess.Scene,
		Title:     sess.Title,
		Messages:  arr,
		CreatedAt: sess.CreatedAt,
		UpdatedAt: sess.UpdatedAt,
	}
}

func (s *memoryStore) setRecommendations(userID uint64, scene string, recs []recommendationRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := recommendationKey(userID, scene)
	s.recommendations[key] = recs
}

func (s *memoryStore) getRecommendations(userID uint64, scene string, limit int) []recommendationRecord {
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

func normalizeScene(scene string) string {
	v := strings.TrimSpace(scene)
	if v == "" {
		return "global"
	}
	return v
}

func (s *memoryStore) newApproval(uid uint64, metaTool approvalTicket) *approvalTicket {
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

func (s *memoryStore) getApproval(id string) (*approvalTicket, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.approvals[id]
	if !ok {
		return nil, false
	}
	cp := *t
	return &cp, true
}

func (s *memoryStore) setApprovalStatus(id, status string, reviewUID uint64) (*approvalTicket, bool) {
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

func (s *memoryStore) saveExecution(rec *executionRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.executions[rec.ID] = rec
}

func (s *memoryStore) getExecution(id string) (*executionRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rec, ok := s.executions[id]
	if !ok {
		return nil, false
	}
	cp := *rec
	return &cp, true
}

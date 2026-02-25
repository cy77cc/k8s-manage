package ai

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/cy77cc/k8s-manage/internal/model"
	"gorm.io/gorm"
)

func (s *memoryStore) appendMessage(userID uint64, scene, sessionID string, message map[string]any) (*aiSession, error) {
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
		return &aiSession{ID: sid, Scene: scene, Title: "AI Session", Messages: []map[string]any{message}, CreatedAt: now, UpdatedAt: now}, nil
	}

	var sess model.AIChatSession
	err := s.db.Where("id = ? AND user_id = ?", sid, userID).First(&sess).Error
	switch {
	case err == nil:
		if sess.Scene == "" {
			sess.Scene = scene
		}
		sess.UpdatedAt = now
		if saveErr := s.db.Save(&sess).Error; saveErr != nil {
			return nil, saveErr
		}
	case errors.Is(err, gorm.ErrRecordNotFound):
		var exists int64
		if countErr := s.db.Model(&model.AIChatSession{}).Where("id = ?", sid).Count(&exists).Error; countErr != nil {
			return nil, countErr
		}
		if exists > 0 {
			return nil, errors.New("session not found")
		}
		sess = model.AIChatSession{ID: sid, UserID: userID, Scene: scene, Title: "AI Session", CreatedAt: now, UpdatedAt: now}
		if createErr := s.db.Create(&sess).Error; createErr != nil {
			return nil, createErr
		}
	default:
		return nil, err
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
	if err := s.db.Create(&model.AIChatMessage{
		ID:        msgID,
		SessionID: sid,
		Role:      role,
		Content:   content,
		Thinking:  thinking,
		CreatedAt: createdAt,
	}).Error; err != nil {
		return nil, err
	}

	loaded := s.mustLoadSession(userID, sid)
	if loaded == nil {
		return nil, errors.New("session not found")
	}
	return loaded, nil
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
	_ = s.db.Transaction(func(tx *gorm.DB) error {
		var sess model.AIChatSession
		if err := tx.Where("id = ? AND user_id = ?", id, userID).First(&sess).Error; err != nil {
			return nil
		}
		if err := tx.Where("session_id = ?", sess.ID).Delete(&model.AIChatMessage{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ? AND user_id = ?", sess.ID, userID).Delete(&model.AIChatSession{}).Error
	})
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

func toolParamKey(userID uint64, scene, tool string) string {
	return fmt.Sprintf("%d:%s:%s", userID, normalizeScene(scene), strings.TrimSpace(tool))
}

func normalizeScene(scene string) string {
	v := strings.TrimSpace(scene)
	if v == "" {
		return "global"
	}
	return v
}

type toolMemoryAccessor struct {
	store *memoryStore
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

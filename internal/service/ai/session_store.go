package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const sessionCacheTTL = 30 * time.Minute

type SessionStore struct {
	db  *gorm.DB
	rdb redis.UniversalClient
	ttl time.Duration
}

func NewSessionStore(db *gorm.DB, rdb redis.UniversalClient) *SessionStore {
	return &SessionStore{db: db, rdb: rdb, ttl: sessionCacheTTL}
}

func (s *SessionStore) dbEnabled() bool { return s != nil && s.db != nil }

func (s *SessionStore) AppendMessage(userID uint64, scene, sessionID string, message map[string]any) (*aiSession, error) {
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
		return &aiSession{ID: sid, Scene: scene, Title: defaultAISessionTitle, Messages: []map[string]any{message}, CreatedAt: now, UpdatedAt: now}, nil
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
		sess = model.AIChatSession{ID: sid, UserID: userID, Scene: scene, Title: defaultAISessionTitle, CreatedAt: now, UpdatedAt: now}
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

	s.invalidateSessionCaches(userID, sid, scene)
	loaded := s.mustLoadSession(userID, sid)
	if loaded == nil {
		return nil, errors.New("session not found")
	}
	s.cacheSession(userID, loaded)
	return loaded, nil
}

func (s *SessionStore) BranchSession(userID uint64, sourceSessionID, anchorMessageID, title string) (*aiSession, error) {
	if !s.dbEnabled() {
		return nil, errors.New("db unavailable")
	}
	sourceID := strings.TrimSpace(sourceSessionID)
	if sourceID == "" {
		return nil, errors.New("source session id is required")
	}
	source, ok := s.GetSession(userID, sourceID)
	if !ok || source == nil {
		return nil, gorm.ErrRecordNotFound
	}

	cutoff := len(source.Messages) - 1
	anchorID := strings.TrimSpace(anchorMessageID)
	if anchorID != "" {
		found := -1
		for i := range source.Messages {
			if strings.TrimSpace(toString(source.Messages[i]["id"])) == anchorID {
				found = i
				break
			}
		}
		if found < 0 {
			return nil, errors.New("anchor message not found")
		}
		cutoff = found
	}
	if cutoff < 0 {
		return nil, errors.New("source session has no messages")
	}

	newID := fmt.Sprintf("sess-%d", time.Now().UnixNano())
	newTitle := normalizeSessionTitle(title)
	if newTitle == "" {
		base := normalizeSessionTitle(source.Title)
		if base == "" {
			base = defaultAISessionTitle
		}
		newTitle = normalizeSessionTitle("分支: " + base)
		if newTitle == "" {
			newTitle = "Branch Session"
		}
	}
	now := time.Now()

	err := s.db.Transaction(func(tx *gorm.DB) error {
		sessionModel := model.AIChatSession{
			ID:        newID,
			UserID:    userID,
			Scene:     source.Scene,
			Title:     newTitle,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := tx.Create(&sessionModel).Error; err != nil {
			return err
		}
		for i := 0; i <= cutoff; i++ {
			msg := source.Messages[i]
			msgTime := now
			if t, ok := msg["timestamp"].(time.Time); ok && !t.IsZero() {
				msgTime = t
			}
			msgModel := model.AIChatMessage{
				ID:        fmt.Sprintf("msg-%d-%d", time.Now().UnixNano(), i+1),
				SessionID: newID,
				Role:      strings.TrimSpace(toString(msg["role"])),
				Content:   toString(msg["content"]),
				Thinking:  toString(msg["thinking"]),
				CreatedAt: msgTime,
			}
			if err := tx.Create(&msgModel).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.invalidateListCaches(userID, source.Scene)
	branched := s.mustLoadSession(userID, newID)
	if branched == nil {
		return nil, errors.New("branch session created but failed to load")
	}
	s.cacheSession(userID, branched)
	return branched, nil
}

func (s *SessionStore) CurrentSession(userID uint64, scene string) (*aiSession, bool) {
	if !s.dbEnabled() {
		return nil, false
	}
	key := s.currentSessionCacheKey(userID, scene)
	if out, ok := s.loadCachedSession(key); ok {
		return out, true
	}
	var sess model.AIChatSession
	if err := s.db.Where("user_id = ? AND scene = ?", userID, normalizeScene(scene)).Order("updated_at DESC").First(&sess).Error; err != nil {
		return nil, false
	}
	loaded := s.mustLoadSession(userID, sess.ID)
	if loaded == nil {
		return nil, false
	}
	s.cacheSessionWithKey(key, loaded)
	s.cacheSession(userID, loaded)
	return loaded, true
}

func (s *SessionStore) ListSessions(userID uint64, scene string) []*aiSession {
	if !s.dbEnabled() {
		return nil
	}
	key := s.sessionListCacheKey(userID, scene)
	if out, ok := s.loadCachedSessionList(key); ok {
		return out
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
		if loaded := s.mustLoadSession(userID, sessions[i].ID); loaded != nil {
			out = append(out, loaded)
			s.cacheSession(userID, loaded)
		}
	}
	s.cacheSessionList(key, out)
	return out
}

func (s *SessionStore) GetSession(userID uint64, id string) (*aiSession, bool) {
	if !s.dbEnabled() {
		return nil, false
	}
	key := s.sessionCacheKey(userID, id)
	if out, ok := s.loadCachedSession(key); ok {
		return out, true
	}
	var sess model.AIChatSession
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&sess).Error; err != nil {
		return nil, false
	}
	loaded := s.mustLoadSession(userID, id)
	if loaded == nil {
		return nil, false
	}
	s.cacheSession(userID, loaded)
	return loaded, true
}

func (s *SessionStore) DeleteSession(userID uint64, id string) {
	if !s.dbEnabled() {
		return
	}
	scene := ""
	_ = s.db.Transaction(func(tx *gorm.DB) error {
		var sess model.AIChatSession
		if err := tx.Where("id = ? AND user_id = ?", id, userID).First(&sess).Error; err != nil {
			return nil
		}
		scene = sess.Scene
		if err := tx.Where("session_id = ?", sess.ID).Delete(&model.AIChatMessage{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ? AND user_id = ?", sess.ID, userID).Delete(&model.AIChatSession{}).Error
	})
	s.invalidateSessionCaches(userID, id, scene)
}

func (s *SessionStore) UpdateSessionTitle(userID uint64, id, title string) (*aiSession, error) {
	if !s.dbEnabled() {
		return nil, errors.New("db unavailable")
	}
	sid := strings.TrimSpace(id)
	if sid == "" {
		return nil, errors.New("session id is required")
	}
	nextTitle := normalizeSessionTitle(title)
	if nextTitle == "" {
		return nil, errors.New("title is required")
	}
	var sess model.AIChatSession
	if err := s.db.Where("id = ? AND user_id = ?", sid, userID).First(&sess).Error; err != nil {
		return nil, err
	}
	sess.Title = nextTitle
	if err := s.db.Save(&sess).Error; err != nil {
		return nil, err
	}
	s.invalidateSessionCaches(userID, sid, sess.Scene)
	loaded := s.mustLoadSession(userID, sid)
	if loaded == nil {
		return nil, errors.New("session not found")
	}
	s.cacheSession(userID, loaded)
	return loaded, nil
}

func (s *SessionStore) getOrCreateCurrentSessionID(userID uint64, scene string) string {
	if !s.dbEnabled() {
		return ""
	}
	var sess model.AIChatSession
	if err := s.db.Where("user_id = ? AND scene = ?", userID, scene).Order("updated_at DESC").First(&sess).Error; err == nil {
		return sess.ID
	}
	return ""
}

func (s *SessionStore) mustLoadSession(userID uint64, id string) *aiSession {
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

func (s *SessionStore) sessionCacheKey(userID uint64, id string) string {
	return fmt.Sprintf("ai:session:%d:%s", userID, strings.TrimSpace(id))
}

func (s *SessionStore) sessionListCacheKey(userID uint64, scene string) string {
	return fmt.Sprintf("ai:session:list:%d:%s", userID, normalizeScene(scene))
}

func (s *SessionStore) currentSessionCacheKey(userID uint64, scene string) string {
	return fmt.Sprintf("ai:session:current:%d:%s", userID, normalizeScene(scene))
}

func (s *SessionStore) cacheSession(userID uint64, sess *aiSession) {
	if sess == nil {
		return
	}
	s.cacheSessionWithKey(s.sessionCacheKey(userID, sess.ID), sess)
	s.cacheSessionWithKey(s.currentSessionCacheKey(userID, sess.Scene), sess)
}

func (s *SessionStore) cacheSessionWithKey(key string, sess *aiSession) {
	if s == nil || s.rdb == nil || sess == nil {
		return
	}
	raw, err := json.Marshal(sess)
	if err != nil {
		return
	}
	_ = s.rdb.Set(context.Background(), key, raw, s.ttl).Err()
}

func (s *SessionStore) cacheSessionList(key string, sessions []*aiSession) {
	if s == nil || s.rdb == nil {
		return
	}
	raw, err := json.Marshal(sessions)
	if err != nil {
		return
	}
	_ = s.rdb.Set(context.Background(), key, raw, s.ttl).Err()
}

func (s *SessionStore) loadCachedSession(key string) (*aiSession, bool) {
	if s == nil || s.rdb == nil {
		return nil, false
	}
	raw, err := s.rdb.Get(context.Background(), key).Bytes()
	if err != nil {
		return nil, false
	}
	var out aiSession
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, false
	}
	return &out, true
}

func (s *SessionStore) loadCachedSessionList(key string) ([]*aiSession, bool) {
	if s == nil || s.rdb == nil {
		return nil, false
	}
	raw, err := s.rdb.Get(context.Background(), key).Bytes()
	if err != nil {
		return nil, false
	}
	var out []*aiSession
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, false
	}
	return out, true
}

func (s *SessionStore) invalidateSessionCaches(userID uint64, id, scene string) {
	if s == nil || s.rdb == nil {
		return
	}
	keys := []string{
		s.sessionCacheKey(userID, id),
		s.sessionListCacheKey(userID, scene),
		s.currentSessionCacheKey(userID, scene),
	}
	_ = s.rdb.Del(context.Background(), keys...).Err()
}

func (s *SessionStore) invalidateListCaches(userID uint64, scene string) {
	if s == nil || s.rdb == nil {
		return
	}
	keys := []string{
		s.sessionListCacheKey(userID, scene),
		s.currentSessionCacheKey(userID, scene),
	}
	_ = s.rdb.Del(context.Background(), keys...).Err()
}

package logic

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type SessionStore struct {
	db       *gorm.DB
	rdb      redis.UniversalClient
	mu       sync.RWMutex
	sessions map[string]*AISession
}

func NewSessionStore(db *gorm.DB, rdb redis.UniversalClient) *SessionStore {
	return &SessionStore{
		db:       db,
		rdb:      rdb,
		sessions: make(map[string]*AISession),
	}
}

func (s *SessionStore) Ensure(userID uint64, scene string) *AISession {
	s.mu.Lock()
	defer s.mu.Unlock()
	scene = NormalizeScene(scene)
	for _, sess := range s.sessions {
		if sess.UserID == userID && sess.Scene == scene {
			sess.UpdatedAt = time.Now()
			return cloneSession(sess)
		}
	}
	now := time.Now()
	session := &AISession{
		ID:        "sess-" + uuid.NewString(),
		UserID:    userID,
		Scene:     scene,
		Title:     fmt.Sprintf("%s session", scene),
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.sessions[session.ID] = session
	return cloneSession(session)
}

func (s *SessionStore) List(userID uint64, scene string) []*AISession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	scene = NormalizeScene(scene)
	out := make([]*AISession, 0)
	for _, sess := range s.sessions {
		if sess.UserID != userID {
			continue
		}
		if scene != "global" && scene != "" && sess.Scene != scene {
			continue
		}
		out = append(out, cloneSession(sess))
	}
	return out
}

func (s *SessionStore) Get(userID uint64, id string) (*AISession, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[id]
	if !ok || sess.UserID != userID {
		return nil, false
	}
	return cloneSession(sess), true
}

func (s *SessionStore) Put(session *AISession) {
	if s == nil || session == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	c := cloneSession(session)
	s.sessions[c.ID] = c
}

func (s *SessionStore) Delete(userID uint64, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sess, ok := s.sessions[id]; ok && sess.UserID == userID {
		delete(s.sessions, id)
	}
}

func cloneSession(in *AISession) *AISession {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}

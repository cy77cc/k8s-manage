package ai

import (
	"fmt"
	"time"
)

func (s *memoryStore) appendMessage(sessionID string, message map[string]any) *aiSession {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	ss, ok := s.sessions[sessionID]
	if !ok {
		ss = &aiSession{
			ID:        sessionID,
			Title:     "AI Session",
			CreatedAt: now,
		}
		s.sessions[sessionID] = ss
	}
	ss.Messages = append(ss.Messages, message)
	ss.UpdatedAt = now
	return cloneSession(ss)
}

func (s *memoryStore) listSessions() []*aiSession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*aiSession, 0, len(s.sessions))
	for _, session := range s.sessions {
		out = append(out, cloneSession(session))
	}
	return out
}

func (s *memoryStore) getSession(id string) (*aiSession, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ss, ok := s.sessions[id]
	if !ok {
		return nil, false
	}
	return cloneSession(ss), true
}

func (s *memoryStore) deleteSession(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, id)
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

func cloneSession(in *aiSession) *aiSession {
	out := *in
	out.Messages = make([]map[string]any, 0, len(in.Messages))
	for _, m := range in.Messages {
		cm := make(map[string]any, len(m))
		for k, v := range m {
			cm[k] = v
		}
		out.Messages = append(out.Messages, cm)
	}
	return &out
}

package aiv2

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type PendingStore struct {
	rdb    redis.UniversalClient
	prefix string
	ttl    time.Duration
}

func NewPendingStore(rdb redis.UniversalClient, prefix string, ttl time.Duration) *PendingStore {
	return &PendingStore{rdb: rdb, prefix: prefix, ttl: ttl}
}

func (s *PendingStore) key(sessionID string) string {
	return s.prefix + strings.TrimSpace(sessionID)
}

func (s *PendingStore) Save(ctx context.Context, approval PendingApproval) error {
	if s == nil || s.rdb == nil || strings.TrimSpace(approval.SessionID) == "" {
		return nil
	}
	raw, err := json.Marshal(approval)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, s.key(approval.SessionID), raw, s.ttl).Err()
}

func (s *PendingStore) Get(ctx context.Context, sessionID string) (*PendingApproval, error) {
	if s == nil || s.rdb == nil || strings.TrimSpace(sessionID) == "" {
		return nil, nil
	}
	raw, err := s.rdb.Get(ctx, s.key(sessionID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var approval PendingApproval
	if err := json.Unmarshal(raw, &approval); err != nil {
		return nil, err
	}
	return &approval, nil
}

func (s *PendingStore) Delete(ctx context.Context, sessionID string) error {
	if s == nil || s.rdb == nil || strings.TrimSpace(sessionID) == "" {
		return nil
	}
	return s.rdb.Del(ctx, s.key(sessionID)).Err()
}


package aiv2

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisCheckpointStore struct {
	rdb    redis.UniversalClient
	prefix string
	ttl    time.Duration
}

func newRedisCheckpointStore(rdb redis.UniversalClient, prefix string, ttl time.Duration) *redisCheckpointStore {
	return &redisCheckpointStore{rdb: rdb, prefix: prefix, ttl: ttl}
}

func (s *redisCheckpointStore) key(id string) string {
	return s.prefix + id
}

func (s *redisCheckpointStore) Get(ctx context.Context, checkPointID string) ([]byte, bool, error) {
	if s == nil || s.rdb == nil {
		return nil, false, nil
	}
	raw, err := s.rdb.Get(ctx, s.key(checkPointID)).Bytes()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return raw, true, nil
}

func (s *redisCheckpointStore) Set(ctx context.Context, checkPointID string, checkPoint []byte) error {
	if s == nil || s.rdb == nil {
		return nil
	}
	return s.rdb.Set(ctx, s.key(checkPointID), checkPoint, s.ttl).Err()
}


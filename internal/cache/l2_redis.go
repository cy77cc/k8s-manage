package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisL2 struct {
	client redis.UniversalClient
}

func NewRedisL2(client redis.UniversalClient) *RedisL2 {
	return &RedisL2{client: client}
}

func (r *RedisL2) Get(ctx context.Context, key string) (string, error) {
	if r == nil || r.client == nil {
		return "", ErrCacheMiss
	}
	v, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrCacheMiss
	}
	return v, err
}

func (r *RedisL2) Set(ctx context.Context, key string, val string, ttl time.Duration) error {
	if r == nil || r.client == nil {
		return nil
	}
	return r.client.Set(ctx, key, val, ttl).Err()
}

func (r *RedisL2) Delete(ctx context.Context, keys ...string) error {
	if r == nil || r.client == nil || len(keys) == 0 {
		return nil
	}
	return r.client.Del(ctx, keys...).Err()
}

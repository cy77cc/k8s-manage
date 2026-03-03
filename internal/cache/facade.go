package cache

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
)

var ErrCacheMiss = errors.New("cache miss")

const (
	SourceL1   = "l1"
	SourceL2   = "l2"
	SourceLoad = "load"
)

type L2Store interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, val string, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
}

type Stats struct {
	L1Hits        int64 `json:"l1_hits"`
	L2Hits        int64 `json:"l2_hits"`
	Misses        int64 `json:"misses"`
	FallbackLoads int64 `json:"fallback_loads"`
}

type Facade struct {
	l1 *expirable.LRU[string, string]
	l2 L2Store

	l1Hits        atomic.Int64
	l2Hits        atomic.Int64
	misses        atomic.Int64
	fallbackLoads atomic.Int64
}

func NewFacade(l1 *expirable.LRU[string, string], l2 L2Store) *Facade {
	return &Facade{l1: l1, l2: l2}
}

func (f *Facade) Get(key string) (string, bool) {
	if f == nil || f.l1 == nil {
		return "", false
	}
	v, ok := f.l1.Get(key)
	if ok {
		f.l1Hits.Add(1)
	}
	return v, ok
}

func (f *Facade) Set(ctx context.Context, key, val string, ttl time.Duration) {
	if f == nil {
		return
	}
	if f.l1 != nil {
		f.l1.Add(key, val)
	}
	if f.l2 != nil {
		_ = f.l2.Set(ctx, key, val, ttl)
	}
}

func (f *Facade) Delete(ctx context.Context, keys ...string) {
	if f == nil {
		return
	}
	if f.l1 != nil {
		for _, key := range keys {
			f.l1.Remove(key)
		}
	}
	if f.l2 != nil {
		_ = f.l2.Delete(ctx, keys...)
	}
}

func (f *Facade) GetOrLoad(ctx context.Context, key string, ttl time.Duration, loader func(context.Context) (string, error)) (string, string, error) {
	if v, ok := f.Get(key); ok {
		return v, SourceL1, nil
	}
	f.misses.Add(1)

	if f != nil && f.l2 != nil {
		if v, err := f.l2.Get(ctx, key); err == nil {
			f.l2Hits.Add(1)
			if f.l1 != nil {
				f.l1.Add(key, v)
			}
			return v, SourceL2, nil
		} else if !errors.Is(err, ErrCacheMiss) {
			f.fallbackLoads.Add(1)
		}
	}

	v, err := loader(ctx)
	if err != nil {
		return "", "", err
	}
	f.Set(ctx, key, v, ttl)
	return v, SourceLoad, nil
}

func (f *Facade) Stats() Stats {
	if f == nil {
		return Stats{}
	}
	return Stats{
		L1Hits:        f.l1Hits.Load(),
		L2Hits:        f.l2Hits.Load(),
		Misses:        f.misses.Load(),
		FallbackLoads: f.fallbackLoads.Load(),
	}
}

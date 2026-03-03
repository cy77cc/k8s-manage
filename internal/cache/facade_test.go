package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
)

type testL2 struct {
	data map[string]string
	err  error
}

func (t *testL2) Get(_ context.Context, key string) (string, error) {
	if t.err != nil {
		return "", t.err
	}
	v, ok := t.data[key]
	if !ok {
		return "", ErrCacheMiss
	}
	return v, nil
}

func (t *testL2) Set(_ context.Context, key string, val string, _ time.Duration) error {
	if t.err != nil {
		return t.err
	}
	if t.data == nil {
		t.data = map[string]string{}
	}
	t.data[key] = val
	return nil
}

func (t *testL2) Delete(_ context.Context, keys ...string) error {
	if t.err != nil {
		return t.err
	}
	for _, k := range keys {
		delete(t.data, k)
	}
	return nil
}

func TestFacadeGetOrLoadL1MissThenHit(t *testing.T) {
	l1 := expirable.NewLRU[string, string](100, nil, time.Minute)
	f := NewFacade(l1, nil)

	loads := 0
	load := func(context.Context) (string, error) {
		loads++
		return "value-1", nil
	}

	ctx := context.Background()
	v1, src1, err := f.GetOrLoad(ctx, "k1", time.Minute, load)
	if err != nil {
		t.Fatalf("first get or load failed: %v", err)
	}
	if v1 != "value-1" || src1 != SourceLoad {
		t.Fatalf("unexpected first result: value=%q source=%q", v1, src1)
	}

	v2, src2, err := f.GetOrLoad(ctx, "k1", time.Minute, load)
	if err != nil {
		t.Fatalf("second get or load failed: %v", err)
	}
	if v2 != "value-1" || src2 != SourceL1 {
		t.Fatalf("unexpected second result: value=%q source=%q", v2, src2)
	}
	if loads != 1 {
		t.Fatalf("expected single source load, got %d", loads)
	}

	stats := f.Stats()
	if stats.L1Hits != 1 || stats.Misses != 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
}

func TestFacadeGetOrLoadL2Fallback(t *testing.T) {
	l1 := expirable.NewLRU[string, string](100, nil, time.Minute)
	l2 := &testL2{data: map[string]string{"k2": "from-l2"}}
	f := NewFacade(l1, l2)

	v, src, err := f.GetOrLoad(context.Background(), "k2", time.Minute, func(context.Context) (string, error) {
		return "", errors.New("should not load")
	})
	if err != nil {
		t.Fatalf("get or load with l2 failed: %v", err)
	}
	if v != "from-l2" || src != SourceL2 {
		t.Fatalf("unexpected result: value=%q source=%q", v, src)
	}
}

func TestFacadeDeleteInvalidates(t *testing.T) {
	l1 := expirable.NewLRU[string, string](100, nil, time.Minute)
	f := NewFacade(l1, nil)
	f.Set(context.Background(), "k3", "v3", time.Minute)
	f.Delete(context.Background(), "k3")

	if _, ok := f.Get("k3"); ok {
		t.Fatalf("expected key deleted from l1")
	}
}

func TestFacadeNilSafeL2ErrorFallsBackToLoad(t *testing.T) {
	l1 := expirable.NewLRU[string, string](100, nil, time.Minute)
	l2 := &testL2{err: errors.New("redis down")}
	f := NewFacade(l1, l2)

	v, src, err := f.GetOrLoad(context.Background(), "k4", time.Minute, func(context.Context) (string, error) {
		return "from-db", nil
	})
	if err != nil {
		t.Fatalf("expected fallback load success, got: %v", err)
	}
	if v != "from-db" || src != SourceLoad {
		t.Fatalf("unexpected fallback result: value=%q source=%q", v, src)
	}
	if f.Stats().FallbackLoads == 0 {
		t.Fatalf("expected fallback load metrics increment")
	}
}

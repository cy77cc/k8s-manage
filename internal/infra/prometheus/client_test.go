package prometheus

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientQuery(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/query" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[{"metric":{"__name__":"up"},"value":[1710000000,"1"]}]}}`))
	}))
	defer ts.Close()

	c, err := NewClient(Config{Address: ts.URL, Timeout: 2 * time.Second, RetryCount: 1})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	res, err := c.Query(context.Background(), "up", time.Now())
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if res.ResultType != "vector" {
		t.Fatalf("expected vector, got %s", res.ResultType)
	}
	if len(res.Vector) != 1 {
		t.Fatalf("expected 1 sample, got %d", len(res.Vector))
	}
}

func TestClientQueryRange(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/query_range" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"matrix","result":[{"metric":{"__name__":"cpu_usage"},"values":[[1710000000,"11.2"],[1710000060,"12.6"]]}]}}`))
	}))
	defer ts.Close()

	c, err := NewClient(Config{Address: ts.URL, Timeout: 2 * time.Second, RetryCount: 1})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	res, err := c.QueryRange(context.Background(), "cpu_usage", time.Now().Add(-5*time.Minute), time.Now(), time.Minute)
	if err != nil {
		t.Fatalf("query range: %v", err)
	}
	if res.ResultType != "matrix" {
		t.Fatalf("expected matrix, got %s", res.ResultType)
	}
	if len(res.Matrix) != 1 {
		t.Fatalf("expected 1 series, got %d", len(res.Matrix))
	}
}

func TestClientMetadata(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/metadata" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"success","data":{"up":[{"type":"gauge","help":"up status","unit":""}]}}`))
	}))
	defer ts.Close()

	c, err := NewClient(Config{Address: ts.URL, Timeout: 2 * time.Second, RetryCount: 1})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	items, err := c.Metadata(context.Background(), "up")
	if err != nil {
		t.Fatalf("metadata: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Metric != "up" {
		t.Fatalf("expected metric up, got %s", items[0].Metric)
	}
}

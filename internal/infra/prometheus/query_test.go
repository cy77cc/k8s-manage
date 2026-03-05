package prometheus

import "testing"

func TestQueryBuilderBuildsMetricOnly(t *testing.T) {
	got := NewQueryBuilder("up").Build()
	if got != "up" {
		t.Fatalf("expected up, got %s", got)
	}
}

func TestQueryBuilderBuildsWithLabels(t *testing.T) {
	got := NewQueryBuilder("cpu_usage").
		WithLabel("source", "host").
		WithLabel("host_id", "123").
		Build()

	if got != "cpu_usage{host_id=\"123\",source=\"host\"}" {
		t.Fatalf("unexpected query: %s", got)
	}
}

func TestQueryBuilderBuildsWithRangeAndAgg(t *testing.T) {
	got := NewQueryBuilder("cpu_usage").
		WithRange("5m").
		WithAggregation("avg").
		Build()

	if got != "avg(cpu_usage[5m])" {
		t.Fatalf("unexpected query: %s", got)
	}
}

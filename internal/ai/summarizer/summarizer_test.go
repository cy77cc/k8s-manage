package summarizer

import (
	"context"
	"errors"
	"testing"
)

func TestSummarizerReturnsUnavailableWhenRunnerMissing(t *testing.T) {
	_, err := New(nil).Summarize(context.Background(), Input{})
	if err == nil {
		t.Fatalf("Summarize() error = nil, want UnavailableError")
	}
	var unavailable *UnavailableError
	if !errors.As(err, &unavailable) {
		t.Fatalf("Summarize() error = %v, want UnavailableError", err)
	}
	if unavailable.Code != "summarizer_runner_unavailable" {
		t.Fatalf("Code = %q, want summarizer_runner_unavailable", unavailable.Code)
	}
}

func TestNormalizeSummaryRequiresStructuredFields(t *testing.T) {
	_, err := normalizeSummary(SummaryOutput{})
	if err == nil {
		t.Fatalf("normalizeSummary() error = nil, want invalid output error")
	}
	var unavailable *UnavailableError
	if !errors.As(err, &unavailable) {
		t.Fatalf("normalizeSummary() error = %v, want UnavailableError", err)
	}
	if unavailable.Code != "summarizer_invalid_output" {
		t.Fatalf("Code = %q, want summarizer_invalid_output", unavailable.Code)
	}
}

func TestNormalizeSummaryQualifiesUncertainConclusions(t *testing.T) {
	out, err := normalizeSummary(SummaryOutput{
		Summary:               "日志显示正常",
		Headline:              "服务已经恢复",
		Conclusion:            "服务已经恢复",
		Narrative:             "根据结果可以确认服务恢复",
		KeyFindings:           []string{"日志显示正常", "日志显示正常"},
		Recommendations:       []string{"继续观察", "继续观察"},
		NeedMoreInvestigation: true,
	})
	if err != nil {
		t.Fatalf("normalizeSummary() error = %v", err)
	}
	if !out.NeedMoreInvestigation {
		t.Fatalf("NeedMoreInvestigation = false, want true")
	}
	if out.Conclusion == "服务已经恢复" {
		t.Fatalf("Conclusion should be qualified, got %q", out.Conclusion)
	}
	if out.Narrative == "根据结果可以确认服务恢复" {
		t.Fatalf("Narrative should be qualified, got %q", out.Narrative)
	}
	if len(out.KeyFindings) != 1 || len(out.Recommendations) != 1 {
		t.Fatalf("summary fields should be deduped: %#v", out)
	}
	if out.RawOutputPolicy != "summary_only" {
		t.Fatalf("RawOutputPolicy = %q, want summary_only", out.RawOutputPolicy)
	}
}

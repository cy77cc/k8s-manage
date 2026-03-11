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

func TestNormalizeSummaryRequiresRenderableText(t *testing.T) {
	_, err := normalizeSummary("")
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

func TestNormalizeSummaryPreservesPlainText(t *testing.T) {
	out, err := normalizeSummary("日志显示正常，服务可能已经恢复，建议继续观察。")
	if err != nil {
		t.Fatalf("normalizeSummary() error = %v", err)
	}
	if out != "日志显示正常，服务可能已经恢复，建议继续观察。" {
		t.Fatalf("Summary = %q", out)
	}
}

func TestNormalizeSummaryStripsCodeFenceWrapper(t *testing.T) {
	out, err := normalizeSummary("```json\n根据执行结果，payment-api 当前状态正常。\n```")
	if err != nil {
		t.Fatalf("normalizeSummary() error = %v", err)
	}
	if out != "根据执行结果，payment-api 当前状态正常。" {
		t.Fatalf("Summary = %q", out)
	}
}

func TestSummarizeFallsBackToPlainTextModelOutput(t *testing.T) {
	s := NewWithFunc(func(_ context.Context, _ Input, _ func(string)) (string, error) {
		return normalizeSummary("服务运行正常，根分区使用率 27%，当前无需处理。")
	})
	out, err := s.Summarize(context.Background(), Input{})
	if err != nil {
		t.Fatalf("Summarize() error = %v", err)
	}
	if out != "服务运行正常，根分区使用率 27%，当前无需处理。" {
		t.Fatalf("Summary = %q", out)
	}
}

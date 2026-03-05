package rag

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeIngestionRunner struct {
	checkpoints map[string]time.Time

	toolStats  IngestionStats
	assetStats IngestionStats
	caseStats  IngestionStats

	toolErr  error
	assetErr error
	caseErr  error

	toolCalls  int
	assetCalls int
	caseCalls  int
}

func newFakeIngestionRunner() *fakeIngestionRunner {
	return &fakeIngestionRunner{checkpoints: make(map[string]time.Time)}
}

func (f *fakeIngestionRunner) IngestToolExamples(_ context.Context, _ time.Time) (IngestionStats, error) {
	f.toolCalls++
	if f.toolErr != nil {
		return IngestionStats{}, f.toolErr
	}
	return f.toolStats, nil
}

func (f *fakeIngestionRunner) IngestPlatformAssets(_ context.Context, _ time.Time) (IngestionStats, error) {
	f.assetCalls++
	if f.assetErr != nil {
		return IngestionStats{}, f.assetErr
	}
	return f.assetStats, nil
}

func (f *fakeIngestionRunner) IngestTroubleshootingCases(_ context.Context, _ []TroubleshootingCaseInput, _ time.Time) (IngestionStats, error) {
	f.caseCalls++
	if f.caseErr != nil {
		return IngestionStats{}, f.caseErr
	}
	return f.caseStats, nil
}

func (f *fakeIngestionRunner) Checkpoint(source string) time.Time {
	return f.checkpoints[source]
}

func (f *fakeIngestionRunner) SetCheckpoint(source string, ts time.Time) {
	f.checkpoints[source] = ts
}

func TestRAGSchedulerRunJobs(t *testing.T) {
	runner := newFakeIngestionRunner()
	now := time.Now().UTC()
	runner.toolStats = IngestionStats{LatestUpdate: now.Add(-2 * time.Minute)}
	runner.assetStats = IngestionStats{LatestUpdate: now.Add(-1 * time.Minute)}
	runner.caseStats = IngestionStats{LatestUpdate: now}

	s := NewRAGScheduler(runner, DefaultSchedulerConfig(), func(context.Context) ([]TroubleshootingCaseInput, error) {
		return []TroubleshootingCaseInput{{Title: "case"}}, nil
	}, nil)

	if err := s.RunToolExamplesJob(context.Background()); err != nil {
		t.Fatalf("run tool job: %v", err)
	}
	if err := s.RunPlatformAssetsJob(context.Background()); err != nil {
		t.Fatalf("run asset job: %v", err)
	}
	if err := s.RunTroubleshootingJob(context.Background()); err != nil {
		t.Fatalf("run troubleshooting job: %v", err)
	}

	if runner.toolCalls != 1 || runner.assetCalls != 1 || runner.caseCalls != 1 {
		t.Fatalf("unexpected job call counts: tool=%d asset=%d case=%d", runner.toolCalls, runner.assetCalls, runner.caseCalls)
	}
	m := s.Metrics()
	if m.ToolExamples.Successes != 1 || m.PlatformAssets.Successes != 1 || m.Troubleshooting.Successes != 1 {
		t.Fatalf("unexpected scheduler metrics: %+v", m)
	}
	if runner.Checkpoint(checkpointToolExamples).IsZero() || runner.Checkpoint(checkpointPlatformAssets).IsZero() || runner.Checkpoint(checkpointTroubleshooting).IsZero() {
		t.Fatalf("expected checkpoints to be updated")
	}
}

func TestRAGSchedulerAlertOnFailure(t *testing.T) {
	runner := newFakeIngestionRunner()
	runner.toolErr = errors.New("tool ingest failed")

	alerts := make([]string, 0)
	s := NewRAGScheduler(runner, SchedulerConfig{FailureAlertThreshold: 1}, nil, func(job string, err error) {
		alerts = append(alerts, job+":"+err.Error())
	})

	err := s.RunToolExamplesJob(context.Background())
	if err == nil {
		t.Fatalf("expected tool job error")
	}
	if len(alerts) != 1 {
		t.Fatalf("expected one alert, got %d", len(alerts))
	}
	if alerts[0] != "tool_examples:tool ingest failed" {
		t.Fatalf("unexpected alert payload: %v", alerts[0])
	}
	m := s.Metrics()
	if m.ToolExamples.Failures != 1 || m.ToolExamples.ConsecutiveError != 1 {
		t.Fatalf("unexpected failure metrics: %+v", m.ToolExamples)
	}
}

func TestChooseCheckpoint(t *testing.T) {
	fallback := time.Now().UTC()
	if got := chooseCheckpoint(time.Time{}, time.Time{}, fallback); got != fallback {
		t.Fatalf("expected fallback checkpoint")
	}
	p := fallback.Add(-1 * time.Hour)
	l := fallback.Add(-30 * time.Minute)
	if got := chooseCheckpoint(p, l, fallback); got != l {
		t.Fatalf("expected latest checkpoint")
	}
	if got := chooseCheckpoint(l, p, fallback); got != l {
		t.Fatalf("expected previous checkpoint when newer")
	}
}

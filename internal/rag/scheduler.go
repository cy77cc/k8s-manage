package rag

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type ingestionRunner interface {
	IngestToolExamples(ctx context.Context, since time.Time) (IngestionStats, error)
	IngestPlatformAssets(ctx context.Context, since time.Time) (IngestionStats, error)
	IngestTroubleshootingCases(ctx context.Context, cases []TroubleshootingCaseInput, since time.Time) (IngestionStats, error)
	Checkpoint(source string) time.Time
	SetCheckpoint(source string, ts time.Time)
}

type TroubleshootingCasesLoader func(ctx context.Context) ([]TroubleshootingCaseInput, error)

type AlertHandler func(job string, err error)

type SchedulerConfig struct {
	ToolExamplesInterval    time.Duration
	PlatformAssetsInterval  time.Duration
	TroubleshootingInterval time.Duration
	FailureAlertThreshold   int
}

func DefaultSchedulerConfig() SchedulerConfig {
	return SchedulerConfig{
		ToolExamplesInterval:    time.Hour,
		PlatformAssetsInterval:  24 * time.Hour,
		TroubleshootingInterval: time.Hour,
		FailureAlertThreshold:   1,
	}
}

type JobMetrics struct {
	Runs             int
	Successes        int
	Failures         int
	ConsecutiveError int
	LastError        string
	LastRunAt        time.Time
	LastSuccessAt    time.Time
}

type SchedulerMetrics struct {
	ToolExamples    JobMetrics
	PlatformAssets  JobMetrics
	Troubleshooting JobMetrics
}

type RAGScheduler struct {
	runner ingestionRunner
	loadFn TroubleshootingCasesLoader
	alert  AlertHandler
	cfg    SchedulerConfig

	mu      sync.RWMutex
	metrics SchedulerMetrics
}

func NewRAGScheduler(runner ingestionRunner, cfg SchedulerConfig, loadFn TroubleshootingCasesLoader, alert AlertHandler) *RAGScheduler {
	if cfg.ToolExamplesInterval <= 0 {
		cfg.ToolExamplesInterval = time.Hour
	}
	if cfg.PlatformAssetsInterval <= 0 {
		cfg.PlatformAssetsInterval = 24 * time.Hour
	}
	if cfg.TroubleshootingInterval <= 0 {
		cfg.TroubleshootingInterval = time.Hour
	}
	if cfg.FailureAlertThreshold <= 0 {
		cfg.FailureAlertThreshold = 1
	}
	return &RAGScheduler{
		runner: runner,
		loadFn: loadFn,
		alert:  alert,
		cfg:    cfg,
	}
}

func (s *RAGScheduler) Start(ctx context.Context) error {
	if s == nil || s.runner == nil {
		return fmt.Errorf("rag scheduler is not initialized")
	}
	go s.loop(ctx, s.cfg.ToolExamplesInterval, s.RunToolExamplesJob)
	go s.loop(ctx, s.cfg.PlatformAssetsInterval, s.RunPlatformAssetsJob)
	go s.loop(ctx, s.cfg.TroubleshootingInterval, s.RunTroubleshootingJob)
	return nil
}

func (s *RAGScheduler) loop(ctx context.Context, interval time.Duration, job func(context.Context) error) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = job(ctx)
		}
	}
}

func (s *RAGScheduler) RunToolExamplesJob(ctx context.Context) error {
	since := s.runner.Checkpoint(checkpointToolExamples)
	stats, err := s.runner.IngestToolExamples(ctx, since)
	if err != nil {
		s.recordFailure("tool_examples", err)
		return err
	}
	s.runner.SetCheckpoint(checkpointToolExamples, chooseCheckpoint(since, stats.LatestUpdate, time.Now().UTC()))
	s.recordSuccess("tool_examples")
	return nil
}

func (s *RAGScheduler) RunPlatformAssetsJob(ctx context.Context) error {
	stats, err := s.runner.IngestPlatformAssets(ctx, time.Time{})
	if err != nil {
		s.recordFailure("platform_assets", err)
		return err
	}
	s.runner.SetCheckpoint(checkpointPlatformAssets, chooseCheckpoint(time.Time{}, stats.LatestUpdate, time.Now().UTC()))
	s.recordSuccess("platform_assets")
	return nil
}

func (s *RAGScheduler) RunTroubleshootingJob(ctx context.Context) error {
	var cases []TroubleshootingCaseInput
	if s.loadFn != nil {
		loaded, err := s.loadFn(ctx)
		if err != nil {
			s.recordFailure("troubleshooting_cases", err)
			return err
		}
		cases = loaded
	}
	since := s.runner.Checkpoint(checkpointTroubleshooting)
	stats, err := s.runner.IngestTroubleshootingCases(ctx, cases, since)
	if err != nil {
		s.recordFailure("troubleshooting_cases", err)
		return err
	}
	s.runner.SetCheckpoint(checkpointTroubleshooting, chooseCheckpoint(since, stats.LatestUpdate, time.Now().UTC()))
	s.recordSuccess("troubleshooting_cases")
	return nil
}

func (s *RAGScheduler) Metrics() SchedulerMetrics {
	if s == nil {
		return SchedulerMetrics{}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.metrics
}

func (s *RAGScheduler) recordSuccess(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	m := s.jobMetrics(job)
	m.Runs++
	m.Successes++
	m.ConsecutiveError = 0
	m.LastError = ""
	m.LastRunAt = now
	m.LastSuccessAt = now
}

func (s *RAGScheduler) recordFailure(job string, err error) {
	s.mu.Lock()
	now := time.Now().UTC()
	m := s.jobMetrics(job)
	m.Runs++
	m.Failures++
	m.ConsecutiveError++
	m.LastRunAt = now
	if err != nil {
		m.LastError = err.Error()
	}
	shouldAlert := m.ConsecutiveError >= s.cfg.FailureAlertThreshold && s.alert != nil
	s.mu.Unlock()

	if shouldAlert {
		s.alert(job, err)
	}
}

func (s *RAGScheduler) jobMetrics(job string) *JobMetrics {
	switch job {
	case "tool_examples":
		return &s.metrics.ToolExamples
	case "platform_assets":
		return &s.metrics.PlatformAssets
	default:
		return &s.metrics.Troubleshooting
	}
}

func chooseCheckpoint(previous, latest, fallback time.Time) time.Time {
	next := previous
	if latest.After(next) {
		next = latest
	}
	if next.IsZero() {
		next = fallback
	}
	return next
}

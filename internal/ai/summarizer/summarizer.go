package summarizer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cy77cc/OpsPilot/internal/ai/availability"
	"github.com/cy77cc/OpsPilot/internal/ai/executor"
	"github.com/cy77cc/OpsPilot/internal/ai/planner"
	"github.com/cy77cc/OpsPilot/internal/ai/runtime"
)

type Input struct {
	Message string
	Plan    *planner.ExecutionPlan
	State   runtime.ExecutionState
	Steps   []executor.StepResult
}

type ReplanHint struct {
	Reason          string   `json:"reason,omitempty"`
	Focus           string   `json:"focus,omitempty"`
	MissingEvidence []string `json:"missing_evidence,omitempty"`
}

type SummaryOutput struct {
	Summary               string      `json:"summary"`
	Headline              string      `json:"headline,omitempty"`
	Conclusion            string      `json:"conclusion,omitempty"`
	KeyFindings           []string    `json:"key_findings,omitempty"`
	ResourceSummaries     []string    `json:"resource_summaries,omitempty"`
	Recommendations       []string    `json:"recommendations,omitempty"`
	RawOutputPolicy       string      `json:"raw_output_policy,omitempty"`
	NextActions           []string    `json:"next_actions,omitempty"`
	NeedMoreInvestigation bool        `json:"need_more_investigation"`
	Narrative             string      `json:"narrative"`
	ReplanHint            *ReplanHint `json:"replan_hint,omitempty"`
}

type Summarizer struct {
	runner *adk.Runner
	runFn  func(context.Context, Input, func(string)) (SummaryOutput, error)
}

func New(runner *adk.Runner) *Summarizer {
	return &Summarizer{runner: runner}
}

func NewWithFunc(runFn func(context.Context, Input, func(string)) (SummaryOutput, error)) *Summarizer {
	return &Summarizer{runFn: runFn}
}

type UnavailableError struct {
	Code              string
	UserVisibleReason string
	Cause             error
}

func (e *UnavailableError) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", strings.TrimSpace(e.Code), e.Cause)
	}
	return firstNonEmpty(e.Code, "summarizer_unavailable")
}

func (e *UnavailableError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func (e *UnavailableError) UserVisibleMessage() string {
	if e == nil {
		return availability.UnavailableMessage(availability.LayerSummarizer)
	}
	return firstNonEmpty(e.UserVisibleReason, availability.UnavailableMessage(availability.LayerSummarizer))
}

func (s *Summarizer) Summarize(ctx context.Context, in Input) (SummaryOutput, error) {
	return s.summarize(ctx, in, nil)
}

func (s *Summarizer) SummarizeStream(ctx context.Context, in Input, onDelta func(string)) (SummaryOutput, error) {
	return s.summarize(ctx, in, onDelta)
}

func (s *Summarizer) summarize(ctx context.Context, in Input, onDelta func(string)) (SummaryOutput, error) {
	if s != nil && s.runFn != nil {
		return s.runFn(ctx, in, onDelta)
	}
	if s == nil || s.runner == nil {
		return SummaryOutput{}, &UnavailableError{
			Code:              "summarizer_runner_unavailable",
			UserVisibleReason: availability.UnavailableMessage(availability.LayerSummarizer),
		}
	}
	raw, err := runADKSummarizer(ctx, s.runner, buildPromptInput(in), onDelta)
	if err != nil {
		return SummaryOutput{}, &UnavailableError{
			Code:              "summarizer_model_unavailable",
			UserVisibleReason: availability.UnavailableMessage(availability.LayerSummarizer),
			Cause:             err,
		}
	}
	var parsed SummaryOutput
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &parsed); err != nil {
		return SummaryOutput{}, &UnavailableError{
			Code:              "summarizer_invalid_json",
			UserVisibleReason: availability.InvalidOutputMessage(availability.LayerSummarizer),
			Cause:             err,
		}
	}
	return normalizeSummary(parsed)
}

func buildPromptInput(in Input) string {
	data, _ := json.Marshal(in)
	return string(data)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func normalizeSummary(parsed SummaryOutput) (SummaryOutput, error) {
	parsed.Summary = firstNonEmpty(parsed.Summary, parsed.Headline, parsed.Conclusion)
	parsed.Headline = firstNonEmpty(parsed.Headline, parsed.Summary)
	parsed.Conclusion = firstNonEmpty(parsed.Conclusion, parsed.Summary)
	if parsed.Summary == "" || parsed.Headline == "" || parsed.Conclusion == "" {
		return SummaryOutput{}, &UnavailableError{
			Code:              "summarizer_invalid_output",
			UserVisibleReason: availability.InvalidOutputMessage(availability.LayerSummarizer),
			Cause:             fmt.Errorf("summary output missing summary/headline/conclusion"),
		}
	}
	if strings.TrimSpace(parsed.RawOutputPolicy) == "" {
		parsed.RawOutputPolicy = "summary_only"
	}
	parsed.KeyFindings = dedupe(parsed.KeyFindings)
	parsed.ResourceSummaries = dedupe(parsed.ResourceSummaries)
	parsed.Recommendations = dedupe(parsed.Recommendations)
	parsed.NextActions = dedupe(parsed.NextActions)
	if parsed.NeedMoreInvestigation {
		parsed.Conclusion = qualifyUncertainConclusion(parsed.Conclusion)
		parsed.Narrative = qualifyUncertainNarrative(parsed.Narrative)
		parsed.Headline = qualifyHeadline(parsed.Headline)
	}
	return parsed, nil
}

func qualifyUncertainConclusion(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return "当前仅能基于现有证据给出初步判断，仍需进一步调查。"
	}
	if containsUncertaintyMarker(text) {
		return text
	}
	return text + " 当前仅为基于现有证据的初步判断，仍需进一步调查。"
}

func qualifyUncertainNarrative(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return "当前叙述仅基于已完成步骤和现有证据，不足部分仍待补充确认。"
	}
	if containsUncertaintyMarker(text) && strings.Contains(text, "证据") {
		return text
	}
	if !strings.Contains(text, "证据") {
		text += " 以上内容仅基于当前执行证据。"
	}
	if !containsUncertaintyMarker(text) {
		text += " 仍存在待确认的不确定性。"
	}
	return strings.TrimSpace(text)
}

func qualifyHeadline(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return "当前仅能给出初步判断"
	}
	if containsUncertaintyMarker(text) {
		return text
	}
	return text + "（初步判断）"
}

func containsUncertaintyMarker(text string) bool {
	text = strings.ToLower(strings.TrimSpace(text))
	markers := []string{"可能", "初步", "待确认", "不确定", "证据不足", "进一步调查", "尚不能", "仍需"}
	for _, marker := range markers {
		if strings.Contains(text, marker) {
			return true
		}
	}
	return false
}

func dedupe(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

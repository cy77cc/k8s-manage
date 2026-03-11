package ai

import (
	"strings"

	"github.com/cy77cc/OpsPilot/internal/ai/executor"
	"github.com/cy77cc/OpsPilot/internal/ai/planner"
	"github.com/cy77cc/OpsPilot/internal/ai/summarizer"
)

type finalAnswerRenderer struct{}

func newFinalAnswerRenderer() *finalAnswerRenderer {
	return &finalAnswerRenderer{}
}

func (r *finalAnswerRenderer) Render(message string, plan *planner.ExecutionPlan, result *executor.Result, summaryOut summarizer.SummaryOutput) []string {
	_ = message
	_ = plan

	paragraphs := []string{}
	headline := sanitizeAnswerText(firstNonEmpty(summaryOut.Headline, summaryOut.Summary))
	if headline != "" {
		paragraphs = append(paragraphs, headline)
	}
	conclusion := sanitizeAnswerText(firstNonEmpty(summaryOut.Conclusion, summaryOut.Summary))
	if conclusion != "" && conclusion != headline {
		paragraphs = append(paragraphs, conclusion)
	}

	findings := sanitizeLines(append(append([]string(nil), summaryOut.KeyFindings...), summaryOut.ResourceSummaries...))
	if len(findings) > 0 {
		paragraphs = append(paragraphs, "关键依据：\n- "+strings.Join(findings, "\n- "))
	}

	recommendations := sanitizeLines(summaryOut.Recommendations)
	if len(recommendations) == 0 {
		recommendations = sanitizeLines(summaryOut.NextActions)
	}
	if len(recommendations) > 0 {
		paragraphs = append(paragraphs, "建议：\n- "+strings.Join(recommendations, "\n- "))
	}

	if shouldIncludeEvidence(summaryOut) {
		evidence := collectEvidenceLines(result, 6)
		if len(evidence) > 0 {
			paragraphs = append(paragraphs, "原始执行证据：\n- "+strings.Join(evidence, "\n- "))
		}
	}
	return compactParagraphs(paragraphs)
}

func shouldIncludeEvidence(summaryOut summarizer.SummaryOutput) bool {
	policy := strings.ToLower(strings.TrimSpace(summaryOut.RawOutputPolicy))
	return policy == "include_evidence" || policy == "raw_evidence"
}

func collectEvidenceLines(result *executor.Result, limit int) []string {
	if result == nil || limit <= 0 {
		return nil
	}
	out := make([]string, 0, limit)
	for _, step := range result.Steps {
		if summary := sanitizeAnswerText(step.Summary); summary != "" {
			out = append(out, summary)
			if len(out) >= limit {
				return dedupeStrings(out)
			}
		}
		for _, evidence := range step.Evidence {
			facts, ok := evidence.Data["observed_facts"].([]string)
			if ok {
				for _, fact := range facts {
					fact = sanitizeAnswerText(fact)
					if fact == "" {
						continue
					}
					out = append(out, fact)
					if len(out) >= limit {
						return dedupeStrings(out)
					}
				}
				continue
			}
			if rawFacts, ok := evidence.Data["observed_facts"].([]any); ok {
				for _, fact := range rawFacts {
					text := sanitizeAnswerText(asString(fact))
					if text == "" {
						continue
					}
					out = append(out, text)
					if len(out) >= limit {
						return dedupeStrings(out)
					}
				}
			}
		}
	}
	return dedupeStrings(out)
}

func sanitizeLines(items []string) []string {
	cleaned := make([]string, 0, len(items))
	for _, item := range items {
		item = sanitizeAnswerText(item)
		if item == "" {
			continue
		}
		cleaned = append(cleaned, item)
	}
	return dedupeStrings(cleaned)
}

func sanitizeAnswerText(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	lower := strings.ToLower(text)
	if strings.Contains(text, "```") {
		return ""
	}
	if strings.Contains(lower, "完整输出如下") || strings.Contains(lower, "raw output") {
		return ""
	}
	if strings.Contains(lower, "filesystem") && strings.Contains(lower, "mounted on") {
		return ""
	}
	text = strings.ReplaceAll(text, "`", "")
	text = strings.ReplaceAll(text, "***", "")
	text = strings.ReplaceAll(text, "**", "")
	return strings.TrimSpace(text)
}

func compactParagraphs(in []string) []string {
	out := make([]string, 0, len(in))
	for _, item := range in {
		text := strings.TrimSpace(item)
		if text != "" {
			out = append(out, text)
		}
	}
	return out
}

func dedupeStrings(in []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, item := range in {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

func asString(v any) string {
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x)
	default:
		return ""
	}
}

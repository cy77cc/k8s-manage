package summarizer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/adk"
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
	Conclusion            string      `json:"conclusion,omitempty"`
	NextActions           []string    `json:"next_actions,omitempty"`
	NeedMoreInvestigation bool        `json:"need_more_investigation"`
	Narrative             string      `json:"narrative"`
	ReplanHint            *ReplanHint `json:"replan_hint,omitempty"`
}

type Summarizer struct {
	runner *adk.Runner
}

func New(runner *adk.Runner) *Summarizer {
	return &Summarizer{runner: runner}
}

func (s *Summarizer) Summarize(ctx context.Context, in Input) (SummaryOutput, error) {
	base := buildBaseSummary(in)
	if s == nil || s.runner == nil {
		return base, nil
	}
	raw, err := runADKSummarizer(ctx, s.runner, buildPromptInput(in))
	if err != nil {
		return base, nil
	}
	var parsed SummaryOutput
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &parsed); err != nil {
		return base, nil
	}
	return normalizeSummary(base, parsed), nil
}

func buildBaseSummary(in Input) SummaryOutput {
	output := SummaryOutput{
		Summary:    "本轮执行已结束，结果已交给 Summarizer 生成结构化结论。",
		Conclusion: "可继续查看正文回答获取自然语言结论。",
		Narrative:  "Summarizer 模型不可用时，使用当前执行状态生成最小结构化摘要。",
	}
	if in.State.PendingApproval != nil && in.State.PendingApproval.Status == "pending" {
		title := firstNonEmpty(in.State.PendingApproval.Title, in.State.PendingApproval.StepID)
		output.Summary = fmt.Sprintf("步骤 %q 正在等待审批。", title)
		output.Conclusion = "当前计划已暂停，等待你确认后继续执行。"
		output.NextActions = []string{"确认当前审批请求"}
		output.Narrative = "执行链到达审批节点，当前只能回报暂停状态。"
		return output
	}

	completed, failed, blocked := summarizeStepCounts(in.Steps)

	if failed > 0 || blocked > 0 {
		output.Summary = "当前证据不足以形成稳定结论。"
		output.Conclusion = "执行链中存在失败或阻断步骤，建议补充调查后再继续决策。"
		output.NextActions = []string{"检查失败步骤的错误详情", "补充缺失证据后重新规划"}
		output.NeedMoreInvestigation = true
		output.ReplanHint = &ReplanHint{
			Reason:          "executor_has_failed_or_blocked_steps",
			Focus:           "补充失败步骤所需的资源信息或调查证据",
			MissingEvidence: []string{"failed_or_blocked_step_evidence"},
		}
		output.Narrative = "当前执行状态显示仍有失败或阻断步骤。"
		return output
	}

	goal := strings.TrimSpace(in.Message)
	if in.Plan != nil && strings.TrimSpace(in.Plan.Goal) != "" {
		goal = strings.TrimSpace(in.Plan.Goal)
	}
	output.Summary = fmt.Sprintf("已围绕目标“%s”完成 %d 个步骤。", goal, completed)
	output.Conclusion = "当前执行链已经完成本轮计划，可继续查看正文回答获取自然语言说明。"
	output.NextActions = []string{"查看最终结论正文"}
	return output
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

func normalizeSummary(base, parsed SummaryOutput) SummaryOutput {
	if strings.TrimSpace(parsed.Summary) == "" {
		parsed.Summary = base.Summary
	}
	if strings.TrimSpace(parsed.Conclusion) == "" {
		parsed.Conclusion = base.Conclusion
	}
	if len(parsed.NextActions) == 0 {
		parsed.NextActions = base.NextActions
	}
	if strings.TrimSpace(parsed.Narrative) == "" {
		parsed.Narrative = base.Narrative
	}
	if parsed.ReplanHint == nil && base.ReplanHint != nil {
		parsed.ReplanHint = base.ReplanHint
	}
	if parsed.NeedMoreInvestigation {
		return parsed
	}
	parsed.NeedMoreInvestigation = base.NeedMoreInvestigation
	return parsed
}

func summarizeStepCounts(steps []executor.StepResult) (completed, failed, blocked int) {
	for _, step := range steps {
		switch step.Status {
		case runtime.StepCompleted:
			completed++
		case runtime.StepFailed:
			failed++
		case runtime.StepBlocked, runtime.StepCancelled:
			blocked++
		}
	}
	return completed, failed, blocked
}

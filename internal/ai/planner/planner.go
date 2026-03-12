// Package planner 实现 AI 编排的规划阶段。
//
// 架构概览:
//
//	┌─────────────────────────────────────────────────────────────┐
//	│                         Planner                             │
//	│                                                             │
//	│   Input (message + rewrite) ──▶ LLM ──▶ Decision           │
//	│                                                             │
//	│   Decision types:                                           │
//	│     clarify      → 需要用户澄清                              │
//	│     direct_reply → 直接回复，无需执行计划                     │
//	│     plan         → 生成 ExecutionPlan 并交给 Executor        │
//	└─────────────────────────────────────────────────────────────┘
//
// 文件分布:
//   - planner.go  : Planner struct + Plan/PlanStream 入口
//   - types.go    : 所有导出类型定义
//   - parse.go    : LLM 宽松 JSON 解析
//   - normalize.go: 计划规范化、步骤输入填充、执行前提校验
//   - collect.go  : 从 rewrite.Output 提取资源引用
//   - support.go  : 平台 DB 资源解析辅助
package planner

import (
	"context"
	"errors"
	"strings"

	"github.com/cloudwego/eino/adk"

	"github.com/cy77cc/OpsPilot/internal/ai/availability"
)

// Planner 是规划器核心，负责生成执行计划。
type Planner struct {
	runner   *adk.Runner                                                  // ADK 运行器
	runFn    func(context.Context, Input, func(string)) (Decision, error) // 执行函数
	runRawFn func(context.Context, string, func(string)) (string, error)
}

const maxPlannerReplanAttempts = 2

// New 创建新的规划器实例。
func New(runner *adk.Runner) *Planner {
	return &Planner{runner: runner}
}

// NewWithFunc 使用自定义执行函数创建规划器。
func NewWithFunc(runFn func(context.Context, Input, func(string)) (Decision, error)) *Planner {
	return &Planner{runFn: runFn}
}

// Plan 执行规划，返回决策结果。
func (p *Planner) Plan(ctx context.Context, in Input) (Decision, error) {
	return p.plan(ctx, in, nil)
}

// PlanStream 执行规划并支持流式输出。
func (p *Planner) PlanStream(ctx context.Context, in Input, onDelta func(string)) (Decision, error) {
	return p.plan(ctx, in, onDelta)
}

// plan 执行规划的核心逻辑。
func (p *Planner) plan(ctx context.Context, in Input, onDelta func(string)) (Decision, error) {
	if p != nil && p.runFn != nil {
		return p.runFn(ctx, in, onDelta)
	}
	if p == nil || (p.runner == nil && p.runRawFn == nil) {
		return Decision{}, &PlanningError{
			Code:              "planner_runner_unavailable",
			UserVisibleReason: availability.UnavailableMessage(availability.LayerPlanner),
		}
	}
	runRaw := p.runRawFn
	if runRaw == nil {
		runRaw = func(ctx context.Context, prompt string, onDelta func(string)) (string, error) {
			return runADKPlanner(ctx, p.runner, prompt, onDelta)
		}
	}

	base := buildBasePlanContext(in)
	prompt := buildPromptInput(in)
	var lastErr error
	var raw string
	for attempt := 0; attempt <= maxPlannerReplanAttempts; attempt++ {
		rawOut, runErr := runRaw(ctx, prompt, onDelta)
		raw = strings.TrimSpace(rawOut)
		if runErr != nil {
			if isRetryablePlannerRunError(runErr) {
				lastErr = &PlanningError{
					Code:              "planner_invalid_json",
					UserVisibleReason: availability.InvalidOutputMessage(availability.LayerPlanner),
					Cause:             runErr,
				}
			} else {
				return Decision{}, &PlanningError{
					Code:              "planner_model_unavailable",
					UserVisibleReason: availability.UnavailableMessage(availability.LayerPlanner),
					Cause:             runErr,
				}
			}
		} else {
			decision, decodeErr := decodePlannerDecision(base, raw)
			if decodeErr == nil {
				return decision, nil
			}
			lastErr = decodeErr
		}
		if !isRetryablePlanningError(lastErr) || attempt == maxPlannerReplanAttempts {
			return Decision{}, lastErr
		}
		if in.OnReplan != nil {
			in.OnReplan(ReplanAttempt{
				Attempt:           attempt + 1,
				MaxAttempts:       maxPlannerReplanAttempts,
				Reason:            plannerRepairReason(lastErr),
				PreviousErrorCode: plannerErrorCode(lastErr),
				PreviousOutput:    raw,
			})
		}
		prompt = buildRepairPromptInput(in, raw, plannerRepairReason(lastErr), attempt+1, maxPlannerReplanAttempts)
	}
	return Decision{}, lastErr
}

func decodePlannerDecision(base *ExecutionPlan, raw string) (Decision, error) {
	parsed, err := ParseDecision(strings.TrimSpace(raw))
	if err != nil {
		return Decision{}, &PlanningError{
			Code:              "planner_invalid_json",
			UserVisibleReason: availability.InvalidOutputMessage(availability.LayerPlanner),
			Cause:             err,
		}
	}
	decision, err := normalizeDecision(base, parsed)
	if err != nil {
		return Decision{}, err
	}
	return decision, nil
}

func isRetryablePlannerRunError(err error) bool {
	return errors.Is(err, errPlannerEmptyOutput)
}

func isRetryablePlanningError(err error) bool {
	var planningErr *PlanningError
	if !errors.As(err, &planningErr) {
		return false
	}
	switch strings.TrimSpace(planningErr.Code) {
	case "planner_invalid_json", "planning_invalid":
		return true
	default:
		return false
	}
}

func plannerRepairReason(err error) string {
	var planningErr *PlanningError
	if !errors.As(err, &planningErr) {
		return "planner output did not satisfy the required schema"
	}
	cause := ""
	if planningErr.Cause != nil {
		cause = strings.TrimSpace(planningErr.Cause.Error())
	}
	switch strings.TrimSpace(planningErr.Code) {
	case "planner_invalid_json":
		if cause != "" {
			return "Planner output must be valid final-decision JSON: " + cause
		}
		return "Planner output must be valid final-decision JSON."
	case "planning_invalid":
		if cause != "" {
			return "Planner output is structurally invalid for execution: " + cause
		}
		return "Planner output is structurally invalid for execution."
	default:
		if cause != "" {
			return cause
		}
		return availability.InvalidOutputMessage(availability.LayerPlanner)
	}
}

func plannerErrorCode(err error) string {
	var planningErr *PlanningError
	if errors.As(err, &planningErr) {
		return strings.TrimSpace(planningErr.Code)
	}
	return ""
}

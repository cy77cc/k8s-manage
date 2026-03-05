package skills

import (
	"context"
	"fmt"
	"strings"

	"github.com/cy77cc/k8s-manage/internal/ai/tools"
)

type ToolInvokeFunc func(ctx context.Context, toolName string, params map[string]any) (tools.ToolResult, error)

type ApprovalStepHandler func(ctx context.Context, skill Skill, step SkillStep, state ExecutionState) error

type ResolverStepHandler func(ctx context.Context, skill Skill, step SkillStep, state ExecutionState) (map[string]any, error)

type ExecutionState struct {
	Params      map[string]any
	StepResults map[string]any
}

type ExecutionResult struct {
	SkillName   string
	Success     bool
	StepResults map[string]any
	FinalOutput any
}

type Executor struct {
	invokeTool       ToolInvokeFunc
	approvalHandler  ApprovalStepHandler
	resolverHandler  ResolverStepHandler
}

func NewExecutor(invokeTool ToolInvokeFunc, approvalHandler ApprovalStepHandler, resolverHandler ResolverStepHandler) *Executor {
	return &Executor{
		invokeTool:      invokeTool,
		approvalHandler: approvalHandler,
		resolverHandler: resolverHandler,
	}
}

func (e *Executor) Execute(ctx context.Context, skill Skill, params map[string]any) (*ExecutionResult, error) {
	if e == nil {
		return nil, fmt.Errorf("skill executor is nil")
	}
	if strings.TrimSpace(skill.Name) == "" {
		return nil, fmt.Errorf("skill name is required")
	}
	normalized, err := validateParams(skill.Parameters, params)
	if err != nil {
		return nil, err
	}

	state := ExecutionState{
		Params:      normalized,
		StepResults: make(map[string]any, len(skill.Steps)),
	}
	result := &ExecutionResult{
		SkillName:   skill.Name,
		Success:     false,
		StepResults: make(map[string]any, len(skill.Steps)),
	}

	for _, step := range skill.Steps {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		stepType := strings.ToLower(strings.TrimSpace(step.Type))
		switch stepType {
		case "tool":
			stepResult, err := e.executeToolStep(ctx, skill, step, state)
			if err != nil {
				return nil, err
			}
			state.StepResults[step.Name] = stepResult
			result.StepResults[step.Name] = stepResult
			result.FinalOutput = stepResult
		case "approval":
			if err := e.executeApprovalStep(ctx, skill, step, state); err != nil {
				return nil, err
			}
			state.StepResults[step.Name] = map[string]any{"status": "approved"}
			result.StepResults[step.Name] = state.StepResults[step.Name]
			result.FinalOutput = state.StepResults[step.Name]
		case "resolver":
			resolved, err := e.executeResolverStep(ctx, skill, step, state)
			if err != nil {
				return nil, err
			}
			if resolved != nil {
				for k, v := range resolved {
					state.Params[k] = v
				}
			}
			state.StepResults[step.Name] = resolved
			result.StepResults[step.Name] = resolved
			result.FinalOutput = resolved
		default:
			return nil, fmt.Errorf("unsupported step type: %s", step.Type)
		}
	}
	result.Success = true
	return result, nil
}

func (e *Executor) ExecuteFromMessage(ctx context.Context, skill Skill, message string, enumProvider EnumValueProvider) (*ExecutionResult, error) {
	params, err := extractParams(message, skill.Parameters, enumProvider)
	if err != nil {
		return nil, err
	}
	return e.Execute(ctx, skill, params)
}

func (e *Executor) executeToolStep(ctx context.Context, skill Skill, step SkillStep, state ExecutionState) (tools.ToolResult, error) {
	if e.invokeTool == nil {
		return tools.ToolResult{}, fmt.Errorf("tool invoker is not configured")
	}
	if strings.TrimSpace(step.Tool) == "" {
		return tools.ToolResult{}, fmt.Errorf("tool step %s missing tool name", step.Name)
	}
	rendered, err := renderParamsTemplate(step.ParamsTemplate, state.Params, state.StepResults)
	if err != nil {
		return tools.ToolResult{}, err
	}
	result, err := e.invokeTool(ctx, step.Tool, rendered)
	if err != nil {
		return tools.ToolResult{}, err
	}
	if !result.OK {
		return result, fmt.Errorf("tool step %s failed: %s", step.Name, result.Error)
	}
	return result, nil
}

func (e *Executor) executeApprovalStep(ctx context.Context, skill Skill, step SkillStep, state ExecutionState) error {
	if e.approvalHandler == nil {
		return nil
	}
	return e.approvalHandler(ctx, skill, step, state)
}

func (e *Executor) executeResolverStep(ctx context.Context, skill Skill, step SkillStep, state ExecutionState) (map[string]any, error) {
	if e.resolverHandler == nil {
		return nil, nil
	}
	return e.resolverHandler(ctx, skill, step, state)
}

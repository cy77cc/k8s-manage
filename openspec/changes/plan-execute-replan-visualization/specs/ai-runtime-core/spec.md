## MODIFIED Requirements

### REQ-RT-003: SSE 流式输出 (Enhanced)

系统 SHALL 通过 SSE 流式输出 Agent 执行事件。

**验收标准:**
- GIVEN 用户发送对话请求
- WHEN Agent 开始执行
- THEN 系统通过 SSE 流式输出事件
- AND 支持的事件类型包括: meta, delta, thinking_delta, tool_call, tool_result, approval_required, done, error
- **AND 支持新增的阶段事件: phase_started, phase_complete, plan_generated, step_started, step_complete, replan_triggered**

## NEW Requirements

### REQ-RT-005: Phase Detection

系统 SHALL 在 Orchestrator 中检测当前执行阶段。

**验收标准:**
- GIVEN Agent 事件流
- WHEN 事件包含 RunPath 或 AgentName 信息
- THEN 系统能推断当前阶段 (planning/executing/replanning)
- AND 阶段变化时发送 phase_started 事件

### REQ-RT-006: Plan Parsing

系统 SHALL 从 Planner 输出解析结构化计划。

**验收标准:**
- GIVEN Planner 输出包含 JSON 格式的步骤列表
- WHEN 解析成功
- THEN 发送 plan_generated 事件
- AND 事件包含结构化的 steps 数组

## NEW Interfaces

### PhaseDetector

```go
// PhaseDetector 从 ADK 事件推断当前执行阶段
type PhaseDetector struct {
    currentPhase string
    stepCounter  int
}

// Detect 从事件推断当前阶段
// 返回: planning, executing, replanning
func (d *PhaseDetector) Detect(event *adk.AgentEvent) string

// NextStepID 生成下一个步骤ID
func (d *PhaseDetector) NextStepID() string
```

### PlanParser

```go
// PlanParser 从 LLM 输出解析结构化计划
type PlanParser struct{}

// Parse 从 Planner 输出解析计划步骤
func (p *PlanParser) Parse(event *adk.AgentEvent) *ParsedPlan

type ParsedPlan struct {
    PlanID string
    Steps  []PlanStep
}

type PlanStep struct {
    ID       string
    Content  string
    ToolHint string
}
```

### StepInfo

```go
// StepInfo 步骤执行信息
type StepInfo struct {
    StepID    string
    Title     string
    ToolName  string
    Params    map[string]any
    Status    string  // running, success, error
    StartedAt string
    Summary   string
}
```

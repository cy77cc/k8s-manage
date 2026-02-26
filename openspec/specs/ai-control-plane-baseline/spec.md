# ai-control-plane-baseline Specification

## Purpose
TBD - created by archiving change migrate-docs-to-openspec-baseline. Update Purpose after archive.
## Requirements
### Requirement: AI Chat SHALL Support SSE Streaming Contract
AI chat capability SHALL be specified with an SSE event contract that includes message streaming and completion signaling.

#### Scenario: SSE event family is defined
- **WHEN** reviewers inspect the AI baseline
- **THEN** the spec SHALL include `meta`, `delta`, `thinking_delta`, `tool_call`, `tool_result`, `approval_required`, `done`, and `error` as baseline stream events

### Requirement: AI Tooling Control Plane SHALL Be Baseline Documented
The baseline SHALL document control-plane APIs for capabilities discovery, tool preview/execute, approval create/confirm, and execution query.

#### Scenario: Control-plane endpoint coverage
- **WHEN** maintainers compare baseline spec with AI routes
- **THEN** the spec SHALL cover endpoints defined in `internal/service/ai/routes.go` for capabilities/tools/approvals/executions

### Requirement: Mutating Tool Execution SHALL Require Approval Token
The baseline SHALL specify that mutating tools require approval prior to execution, while readonly tools can execute without approval.

#### Scenario: Approval gating behavior
- **WHEN** a mutating tool is requested without valid approval
- **THEN** execution SHALL be blocked and approval-required state SHALL be surfaced to caller

### Requirement: AI Session Persistence SHALL Be Reflected In Baseline
The baseline SHALL include session-oriented chat behavior and persisted chat/session records as part of current capability status.

#### Scenario: Session capability reflected
- **WHEN** reviewers inspect AI progress baseline
- **THEN** session list/current/get/delete and persisted message history behavior SHALL be represented as implemented baseline items


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
The baseline SHALL document control-plane APIs for capabilities discovery, tool preview/execute, approval create/confirm, execution query, command preview/execute/history, and AI action execution surfaces that include host-operation approval context.

#### Scenario: Control-plane endpoint coverage
- **WHEN** maintainers compare baseline spec with AI routes
- **THEN** the spec SHALL cover endpoints defined in `internal/service/ai/routes.go` for capabilities, tools, approvals, executions, and command bridge flows

#### Scenario: Host-operation approval context coverage
- **WHEN** reviewers inspect AI baseline for host mutating operations
- **THEN** the baseline SHALL document approval-required signaling and approval-token propagation requirements across preview and execute APIs

### Requirement: Mutating Tool Execution SHALL Require Approval Token
The baseline SHALL specify that mutating tools require approval prior to execution, while readonly tools can execute without approval, and high-risk command execution SHALL require explicit execution confirmation in addition to approval.

#### Scenario: Approval gating behavior
- **WHEN** a mutating tool is requested without valid approval
- **THEN** execution SHALL be blocked and approval-required state SHALL be surfaced to caller

#### Scenario: High-risk command confirm behavior
- **WHEN** a high-risk command execution request is sent without explicit confirm flag
- **THEN** execution SHALL be rejected before tool invocation

#### Scenario: Reviewer authorization behavior
- **WHEN** approval confirmation is requested by a user without approval-review permission
- **THEN** the confirmation SHALL be denied and approval status SHALL remain unchanged

### Requirement: AI Approval Interactions SHALL Support Multi-surface Actions
The baseline SHALL include approval interaction expectations for chat traces and notification entries to approve or reject pending mutating operations.

#### Scenario: Multi-surface approval consistency
- **WHEN** a pending approval is resolved from chat or notification surface
- **THEN** the approval status SHALL be synchronized to command execution and history views

### Requirement: AI Session Persistence SHALL Be Reflected In Baseline
The baseline SHALL include session-oriented chat behavior and persisted chat/session records as part of current capability status.

#### Scenario: Session capability reflected
- **WHEN** reviewers inspect AI progress baseline
- **THEN** session list/current/get/delete and persisted message history behavior SHALL be represented as implemented baseline items

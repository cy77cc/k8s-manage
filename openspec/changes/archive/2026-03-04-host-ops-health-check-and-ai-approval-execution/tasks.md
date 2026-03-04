## 1. Data Model And Contracts

- [x] 1.1 Design and add migrations for host health snapshots, maintenance metadata, and AI host execution records with Up/Down scripts in `storage/migrations`.
- [x] 1.2 Update backend API contracts under `api/*/v1` for host health diagnostics, maintenance lifecycle fields, and governed AI host execution payloads.
- [x] 1.3 Update frontend API module typings in `web/src/api/modules/*` for new host health states, maintenance metadata, approval interaction payloads, and execution history responses.

## 2. Host Health Diagnostics Backend

- [x] 2.1 Implement standardized host health check runner (connectivity/resource/system dimensions) under `internal/service/host/logic`.
- [x] 2.2 Implement scheduled snapshot pipeline with bounded concurrency, timeout, and retry/jitter strategy for host health collection.
- [x] 2.3 Implement on-demand host health check API path in `internal/service/host/handler` and wire routes in `internal/service/host/routes.go`.
- [x] 2.4 Replace placeholder `Metrics` behavior with persisted/derived diagnostics responses and consistent `healthy/degraded/critical/unknown` mapping.

## 3. Host Maintenance Lifecycle Backend

- [x] 3.1 Extend host action logic to support maintenance enter/exit metadata (`reason`, `operator`, `started_at`, `until`) instead of status-only updates.
- [x] 3.2 Implement shared host eligibility predicate so maintenance hosts are excluded from cluster/deployment candidate selection flows.
- [x] 3.3 Integrate maintenance-aware behavior into non-essential automation paths to skip maintenance hosts with explicit reason output.
- [x] 3.4 Emit audit and notification events for maintenance lifecycle transitions.

## 4. AI Governed Host Execution Backend

- [x] 4.1 Extend AI command/tool preview responses to include host execution plan details (targets, risk, script/command summary, limits) without mutating execution.
- [x] 4.2 Enforce dual-gate execution checks (explicit execution confirm + approved ticket for mutating/high-risk host operations).
- [x] 4.3 Implement script upload-and-run workflow to controlled path (ticket/execution scoped) with timeout and per-host isolation.
- [x] 4.4 Persist and expose per-host execution results (stdout/stderr/exit code/timestamps) and replay context via AI history endpoints.

## 5. Policy And Security Enforcement

- [x] 5.1 Implement centralized host-operation policy checks (scope restrictions, denylist/allowlist, concurrency caps, timeout policies) in AI tool execution path.
- [x] 5.2 Add output redaction for sensitive patterns in execution results before persistence and API response.
- [x] 5.3 Enforce RBAC checks for approval-review and host mutating operations across APIs and UI actions.

## 6. Chat And Notification Approval UX

- [x] 6.1 Add approve/reject actions to AI chat `approval_required` trace cards for authorized reviewers in `web/src/components/AI`.
- [x] 6.2 Add approval-type tab/filter and approve/reject item actions in notification center components and API module.
- [x] 6.3 Implement cross-surface state sync so approval status updates propagate consistently to chat, notification, and command/history views.

## 7. Host Management UX Updates

- [x] 7.1 Update host list/detail pages to display health state, key diagnostic signals, and maintenance metadata clearly.
- [x] 7.2 Replace prompt-based high-risk operation triggers with structured confirmation dialogs showing risk, scope, and expected impact.
- [x] 7.3 Add manual health-check trigger UI with result rendering and error diagnostics feedback.

## 8. Testing And Verification

- [x] 8.1 Add backend unit/integration tests for health-state mapping, maintenance side effects, approval gating, and policy enforcement.
- [x] 8.2 Add frontend tests for chat approval actions, notification approval workflows, and host maintenance/health UI behavior.
- [x] 8.3 Run `openspec validate --changes --json`, targeted backend tests, and relevant frontend test suites; fix issues before apply phase completion.

## 9. Rollout And Operations

- [x] 9.1 Add feature flags for health diagnostics, maintenance semantics, and AI governed host execution pathways.
- [x] 9.2 Define observability metrics and dashboards for health check latency/failure, approval turnaround, and execution policy rejection counts.
- [x] 9.3 Prepare rollout and rollback runbook covering staged enablement, emergency disable steps, and audit trace verification.

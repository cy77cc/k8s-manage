## Context

Current host operations expose health-check and maintenance actions but lack operational semantics: `Action` only updates host status, and host metrics/audits endpoints still return placeholder data. On AI side, command/tool preview-execute-approval primitives already exist, but approval interaction is not fully integrated into chat and notification surfaces for direct approve/reject workflows.

This change spans backend (`internal/service/host/*`, `internal/service/ai/*`, `internal/ai/tools/*`), frontend (`web/src/pages/Hosts/*`, `web/src/components/AI/*`, `web/src/components/Notification/*`), and persistence/audit flows, and must preserve RBAC and approval gating for mutating operations.

## Goals / Non-Goals

**Goals:**
- Define and implement host health diagnostics as first-class behavior (not mock metrics), with standardized states and check dimensions.
- Define maintenance lifecycle semantics with explicit metadata, scheduling impact, and audit/notification side effects.
- Establish governed AI host execution workflow for command/script operations: preview, risk, two-step confirmation, approval gating, execution, and replayable results.
- Enable approval interactions directly from chat traces and notification center for host-operation approvals.
- Introduce policy guardrails for AI-initiated host operations (scope, allow/deny controls, timeout/concurrency, redaction).

**Non-Goals:**
- Replacing current AI runtime/LLM orchestration architecture.
- Building a full remote configuration management platform (Ansible/Salt-level feature parity).
- Delivering auto-remediation workflows in this change (manual approved execution only).
- Redesigning the entire notification system beyond approval interaction capabilities required by this scope.

## Decisions

### 1) Health model: asynchronous snapshot + on-demand probe
- Decision: Use two health data paths:
  - periodic snapshot checks persisted per host for list/detail rendering,
  - on-demand deep checks triggered by explicit user action.
- Rationale: list page requires low-latency stable state; deep checks may be expensive and should be explicit.
- Alternatives considered:
  - pure on-demand checks only: simple but causes slow list UX and unstable status.
  - pure scheduled checks only: cannot provide operator-triggered instant verification.

### 2) Host health state contract
- Decision: Standardize host health to `healthy | degraded | critical | unknown`, independent from operational status (`online/offline/maintenance`).
- Rationale: status and health represent different concerns; conflating them blocks accurate decisioning.
- Alternatives considered:
  - keep `online/offline/maintenance` only: insufficient granularity for diagnostics and automation policies.

### 3) Maintenance lifecycle semantics
- Decision: Introduce maintenance metadata (`reason`, `operator`, `started_at`, `until`, `scope`) and side effects:
  - exclude host from deployment/scheduling candidate sets,
  - pause non-essential automation on the host,
  - emit audit and notification events.
- Rationale: maintenance must affect behavior, not just display.
- Alternatives considered:
  - status-only update: easiest, but operationally unsafe and opaque.

### 4) AI host execution flow
- Decision: enforce flow:
  - AI generates execution plan/script preview (read-only),
  - user confirms execution intent,
  - approver approves ticket (required for mutating/high-risk host ops),
  - executor uploads script (when needed) and runs command/script with policy checks,
  - per-host results stored and replayable.
- Rationale: aligns existing approval primitives with host execution risk model.
- Alternatives considered:
  - single-click execution after preview: lower friction but unacceptable risk.
  - approval-only without explicit user confirm: weak user intent confirmation.

### 5) Policy enforcement layer
- Decision: centralize host operation policy enforcement in AI tool execution path:
  - host scope allowlist (project/env/tag-based),
  - command/script denylist + optional allowlist mode,
  - timeout/concurrency caps,
  - output redaction rules.
- Rationale: consistency across chat-triggered and panel-triggered execution.
- Alternatives considered:
  - enforce per-endpoint ad hoc: prone to drift and bypass.

### 6) Approval interaction surfaces
- Decision: add approval actions in:
  - AI chat trace cards (`approval_required` entries gain approve/reject controls for authorized roles),
  - notification center (`approval` type tab with direct approve/reject).
- Rationale: reduce context switching and improve closure rate of pending approvals.
- Alternatives considered:
  - keep approval in separate page only: secure but slower operational response.

## Risks / Trade-offs

- [Risk] Health checks increase host connection load and may affect SSH limits.
  - Mitigation: rate limiting, jittered schedules, bounded concurrency, fast-fail timeouts.
- [Risk] Maintenance exclusion logic may diverge across deployment and automation modules.
  - Mitigation: shared host-eligibility predicate used by all scheduling paths.
- [Risk] AI-generated scripts may include unsafe operations.
  - Mitigation: denylist/allowlist policy gate, mandatory approval for mutating ops, execution scope constraints.
- [Risk] Dual confirmation may reduce operator efficiency.
  - Mitigation: optimize approval UX in chat/notification and support reusable templates for common ops.
- [Risk] Approval token leakage could enable unauthorized execution.
  - Mitigation: short TTL, single-use tokens, strict reviewer permission checks, audit correlation.

## Migration Plan

1. Introduce data model/API contract changes for health snapshots and maintenance metadata (backward-compatible reads).
2. Implement backend health check pipeline and maintenance lifecycle semantics behind feature flags.
3. Integrate AI host execution governance path and policy checks; keep existing command panel flow compatible.
4. Add chat and notification approval interaction UI and backend linkage.
5. Enable feature flags progressively by environment; monitor audit/latency/error metrics.

Rollback:
- Disable feature flags to return to current behavior.
- Keep additive schema fields/tables; stop writing new semantics if rollback required.
- Preserve audit records generated during rollout for traceability.

## Open Questions

- Should medium-risk host operations require approval by policy default, or only high-risk operations?
- What is the minimal initial denylist/allowlist policy set acceptable for production rollout?
- Which modules must immediately consume maintenance exclusion (deployment only vs deployment + automation + CMDB actions)?
- Should script upload execution support only ephemeral `/tmp/opsx/<ticket>/` paths in v1, or configurable remote directories?

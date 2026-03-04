## Why

Host management currently exposes health check and maintenance actions in UI, but behavior is mostly status toggling and mock metrics without operational semantics. At the same time, AI-assisted host operations need a safer execution path with explicit second confirmation and approval from chat or notifications to reduce production risk.

## What Changes

- Implement real host health checks with standardized dimensions and state model (`healthy/degraded/critical/unknown`) instead of placeholder metrics.
- Introduce maintenance mode lifecycle semantics (enter/exit reason, owner, expiry, scheduling exclusion, and audit/notification side effects), not only a status field update.
- Add governed AI-to-host execution flow for command/script operations: plan preview, risk classification, two-step confirmation, approval gating, scoped execution, and result replay.
- Enable approval action entrypoints directly from AI chat traces and notification center for pending AI operation tickets.
- Add policy controls for host operation safety: command/script allow/deny controls, host scope constraints, concurrency/timeout limits, and output redaction.

## Capabilities

### New Capabilities

- `host-health-diagnostics`: Define health check dimensions, probe lifecycle, thresholds, and health-state transitions for managed hosts.
- `host-maintenance-lifecycle`: Define maintenance enter/exit contract with reason metadata, automation side effects, and audit/notification requirements.
- `ai-governed-host-execution`: Define AI-generated command/script execution workflow on hosts with preview, risk, approvals, and per-host execution result contract.
- `ai-approval-interaction-surface`: Define approval interaction requirements in chat and notification surfaces, including approve/reject actions and state synchronization.

### Modified Capabilities

- `deployment-infrastructure-management`: Host operations requirements change from simple action endpoints to semantic health and maintenance lifecycle behavior.
- `ai-control-plane-baseline`: AI control-plane requirements change to include host operation approval interactions and execution governance constraints.

## Impact

- Backend domains:
  - `internal/service/host/*` (health checks, maintenance lifecycle, host action semantics, metrics/audit APIs)
  - `internal/service/ai/*` and `internal/ai/tools/*` (approval-gated host execution workflow and policy enforcement)
  - notification and audit related handlers/stores
- Frontend domains:
  - `web/src/pages/Hosts/*` (health/maintenance UX)
  - `web/src/components/AI/*` (approval actions in chat)
  - `web/src/components/Notification/*` and `web/src/api/modules/*` (approval tab and confirm/reject flow)
- Data/contracts:
  - new/updated API contracts under `/api/v1/hosts` and `/api/v1/ai`
  - host health snapshots, maintenance metadata, approval linkage, and execution records
- Security/operations:
  - RBAC and approval policy updates for mutating host operations
  - stronger auditability and runtime guardrails for AI-initiated operations

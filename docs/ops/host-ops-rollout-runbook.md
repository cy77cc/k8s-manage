# Host Ops Governance Rollout / Rollback Runbook

## Feature Flags

Config path: `feature_flags`

- `host_health_diagnostics`
- `host_maintenance_mode`
- `ai_governed_host_execution`

## Rollout Strategy

1. Stage 0 (dark launch)
   - Enable in dev/staging only.
   - Validate health checks write snapshots and no scheduler overload.
2. Stage 1 (limited production)
   - Enable `host_health_diagnostics` in production.
   - Keep `host_maintenance_mode` and `ai_governed_host_execution` disabled.
3. Stage 2 (maintenance semantics)
   - Enable `host_maintenance_mode`.
   - Verify deployment/automation exclusion logs.
4. Stage 3 (AI governed execution)
   - Enable `ai_governed_host_execution`.
   - Validate dual-confirm + approval token flow in chat and notifications.
5. Stage 4 (full rollout)
   - Keep all flags enabled.
   - Monitor observability dashboards continuously for 24h.

## Verification Checklist

- Health:
  - Manual check endpoint returns diagnostics.
  - Scheduled snapshots are written on expected cadence.
- Maintenance:
  - Enter/exit actions generate audit events.
  - Maintenance hosts are excluded from deployment/automation candidate sets.
- AI execution:
  - Missing `confirm=true` is rejected.
  - Missing approval token for mutating host intent is rejected.
  - Approved request executes and writes per-host execution records.

## Emergency Disable

If incident detected, disable in this order:

1. `ai_governed_host_execution = false`
2. `host_maintenance_mode = false` (only if maintenance flow causes incorrect scheduling)
3. `host_health_diagnostics = false` (only if check load impacts production SSH availability)

Then restart service and verify:

- AI host execution intents return `feature disabled` policy message.
- Maintenance action does not enter new semantic maintenance mode.
- Scheduled health collector no longer starts.

## Rollback Data Notes

- Database migrations are additive; no immediate down migration required for emergency stop.
- Keep generated audit/notification/execution records for postmortem traceability.

## Audit Trace Validation

Before closing incident:

1. Sample maintenance lifecycle audits:
   - `host_maintenance_entered`
   - `host_maintenance_exited`
2. Sample AI command records:
   - approval context (`approval_token`, `approved_by`)
   - host execution result records (`exit_code`, `status`)
3. Confirm notification and chat surfaces show consistent approval terminal state.

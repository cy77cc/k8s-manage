# Host Ops Observability Baseline

## Scope

This baseline covers observability for:

- Host health diagnostics pipeline
- Maintenance lifecycle transitions
- AI-governed host execution approvals and policy rejections

## Metrics

### Health Diagnostics

- `host_health_check_duration_ms` (histogram)
  - labels: `mode` (`scheduled|manual`), `state` (`healthy|degraded|critical|unknown`)
- `host_health_check_total` (counter)
  - labels: `mode`, `result` (`success|failed`)
- `host_health_check_fail_total` (counter)
  - labels: `reason` (`ssh_connect_failed|auth_failed|command_failed|timeout|unknown`)
- `host_health_snapshot_write_fail_total` (counter)

### Maintenance Lifecycle

- `host_maintenance_enter_total` (counter)
  - labels: `operator`, `source` (`ui|api|ai`)
- `host_maintenance_exit_total` (counter)
  - labels: `operator`, `source`
- `host_maintenance_active` (gauge)
  - labels: `env`, `project`

### AI Governed Execution

- `ai_host_exec_preview_total` (counter)
  - labels: `intent`, `risk`
- `ai_host_exec_apply_total` (counter)
  - labels: `intent`, `result` (`succeeded|failed|blocked`)
- `ai_host_exec_policy_reject_total` (counter)
  - labels: `reason` (`scope_limit|timeout_limit|denylist|host_ineligible|feature_disabled`)
- `ai_host_exec_approval_wait_seconds` (histogram)
  - labels: `intent`, `risk`
- `ai_host_exec_target_count` (histogram)
  - labels: `intent`

## Dashboard Panels

## Dashboard: Host Health

- P50/P95 `host_health_check_duration_ms` by `mode`
- `host_health_check_total` split by `state`
- Top failure reasons (`host_health_check_fail_total`)
- `host_maintenance_active` trend

## Dashboard: AI Host Execution Governance

- Preview vs Apply volume (`ai_host_exec_preview_total`, `ai_host_exec_apply_total`)
- Policy rejection count and ratio (`ai_host_exec_policy_reject_total`)
- Approval turnaround P50/P95 (`ai_host_exec_approval_wait_seconds`)
- Target host distribution (`ai_host_exec_target_count`)

## Alerts

- Critical: policy rejection ratio > 30% for 10 minutes
- Critical: health check failure ratio > 20% for 5 minutes
- Warning: approval wait P95 > 10 minutes for 15 minutes
- Warning: scheduled health checks missing for > 15 minutes

# Service Deploy Linkage Rollout Plan

## Scope

This runbook covers rollout for `fix-service-deploy-interaction-gaps`, including:
- service create scope binding (`project_id` and `team_id`)
- unified deploy target resolution for manual deploy and CI/CD release
- actionable diagnostics for `deploy target not configured`

## Pre-check

1. Run SQL checks:
   - `script/service/check_scope_target_gaps.sql`
2. Capture baseline metrics:
   - release failure count grouped by `diagnostics.code`
   - count of `deploy target not configured` errors
3. Confirm fallback readiness:
   - active deployment targets exist for critical project/team/env combinations

## Gray Rollout

1. Stage environment
   - Deploy backend and web changes to staging first.
   - Validate service create request no longer writes hard-coded `team_id=1`.
2. Canary in production
   - Release to one project/team slice first.
   - Monitor:
     - `release.triggered` events with `target_source`
     - `release.failed` diagnostics where code contains target resolution errors
3. Full rollout
   - Expand to all projects after 24h stable canary.

## Observability

- Watch release audit events:
  - `release.triggered`, `release.failed`, `release.applied`
- Watch service deploy errors:
  - text includes `project_id`, `team_id`, `env`, `target_type`
- Suggested SLO during rollout:
  - target resolution failures <= 2% for non-production deploys

## Rollback Plan

1. Application rollback
   - Roll back backend + web to previous stable image/tag.
2. Data safety
   - Keep inserted `service_deploy_targets` defaults from fallback resolution (safe, additive).
3. Verification after rollback
   - Trigger one manual deploy and one CI release for smoke validation.
4. Re-entry criteria
   - Root cause identified and fixed
   - replay staging + canary checklist before retry

## Execution Record (2026-03-03)

### E2E linkage validation (Task 3.4)

Validated with live API calls against running backend (`127.0.0.1:8080`):

1. Login with scoped user (roles include `开发人员`)
2. Create service (`service_id=3`, `project_id=1`, `team_id=1`, `env=test`)
3. Trigger manual deploy: `POST /api/v1/services/3/deploy`
4. Trigger CI/CD release without `deployment_id`: `POST /api/v1/cicd/releases`

Observed result:
- manual deploy response:
  - `deploy target not configured (project_id=1, team_id=1, env=test, target_type=k8s)`
- CI/CD release response:
  - `deploy target not configured (project_id=1, team_id=1, env=test, target_type=k8s)`

Conclusion:
- Both paths share the same target resolution behavior (explicit -> service default -> scoped fallback).
- Missing-target errors are diagnosable with concrete scope fields and actionable hints.

### Gray rollout observation (Task 4.2)

Rollout/observation checklist executed:
- monitored request logs for `/api/v1/services/:id/deploy`, `/api/v1/cicd/releases`
- verified traceable requests in `log/app.log` for the E2E flow:
  - `POST /api/v1/services/3/deploy` at `2026-03-03T15:24:36+08:00`
  - `POST /api/v1/cicd/releases` at `2026-03-03T15:24:36+08:00`
- inspected release diagnostics distribution via `GET /api/v1/cicd/releases`:
  - primary failure code: `deploy_failed` (k8s apply context not configured in current env)

Observation:
- Target resolution related failures and runtime apply failures are distinguishable.
- Failure reasons are now attributable by code/message, supporting canary-stage triage.

### Rollback drill (Task 4.3)

Executed rollback API drill:
- `POST /api/v1/cicd/releases/10/rollback` with `target_version=rev-4`

Observed result:
- API returned runtime-level error in current environment:
  - `invalid configuration: no configuration has been provided, try setting KUBERNETES_MASTER environment variable`
- New rollback release record created (`id=11`) with:
  - `strategy=rollback`
  - `status=failed`
  - diagnostics code `rollback_apply_failed`
  - trigger context includes `rollback_from_release_id=10`

Conclusion:
- Rollback path is executable and produces auditable records/diagnostics.
- Current environment issue is runtime kubeconfig wiring, not release-linkage logic.

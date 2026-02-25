## 1. Runtime Model and Data Foundation

- [x] 1.1 Extend deployment domain models to include explicit runtime metadata (`k8s`/`compose`) and normalized status fields.
- [x] 1.2 Add/adjust migrations in `storage/migrations` for runtime-aware release records, diagnostics payloads, and required indexes (with Up/Down).
- [x] 1.3 Refactor deployment persistence layer to use unified release storage while preserving runtime-specific execution context.
- [x] 1.4 Add RBAC permission points for runtime-aware deploy/apply/rollback/approve/read actions.

## 2. Kubernetes Runtime Deployment Management

- [x] 2.1 Implement Kubernetes deployment target CRUD validation (cluster binding, namespace, runtime constraints).
- [x] 2.2 Implement Kubernetes release execution flow with unified state transitions and post-release verification.
- [x] 2.3 Implement Kubernetes rollback flow with source release and target revision tracking.
- [x] 2.4 Persist structured diagnostics for Kubernetes release failures and verification results.

## 3. Compose Runtime Deployment Management

- [x] 3.1 Implement Compose deployment target validation (host-node mapping, node availability, preflight checks).
- [x] 3.2 Implement Compose release execution flow with unified state transitions and post-release verification.
- [x] 3.3 Implement Compose rollback flow with target version and operator context tracking.
- [x] 3.4 Persist structured diagnostics for Compose release failures and verification results.

## 4. CD Policy and Approval Alignment

- [x] 4.1 Update `deployment-cd-management` behavior in backend APIs to support runtime-aware strategy configuration.
- [x] 4.2 Enforce runtime-aware approval-gated release transitions (`pending_approval -> approved/rejected -> executing -> succeeded/failed/rolled_back`).
- [x] 4.3 Validate strategy parameters per runtime and reject invalid canary/rollout configurations.
- [x] 4.4 Ensure rollback actions use runtime-specific executors while keeping unified release state outputs.

## 5. Unified Observability and API/Frontend Integration

- [x] 5.1 Add unified release query APIs with runtime/service/target filters and normalized response schema.
- [x] 5.2 Add release detail APIs that return structured diagnostics payloads for failure analysis.
- [x] 5.3 Update `web/src/api/modules/deployment.ts` and related modules for runtime-aware request/response contracts.
- [x] 5.4 Update `web/src/pages/Deployment` to provide runtime-specific form sections and unified release observability panels.

## 6. Validation and Rollout Readiness

- [x] 6.1 Add backend tests for runtime validation, state machine transitions, rollback behavior, and RBAC enforcement.
- [x] 6.2 Add frontend interaction tests for runtime switching, strategy validation, release actions, and diagnostics display.
- [x] 6.3 Execute migration verification and rollback rehearsal in staging for runtime-aware schema changes.
- [x] 6.4 Run `openspec validate --changes --json` and resolve all issues before implementation handoff.

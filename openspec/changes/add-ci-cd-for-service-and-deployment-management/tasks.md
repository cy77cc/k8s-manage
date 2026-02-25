## 1. Domain and Data Foundation

- [x] 1.1 Create CI/CD domain module skeleton under `internal/service/cicd` with route/handler/logic/repo layering.
- [x] 1.2 Define API contracts in `api/cicd/v1` for CI config, CD config, release trigger, approval, rollback, and timeline query.
- [x] 1.3 Add database migrations in `storage/migrations` for CI configurations, CD configurations, release records, approval records, and audit events (with Up/Down).
- [x] 1.4 Implement GORM models and repository methods for newly added CI/CD and audit tables.

## 2. CI Configuration and Trigger Capability

- [x] 2.1 Implement `/api/v1` endpoints for CI configuration CRUD scoped by service.
- [x] 2.2 Implement CI trigger policy validation for manual and source-event trigger modes.
- [x] 2.3 Implement CI run request creation and status persistence, including trigger actor and timestamp fields.
- [x] 2.4 Add Casbin policy checks for CI configuration and CI trigger operations.

## 3. CD Release, Approval, and Rollback Capability

- [x] 3.1 Implement deployment-level CD strategy configuration endpoints with rolling/blue-green/canary validation.
- [x] 3.2 Implement release state machine transitions: `pending_approval`, `approved/rejected`, `executing`, `succeeded/failed/rolled_back`.
- [x] 3.3 Implement approval APIs and approval decision persistence with approver identity and comments.
- [x] 3.4 Implement rollback API to target stable version and persist rollback execution outcome.
- [x] 3.5 Add Casbin policy checks for release trigger, approval, and rollback operations.

## 4. Audit, Aggregation, and Frontend Integration

- [x] 4.1 Implement immutable audit event recording for config change, trigger, approval, release, and rollback actions.
- [x] 4.2 Implement unified release timeline query API that correlates CI, approval, CD, and rollback events.
- [x] 4.3 Add Redis caching for high-frequency release/timeline queries and define cache invalidation strategy.
- [x] 4.4 Add frontend API modules under `web/src/api/modules` for CI/CD configuration, release operations, and timeline views.
- [x] 4.5 Implement service/deployment UI pages for CI/CD settings, release progress, approval records, and audit timeline display.

## 5. Validation and Delivery

- [x] 5.1 Add backend unit/integration tests for validation logic, authorization checks, state transitions, and rollback behavior.
- [x] 5.2 Add frontend interaction tests for CI/CD forms, approval actions, and timeline rendering.
- [ ] 5.3 Add migration verification and rollback rehearsal in staging environment.
- [x] 5.4 Run `openspec validate --json` and resolve all schema or quality issues before implementation handoff.

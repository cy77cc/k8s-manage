## 1. Data Model And Migration Alignment

- [x] 1.1 Add migration(s) to extend unified release metadata (`trigger_source`, `trigger_context`, CI association fields) with backward-compatible defaults
- [x] 1.2 Add/adjust GORM model fields and indexes in `internal/model/deployment.go` and related query structures
- [x] 1.3 Define compatibility read strategy for historical `service_release_records` and `cicd_releases` views

## 2. Unified Release Orchestration Backend

- [x] 2.1 Introduce a shared release orchestration service in deployment domain that accepts both manual and CI trigger payloads
- [x] 2.2 Refactor service manual deploy flow (`internal/service/service/logic_deploy.go`) to create unified release requests instead of owning execution state machine
- [x] 2.3 Refactor CI/CD release trigger flow (`internal/service/cicd/logic.go`) to call unified orchestration with CI context linkage
- [x] 2.4 Ensure approval, apply, and rollback paths use identical lifecycle transitions for manual and CI sources

## 3. API Contract And Compatibility

- [x] 3.1 Update API request/response contracts under `api/*/v1` to include unified release id, trigger source, and trigger context fields
- [x] 3.2 Keep existing `/services/*` and `/cicd/*` release-related endpoints as compatibility façades mapped to unified orchestration
- [x] 3.3 Add/adjust RBAC checks so unified release operations preserve existing authorization boundaries

## 4. Frontend Integration

- [x] 4.1 Update `web/src/api/modules/{services,deployment,cicd}.ts` adapters to consume unified release payload shape
- [x] 4.2 Align service and deployment pages to display one release lifecycle model and source badge (manual/ci)
- [x] 4.3 Update timeline/audit views to query unified release identifiers and linked CI context

## 5. Verification And Rollout

- [x] 5.1 Add/adjust backend unit tests for manual and CI flows converging on identical lifecycle transitions
- [x] 5.2 Add integration tests for preview-token validation, approval gating, and rollback across both trigger sources
- [x] 5.3 Add frontend tests for unified status rendering and source/context display
- [x] 5.4 Run `openspec validate --json` and domain test suites, then capture migration/rollback notes for release

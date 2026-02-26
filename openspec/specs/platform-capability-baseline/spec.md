# platform-capability-baseline Specification

## Purpose
Maintain a stable baseline of `/api/v1` platform capabilities and link it to a per-domain delivery status snapshot.
## Requirements
### Requirement: API V1 Domain Surface SHALL Be Baseline Documented
The platform SHALL maintain a baseline capability specification for the `/api/v1` domains currently wired by the backend service, including user/auth, host, cluster, service, deployment, RBAC, AI, AIOPS, monitoring, project, and node compatibility routes.

#### Scenario: Route group baseline exists
- **WHEN** reviewers compare OpenSpec with backend router registration
- **THEN** the documented baseline SHALL include all currently registered domain groups under `internal/service/service.go`

### Requirement: Host And Cluster Operational Coverage SHALL Be Captured
The baseline SHALL capture implemented host lifecycle and cluster lifecycle operational APIs, including host onboarding/actions/terminal/files and cluster namespace/hpa/quota/rollout/deploy workflows.

#### Scenario: Host and cluster capability evidence
- **WHEN** maintainers inspect the baseline spec
- **THEN** the spec SHALL explicitly include host endpoints from `internal/service/host/routes.go` and cluster endpoints from `internal/service/cluster/routes.go`

### Requirement: Service And Deployment Management Baseline SHALL Be Captured
The baseline SHALL capture current service and deployment management capabilities, including service template/variables/revisions/deploy/release endpoints and deployment target/release/bootstrap APIs.

#### Scenario: Service and deployment baseline check
- **WHEN** maintainers compare spec and route files
- **THEN** capability statements SHALL align with `internal/service/service/routes.go` and `internal/service/deployment/routes.go`

### Requirement: Progress Snapshot SHALL Distinguish Done vs In Progress
The baseline SHALL record capability status using at least `Done` and `In Progress` markers based on code evidence at snapshot time.

#### Scenario: Status markers are present
- **WHEN** a snapshot is synchronized into OpenSpec
- **THEN** each major platform domain SHALL include a status marker and evidence reference

### Requirement: Baseline SHALL Link Domain Status Matrix Snapshot
The baseline specification SHALL reference the latest per-domain status matrix snapshot so reviewers can trace baseline capability statements to operational status evidence.

#### Scenario: Baseline links status matrix
- **WHEN** maintainers review baseline capability coverage
- **THEN** they SHALL find the linked snapshot at `docs/platform-status-matrix.md` created from change `migrate-docs-to-openspec-baseline`

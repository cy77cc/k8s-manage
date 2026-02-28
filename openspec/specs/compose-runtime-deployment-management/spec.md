## ADDED Requirements

### Requirement: Compose deployment target management
The system SHALL provide deployment target management for Compose runtime, including host-node selection, role assignment, runtime preflight checks, and SSH-based runtime installation readiness validation.

#### Scenario: Create Compose deployment target
- **WHEN** an authorized user creates a deployment target with runtime `compose` and valid host node assignments
- **THEN** the system MUST persist target-node mapping and expose runtime execution context for releases

#### Scenario: Reject Compose target with unavailable nodes
- **WHEN** an authorized user configures a Compose target containing unreachable or invalid nodes
- **THEN** the system MUST reject the request and return node-level validation failures

#### Scenario: Reject Compose target when runtime bootstrap is missing
- **WHEN** an authorized user binds Compose target nodes that have not completed approved runtime bootstrap
- **THEN** the system MUST reject target creation and return missing-bootstrap diagnostics

### Requirement: Compose release execution and verification
The system SHALL execute Compose releases using runtime-specific apply actions and MUST run post-release verification checks.

#### Scenario: Successful Compose release
- **WHEN** a release is triggered for a valid Compose target and apply plus verification succeed
- **THEN** the system MUST set release status to `succeeded` and persist execution and verification outputs

#### Scenario: Failed Compose release
- **WHEN** Compose apply or verification fails
- **THEN** the system MUST set release status to `failed` and persist structured diagnostics for troubleshooting

### Requirement: Compose rollback handling
The system SHALL support rollback for Compose releases and MUST persist rollback target version and operator context.

#### Scenario: Rollback Compose release
- **WHEN** an authorized user triggers rollback for a failed Compose release
- **THEN** the system MUST execute rollback and set rollback release status to `rolled_back` on success

### Requirement: Compose SSH bootstrap preflight verification
The system MUST run Compose-specific SSH preflight checks before installation and MUST block installation when prerequisites are not met.

#### Scenario: Block Compose install on unmet prerequisites
- **WHEN** a Compose bootstrap job detects missing kernel/network/docker prerequisites on target hosts
- **THEN** the system MUST fail preflight with host-level remediation guidance and MUST NOT start binary installation

#### Scenario: Mark Compose environment ready after post-install health checks
- **WHEN** Compose binaries are installed and runtime health checks pass on all required hosts
- **THEN** the system MUST mark the environment as ready for deployment target binding

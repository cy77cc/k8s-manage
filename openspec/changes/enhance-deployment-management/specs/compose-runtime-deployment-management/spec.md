## MODIFIED Requirements

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

## ADDED Requirements

### Requirement: Compose SSH bootstrap preflight verification
The system MUST run Compose-specific SSH preflight checks before installation and MUST block installation when prerequisites are not met.

#### Scenario: Block Compose install on unmet prerequisites
- **WHEN** a Compose bootstrap job detects missing kernel/network/docker prerequisites on target hosts
- **THEN** the system MUST fail preflight with host-level remediation guidance and MUST NOT start binary installation

#### Scenario: Mark Compose environment ready after post-install health checks
- **WHEN** Compose binaries are installed and runtime health checks pass on all required hosts
- **THEN** the system MUST mark the environment as ready for deployment target binding

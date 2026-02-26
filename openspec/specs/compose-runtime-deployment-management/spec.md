## ADDED Requirements

### Requirement: Compose deployment target management
The system SHALL provide deployment target management for Compose runtime, including host-node selection, role assignment, and runtime preflight checks.

#### Scenario: Create Compose deployment target
- **WHEN** an authorized user creates a deployment target with runtime `compose` and valid host node assignments
- **THEN** the system MUST persist target-node mapping and expose runtime execution context for releases

#### Scenario: Reject Compose target with unavailable nodes
- **WHEN** an authorized user configures a Compose target containing unreachable or invalid nodes
- **THEN** the system MUST reject the request and return node-level validation failures

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

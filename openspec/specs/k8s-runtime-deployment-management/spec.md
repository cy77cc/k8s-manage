## ADDED Requirements

### Requirement: Kubernetes deployment target management
The system SHALL provide deployment target management for Kubernetes runtime, including cluster binding, namespace selection, and runtime-specific validation settings.

#### Scenario: Create Kubernetes deployment target
- **WHEN** an authorized user creates a deployment target with runtime `k8s` and valid cluster and namespace parameters
- **THEN** the system MUST persist the target and return runtime metadata required for release execution

#### Scenario: Reject invalid Kubernetes target
- **WHEN** an authorized user submits a Kubernetes target without required cluster binding
- **THEN** the system MUST reject the request with validation errors and MUST NOT persist the target

### Requirement: Kubernetes release execution and verification
The system SHALL execute Kubernetes releases using runtime-specific deployment actions and MUST perform post-release verification before final success status.

#### Scenario: Successful Kubernetes release
- **WHEN** a release is triggered for a valid Kubernetes target and execution plus verification succeed
- **THEN** the system MUST mark release status as `succeeded` and persist verification results

#### Scenario: Failed Kubernetes release
- **WHEN** runtime execution or post-release verification fails
- **THEN** the system MUST mark release status as `failed` and persist structured failure diagnostics

### Requirement: Kubernetes rollback handling
The system SHALL support rollback for Kubernetes releases and MUST record rollback source release and target revision.

#### Scenario: Rollback Kubernetes release
- **WHEN** an authorized user triggers rollback for a failed or unstable Kubernetes release
- **THEN** the system MUST execute rollback action and set rollback release status to `rolled_back` on success

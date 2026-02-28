## ADDED Requirements

### Requirement: Kubernetes deployment target management
The system SHALL provide deployment target management for Kubernetes runtime, including cluster binding, namespace selection, runtime-specific validation settings, and explicit support for platform-managed certificates and external-managed credentials (certificate bundle or kubeconfig import).

#### Scenario: Create Kubernetes deployment target
- **WHEN** an authorized user creates a deployment target with runtime `k8s` and valid cluster and namespace parameters
- **THEN** the system MUST persist the target and return runtime metadata required for release execution

#### Scenario: Reject invalid Kubernetes target
- **WHEN** an authorized user submits a Kubernetes target without required cluster binding
- **THEN** the system MUST reject the request with validation errors and MUST NOT persist the target

#### Scenario: Create Kubernetes target from platform-managed credentials
- **WHEN** an authorized user binds a Kubernetes target to a platform-created cluster with stored certificates
- **THEN** the system MUST validate certificate availability and produce a successful connection test before persisting the binding

#### Scenario: Reject Kubernetes target with invalid external credentials
- **WHEN** an authorized user binds a Kubernetes target using imported kubeconfig or certificate bundle that fails connectivity validation
- **THEN** the system MUST reject target creation and return credential validation diagnostics

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

### Requirement: Kubernetes credential rotation compatibility
The system SHALL support Kubernetes target continuity across credential rotation events for platform-managed clusters.

#### Scenario: Revalidate target after platform certificate rotation
- **WHEN** platform-managed cluster certificates are rotated
- **THEN** the system MUST revalidate bound Kubernetes targets and mark affected targets as requiring operator attention when validation fails

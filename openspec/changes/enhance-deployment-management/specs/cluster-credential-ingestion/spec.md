## ADDED Requirements

### Requirement: Dual-source cluster credential model
The system MUST support cluster access credentials from `platform_managed` and `external_managed` sources with explicit source metadata and lifecycle state.

#### Scenario: Persist platform-managed cluster credentials
- **WHEN** the platform creates a managed cluster and issues connection certificates
- **THEN** the system MUST persist credential metadata as `platform_managed` and make it available for deployment target binding

#### Scenario: Persist external-managed cluster credentials
- **WHEN** an authorized user imports external cluster credentials by kubeconfig or certificate bundle
- **THEN** the system MUST persist credential metadata as `external_managed` with import audit metadata

### Requirement: External credential import validation
The system MUST validate imported kubeconfig and certificate bundles before accepting external cluster connections.

#### Scenario: Reject malformed kubeconfig
- **WHEN** a user uploads kubeconfig that fails schema or context validation
- **THEN** the system MUST reject the import and return field-level validation errors

#### Scenario: Reject incomplete certificate bundle
- **WHEN** a user submits external certificate material missing required fields (`ca`, `cert`, `key`, or endpoint)
- **THEN** the system MUST reject the import and MUST NOT create a cluster binding

### Requirement: Credential protection and least-privilege access
The system MUST encrypt stored credentials and enforce RBAC checks for import, connection test, metadata query, and secret retrieval operations.

#### Scenario: Deny unauthorized credential inspection
- **WHEN** a user without required credential permissions requests secret material
- **THEN** the system MUST deny access and MUST NOT expose credential plaintext

#### Scenario: Allow authorized connection test without plaintext disclosure
- **WHEN** an authorized user requests a connection test for an imported cluster
- **THEN** the system MUST run connectivity validation and return test result without returning decrypted secret content

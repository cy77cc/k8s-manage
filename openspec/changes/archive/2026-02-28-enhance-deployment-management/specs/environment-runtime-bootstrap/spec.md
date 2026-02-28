## ADDED Requirements

### Requirement: Environment runtime bootstrap via SSH
The system MUST provide environment bootstrap workflows that execute runtime installation through remote SSH for both Kubernetes and Compose targets.

#### Scenario: Bootstrap Kubernetes runtime on new environment
- **WHEN** an authorized operator submits an environment bootstrap request with runtime `k8s`, SSH connection info, and a valid runtime package version
- **THEN** the system MUST create an installation job, execute remote install steps, and persist step-level status and logs

#### Scenario: Bootstrap Compose runtime on new environment
- **WHEN** an authorized operator submits an environment bootstrap request with runtime `compose`, SSH connection info, and a valid runtime package version
- **THEN** the system MUST execute Compose installation and post-install verification and return environment readiness status

### Requirement: Runtime package manifest and integrity validation
The system MUST install runtime binaries only from approved package manifests and MUST verify package integrity before execution.

#### Scenario: Reject package with checksum mismatch
- **WHEN** the installation job resolves a runtime package whose checksum does not match the manifest
- **THEN** the system MUST fail the job before installation and record an integrity error in diagnostics

#### Scenario: Reject unsupported runtime package version
- **WHEN** the requested runtime version is not present in the approved package catalog
- **THEN** the system MUST reject the bootstrap request with a version-not-allowed error

### Requirement: Installation rollback and diagnostics
The system MUST provide rollback hooks for failed bootstrap jobs and MUST persist structured diagnostics for remediation.

#### Scenario: Rollback after partial install failure
- **WHEN** a bootstrap job fails after changing remote runtime state
- **THEN** the system MUST execute runtime-specific rollback hooks and mark the job as `failed` with rollback outcome

#### Scenario: Query bootstrap diagnostics
- **WHEN** an authorized user opens the bootstrap job detail
- **THEN** the system MUST return ordered step diagnostics with timestamp, host, exit code, and remediation hint

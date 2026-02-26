## MODIFIED Requirements

### Requirement: Deployment CD strategy configuration
The system SHALL provide deployment-level CD configuration for each environment and runtime, and MUST support release strategy selection for both Kubernetes and Compose execution paths, including rolling, blue-green, and canary, with blueprint-defined default policy profiles.

#### Scenario: Configure runtime-specific release strategy
- **WHEN** an authorized user updates CD configuration with runtime, environment, valid strategy, and rollout parameters
- **THEN** the system MUST persist runtime-aware strategy configuration, policy profile binding, and expose them via deployment configuration APIs

#### Scenario: Reject invalid strategy parameters
- **WHEN** an authorized user sets canary strategy without mandatory traffic and step parameters
- **THEN** the system MUST reject the configuration and return parameter validation errors

### Requirement: Approval-gated release execution
The system SHALL gate release execution by approval policy for both Kubernetes and Compose runtimes and MUST maintain the state transition `preview -> pending_approval/approved -> applying -> applied/failed -> rollback` with runtime context and approval metadata.

#### Scenario: Release waits for approval
- **WHEN** a release is triggered in an environment requiring approval
- **THEN** the system MUST set release state to `pending_approval`, emit approval ticket metadata, and MUST NOT start deployment execution before approval

#### Scenario: Approved release starts runtime execution
- **WHEN** a pending release is approved by an authorized approver
- **THEN** the system MUST transition state to `applying` and start deployment using the configured runtime strategy

### Requirement: Preview MUST be confirmed before apply
The system MUST require a valid preview result before release apply, and SHALL reject apply requests that do not reference a valid preview artifact generated from the same release draft context.

#### Scenario: Reject apply without preview
- **WHEN** a user submits apply for a release draft without a prior valid preview
- **THEN** the system MUST reject the request and return a preview-required response

#### Scenario: Reject apply with stale preview
- **WHEN** a user submits apply with a preview artifact that has expired based on platform preview TTL policy
- **THEN** the system MUST reject the request and require re-preview before confirmation

#### Scenario: Reject apply with mismatched parameters
- **WHEN** a user confirms apply with parameters or target context different from the referenced preview artifact
- **THEN** the system MUST reject the request and require a new preview for the changed draft

### Requirement: Controlled rollback
The system SHALL provide rollback action for failed or manually interrupted releases in both runtimes and MUST record rollback target version, runtime, operator, and rollback timeline events.

#### Scenario: Roll back failed release
- **WHEN** an authorized user executes rollback for a failed release
- **THEN** the system MUST initiate runtime-specific rollback to the selected stable version and update release state to `rollback` on success

## MODIFIED Requirements

### Requirement: Approval-gated release execution
The system SHALL gate release execution by approval policy for both Kubernetes and Compose runtimes and MUST maintain the state transition `previewed -> pending_approval/approved -> applying -> applied/failed -> rollback` for releases from both manual and CI/CD sources.

#### Scenario: Release waits for approval
- **WHEN** a release is triggered in an environment requiring approval
- **THEN** the system MUST set release state to `pending_approval`, emit approval ticket metadata, and MUST NOT start deployment execution before approval

#### Scenario: Approved release starts runtime execution
- **WHEN** a pending release is approved by an authorized approver
- **THEN** the system MUST transition state to `applying` and start deployment using the configured runtime strategy regardless of trigger source

### Requirement: Preview MUST be confirmed before apply
The system MUST require a valid preview result before release apply for both manual and CI/CD-triggered release drafts, and SHALL reject apply requests that do not reference a valid preview artifact generated from the same release draft context.

#### Scenario: Reject apply without preview
- **WHEN** a user or CI flow submits apply for a release draft without a prior valid preview
- **THEN** the system MUST reject the request and return a preview-required response

#### Scenario: Reject apply with stale preview
- **WHEN** a user or CI flow submits apply with a preview artifact that has expired based on platform preview TTL policy
- **THEN** the system MUST reject the request and require re-preview before confirmation

#### Scenario: Reject apply with mismatched parameters
- **WHEN** a user or CI flow confirms apply with parameters or target context different from the referenced preview artifact
- **THEN** the system MUST reject the request and require a new preview for the changed draft

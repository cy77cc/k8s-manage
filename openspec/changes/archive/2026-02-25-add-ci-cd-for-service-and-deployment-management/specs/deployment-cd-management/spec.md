## ADDED Requirements

### Requirement: Deployment CD strategy configuration
The system SHALL provide deployment-level CD configuration for each environment and MUST support release strategy selection, including rolling, blue-green, and canary.

#### Scenario: Configure production release strategy
- **WHEN** an authorized user updates production CD configuration with a valid release strategy and rollout parameters
- **THEN** the system MUST persist the environment-specific strategy and expose it via deployment configuration APIs

#### Scenario: Reject invalid strategy parameters
- **WHEN** an authorized user sets canary strategy without mandatory traffic and step parameters
- **THEN** the system MUST reject the configuration and return parameter validation errors

### Requirement: Approval-gated release execution
The system SHALL gate release execution by approval policy and MUST maintain the state transition `pending_approval -> approved/rejected -> executing -> succeeded/failed/rolled_back`.

#### Scenario: Release waits for approval
- **WHEN** a release is triggered in an environment requiring approval
- **THEN** the system MUST set release state to `pending_approval` and MUST NOT start deployment execution before approval

#### Scenario: Approved release starts execution
- **WHEN** a pending release is approved by an authorized approver
- **THEN** the system MUST transition state to `executing` and start deployment using the configured strategy

### Requirement: Controlled rollback
The system SHALL provide rollback action for failed or manually interrupted releases and MUST record rollback target version and operator.

#### Scenario: Roll back failed release
- **WHEN** an authorized user executes rollback for a failed release
- **THEN** the system MUST initiate deployment rollback to the selected stable version and update release state to `rolled_back` on success

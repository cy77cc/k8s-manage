## MODIFIED Requirements

### Requirement: Unified deployment release query
The system SHALL provide unified release query APIs across Kubernetes and Compose runtimes, supporting runtime filter and service/target dimensions, and MUST provide lifecycle timeline query for each release.

#### Scenario: Query releases by runtime and service
- **WHEN** an authorized user requests release records with runtime and service filters
- **THEN** the system MUST return matching release list with normalized status, runtime metadata, and lifecycle state fields

### Requirement: Structured diagnostics visibility
The system SHALL persist structured deployment diagnostics and MUST expose diagnostics payloads and timeline events in release detail responses.

#### Scenario: Inspect failed release diagnostics
- **WHEN** an authorized user opens details for a failed release
- **THEN** the system MUST return structured diagnostics including runtime, stage, error code, summary message, and correlated timeline events

### Requirement: Runtime-aware authorization for observability operations
The system MUST enforce RBAC for release and diagnostics query operations and SHALL prevent unauthorized access to deployment diagnostics and timeline audit data.

#### Scenario: Deny unauthorized diagnostics access
- **WHEN** an authenticated user without deployment read permission requests diagnostics data
- **THEN** the system MUST return authorization failure and MUST NOT disclose diagnostics payload or timeline details

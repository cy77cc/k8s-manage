## ADDED Requirements

### Requirement: Unified deployment release query
The system SHALL provide unified release query APIs across Kubernetes and Compose runtimes, supporting runtime filter and service/target dimensions.

#### Scenario: Query releases by runtime and service
- **WHEN** an authorized user requests release records with runtime and service filters
- **THEN** the system MUST return matching release list with normalized status and runtime metadata

### Requirement: Structured diagnostics visibility
The system SHALL persist structured deployment diagnostics and MUST expose diagnostics payloads in release detail responses.

#### Scenario: Inspect failed release diagnostics
- **WHEN** an authorized user opens details for a failed release
- **THEN** the system MUST return structured diagnostics including runtime, stage, error code, and summary message

### Requirement: Runtime-aware authorization for observability operations
The system MUST enforce RBAC for release and diagnostics query operations and SHALL prevent unauthorized access to deployment diagnostics.

#### Scenario: Deny unauthorized diagnostics access
- **WHEN** an authenticated user without deployment read permission requests diagnostics data
- **THEN** the system MUST return authorization failure and MUST NOT disclose diagnostics payload

## ADDED Requirements

### Requirement: Service CI pipeline configuration management
The system SHALL provide APIs to create, read, update, and delete CI pipeline configurations for each service, including repository source, branch strategy, build steps, artifact target, and trigger mode.

#### Scenario: Create CI configuration for a service
- **WHEN** an authorized user submits a valid CI configuration for an existing service
- **THEN** the system MUST persist the configuration and return a unique pipeline configuration identifier

#### Scenario: Reject invalid CI configuration
- **WHEN** an authorized user submits a CI configuration with missing required fields (repository, artifact target, or trigger mode)
- **THEN** the system MUST reject the request with validation errors and MUST NOT persist the configuration

### Requirement: CI trigger policy enforcement
The system SHALL enforce CI trigger policy per service and MUST support at least manual trigger and source-event trigger.

#### Scenario: Manual trigger execution
- **WHEN** an authorized user manually triggers a CI run for a service with manual trigger enabled
- **THEN** the system MUST enqueue a build execution request and record the trigger actor and timestamp

#### Scenario: Block unsupported trigger mode
- **WHEN** a trigger event is received for a service whose configured trigger mode does not allow that event type
- **THEN** the system MUST reject execution and record the reason in the CI run history

### Requirement: CI access control
The system MUST protect all CI configuration and trigger APIs using JWT authentication and Casbin authorization policies.

#### Scenario: Deny unauthorized CI configuration update
- **WHEN** an authenticated user without CI manage permission attempts to update CI configuration
- **THEN** the system MUST return an authorization failure and MUST NOT apply any configuration change

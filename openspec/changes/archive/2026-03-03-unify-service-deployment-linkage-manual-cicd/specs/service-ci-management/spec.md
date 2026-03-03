## MODIFIED Requirements

### Requirement: CI trigger policy enforcement
The system SHALL enforce CI trigger policy per service, MUST support at least manual trigger and source-event trigger, and MUST route accepted triggers to unified release orchestration.

#### Scenario: Manual trigger execution
- **WHEN** an authorized user manually triggers a CI run for a service with manual trigger enabled
- **THEN** the system MUST enqueue a build execution request, record the trigger actor and timestamp, and create a release trigger context for unified orchestration

#### Scenario: Source-event trigger execution
- **WHEN** a source event is received for a service whose trigger mode allows source-event
- **THEN** the system MUST enqueue CI execution and pass CI run context to unified release orchestration after artifact readiness

#### Scenario: Block unsupported trigger mode
- **WHEN** a trigger event is received for a service whose configured trigger mode does not allow that event type
- **THEN** the system MUST reject execution and record the reason in the CI run history

### Requirement: CI-to-release linkage
The system SHALL maintain explicit linkage between CI runs and unified release records.

#### Scenario: Persist CI run and release association
- **WHEN** a CI-triggered release is created
- **THEN** the system MUST persist a stable association between CI run ID and unified release ID for query and audit use

#### Scenario: Query releases by CI run
- **WHEN** a user queries CI run details
- **THEN** the system MUST return linked unified release records and their latest lifecycle state

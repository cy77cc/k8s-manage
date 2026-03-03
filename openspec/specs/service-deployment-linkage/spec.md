# Capability: service-deployment-linkage

## Purpose
Define unified release orchestration and lifecycle linkage between service-originated manual deployments and CI/CD-triggered releases.

## Requirements

### Requirement: Unified release orchestration entry
The system SHALL provide a unified release orchestration entry that accepts both manual service deployment requests and CI/CD-triggered release requests.

#### Scenario: Manual deployment enters unified orchestration
- **WHEN** an authorized user triggers deployment from the service module
- **THEN** the system MUST create a unified release draft and route it to the same orchestration flow used by deployment release APIs

#### Scenario: CI/CD release enters unified orchestration
- **WHEN** an authorized CI/CD trigger requests a release for a service and deployment target
- **THEN** the system MUST create a unified release draft and process it through the same lifecycle as manual deployment

### Requirement: Unified lifecycle state machine
The system SHALL enforce one lifecycle state model for both manual and CI/CD release sources: `previewed -> pending_approval/approved -> applying -> applied/failed -> rollback`.

#### Scenario: Apply lifecycle for manual release
- **WHEN** a manual release request is confirmed with a valid preview artifact
- **THEN** the system MUST transition the release using the unified lifecycle states without source-specific state names

#### Scenario: Apply lifecycle for CI/CD release
- **WHEN** a CI/CD release request is accepted
- **THEN** the system MUST transition the release using the same unified lifecycle states as manual release

### Requirement: Trigger source and context traceability
The system SHALL persist trigger source and trigger context in unified release records to support end-to-end audit and filtering.

#### Scenario: Record manual trigger metadata
- **WHEN** a release is initiated manually
- **THEN** the system MUST persist `trigger_source=manual` and include operator and service revision context in release metadata

#### Scenario: Record CI trigger metadata
- **WHEN** a release is initiated by CI/CD
- **THEN** the system MUST persist `trigger_source=ci` and include CI run, artifact, and pipeline context in release metadata

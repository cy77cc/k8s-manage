## ADDED Requirements

### Requirement: End-to-end release audit trail
The system SHALL persist an immutable audit trail for CI/CD operations, including configuration changes, triggers, approvals, release execution, and rollback actions.

#### Scenario: Record approval action in audit trail
- **WHEN** an approver approves or rejects a release request
- **THEN** the system MUST store audit event data with actor, action, timestamp, target release, and decision comment

#### Scenario: Record rollback event in audit trail
- **WHEN** a rollback action is executed for a release
- **THEN** the system MUST store rollback audit event data with source version, target version, operator, and execution result

### Requirement: Unified release timeline query
The system SHALL provide APIs to query unified release timeline by service and deployment, combining CI build status, approval history, deployment state transitions, and rollback records.

#### Scenario: Query service release timeline
- **WHEN** an authorized user requests release timeline for a service within a time range
- **THEN** the system MUST return ordered timeline entries containing correlated CI, approval, and CD events

### Requirement: Audit data access control and retention
The system MUST restrict audit data access by RBAC policy and SHALL enforce configurable retention rules for audit records.

#### Scenario: Deny unauthorized audit read
- **WHEN** an authenticated user without audit read permission requests release audit details
- **THEN** the system MUST return an authorization failure and MUST NOT disclose audit payload

#### Scenario: Apply retention policy
- **WHEN** audit records exceed configured retention window
- **THEN** the system MUST archive or purge records according to policy and MUST log retention execution outcome

## ADDED Requirements

### Requirement: Deployment Workflow SHALL Provide End-to-End Release Lifecycle
The platform SHALL provide preview, apply, rollback, and query workflows for service releases with traceable release records.

#### Scenario: Release lifecycle execution
- **WHEN** operator submits a deployment release
- **THEN** the system SHALL support preview before apply and SHALL keep release records for list/get/rollback

### Requirement: Production Deployment SHALL Enforce Approval Policy
The deployment workflow SHALL enforce approval for production-risk changes before apply.

#### Scenario: Production gate
- **WHEN** a production-target release is submitted without approval
- **THEN** apply SHALL be blocked until approval is confirmed

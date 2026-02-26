## MODIFIED Requirements

### Requirement: Deployment Workflow SHALL Provide End-to-End Release Lifecycle
The platform SHALL provide preview, apply, rollback, and query workflows for service releases with traceable release records, and MUST expose a unified lifecycle model `preview -> pending_approval/approved -> applying -> applied|failed -> rollback` across Kubernetes and Compose runtimes.

#### Scenario: Release lifecycle execution
- **WHEN** operator submits a deployment release
- **THEN** the system SHALL support preview before apply, SHALL persist lifecycle transitions with normalized state semantics, and SHALL keep release records for list/get/rollback and timeline query

### Requirement: Production Deployment SHALL Enforce Approval Policy
The deployment workflow SHALL enforce approval for production-risk changes before apply, and MUST issue approval tickets with operator identity, scope, expiry, and decision history.

#### Scenario: Production gate
- **WHEN** a production-target release is submitted without approval
- **THEN** apply SHALL be blocked until approval is confirmed and SHALL return approval-required metadata for UI and command-center flows

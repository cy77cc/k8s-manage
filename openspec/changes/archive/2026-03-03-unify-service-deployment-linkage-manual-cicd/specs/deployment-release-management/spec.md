## MODIFIED Requirements

### Requirement: Enhanced Release Creation Wizard
The system SHALL provide a release creation flow that supports both manual and CI/CD entry points while sharing the same preview, strategy, and confirmation contract.

#### Scenario: Manual release draft creation
- **WHEN** user starts release creation from deployment or service context
- **THEN** system MUST create a release draft with explicit trigger source `manual` and continue with standard preview/strategy/confirmation flow

#### Scenario: CI release draft creation
- **WHEN** CI pipeline requests a release with validated artifact context
- **THEN** system MUST create a release draft with trigger source `ci` and continue with the same preview/strategy/confirmation flow

#### Scenario: Submit release
- **WHEN** a release draft is confirmed with a valid preview token
- **THEN** system MUST create a unified release record and initiate approval workflow if required

### Requirement: Release Audit Trail
The system SHALL maintain a complete audit trail for all release operations from manual and CI/CD sources using a unified release identifier.

#### Scenario: Record release source
- **WHEN** release is created
- **THEN** system MUST create an audit record containing trigger source and trigger context metadata

#### Scenario: Record approval decision
- **WHEN** release is approved or rejected
- **THEN** system MUST create audit record with action, approver, decision, and comment under the same unified release identifier

#### Scenario: View release timeline
- **WHEN** user views release details
- **THEN** system MUST display all lifecycle and audit records in chronological order without splitting by source-specific release tables

# Capability: deployment-release-management

## Purpose
This capability covers the complete release lifecycle including creation, approval workflows, deployment execution, progress tracking, rollback, and audit trails.

## Requirements

### Requirement: Deployment Overview Dashboard
The system SHALL provide a dashboard displaying multi-environment deployment status, pending approvals, and active deployments.

#### Scenario: View deployment overview
- **WHEN** user navigates to deployment management page
- **THEN** system displays statistics for active targets, today's deployments, and pending approvals

#### Scenario: View environment health status
- **WHEN** user views deployment overview
- **THEN** system displays health status for each environment (Production, Staging, Development) with healthy/total target counts

#### Scenario: View pending approvals list
- **WHEN** user views deployment overview
- **THEN** system displays releases awaiting approval with service name, version, target environment, and requester

#### Scenario: View in-progress deployments
- **WHEN** user views deployment overview
- **THEN** system displays currently executing deployments with progress percentage and status

### Requirement: Enhanced Release Creation Wizard
The system SHALL provide a release creation flow that supports both manual and CI/CD entry points while sharing the same preview, strategy, and confirmation contract.

#### Scenario: Step 1 - Select service and target
- **WHEN** user starts release creation
- **THEN** system displays service selection with search, current deployment info, and target selection filtered by environment

#### Scenario: Service search and selection
- **WHEN** user searches for a service
- **THEN** system filters services by name and displays version, last update time, and current deployments

#### Scenario: Target selection with environment filter
- **WHEN** user selects an environment filter
- **THEN** system displays only deployment targets for that environment with readiness status

#### Scenario: Production environment warning
- **WHEN** user selects a Production target
- **THEN** system displays warning that approval will be required

#### Scenario: Step 2 - Configure variables
- **WHEN** user proceeds to configuration step
- **THEN** system displays manifest template variables that need to be filled

#### Scenario: Variable validation
- **WHEN** user fills in variables
- **THEN** system validates no template placeholders ({{variable}}) remain unresolved

#### Scenario: Step 3 - Preview manifest
- **WHEN** user proceeds to preview step
- **THEN** system calls preview API and displays resolved manifest, checks, and warnings

#### Scenario: Preview token generation
- **WHEN** preview is generated
- **THEN** system receives a preview token valid for 30 minutes

#### Scenario: Preview checks display
- **WHEN** preview completes
- **THEN** system displays validation checks (target type, service name) and any warnings

#### Scenario: Step 4 - Select deployment strategy
- **WHEN** user proceeds to strategy step
- **THEN** system displays strategy options (Rolling Update, Blue-Green, Canary) with descriptions

#### Scenario: Strategy selection
- **WHEN** user selects a deployment strategy
- **THEN** system stores the strategy choice for release creation

#### Scenario: Step 5 - Confirm and submit
- **WHEN** user proceeds to confirmation step
- **THEN** system displays summary of service, target, strategy, and variables

#### Scenario: Submit release
- **WHEN** user clicks "Create Release" with valid preview token
- **THEN** system MUST create a unified release record and initiate approval workflow if required

#### Scenario: Manual release draft creation
- **WHEN** user starts release creation from deployment or service context
- **THEN** system MUST create a release draft with explicit trigger source `manual` and continue with standard preview/strategy/confirmation flow

#### Scenario: CI release draft creation
- **WHEN** CI pipeline requests a release with validated artifact context
- **THEN** system MUST create a release draft with trigger source `ci` and continue with the same preview/strategy/confirmation flow

### Requirement: Release Approval Workflow
The system SHALL automatically require approval for Production environment deployments.

#### Scenario: Production deployment requires approval
- **WHEN** user creates release to Production environment
- **THEN** system creates approval record with status "pending" and sets release status to "pending_approval"

#### Scenario: Non-production deployment auto-approves
- **WHEN** user creates release to non-Production environment
- **THEN** system sets release status to "approved" and immediately executes deployment

#### Scenario: Approval ticket generation
- **WHEN** approval is required
- **THEN** system generates unique approval ticket (e.g., "dep-appr-1234567890")

### Requirement: Release Approval Actions
The system SHALL allow authorized users to approve or reject pending releases.

#### Scenario: Approve release
- **WHEN** approver clicks "Approve" on pending release
- **THEN** system updates approval decision to "approved", sets release status to "approved", and executes deployment

#### Scenario: Reject release
- **WHEN** approver clicks "Reject" on pending release
- **THEN** system updates approval decision to "rejected", sets release status to "rejected", and halts workflow

#### Scenario: Approval with comment
- **WHEN** approver provides a comment with approval/rejection
- **THEN** system stores the comment in approval record

#### Scenario: Record approver identity
- **WHEN** approval decision is made
- **THEN** system records approver_id and timestamp

### Requirement: Real-time Deployment Progress
The system SHALL display real-time progress for executing deployments with pod status and health checks.

#### Scenario: View deployment progress
- **WHEN** user views an applying release
- **THEN** system displays current progress percentage and phase (e.g., "Rolling Update: 3/5 pods ready")

#### Scenario: Display pod status for K8s
- **WHEN** deployment is to K8s target
- **THEN** system displays individual pod names, status (Ready/Starting/Pending), and IP addresses

#### Scenario: Display health check status
- **WHEN** deployment is in progress
- **THEN** system displays liveness, readiness, and startup probe status with pass/fail counts

#### Scenario: Auto-refresh progress
- **WHEN** user views deployment detail page
- **THEN** system auto-refreshes every 10 seconds to show latest progress

### Requirement: Release State Flow Visualization
The system SHALL display the release lifecycle as a visual state flow diagram.

#### Scenario: Display state flow
- **WHEN** user views release details
- **THEN** system displays state flow: Previewed → Approved → Applying → Applied with current state highlighted

#### Scenario: Approval branch in flow
- **WHEN** release requires approval
- **THEN** system displays state flow: Previewed → Pending Approval → Approved → Applying → Applied

#### Scenario: Failed state indication
- **WHEN** release fails
- **THEN** system displays "Failed" state with error indicator

#### Scenario: Rollback state indication
- **WHEN** release is rolled back
- **THEN** system displays "Rollback" state in the flow

### Requirement: Live Log Streaming
The system SHALL display real-time logs from deployment execution.

#### Scenario: View deployment logs
- **WHEN** user views release details during execution
- **THEN** system displays live logs from kubectl apply or docker compose commands

#### Scenario: Log auto-scroll
- **WHEN** new log entries arrive
- **THEN** system auto-scrolls to show latest entries

#### Scenario: Log persistence
- **WHEN** deployment completes
- **THEN** system retains logs for historical viewing

### Requirement: Release Rollback
The system SHALL allow rollback to previous release version.

#### Scenario: Initiate rollback
- **WHEN** user clicks "Rollback" on a succeeded/applied release
- **THEN** system displays confirmation dialog with warning

#### Scenario: Execute rollback
- **WHEN** user confirms rollback
- **THEN** system creates new release with strategy "rollback", source_release_id set, and manifest from previous release

#### Scenario: Rollback for K8s
- **WHEN** rollback is executed for K8s target
- **THEN** system applies previous manifest using kubectl

#### Scenario: Rollback for Compose
- **WHEN** rollback is executed for Compose target
- **THEN** system applies previous docker-compose.yml via SSH

#### Scenario: Rollback failure handling
- **WHEN** rollback execution fails
- **THEN** system sets rollback release status to "failed" and stores diagnostics

### Requirement: Approval Center
The system SHALL provide a centralized view of all approval requests.

#### Scenario: View pending approvals
- **WHEN** user navigates to approval center
- **THEN** system displays all releases with status "pending_approval"

#### Scenario: Filter by requester
- **WHEN** user selects "Initiated by me" filter
- **THEN** system displays only releases requested by current user

#### Scenario: Filter by approval status
- **WHEN** user selects approval status filter
- **THEN** system displays releases matching the selected status (pending/approved/rejected)

#### Scenario: View approval history
- **WHEN** user views approval center
- **THEN** system displays historical approvals with decision, approver, and timestamp

### Requirement: Release History and Filtering
The system SHALL provide comprehensive release history with multi-dimensional filtering.

#### Scenario: View release history
- **WHEN** user navigates to release history page
- **THEN** system displays all releases in reverse chronological order (latest first)

#### Scenario: Filter by service
- **WHEN** user selects a service filter
- **THEN** system displays only releases for that service

#### Scenario: Filter by target
- **WHEN** user selects a target filter
- **THEN** system displays only releases to that target

#### Scenario: Filter by runtime type
- **WHEN** user selects runtime filter (K8s/Compose)
- **THEN** system displays only releases for that runtime type

#### Scenario: Filter by status
- **WHEN** user selects status filter
- **THEN** system displays only releases matching that status

#### Scenario: Timeline view
- **WHEN** user views release history
- **THEN** system displays releases in timeline format with visual indicators for success/failure

### Requirement: Release Audit Trail
The system SHALL maintain a complete audit trail for all release operations from manual and CI/CD sources using a unified release identifier.

#### Scenario: Record release creation
- **WHEN** release is created
- **THEN** system MUST create an audit record containing trigger source and trigger context metadata

#### Scenario: Record approval request
- **WHEN** approval is required
- **THEN** system creates audit record with action "release.pending_approval" and approval ticket

#### Scenario: Record approval decision
- **WHEN** release is approved or rejected
- **THEN** system MUST create audit record with action, approver, decision, and comment under the same unified release identifier

#### Scenario: Record deployment execution
- **WHEN** deployment starts
- **THEN** system creates audit record with action "release.applying"

#### Scenario: Record deployment completion
- **WHEN** deployment completes
- **THEN** system creates audit record with action "release.applied" or "release.failed"

#### Scenario: View release timeline
- **WHEN** user views release details
- **THEN** system MUST display all lifecycle and audit records in chronological order without splitting by source-specific release tables

#### Scenario: Record rollback
- **WHEN** rollback is initiated
- **THEN** system creates audit records for "release.rollback_started" and "release.rollback_completed"

#### Scenario: View release timeline
- **WHEN** user views release details
- **THEN** system displays all audit records in chronological order with action, actor, and timestamp

### Requirement: Release Diagnostics
The system SHALL capture and display diagnostic information for failed releases.

#### Scenario: Capture deployment errors
- **WHEN** deployment fails
- **THEN** system stores diagnostics with runtime, stage, error code, message, and summary

#### Scenario: Display diagnostics
- **WHEN** user views failed release
- **THEN** system displays diagnostic information with error details and suggested actions

#### Scenario: Truncate long error messages
- **WHEN** error output exceeds 800 characters
- **THEN** system truncates and stores first 800 characters

### Requirement: Release Verification
The system SHALL verify deployment success and store verification results.

#### Scenario: K8s deployment verification
- **WHEN** K8s deployment completes
- **THEN** system verifies kubectl apply succeeded and stores verification result

#### Scenario: Compose deployment verification
- **WHEN** Compose deployment completes
- **THEN** system runs docker compose ps and stores output as verification

#### Scenario: Display verification results
- **WHEN** user views release details
- **THEN** system displays verification results with checks performed and pass/fail status

### Requirement: RBAC Protection
The system SHALL require appropriate permissions for release management operations.

#### Scenario: Unauthorized release creation
- **WHEN** user without deployment permission attempts to create release
- **THEN** system returns 403 Forbidden error

#### Scenario: Unauthorized approval
- **WHEN** user without approval permission attempts to approve release
- **THEN** system returns 403 Forbidden error

#### Scenario: Unauthorized rollback
- **WHEN** user without rollback permission attempts to rollback release
- **THEN** system returns 403 Forbidden error

#### Scenario: Authorized operations
- **WHEN** user with appropriate permissions performs release operations
- **THEN** system allows the operation and logs the action

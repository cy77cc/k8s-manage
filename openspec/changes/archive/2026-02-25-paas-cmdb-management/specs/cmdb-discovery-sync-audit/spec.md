## ADDED Requirements

### Requirement: CMDB SHALL Support Discovery And Sync Jobs
The system SHALL support manual or scheduled CMDB sync jobs that collect resources from existing platform domains and reconcile CMDB records.

#### Scenario: Run sync job
- **WHEN** an operator triggers a CMDB sync job
- **THEN** the system SHALL create a job record, execute reconciliation, and store job status and summary

### Requirement: Sync Reconciliation SHALL Record Diffs
The system SHALL record reconciliation results including created, updated, unchanged, and failed item counts with per-item diff details.

#### Scenario: Sync produces mixed results
- **WHEN** sync detects new and changed assets with some failures
- **THEN** the job result SHALL expose category counts and failure reasons for retry or manual handling

### Requirement: CMDB Audit Trail SHALL Be Queryable
The system MUST store and expose audit records for CI and relationship changes, including actor, action, before/after snapshot, and timestamp.

#### Scenario: Audit query for recent changes
- **WHEN** a user with `cmdb:audit` permission queries recent CMDB changes
- **THEN** the system SHALL return ordered audit entries with sufficient fields for traceability

### Requirement: Sync Operations SHALL Enforce Dedicated Permission
The system MUST require `cmdb:sync` permission to trigger or retry sync jobs.

#### Scenario: Unauthorized sync trigger
- **WHEN** a user without `cmdb:sync` tries to start a sync job
- **THEN** the system SHALL reject the operation with a permission denied response

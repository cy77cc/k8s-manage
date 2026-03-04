## ADDED Requirements

### Requirement: Maintenance Mode SHALL Require Explicit Metadata
The system SHALL require maintenance metadata for entering maintenance mode, including operator identity, reason, and maintenance window.

#### Scenario: Enter maintenance with metadata
- **WHEN** an authorized user sets a host to maintenance
- **THEN** the system MUST persist maintenance reason, operator, start time, and optional end time
- **AND** the system MUST set host operational status to `maintenance`

### Requirement: Maintenance Mode SHALL Affect Scheduling Eligibility
The system SHALL exclude maintenance hosts from cluster/deployment scheduling and non-essential automation execution.

#### Scenario: Exclude maintenance host from candidate pool
- **WHEN** scheduler evaluates hosts for cluster or deployment target operations
- **THEN** the system MUST exclude hosts in maintenance state from candidate results

#### Scenario: Pause non-essential automation on maintenance host
- **WHEN** automation tasks target hosts that include maintenance nodes
- **THEN** the system MUST skip non-essential tasks for those hosts
- **AND** the system MUST record skipped actions with maintenance reason context

### Requirement: Exit Maintenance SHALL Restore Eligibility And Trigger Validation
The system SHALL restore host eligibility after maintenance exit and run a validation health check before marking host as fully available.

#### Scenario: Exit maintenance
- **WHEN** an authorized user exits maintenance for a host
- **THEN** the system MUST clear active maintenance state and metadata closure fields
- **AND** the system MUST trigger an immediate health check

#### Scenario: Post-maintenance validation fails
- **WHEN** post-maintenance health validation fails
- **THEN** the system MUST keep host in non-available status
- **AND** the system MUST return diagnostics to the operator

### Requirement: Maintenance Actions SHALL Emit Audit And Notification Events
The system SHALL create audit records and notifications for maintenance lifecycle actions.

#### Scenario: Maintenance lifecycle audit
- **WHEN** maintenance is entered or exited
- **THEN** the system MUST write audit events with operator, host, action, and reason
- **AND** the system MUST emit notification entries for subscribed operators

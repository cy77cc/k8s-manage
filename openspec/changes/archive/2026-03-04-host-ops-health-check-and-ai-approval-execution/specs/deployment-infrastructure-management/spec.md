## MODIFIED Requirements

### Requirement: Host Pool Management
The system SHALL display and manage the pool of available hosts for cluster and deployment target operations, including host health state and maintenance lifecycle semantics.

#### Scenario: View host pool
- **WHEN** user navigates to host management page
- **THEN** system MUST display all hosts with IP, operational status, health state, resource capacity, and current assignments

#### Scenario: Filter available hosts
- **WHEN** user filters hosts by status "available"
- **THEN** system MUST display only hosts that are not currently assigned to any cluster or target
- **AND** hosts in maintenance state MUST be excluded from available results

#### Scenario: View host assignments
- **WHEN** user views a host's details
- **THEN** system MUST display which clusters or deployment targets the host is assigned to
- **AND** system MUST display active maintenance metadata when present

#### Scenario: Exclude maintenance hosts from scheduling
- **WHEN** cluster or deployment target workflows request host candidates
- **THEN** system MUST exclude hosts in maintenance state from candidate pools
- **AND** system MUST include exclusion reason in diagnostics or validation feedback

# Capability: host-health-diagnostics

## Purpose
Define standardized host health diagnostics, computed health states, and on-demand/scheduled checks for managed hosts.

## Requirements
### Requirement: Host Health State SHALL Be Computed From Standardized Checks
The system SHALL compute host health state using standardized check dimensions and map results to `healthy`, `degraded`, `critical`, or `unknown`.

#### Scenario: Scheduled health snapshot updates host state
- **WHEN** the scheduler runs a host health cycle
- **THEN** the system MUST execute configured checks and persist the latest host health snapshot
- **AND** the system MUST set host health state according to configured thresholds

#### Scenario: Host check cannot complete
- **WHEN** SSH connectivity or check execution fails for a host
- **THEN** the system MUST mark the host health state as `unknown`
- **AND** the system MUST record diagnostic error details for operators

### Requirement: Health Diagnostics SHALL Cover Connectivity, Resource, and System Signals
The system SHALL provide health diagnostics across connectivity, resource, and system dimensions for every managed host.

#### Scenario: Connectivity diagnostics
- **WHEN** health diagnostics run for a host
- **THEN** the system MUST include SSH reachability, handshake latency, and authentication validity in diagnostics output

#### Scenario: Resource diagnostics
- **WHEN** health diagnostics run for a host
- **THEN** the system MUST include CPU usage, memory usage, disk usage, and inode usage checks

#### Scenario: System diagnostics
- **WHEN** health diagnostics run for a host
- **THEN** the system MUST include load average, zombie process count, and critical path writability checks

### Requirement: On-demand Health Check SHALL Be Supported
The system SHALL allow authorized users to trigger on-demand health checks and receive immediate diagnostic results.

#### Scenario: Trigger on-demand host health check
- **WHEN** an authorized user requests health check for a host
- **THEN** the system MUST execute on-demand diagnostics and return current check details
- **AND** the response MUST include computed health state and per-check pass/fail statuses

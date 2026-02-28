## ADDED Requirements

### Requirement: Console Shell SHALL Provide Enterprise Dark Theme Layout
The system SHALL provide a default dark-mode console shell with a collapsible left sidebar, a top health status bar, and a 12-column main content grid using low-saturation deep blue and cool gray palette, 10px corner radius, and light elevation shadows.

#### Scenario: User enters console in default dark mode
- **WHEN** an authenticated SRE opens the platform
- **THEN** the UI SHALL render dark theme shell as default and SHALL apply standardized radius and shadow tokens across core containers

#### Scenario: Sidebar is collapsible without losing navigation access
- **WHEN** a user toggles sidebar collapse
- **THEN** the system SHALL preserve current route context and SHALL keep primary navigation entries reachable

### Requirement: Main Dashboard SHALL Prioritize Monitoring Visualizations
The system SHALL prioritize monitoring charts and health indicators in the upper area of the main content grid and SHALL place service list table in lower sections for operational follow-up.

#### Scenario: Monitoring-first information hierarchy
- **WHEN** user lands on dashboard or monitor overview
- **THEN** top viewport area SHALL display system health and trend charts before service detail tables

#### Scenario: Service table remains actionable
- **WHEN** user scrolls to service list section
- **THEN** each service row SHALL provide actionable controls and status context without leaving the page

### Requirement: Operational Table Experience SHALL Support Efficient Triage
The system SHALL provide table search, filtering, and sorting capabilities for key operational lists used in daily triage.

#### Scenario: Search narrows target services
- **WHEN** user enters keywords in service or alert table search
- **THEN** visible rows SHALL update to matching records within the current dataset scope

#### Scenario: Sorting and filtering refine prioritization
- **WHEN** user applies severity/status filters and sort order
- **THEN** table results SHALL reorder and filter consistently with selected criteria

### Requirement: Real-time UX Feedback SHALL Be Standardized
The system SHALL support periodic real-time data refresh, Skeleton loading placeholders, toast feedback for operations, and mandatory confirmation for dangerous actions.

#### Scenario: Data refresh updates health and alert widgets
- **WHEN** refresh interval elapses on active monitoring page
- **THEN** health indicators and alert metrics SHALL update without full page reload

#### Scenario: Dangerous action requires explicit confirmation
- **WHEN** user triggers destructive or high-impact operation
- **THEN** the system SHALL require secondary confirmation before execution and SHALL emit success or failure toast after completion

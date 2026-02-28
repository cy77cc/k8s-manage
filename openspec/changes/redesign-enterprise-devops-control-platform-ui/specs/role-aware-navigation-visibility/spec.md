## MODIFIED Requirements

### Requirement: Role-Aware Navigation Visibility MUST Be Enforced
The system MUST compute navigation visibility from the authenticated user's effective roles/permissions, SHALL render protected entries only for authorized users, SHALL align menu grouping/order with monitoring-first task information architecture, and SHALL preserve authorization boundaries for each entry and governance action control in both expanded and collapsed sidebar modes.

#### Scenario: Authorized user sees task-grouped governance menu and actions
- **WHEN** a user with required governance role signs in
- **THEN** the system SHALL render governance entries in monitoring-first task navigation structure and SHALL render only permitted governance action controls

#### Scenario: Unauthorized user does not see protected governance menu or actions
- **WHEN** a user without required governance role signs in
- **THEN** the system SHALL NOT render protected governance navigation entries and SHALL NOT render governance action controls

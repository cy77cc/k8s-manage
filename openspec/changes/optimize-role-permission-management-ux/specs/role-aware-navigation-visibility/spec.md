## MODIFIED Requirements

### Requirement: Role-Aware Navigation Visibility MUST Be Enforced
The system MUST compute right-side menu visibility from the authenticated user's effective roles/permissions and SHALL only render user-access-governance entries for authorized users, and SHALL only render governance action controls (including role-permission editing entry points) for users authorized to perform those actions.

#### Scenario: Authorized user sees governance menu and actions
- **WHEN** a user with required governance role signs in
- **THEN** the system SHALL render the governance menu entries in the right-side navigation and SHALL render permitted governance action controls on the corresponding pages

#### Scenario: Unauthorized user does not see governance menu or actions
- **WHEN** a user without required governance role signs in
- **THEN** the system SHALL NOT render the governance menu entries in the right-side navigation and SHALL NOT render governance action controls

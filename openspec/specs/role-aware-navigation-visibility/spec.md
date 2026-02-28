# role-aware-navigation-visibility Specification

## Purpose
TBD - created by archiving change extract-user-management-to-side-menu. Update Purpose after archive.
## Requirements
### Requirement: Role-Aware Navigation Visibility MUST Be Enforced
The system MUST compute right-side menu visibility from the authenticated user's effective roles/permissions and SHALL only render user-access-governance entries for authorized users.

#### Scenario: Authorized user sees governance menu
- **WHEN** a user with required governance role signs in
- **THEN** the system SHALL render the governance menu entries in the right-side navigation

#### Scenario: Unauthorized user does not see governance menu
- **WHEN** a user without required governance role signs in
- **THEN** the system SHALL NOT render the governance menu entries in the right-side navigation

### Requirement: Route And API Access MUST Be Denied For Unauthorized Users
The system MUST deny unauthorized direct access to governance routes and `/api/v1` governance endpoints even if the user manually enters URLs or calls APIs outside the UI.

#### Scenario: Unauthorized direct route access is blocked
- **WHEN** an unauthorized user requests a governance route URL directly
- **THEN** the system SHALL return an access-denied page and SHALL NOT expose governance page data

#### Scenario: Unauthorized API access is blocked and auditable
- **WHEN** an unauthorized user calls governance-related API endpoints
- **THEN** the backend SHALL return HTTP 403 and SHALL emit an audit record with actor, resource, action, and timestamp


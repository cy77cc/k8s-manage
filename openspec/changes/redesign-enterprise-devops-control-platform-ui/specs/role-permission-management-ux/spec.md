## MODIFIED Requirements

### Requirement: Governance Pages SHALL Provide Complete Action Entry Points
The system SHALL provide explicit and discoverable action entry points for high-frequency governance operations on Users, Roles, and Permissions pages, SHALL present those entry points using dark-mode enterprise design patterns, and SHALL ensure core operations are not hidden behind implicit gestures only.

#### Scenario: User management shows explicit edit action in unified interaction pattern
- **WHEN** an authorized user opens the Users management page
- **THEN** each user row SHALL expose an explicit `Edit` entry alongside destructive actions using consistent button/link patterns defined by the global UI foundation

#### Scenario: Role and permission pages expose complete operation set
- **WHEN** an authorized user opens Roles or Permissions page
- **THEN** the page SHALL expose required non-destructive and destructive operations through discoverable controls with consistent visual hierarchy

### Requirement: Role List SHALL Expose Explicit Permission-Editing Entry
The system SHALL provide explicit and discoverable action controls in the role list for viewing role details and editing role permissions, SHALL keep those controls accessible in responsive table layouts, and SHALL NOT require users to rely solely on implicit row-click behavior to find permission-editing functionality.

#### Scenario: Role list shows explicit actions with redesigned table patterns
- **WHEN** an authorized user opens the Roles management page
- **THEN** each role row SHALL display explicit actions for `View Details` and `Edit Permissions` with consistent alignment and affordance

#### Scenario: Implicit interaction is not the only path
- **WHEN** a user is unfamiliar with role management interactions
- **THEN** the user SHALL be able to discover permission-editing entry without trial-and-error clicking on table rows

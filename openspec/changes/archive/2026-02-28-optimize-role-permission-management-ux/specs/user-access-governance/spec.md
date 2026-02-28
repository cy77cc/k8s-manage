## MODIFIED Requirements

### Requirement: User Role Permission Pages SHALL Support Complete Governance Operations
The system SHALL provide complete core governance operations across Users, Roles, and Permissions pages for authorized users, including user editing, role assignment adjustment, and permission maintenance actions required by platform administration.

#### Scenario: User record can be edited
- **WHEN** an authorized user opens a user record from Users page
- **THEN** the system SHALL allow editing of user profile fields and role bindings through an explicit edit flow

#### Scenario: Users are not restricted to delete-only management
- **WHEN** an authorized user manages users
- **THEN** the system SHALL provide non-destructive management operations in addition to delete

#### Scenario: Role and permission operations are complete and consistent
- **WHEN** an authorized user enters Roles or Permissions page
- **THEN** the system SHALL expose the required role/permission maintenance operations with consistent confirmation and feedback behavior

### Requirement: User Role Permission Pages SHALL Follow Consistent Interaction Pattern
The system SHALL use a consistent management interaction pattern across Users, Roles, and Permissions pages, including searchable list, explicit detail/edit actions, scalable permission-editing controls, explicit save feedback, and risk-operation confirmation.

#### Scenario: Consistent list-detail interaction with explicit entry
- **WHEN** an authorized user navigates between Users, Roles, and Permissions pages
- **THEN** each page SHALL provide searchable list view and explicit detail/edit controls with consistent control placement and state feedback

#### Scenario: Role permission editing supports batch operations
- **WHEN** an authorized user edits role permissions
- **THEN** the system SHALL support batch grant/revoke actions and SHALL display change summary before commit

#### Scenario: Risk operation requires explicit confirmation
- **WHEN** an authorized user performs permission-revoking or role-unbinding operation
- **THEN** the system SHALL require explicit confirmation and SHALL display affected object count before commit

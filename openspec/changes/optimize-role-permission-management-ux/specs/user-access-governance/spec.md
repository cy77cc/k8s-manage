## MODIFIED Requirements

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

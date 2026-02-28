## ADDED Requirements

### Requirement: User Access Governance Navigation SHALL Be Independent
The system SHALL provide an independent right-side navigation group for user access governance, containing user management, role management, and permission management entries, and SHALL remove these governance entries from the system settings group.

#### Scenario: Governance entries are shown in dedicated group
- **WHEN** an authorized user opens the console navigation
- **THEN** the system SHALL display a dedicated governance group with `Users`, `Roles`, and `Permissions` entries

#### Scenario: Legacy settings placement is removed
- **WHEN** any user opens the system settings navigation group
- **THEN** the system SHALL NOT show duplicated `Users`, `Roles`, or `Permissions` entries under settings

### Requirement: User Role Permission Pages SHALL Follow Consistent Interaction Pattern
The system SHALL use a consistent management interaction pattern across Users, Roles, and Permissions pages, including searchable list, detail/edit panel, explicit save feedback, and risk-operation confirmation.

#### Scenario: Consistent list-detail interaction
- **WHEN** an authorized user navigates between Users, Roles, and Permissions pages
- **THEN** each page SHALL provide searchable list view and detail/edit panel with consistent control placement and state feedback

#### Scenario: Risk operation requires explicit confirmation
- **WHEN** an authorized user performs permission-revoking or role-unbinding operation
- **THEN** the system SHALL require explicit confirmation and SHALL display affected object count before commit

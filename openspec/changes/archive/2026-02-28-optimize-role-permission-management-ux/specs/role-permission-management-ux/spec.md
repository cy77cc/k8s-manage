## ADDED Requirements

### Requirement: Governance Pages SHALL Provide Complete Action Entry Points
The system SHALL provide explicit action entry points for high-frequency governance operations on Users, Roles, and Permissions pages, and SHALL ensure core operations are not hidden behind implicit gestures only.

#### Scenario: User management shows explicit edit action
- **WHEN** an authorized user opens the Users management page
- **THEN** each user row SHALL expose an explicit `Edit` entry in addition to destructive actions such as `Delete`

#### Scenario: Role and permission pages expose complete operation set
- **WHEN** an authorized user opens Roles or Permissions page
- **THEN** the page SHALL expose the required non-destructive and destructive operations through discoverable controls

### Requirement: Role List SHALL Expose Explicit Permission-Editing Entry
The system SHALL provide explicit and discoverable action controls in the role list for viewing role details and editing role permissions, and SHALL NOT require users to rely solely on implicit row-click behavior to find permission-editing functionality.

#### Scenario: Role list shows explicit actions
- **WHEN** an authorized user opens the Roles management page
- **THEN** each role row SHALL display explicit actions for `View Details` and `Edit Permissions`

#### Scenario: Implicit interaction is not the only path
- **WHEN** a user is unfamiliar with role management interactions
- **THEN** the user SHALL be able to discover permission-editing entry without trial-and-error clicking on table rows

### Requirement: Permission Editing SHALL Support Scalable Batch Operations
The system SHALL support batch grant and batch revoke operations for role permissions, including group-level select, select-all-in-filtered-results, clear selection, and change summary before commit.

#### Scenario: Batch grant in filtered scope
- **WHEN** an authorized user filters permissions by keyword or module and performs select-all in the filtered result
- **THEN** the system SHALL mark all matched permissions as pending grant in one action

#### Scenario: Batch revoke in grouped scope
- **WHEN** an authorized user clears a module-level permission group
- **THEN** the system SHALL mark all permissions in that group as pending revoke in one action

#### Scenario: Commit with change summary
- **WHEN** an authorized user submits permission changes
- **THEN** the system SHALL present added and removed permission counts before final confirmation and commit in a single update request

### Requirement: Permission Browser SHALL Remain Usable For Large Datasets
The system SHALL provide searchable and grouped permission browsing with responsive interaction for large permission datasets.

#### Scenario: Fast locating in large permission list
- **WHEN** a user enters search criteria in the permission selector
- **THEN** the system SHALL narrow visible permission candidates by keyword and group context

#### Scenario: Large list rendering remains responsive
- **WHEN** the permission dataset size grows significantly
- **THEN** the system SHALL keep selection and scrolling interactions responsive through incremental rendering or equivalent optimization

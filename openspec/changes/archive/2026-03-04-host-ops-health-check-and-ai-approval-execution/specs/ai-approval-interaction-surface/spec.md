## ADDED Requirements

### Requirement: Chat Surface SHALL Support Direct Approval Actions
The system SHALL allow authorized reviewers to approve or reject pending AI operation tickets directly from chat interaction surfaces.

#### Scenario: Approve from chat trace
- **WHEN** chat stream contains an `approval_required` trace with a pending ticket
- **THEN** the UI MUST expose approve/reject actions for users with approval permission
- **AND** the action MUST update ticket status through approval APIs

#### Scenario: Unauthorized reviewer in chat
- **WHEN** a user without approval-review permission views pending ticket in chat
- **THEN** the UI MUST NOT expose approval action controls

### Requirement: Notification Center SHALL Support Approval-type Actions
The system SHALL expose approval-type notifications with direct approve/reject actions and status display.

#### Scenario: Pending approval notification entry
- **WHEN** an AI operation approval ticket is created
- **THEN** the system MUST create notification entries of type `approval`
- **AND** notification center MUST display these entries in an approval-specific filter/tab

#### Scenario: Approve from notification item
- **WHEN** authorized reviewer clicks approve/reject from approval notification
- **THEN** the system MUST update approval ticket status
- **AND** the notification item MUST reflect final state (`approved` or `rejected`)

### Requirement: Approval State SHALL Remain Consistent Across Surfaces
The system SHALL synchronize approval state across chat, notification, and command/history views.

#### Scenario: Cross-surface state synchronization
- **WHEN** a ticket is approved or rejected from any surface
- **THEN** all other surfaces MUST display the latest approval status without stale pending state

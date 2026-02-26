## ADDED Requirements

### Requirement: AI assistant command context in release audit trail
The system SHALL persist AI assistant command context for release-related actions, including command identifier, resolved intent, plan hash, approval context, and execution summary.

#### Scenario: Record AI assistant command execution context
- **WHEN** a release-related action is triggered through AI assistant command bridge
- **THEN** the system MUST store command context including command identifier, resolved intent, plan hash, approval context, and execution summary in the audit payload

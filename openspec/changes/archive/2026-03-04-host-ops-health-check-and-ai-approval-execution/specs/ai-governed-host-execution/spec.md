## ADDED Requirements

### Requirement: AI Host Operations SHALL Provide Plan Preview Before Execution
The system SHALL provide a read-only execution plan preview for AI-generated host command/script operations before any execution begins.

#### Scenario: Preview AI host command plan
- **WHEN** user requests execution of an AI-generated host operation
- **THEN** the system MUST return target hosts, command/script summary, risk level, and execution parameters as preview content
- **AND** the system MUST NOT execute host-side mutations at preview stage

### Requirement: Mutating Host Execution SHALL Require Two-Step Confirmation
The system SHALL require user execution confirmation and approval-ticket authorization for mutating host operations.

#### Scenario: Missing explicit execution confirmation
- **WHEN** user submits mutating host execution without execution confirm flag
- **THEN** the system MUST reject execution request

#### Scenario: Missing approved ticket
- **WHEN** mutating host execution request does not include an approved and valid approval token
- **THEN** the system MUST block execution and return approval-required state

### Requirement: Script-based Execution SHALL Use Controlled Upload-and-Run Workflow
The system SHALL support AI-generated scripts through controlled upload-and-run workflow on target hosts.

#### Scenario: Upload script to controlled path
- **WHEN** approved script execution starts
- **THEN** the system MUST upload script content to a controlled runtime path scoped by approval or execution ID
- **AND** the system MUST execute using explicit interpreter command with timeout constraints

### Requirement: Host Execution SHALL Record Per-host Results And Replay Context
The system SHALL persist execution result details for each target host and provide replayable context.

#### Scenario: Multi-host execution result recording
- **WHEN** command or script executes across multiple hosts
- **THEN** the system MUST persist per-host stdout/stderr/exit code and execution timestamps
- **AND** the system MUST expose execution history and detail query endpoints

### Requirement: Host Execution SHALL Enforce Safety Policies
The system SHALL enforce policy controls for host operation scope and command/script safety.

#### Scenario: Policy blocks unsafe operation
- **WHEN** requested host operation violates scope, denylist, timeout, or concurrency policy
- **THEN** the system MUST reject execution before host mutation begins
- **AND** the system MUST return policy violation reason for operator review

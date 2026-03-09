## ADDED Requirements

### Requirement: AI Chat Stream Events

The system SHALL emit Server-Sent Events (SSE) with the following event types during AI chat streaming:

- `meta`: Session metadata at stream start
- `delta`: Incremental text content
- `thinking_delta`: Model reasoning/thinking content (when available)
- `tool_call`: Tool invocation with name and arguments
- `tool_result`: Tool execution completion with result
- `approval_required`: High-risk tool requiring user approval
- `heartbeat`: Periodic keep-alive event
- `done`: Stream completion
- `error`: Error notification

#### Scenario: Successful chat stream with tool execution
- **WHEN** user sends a message that triggers tool execution
- **THEN** system emits events in order: `meta`, `delta`, `tool_call`, `tool_result`, `done`

#### Scenario: High-risk tool requires approval
- **WHEN** AI attempts to execute a tool marked as high-risk
- **THEN** system emits `approval_required` event with tool name, arguments, and preview
- **AND** stream pauses with checkpoint_id for later resume

#### Scenario: Heartbeat maintains connection
- **WHEN** streaming session lasts longer than 10 seconds
- **THEN** system emits `heartbeat` events every 10 seconds with timestamp

### Requirement: Tool Result Event Content

The system SHALL include the following fields in `tool_result` events:
- `call_id`: Unique identifier matching the `tool_call` event
- `tool_name`: Name of the executed tool
- `result`: Tool execution result (may be truncated for large outputs)
- `status`: "success" or "error"

#### Scenario: Tool execution success
- **WHEN** a tool executes successfully
- **THEN** `tool_result` event contains `status: "success"` and the tool's output

#### Scenario: Tool execution failure
- **WHEN** a tool execution fails
- **THEN** `tool_result` event contains `status: "error"` and error message

### Requirement: Security Aspect Integration

The system SHALL evaluate tool calls through SecurityAspect before execution:
- Check user permissions against tool's required permission
- Interrupt high-risk tools for approval
- Create approval tickets for medium-risk tools

#### Scenario: User lacks tool permission
- **WHEN** user attempts to execute a tool they don't have permission for
- **THEN** system denies execution with permission error
- **AND** emits `error` event with permission denied message

#### Scenario: High-risk tool interruption
- **WHEN** high-risk tool is invoked
- **THEN** SecurityAspect interrupts execution
- **AND** system emits `approval_required` event with checkpoint

### Requirement: Checkpoint-based Resume

The system SHALL support resuming interrupted sessions:
- Store session state at interruption point
- Accept approval decision via resume endpoint
- Continue execution from interruption point after approval

#### Scenario: Resume after approval
- **WHEN** user submits approval for interrupted session
- **THEN** system resumes execution from checkpoint
- **AND** continues emitting SSE events from that point

#### Scenario: Resume after rejection
- **WHEN** user rejects approval for interrupted session
- **THEN** system returns rejection message
- **AND** does not execute the interrupted tool

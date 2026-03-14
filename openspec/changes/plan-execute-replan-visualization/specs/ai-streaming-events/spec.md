## NEW Requirements

### REQ-SE-010: Phase Lifecycle Events

The system SHALL emit phase lifecycle events to indicate the current execution stage.

**Required event types:**
- `phase_started`: Emitted when a new phase begins (planning/executing/replanning)
- `phase_complete`: Emitted when a phase finishes

**Event data structure:**
```json
{
  "type": "phase_started",
  "data": {
    "phase": "planning" | "executing" | "replanning",
    "title": "string",
    "status": "loading"
  }
}
```

#### Scenario: phase events indicate execution progress
- **GIVEN** user sends a chat request
- **WHEN** the runtime begins the planning phase
- **THEN** the system MUST emit `phase_started` with `phase: "planning"`
- **AND** when planning completes, emit `phase_complete` with `phase: "planning"` and `status: "success"`

### REQ-SE-011: Plan Generated Event

The system SHALL emit a structured plan event when the Planner completes.

**Event data structure:**
```json
{
  "type": "plan_generated",
  "data": {
    "plan_id": "string",
    "steps": [
      {
        "id": "string",
        "content": "string",
        "tool_hint": "string (optional)"
      }
    ],
    "total": "number"
  }
}
```

#### Scenario: plan steps are exposed to frontend
- **GIVEN** the Planner generates an execution plan
- **WHEN** planning completes successfully
- **THEN** the system MUST emit `plan_generated` event
- **AND** the event MUST include a structured steps array
- **AND** each step MUST have unique id and content

### REQ-SE-012: Step Lifecycle Events

The system SHALL emit step-level events during execution.

**Required event types:**
- `step_started`: Emitted when a step begins execution
- `step_complete`: Emitted when a step finishes

**Event data structure:**
```json
{
  "type": "step_started",
  "data": {
    "step_id": "string",
    "title": "string",
    "tool_name": "string (optional)",
    "params": "object (optional)",
    "status": "running"
  }
}
```

#### Scenario: step progress is visible to users
- **GIVEN** the Executor begins executing a step
- **WHEN** the step starts
- **THEN** the system MUST emit `step_started` event
- **AND** when the step completes, emit `step_complete` with status and summary

### REQ-SE-013: Replan Triggered Event

The system SHALL emit an event when the Replanner is triggered.

**Event data structure:**
```json
{
  "type": "replan_triggered",
  "data": {
    "reason": "string",
    "completed_steps": "number"
  }
}
```

#### Scenario: replan is communicated to users
- **GIVEN** execution results require plan adjustment
- **WHEN** the Replanner is invoked
- **THEN** the system MUST emit `replan_triggered` event
- **AND** the event MUST include the reason for replanning

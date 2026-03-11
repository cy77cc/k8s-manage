## ADDED Requirements

### Requirement: AI Module SHALL Separate Semantic Authority From Runtime Authority

The AI module architecture SHALL explicitly separate semantic authority from runtime authority.

- Rewrite, Planner, Experts, and Summarizer SHALL own semantic interpretation and language generation.
- Gateway, Orchestrator, Executor, and storage layers SHALL own transport, scheduling, persistence, safety, and observability.

#### Scenario: architecture boundary is preserved
- **WHEN** an AI request flows through the module
- **THEN** semantic interpretation MUST be produced by model-driven stages
- **AND** runtime stages MUST only enforce execution boundaries and safety constraints

#### Scenario: code cannot impersonate AI reasoning
- **WHEN** a model stage is unavailable
- **THEN** the architecture MUST expose that stage failure explicitly
- **AND** the runtime MUST NOT impersonate missing model reasoning with deterministic semantic fallback output

### Requirement: AI Module SHALL Support Layer-Specific Availability Reporting

The architecture SHALL support availability reporting for Rewrite, Planner, Expert, and Summarizer as separate operational layers.

#### Scenario: startup health reflects AI layers
- **WHEN** the server starts
- **THEN** the system MUST perform startup health checks for each configured AI layer
- **AND** logs or status output MUST identify failures by layer name rather than as a single generic model failure

#### Scenario: runtime reports precise layer failure
- **WHEN** one AI layer fails during a request
- **THEN** the system MUST report which layer failed
- **AND** the user-visible response MUST explain that the specific AI capability is temporarily unavailable

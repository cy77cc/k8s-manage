## ADDED Requirements

### Requirement: Model-First AI Runtime

The AI assistant runtime SHALL treat large language models as the primary authority for semantic understanding, planning, execution reasoning, and final summarization.

Runtime code MUST be limited to:

- protocol/schema validation
- approval and RBAC enforcement
- deterministic execution scheduling
- state persistence
- event streaming
- explicit error and unavailability reporting

Runtime code MUST NOT rewrite model semantic decisions such as:

- intent
- expert selection
- task wording
- conclusion wording
- recommendations

#### Scenario: runtime preserves valid model semantic output
- **WHEN** Rewrite, Planner, Expert, or Summarizer returns semantically valid output
- **THEN** the system MUST preserve the model’s semantic fields through the next stage
- **AND** runtime code MUST NOT replace them with code-generated equivalents

#### Scenario: runtime enforces boundary without semantic takeover
- **WHEN** a model output violates schema or execution safety constraints
- **THEN** the runtime MUST reject, block, or mark the output invalid
- **AND** the runtime MUST NOT silently rewrite the semantic meaning to continue execution

### Requirement: Layered Explicit Unavailability

The system SHALL use layered explicit unavailability instead of semantic fallback guessing.

When a stage model is unavailable, the system MUST explicitly tell the user which stage is unavailable and MUST recommend retrying later or using manual operations.

#### Scenario: rewrite unavailable
- **WHEN** the Rewrite model cannot be initialized or invoked
- **THEN** the system MUST report that the AI understanding stage is currently unavailable
- **AND** the system MUST NOT synthesize a substitute request understanding purely from code heuristics

#### Scenario: summarizer unavailable
- **WHEN** the Summarizer model cannot be initialized or invoked after execution completes
- **THEN** the system MUST report that the AI summarization stage is currently unavailable
- **AND** the system MUST still expose execution evidence so the user can inspect raw results manually

### Requirement: Renderer Is Presentation-Only

The final answer renderer SHALL act only as a presentation layer.

The renderer MAY:

- format paragraphs
- fold or expand raw evidence
- redact sensitive output
- apply markdown-safe rendering

The renderer MUST NOT:

- replace model-generated conclusions
- inject new recommendations
- re-rank findings semantically
- convert failure into success wording

#### Scenario: renderer formats without rewriting semantics
- **WHEN** the system renders a model-generated final answer
- **THEN** the renderer MUST preserve the answer’s semantic meaning
- **AND** any filtering or formatting MUST NOT introduce a different conclusion than the one produced by the model

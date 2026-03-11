## ADDED Requirements

### Requirement: Rewrite SHALL Produce RAG-Ready Semantic Output

The Rewrite stage SHALL use a large language model to normalize colloquial user input into a structured semantic representation that is suitable for both planning and retrieval augmentation.

Rewrite output MUST support:

- normalized execution intent
- normalized goal text
- target extraction
- resource hints
- ambiguity reporting
- retrieval intent
- retrieval queries
- retrieval keywords
- knowledge scope

#### Scenario: colloquial request becomes retrieval-ready
- **WHEN** a user sends a natural-language operational request with colloquial phrasing
- **THEN** Rewrite MUST produce a normalized request that can be consumed by Planner
- **AND** Rewrite MUST also produce retrieval-oriented query material suitable for RAG

#### Scenario: rewrite preserves user semantics for retrieval
- **WHEN** the user asks a request that references historical incidents, runbooks, prior cases, or troubleshooting knowledge
- **THEN** Rewrite MUST preserve those semantic cues in retrieval fields
- **AND** the system MUST NOT reduce the request to only a generic execution task string

### Requirement: Rewrite Failure MUST NOT Fall Back To Code Understanding

If Rewrite cannot return a valid model result, the system MUST fail explicitly instead of reconstructing the user intent using code-only heuristics.

#### Scenario: rewrite returns invalid JSON
- **WHEN** the Rewrite model returns invalid or unparseable output
- **THEN** the system MUST mark the Rewrite stage as unavailable or invalid
- **AND** the system MUST NOT generate a substitute semantic interpretation using hard-coded defaults beyond transport-safe error reporting

#### Scenario: rewrite runner unavailable
- **WHEN** the Rewrite runner cannot be initialized
- **THEN** the system MUST surface a user-visible unavailability message
- **AND** downstream planning and retrieval MUST NOT proceed using fabricated normalized semantics

### Requirement: RAG SHALL Consume Rewrite Semantic Fields

The RAG subsystem SHALL consume Rewrite semantic output as its primary retrieval input rather than re-parsing the raw user message independently.

#### Scenario: planner and rag share rewrite contract
- **WHEN** both Planner and RAG are enabled for a request
- **THEN** both components MUST consume the same Rewrite semantic contract
- **AND** they MUST NOT independently derive conflicting normalized interpretations from the raw message

#### Scenario: retrieval scope derives from rewrite
- **WHEN** Rewrite identifies specific resource scope, time range, or troubleshooting knowledge scope
- **THEN** the RAG retrieval query MUST use those fields to bound search and ranking
- **AND** the system MUST NOT broaden retrieval scope beyond the Rewrite semantic envelope unless explicitly requested

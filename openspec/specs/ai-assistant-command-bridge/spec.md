## ADDED Requirements

### Requirement: AI assistant gateway MUST preserve a stable command bridge

The AI assistant gateway MUST preserve a stable bridge between gateway chat requests and AI-core orchestration entrypoints. The bridge SHALL allow the rollout of new orchestration implementations without changing request or SSE transport contracts at the gateway boundary.

#### Scenario: gateway keeps the same bridge while orchestration evolves
- **WHEN** maintainers switch the AI runtime from the legacy agentic path to the multi-domain path
- **THEN** the gateway-facing command bridge MUST remain stable
- **AND** chat requests MUST still delegate into `internal/ai`
- **AND** SSE transport compatibility MUST remain unchanged at the gateway boundary

### Requirement: command bridge rollout MUST be gated by configuration

The command bridge MUST support selecting the multi-domain orchestration path through `ai.use_multi_domain_arch`. The toggle SHALL default to `false` and SHALL preserve the legacy agentic path as a fallback.

#### Scenario: multi-domain bridge path is enabled
- **WHEN** operators set `ai.use_multi_domain_arch=true`
- **THEN** agentic AI requests MUST enter the multi-domain planning path
- **AND** simple-chat requests MUST continue to use the existing simple-chat path
- **AND** disabling the toggle MUST restore the legacy agentic route

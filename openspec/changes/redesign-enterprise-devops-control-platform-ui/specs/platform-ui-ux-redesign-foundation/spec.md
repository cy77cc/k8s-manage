## MODIFIED Requirements

### Requirement: Design Tokens SHALL Define A Unified Visual System
The system SHALL define centralized design tokens for dark-first enterprise color palette, typography, spacing, 10px corner radius baseline, shadow elevation, and semantic states, and MUST consume these tokens consistently in shared layout and component layers across Ant Design and Tailwind usage.

#### Scenario: Shared components consume dark-mode tokens
- **WHEN** shared components render cards, forms, charts, and table containers
- **THEN** visual properties SHALL be derived from centralized dark-first tokens rather than page-local hardcoded values

#### Scenario: Semantic states are consistent in low-saturation palette
- **WHEN** success, warning, error, and info states are displayed across modules
- **THEN** each state SHALL keep consistent semantic color mapping and contrast in dark UI contexts

### Requirement: Global Navigation SHALL Be Task-Oriented And Role-Aware
The system SHALL organize primary navigation by operator task flow, MUST preserve role-aware visibility constraints for protected sections and actions, and SHALL remain usable in expanded and collapsed sidebar states.

#### Scenario: Authorized user sees task-grouped governance entries
- **WHEN** an authorized governance user signs in
- **THEN** the system SHALL render governance entries within task-oriented navigation groups

#### Scenario: Unauthorized user does not see protected entries
- **WHEN** a user lacking required permissions signs in
- **THEN** protected navigation entries SHALL NOT be rendered in any sidebar state

### Requirement: Core Interaction Patterns SHALL Be Explicit And Predictable
The system SHALL provide explicit entry points for core actions, clear loading/empty/error states using Skeleton and status placeholders, and confirmation flow for high-risk mutations with toast feedback.

#### Scenario: Explicit actions are discoverable
- **WHEN** a user lands on a management page
- **THEN** primary and secondary actions SHALL be visible without relying only on hidden gestures

#### Scenario: High-risk action requires confirmation and feedback
- **WHEN** a user triggers destructive or high-impact mutation
- **THEN** the system SHALL require explicit confirmation and SHALL provide post-action feedback via standardized toast messages

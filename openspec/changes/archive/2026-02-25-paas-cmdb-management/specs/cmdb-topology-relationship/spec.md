## ADDED Requirements

### Requirement: CMDB SHALL Manage Typed Relationships Between CIs
The system SHALL manage typed relationships between CIs (e.g., `belongs_to`, `depends_on`, `runs_on`, `exposes_to`) with source and target validation.

#### Scenario: Create valid relationship
- **WHEN** an operator creates a `runs_on` relationship between service CI and cluster CI
- **THEN** the system SHALL validate the source/target types and persist the relation if valid

### Requirement: Relationship Writes SHALL Prevent Invalid Graph States
The system MUST prevent duplicate edges, self-loop edges where disallowed, and type-invalid links.

#### Scenario: Reject invalid relationship
- **WHEN** a request attempts to create an invalid relationship (duplicate or disallowed type pair)
- **THEN** the system SHALL reject the request with a clear validation error

### Requirement: Topology Query SHALL Return Graph View By Scope
The system SHALL provide topology query APIs returning nodes and edges for a project or cluster scope.

#### Scenario: Query topology graph
- **WHEN** a user queries topology for a given project scope
- **THEN** the API SHALL return normalized node and edge structures required for frontend graph rendering

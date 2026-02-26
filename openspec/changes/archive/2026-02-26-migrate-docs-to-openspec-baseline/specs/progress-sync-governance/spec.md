## ADDED Requirements

### Requirement: Progress Synchronization SHALL Use Evidence Priority
Progress synchronization into OpenSpec SHALL follow evidence priority: runtime code and route/module inspection first, then migration files, then legacy progress documents.

#### Scenario: Conflict between docs and code
- **WHEN** legacy progress text conflicts with current implementation
- **THEN** OpenSpec progress status SHALL follow code evidence and include note for the mismatch

### Requirement: Synchronization SHALL Include Snapshot Date
Each synchronized baseline update SHALL include a concrete snapshot date to avoid ambiguity in status interpretation.

#### Scenario: Snapshot date recorded
- **WHEN** maintainers review a baseline sync change
- **THEN** they SHALL find an explicit date anchor for the synchronized status

### Requirement: OpenSpec Tasks SHALL Track Completed And Pending Sync Work
Progress governance SHALL represent synchronization work in checklist tasks, allowing maintainers to continue incremental updates.

#### Scenario: Incremental maintenance visibility
- **WHEN** maintainers open the change task list
- **THEN** they SHALL see completed migration tasks and pending follow-up tasks in checkbox format

### Requirement: Future Product Phases SHALL Enter Through OpenSpec Changes
New roadmap-phase requirements SHALL be introduced through new OpenSpec changes instead of directly adding undocumented behavior.

#### Scenario: New phase requirement intake
- **WHEN** a new product phase item is planned
- **THEN** maintainers SHALL create or update an OpenSpec change before implementation starts

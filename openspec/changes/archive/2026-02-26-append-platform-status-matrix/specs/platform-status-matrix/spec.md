## ADDED Requirements

### Requirement: Domain Status Matrix SHALL Use Standard Columns
The platform status matrix SHALL include domain, status, evidence, risks, and next actions columns for each major capability domain.

#### Scenario: Matrix structure verification
- **WHEN** maintainers update status matrix
- **THEN** each domain row SHALL include all required columns with non-empty values

### Requirement: Status Values SHALL Be Constrained
The status field SHALL only use `Done`, `In Progress`, or `Risk` to keep reporting consistent.

#### Scenario: Status normalization
- **WHEN** a maintainer writes domain status
- **THEN** any status outside the allowed set SHALL be rejected during review

### Requirement: Matrix Updates SHALL Be Periodic
Status matrix updates SHALL be performed at least once per sprint or major release milestone.

#### Scenario: Periodic sync
- **WHEN** sprint/release milestone closes
- **THEN** maintainers SHALL publish a refreshed status matrix update in OpenSpec tasks

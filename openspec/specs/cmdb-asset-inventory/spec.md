# cmdb-asset-inventory Specification

## Purpose
TBD - created by archiving change paas-cmdb-management. Update Purpose after archive.
## Requirements
### Requirement: CMDB SHALL Provide Unified CI Inventory
The system SHALL provide a unified configuration item (CI) inventory for managed PaaS assets, including host, cluster, service, deployment target, and application release entities.

#### Scenario: Create and query CI records
- **WHEN** an operator creates or synchronizes assets into CMDB
- **THEN** each asset SHALL be stored as a CI record with unique identity, type, lifecycle status, owner/project scope, tags, and timestamps

### Requirement: CI Query SHALL Support Multi-Dimensional Filters
The CMDB API SHALL support paginated queries filtered by CI type, status, owner/project, tags, and keyword.

#### Scenario: Filter CI list by type and project
- **WHEN** a user requests CI list with `type=service` and a specific project scope
- **THEN** the system SHALL return only matching CIs with total count and page data

### Requirement: CI Access SHALL Enforce RBAC Permissions
The system MUST enforce `cmdb:read` for listing/detail access and `cmdb:write` for create/update/delete operations.

#### Scenario: Unauthorized write attempt
- **WHEN** a user without `cmdb:write` tries to modify CI metadata
- **THEN** the system SHALL reject the request with a permission denied response


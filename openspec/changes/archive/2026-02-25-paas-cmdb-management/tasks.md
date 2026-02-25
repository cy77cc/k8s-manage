## 1. Domain Modeling And Migration

- [x] 1.1 Define CI, relation, sync job, sync record, and audit data models
- [x] 1.2 Add SQL migrations (Up/Down) for CMDB core tables and indexes
- [x] 1.3 Add API contracts under `api/cmdb/v1` for CI, relation, topology, sync, and audit endpoints

## 2. Backend Implementation

- [x] 2.1 Create `internal/service/cmdb` module with routes/handler/logic split
- [x] 2.2 Implement CI CRUD + list/detail with filter and pagination
- [x] 2.3 Implement relation CRUD + topology graph query APIs
- [x] 2.4 Implement sync job trigger/status/retry and reconciliation record logic
- [x] 2.5 Implement CMDB audit write and query APIs

## 3. Security And Integration

- [x] 3.1 Add RBAC permissions (`cmdb:read|write|sync|audit`) and enforce in handlers
- [x] 3.2 Integrate host/cluster/service/deployment data adapters for initial sync sources
- [x] 3.3 Register CMDB routes in service bootstrap under `/api/v1/cmdb`

## 4. Frontend And Validation

- [x] 4.1 Add `web/src/api/modules/cmdb.ts` and page-level integration for list/detail/topology/sync
- [x] 4.2 Add backend tests for CI/relation/sync/audit critical paths
- [x] 4.3 Run verification: `go test ./...` and `cd web && npm run build`
- [x] 4.4 Update OpenSpec task checkboxes and run `openspec validate --json`

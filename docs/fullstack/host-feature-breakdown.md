# Host Feature Breakdown (Team-Fullstack)

## Backend

- Models:
  - `internal/model/node.go` 扩展 host source/provider 字段
  - 新增 `HostCloudAccount` / `HostImportTask` / `HostVirtualizationTask`
- Logic:
  - `internal/service/host/logic/credentials.go`
  - `internal/service/host/logic/cloud.go`
  - `internal/service/host/logic/virtualization.go`
- Handlers:
  - `credentials_handler.go`
  - `cloud_handler.go`
  - `virtualization_handler.go`
- Routes:
  - `internal/service/host/routes.go` 新增 cloud/kvm/credentials 路由
- Security util:
  - `internal/utils/secret.go` (AES-GCM)

## Frontend

- API 扩展：`web/src/api/modules/hosts.ts`
- 新页面：
  - `HostKeysPage.tsx`
  - `HostCloudImportPage.tsx`
  - `HostVirtualizationPage.tsx`
- 路由接入：`web/src/App.tsx`
- Hosts 入口重构：`HostListPage.tsx` 新增主机下拉入口

## Migration

- 新增 SQL:
  - `storage/migrations/20260224_000003_host_platform_and_key_management.sql`

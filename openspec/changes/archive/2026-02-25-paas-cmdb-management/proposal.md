## Why

当前平台有主机、集群、服务等分散资源数据，但缺少统一 CMDB 视图与标准化资源模型，导致运维排障、变更影响分析和资产追踪成本高。为支撑后续自动化与拓扑能力，需要先建立 PaaS 场景下的 CMDB 基础能力。

## What Changes

- 新增 CMDB 资产台账能力：统一记录主机、集群、服务、部署目标等配置项（CI）及关键元数据。
- 新增 CMDB 拓扑关系能力：定义并维护资产间关系（归属、依赖、运行于、暴露于等）。
- 新增 CMDB 发现/同步/审计能力：支持从现有域（hosts/clusters/services/deploy）定期同步，记录变更审计与差异状态。
- 新增 CMDB API 与权限模型（读/写/同步/审计查询），并提供前端管理页面所需查询接口。

## Capabilities

### New Capabilities
- `cmdb-asset-inventory`: 定义配置项（CI）标准模型、生命周期状态、标签与查询过滤能力。
- `cmdb-topology-relationship`: 定义资产关系模型、关系校验与拓扑查询能力。
- `cmdb-discovery-sync-audit`: 定义资源发现与同步任务、差异处理策略、审计追踪能力。

### Modified Capabilities
- None.

## Impact

- Backend: `internal/service/cmdb/*`（新模块），以及与 `host/cluster/service/deployment` 的同步适配。
- API contracts: `api/cmdb/v1/*`（新增）。
- Storage: `storage/migrations/*` 新增 CMDB 相关表（ci、relation、sync_job、sync_record、audit）。
- Frontend: `web/src/api/modules/cmdb.ts` 与 CMDB 页面（新增）。
- Security: Casbin/RBAC 新增 `cmdb:read|write|sync|audit` 权限点。

## Context

`internal/ai/tools/` 目录当前状态：
- 41 个 Go 文件（含 12 个测试文件）
- 6951 行代码
- 所有工具实现平铺在同一目录

文件分布：
| 类型 | 文件数 | 示例 |
|------|--------|------|
| 核心基础设施 | 9 | tool_contracts.go, tools_common.go, wrapper.go |
| 工具注册 | 1 | tools_registry.go (1002 行) |
| 工具实现 | 14 | tools_host.go, tools_k8s.go, tools_service.go |
| 测试文件 | 12 | *_test.go |
| 小型碎片文件 | 5 | tool_call_id.go (14行), tool_name.go (13行) |

## Goals / Non-Goals

**Goals:**
- 将目录文件数从 41 减少到约 24 个
- 按领域组织工具实现代码
- 合并小型碎片文件
- 符合 `code-organization-convention` 规范

**Non-Goals:**
- 不修改任何工具名称或行为
- 不修改 `configs/scene_mappings.yaml`
- 不添加新工具或删除现有工具
- 不修复重构前已存在的 `scene_mappings.yaml` 与 `tools_registry.go` 历史不一致（`ops_aggregate_status`）

## Decisions

### Decision 1: 创建 `impl/` 子目录按领域分组

**选择：** 创建 `impl/` 子目录，每个领域一个子目录

**理由：**
- 工具实现是独立的功能单元，适合按领域隔离
- 与 `code-organization-convention` 中的 model/service 组织方式一致
- 便于未来添加新领域工具

**备选方案：**
- ❌ 保持平铺：违反规范，不可扩展
- ❌ 全部放在 tools/ 根目录下用前缀命名：无法解决文件数量问题

**领域划分：**
```
impl/
├── kubernetes/   # K8s 相关工具 (tools_k8s.go)
├── host/         # 主机 + OS 工具 (tools_host.go + tools_os.go)
├── service/      # 服务管理工具 (tools_service.go)
├── monitor/      # 监控工具 (tools_monitor.go)
├── cicd/         # CI/CD + Job 工具 (tools_cicd.go + tools_job.go)
├── deployment/   # 部署 + 配置 + 资产清单
├── governance/   # 治理 + 拓扑 (tools_governance.go + tools_topology.go)
├── infrastructure/ # 基础设施 (tools_infrastructure.go)
└── mcp/          # MCP 客户端 + 代理 (mcp_client.go + tools_mcp_proxy.go)
```

### Decision 2: 合并小型碎片文件到 `contracts.go`

**选择：** 合并 tool_call_id.go, tool_name.go, category_helpers.go, tool_contracts_ai_enhancement.go 到 tool_contracts.go

**理由：**
- 这些文件内容紧密相关（类型定义、错误类型、辅助函数）
- 减少文件数量，提高可发现性
- 单个文件约 500-600 行，仍在合理范围

**备选方案：**
- ❌ 保持独立：文件过小，增加认知负担

### Decision 3: 创建 `param/` 子目录

**选择：** 将参数处理相关代码独立到 `param/` 子目录

**理由：**
- 参数解析、验证、提示是独立关注点
- 约 400 行代码，适合独立目录
- 清晰的职责边界

### Decision 4: 重命名 `tools_common.go` 为 `runner.go`

**理由：**
- 该文件包含 `runWithPolicyAndEvent` 核心执行逻辑
- 名称更能反映其职责

## Risks / Trade-offs

### Risk 1: 导入路径变更影响外部代码

**风险：** 重构后导入路径变化，可能影响 `internal/ai/agent/` 等外部调用

**缓解：**
- 在 `tools/` 根目录保留必要的导出（如 `BuildLocalTools`, `PlatformDeps`）
- 使用 `goimports` 自动修复导入
- 重构后运行完整测试验证

### Risk 2: 包循环依赖

**风险：** 子目录可能导致包循环依赖

**缓解：**
- 核心类型（ToolMeta, ToolResult）保留在根目录 `contracts.go`
- 实现包只依赖根包的类型，不反向依赖
- 如遇循环依赖，将共享类型移至 `internal/ai/tools/types/`

### Risk 3: 历史 tool 映射不一致干扰验收

**风险：** `configs/scene_mappings.yaml` 中的 `ops_aggregate_status` 并未在 `tools_registry.go` 注册，可能导致“全量名字完全一致”类验收项失败

**缓解：**
- 将本次验收限定为“重构未引入新的 tool name mismatch”
- 记录该问题为历史遗留，不在本次目录重构范围内
- 如需修复，单独发起后续 change 处理工具实现或映射修正

## Migration Plan

**阶段 1：删除测试文件**
```bash
rm internal/ai/tools/*_test.go
```

**阶段 2：创建目录结构**
```bash
mkdir -p internal/ai/tools/{param,impl/{kubernetes,host,service,monitor,cicd,deployment,governance,infrastructure,mcp}}
```

**阶段 3：合并碎片文件**
1. 将 tool_call_id.go, tool_name.go, category_helpers.go 内容合并到 tool_contracts.go
2. 删除原碎片文件

**阶段 4：移动工具实现**
1. 按领域移动 tools_*.go 到对应 impl/ 子目录
2. 更新包声明和导入

**阶段 5：移动参数处理代码**
1. 移动 param_hints.go, tool_param_resolver.go, tool_param_validator.go 到 param/

**阶段 6：验证**
```bash
go build ./internal/ai/...
go test ./internal/ai/... -short
```

**回滚策略：** 使用 git revert，重构为纯文件移动，无逻辑变更

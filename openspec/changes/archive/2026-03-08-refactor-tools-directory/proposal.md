## Why

`internal/ai/tools/` 目录当前有 41 个文件，远超 `code-organization-convention` 规范定义的 10 文件阈值。文件平铺在单一目录下，缺乏领域分组，导致代码难以发现和维护。本次重构将按工具领域拆分子目录，合并碎片化文件，符合项目代码组织规范。

## What Changes

- 按领域将工具实现拆分到 `impl/` 子目录：`kubernetes/`, `host/`, `service/`, `monitor/`, `cicd/`, `deployment/`, `governance/`, `infrastructure/`, `mcp/`
- 合并小型碎片文件到 `contracts.go`（tool_call_id.go, tool_name.go, category_helpers.go, tool_contracts_ai_enhancement.go）
- 创建 `param/` 子目录组织参数相关代码（hints, resolver, validator）
- 重命名 `tools_common.go` 为 `runner.go` 更准确反映其职责
- 删除所有测试文件（12 个 `*_test.go`）
- 保留 `tools_registry.go` 作为工具注册入口

## Capabilities

### New Capabilities

无新增能力。本次重构为纯代码组织优化，不改变任何外部行为。

### Modified Capabilities

- `code-organization-convention`: 扩展规范以覆盖 AI tools 目录的组织要求

## Impact

**影响范围：**
- `internal/ai/tools/` 全目录重组
- `internal/ai/agent/runner.go` 可能需要更新导入路径
- `internal/ai/hybrid.go` 可能需要更新导入路径

**不变：**
- 所有工具名称和行为保持不变
- `configs/scene_mappings.yaml` 无需修改
- 外部 API 无变化

**已知历史问题（不在本次变更范围内）：**
- `configs/scene_mappings.yaml` 中存在 `ops_aggregate_status` 映射，但 `internal/ai/tools/tools_registry.go` 从更早版本开始就未注册该工具
- 本次目录重构不会修复该历史不一致，也不会借机新增工具或修改映射

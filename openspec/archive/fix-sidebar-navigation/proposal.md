# Proposal: 修复侧边栏导航菜单路径和重复问题

## 问题描述

前端侧边栏导航存在两个主要问题：

1. **路径不对** - 部分菜单项指向的路由路径与实际路由配置不匹配
2. **重复菜单** - 主机管理同时出现在顶级菜单和部署管理>基础设施子菜单中

## 影响范围

- `web/src/components/Layout/AppLayout.tsx` - 菜单配置
- `web/src/routes/AppRoutes.tsx` - 路由配置

## 解决方案

### 1. 菜单结构调整

**移除重复项：**
- 删除顶级菜单中的「主机管理」菜单项
- 保留部署管理>基础设施下的「主机管理」子菜单项

**路由路径修正：**
- CMDB 路由从 `/cmdb/assets` 改为 `/cmdb`

### 2. 补充缺失路由

为以下页面组件添加路由注册：

| 菜单项 | 路由路径 | 页面组件 |
|--------|----------|----------|
| CMDB | `/cmdb` | CMDBPage |
| 自动化 | `/automation` | AutomationPage |
| CI/CD | `/cicd` | CICDPage |
| AI命令中心 | `/ai` | AICommandCenterPage |
| 帮助中心 | `/help` | HelpCenterPage |
| 监控中心 | `/monitor` | MonitorPage |
| 工具箱 | `/tools` | ToolsPage |

## 重构后的菜单结构

```
仪表板 (/)
服务管理 (/services)
CMDB (/cmdb)
自动化 (/automation)
CI/CD (/cicd)
AI命令中心 (/ai)
帮助中心 (/help)
配置中心 (/config)
任务中心 (/tasks)
部署管理 (/deployment)
  ├─ 基础设施
  │   ├─ 集群管理 (/deployment/infrastructure/clusters)
  │   ├─ 凭证管理 (/deployment/infrastructure/credentials)
  │   └─ 主机管理 (/hosts)
  ├─ 部署目标
  │   ├─ 目标列表 (/deployment/targets)
  │   └─ 创建目标 (/deployment/targets/create)
  ├─ 发布管理
  │   ├─ 发布概览 (/deployment/overview)
  │   ├─ 创建发布 (/deployment/create)
  │   ├─ 发布历史 (/deployment)
  │   └─ 审批中心 (/deployment/approvals)
  └─ 可观测性
      ├─ 部署拓扑 (/deployment/observability/topology)
      ├─ 指标仪表板 (/deployment/observability/metrics)
      ├─ 审计日志 (/deployment/observability/audit-logs)
      ├─ 策略管理 (/deployment/observability/policies)
      └─ AIOps 洞察 (/deployment/observability/aiops)
监控中心 (/monitor)
工具箱 (/tools)
系统设置 (/settings)
```

## 风险评估

- **低风险** - 仅涉及菜单配置和路由注册，不改变业务逻辑
- **向后兼容** - 现有路由保持不变，只是添加新路由

## 验收标准

- [ ] 侧边栏菜单无重复项
- [ ] 所有菜单项点击后能正确跳转到对应页面
- [ ] CMDB 菜单跳转到 `/cmdb` 路径
- [ ] 主机管理仅在部署管理>基础设施下显示

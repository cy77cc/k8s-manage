# Tasks: 修复侧边栏导航菜单路径和重复问题

## Task List

- [x] **Task 1**: 修改 AppLayout.tsx 菜单配置 - 删除顶级主机管理菜单项，修正CMDB路径
- [x] **Task 2**: 在 AppRoutes.tsx 中添加缺失路由 - CMDB、自动化、CI/CD、AI命令中心、帮助中心、监控中心、工具箱
- [x] **Task 3**: 验证菜单激活状态映射 - 确认 activeMenuKey 正确处理新增路由

---

## Task Details

### Task 1: 修改 AppLayout.tsx 菜单配置

**文件:** `web/src/components/Layout/AppLayout.tsx`

**改动内容:**

1. 删除 `baseMenuItems` 数组中的顶级「主机管理」菜单项：
   ```typescript
   // 删除这一行
   { key: '/hosts', icon: <DesktopOutlined />, label: t('menu.hosts') },
   ```

2. 修改 CMDB 菜单项的路由路径：
   ```typescript
   // 从
   { key: '/cmdb/assets', icon: <CloudServerOutlined />, label: t('menu.cmdb') },
   // 改为
   { key: '/cmdb', icon: <CloudServerOutlined />, label: t('menu.cmdb') },
   ```

3. 确认部署管理>基础设施下的主机管理子菜单项保留不变（key 为 `/hosts`）

---

### Task 2: 在 AppRoutes.tsx 中添加缺失路由

**文件:** `web/src/routes/AppRoutes.tsx`

**改动内容:**

添加以下懒加载导入和路由配置：

```typescript
// 添加懒加载导入
const CMDBPage = lazy(() => import('../pages/CMDB/CMDBPage'));
const AutomationPage = lazy(() => import('../pages/Automation/AutomationPage'));
const CICDPage = lazy(() => import('../pages/CICD/CICDPage'));
const AICommandCenterPage = lazy(() => import('../pages/AI/AICommandCenterPage'));
const HelpCenterPage = lazy(() => import('../pages/Help/HelpCenterPage'));
const MonitorPage = lazy(() => import('../pages/Monitor/MonitorPage'));
const ToolsPage = lazy(() => import('../pages/Tools/ToolsPage'));
```

路由配置：
- `/cmdb` → CMDBPage
- `/automation` → AutomationPage
- `/cicd` → CICDPage
- `/ai` → AICommandCenterPage
- `/help` → HelpCenterPage
- `/monitor` → MonitorPage
- `/tools` → ToolsPage

---

### Task 3: 验证菜单激活状态映射

**文件:** `web/src/components/Layout/AppLayout.tsx`

**改动内容:**

确认 `activeMenuKey` 的映射逻辑能正确处理新增的路由。当前逻辑已处理 `/hosts` 路径，需要验证是否正确激活部署管理>基础设施>主机管理菜单项。

---

## 验证步骤

1. 启动开发服务器
2. 检查侧边栏菜单无重复的「主机管理」
3. 点击每个菜单项，验证跳转正确
4. 访问 `/hosts/detail/:id`，验证侧边栏正确高亮

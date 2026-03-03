# 任务清单 - 服务管理页面交互优化

## 任务概览

| ID | 任务 | 优先级 | 状态 |
|----|------|--------|------|
| 1 | 列表页视图切换功能 | P0 | ✅ 已完成 |
| 2 | 卡片整体点击交互 | P0 | ✅ 已完成 |
| 3 | 更多操作菜单精简 | P1 | ✅ 已完成 |
| 4 | 创建服务页布局优化 | P1 | ✅ 已完成 |
| 5 | 项目选择器集成 | P1 | ✅ 已完成 |
| 6 | 中英文文案统一 | P2 | ✅ 已完成 |

---

## 任务 1: 列表页视图切换功能

**文件**: `web/src/pages/Services/ServiceListPage.tsx`

**描述**:
- 新增列表/卡片视图切换功能
- 使用 Ant Design Segmented 组件
- 根据服务数量自动选择默认视图（≤8 卡片，>8 列表）
- 用户选择持久化到 localStorage

**实现要点**:
1. 添加 `viewMode` state：`'card' | 'list'`
2. 新增 Segmented 组件在页面头部
3. 实现列表视图 Table 组件
4. 添加自动判断逻辑和 localStorage 持久化

**验收标准**:
- [x] 视图切换按钮显示正常
- [x] 卡片视图功能正常
- [x] 列表视图功能正常
- [x] 默认视图根据数量自动选择
- [x] 用户选择被记住

---

## 任务 2: 卡片整体点击交互

**文件**: `web/src/pages/Services/ServiceListPage.tsx`

**描述**:
- 卡片整体可点击跳转详情页
- Checkbox 区域阻止事件冒泡
- 更多按钮区域阻止事件冒泡

**实现要点**:
1. Card 添加 `onClick` 跳转逻辑
2. Checkbox wrapper 添加 `stopPropagation`
3. Dropdown wrapper 添加 `stopPropagation`
4. 添加 `cursor-pointer` 样式

**验收标准**:
- [x] 点击卡片非交互区域跳转详情页
- [x] Checkbox 可正常勾选，不触发跳转
- [x] 更多按钮可正常展开菜单，不触发跳转

---

## 任务 3: 更多操作菜单精简

**文件**: `web/src/pages/Services/ServiceListPage.tsx`

**描述**:
- 删除"查看详情"菜单项
- 保留"编辑配置"、"启动服务"、"停止服务"、"删除服务"

**实现要点**:
1. 删除 Dropdown menu 中的 `view` 项
2. 调整 onClick 处理逻辑

**验收标准**:
- [x] 菜单项数量正确（5项 → 4项）
- [x] 剩余菜单功能正常

---

## 任务 4: 创建服务页布局优化

**文件**: `web/src/pages/Services/ServiceProvisionPage.tsx`

**描述**:
- 返回按钮移到 Card 外部（页面顶部）
- 按钮样式优化（去掉 type="primary" 大按钮样式）
- 整体布局调整

**实现要点**:
1. 在 Card 外部添加独立的顶部导航区域
2. 移动返回按钮到新区域
3. 调整创建服务按钮样式

**验收标准**:
- [x] 返回按钮在页面左上角
- [x] 按钮样式符合预期
- [x] 整体布局美观

---

## 任务 5: 项目选择器集成

**文件**: `web/src/pages/Services/ServiceProvisionPage.tsx`

**描述**:
- 项目 ID 字段改为项目选择器（显示名称）
- 根据权限判断是否可切换
- 团队 ID 字段隐藏或自动填充

**实现要点**:
1. 加载项目列表 `Api.projects.list()`
2. 检查权限 `Api.rbac.checkPermission()`
3. 有权限显示 Select，无权限显示只读 Input
4. 团队 ID 字段移除或隐藏

**验收标准**:
- [x] 项目显示名称而非 ID
- [x] 高权限用户可切换项目
- [x] 普通用户只能查看当前项目
- [x] 团队字段不再显示

---

## 任务 6: 中英文文案统一

**文件**: `web/src/pages/Services/ServiceProvisionPage.tsx`

**描述**:
- 将所有英文文案改为中文
- 保留通用技术术语（CPU、YAML、K8s、Helm）

**实现要点**:
按照文案映射表修改：
- Service Studio → 服务工作室
- Editor → 编辑器
- Preview → 预览
- Template Variables → 模板变量
- Diagnostics → 诊断信息
- required → 必填
- Service Port → 服务端口
- Container Port → 容器端口
- Memory → 内存
- K8s YAML → K8s 配置
- Compose YAML → Compose 配置
- rendering → 渲染中
- ready → 就绪
- unresolved → 未解析

**验收标准**:
- [x] 所有中文文案统一
- [x] 技术术语保留英文
- [x] 无中英文混用情况

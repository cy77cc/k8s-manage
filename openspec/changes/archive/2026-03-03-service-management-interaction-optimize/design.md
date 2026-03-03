# 设计文档 - 服务管理页面交互优化

## 1. 列表页视图切换

### 1.1 视图模式

```
┌─────────────────────────────────────────────────────────────────────┐
│  服务管理                    [卡片 | 列表]   [刷新] [创建服务]       │
│                               ↑ Segmented 组件                      │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  视图切换逻辑：                                                      │
│  - 服务数量 ≤ 8: 默认卡片视图                                       │
│  - 服务数量 > 8: 默认列表视图                                       │
│  - 用户选择存储到 localStorage，下次访问恢复                        │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 1.2 卡片视图优化

```tsx
// 当前：只有服务名链接可点击
<a onClick={() => navigate(`/services/${service.id}`)}>{service.name}</a>

// 优化后：整个卡片可点击
<Card
  hoverable
  onClick={() => navigate(`/services/${service.id}`)}
  className="cursor-pointer"
>
  <div onClick={(e) => e.stopPropagation()}>  {/* 阻止 checkbox 冒泡 */}
    <Checkbox ... />
  </div>
  <Dropdown ... />  {/* 更多按钮也阻止冒泡 */}
</Card>
```

### 1.3 列表视图设计

使用 Ant Design Table 组件：

| 列名 | 字段 | 说明 |
|------|------|------|
| 服务名 | name | 点击跳转详情 |
| 状态 | status | Tag 展示 |
| 环境 | env | Tag 展示 |
| 运行时 | runtimeType | Tag 展示 |
| 负责人 | owner | 文本 |
| 标签 | labels | 最多显示 3 个 |
| 操作 | - | 启动/停止/删除 |

## 2. 更多操作菜单精简

### 2.1 当前菜单

```
┌────────────────┐
│ 查看详情       │ ← 删除（与点击卡片/服务名重复）
│ 编辑配置       │ ← 保留
│ ─────────────  │
│ 启动服务       │
│ 停止服务       │
│ ─────────────  │
│ 删除服务       │
└────────────────┘
```

### 2.2 优化后菜单

```
┌────────────────┐
│ 编辑配置       │
│ ─────────────  │
│ 启动服务       │
│ 停止服务       │
│ ─────────────  │
│ 删除服务       │
└────────────────┘
```

## 3. 创建服务页面布局优化

### 3.1 当前布局

```
┌─────────────────────────────────────────────────────────────────────┐
│  Service Studio - VSCode Mode              [返回]  ← Card.extra    │
├─────────────────────────────────────────────────────────────────────┤
│  Editor                           │  Preview                        │
│  ─────────────────────────────    │  ───────────────────────────    │
│  ...                              │  ...                            │
│  [创建服务] [刷新预览]             │                                 │
└─────────────────────────────────────────────────────────────────────┘
```

### 3.2 优化后布局

```
┌─────────────────────────────────────────────────────────────────────┐
│  [← 返回]                    ← 独立的页面顶部导航                    │
├─────────────────────────────────────────────────────────────────────┤
│  服务工作室                        ← 标题改为中文                   │
├─────────────────────────────────────────────────────────────────────┤
│  编辑器                           │  预览                           │
│  ─────────────────────────────    │  ───────────────────────────    │
│  项目: [下拉选择项目名]            │  ...                            │
│  团队: [隐藏或自动填充]            │                                 │
│  ...                              │                                 │
│  [创建服务] [刷新预览]             │                                 │
└─────────────────────────────────────────────────────────────────────┘
```

### 3.3 项目/团队选择逻辑

```tsx
// 项目选择器
const [projects, setProjects] = useState<Project[]>([]);
const [currentProjectId, setCurrentProjectId] = useState<string>();
const [canSwitchProject, setCanSwitchProject] = useState(false);

useEffect(() => {
  // 获取项目列表
  const loadProjects = async () => {
    const res = await Api.projects.list();
    setProjects(res.data.list || []);
  };

  // 检查权限
  const checkPermission = async () => {
    const res = await Api.rbac.checkPermission('project', 'switch');
    setCanSwitchProject(res.data.hasPermission);
  };

  loadProjects();
  checkPermission();
}, []);

// 渲染
{canSwitchProject ? (
  <Select
    value={currentProjectId}
    options={projects.map(p => ({ value: p.id, label: p.name }))}
    onChange={setCurrentProjectId}
  />
) : (
  <Input value={currentProjectName} disabled />
)}
```

## 4. 中英文文案统一

### 4.1 文案映射表

| 原文 | 中文 |
|------|------|
| Service Studio - VSCode Mode | 服务工作室 |
| Editor | 编辑器 |
| Preview | 预览 |
| Template Variables | 模板变量 |
| Diagnostics | 诊断信息 |
| required | 必填 |
| Service Port | 服务端口 |
| Container Port | 容器端口 |
| Memory | 内存 |
| K8s YAML | K8s 配置 |
| Compose YAML | Compose 配置 |
| rendering | 渲染中 |
| ready | 就绪 |
| unresolved | 未解析 |

### 4.2 保留英文的技术术语

- CPU（通用技术术语）
- K8s（Kubernetes 简称）
- YAML（技术格式）
- Helm（技术名称）

## 5. 状态管理

### 5.1 视图模式存储

```tsx
const VIEW_MODE_KEY = 'service-list-view-mode';

const [viewMode, setViewMode] = useState<'card' | 'list'>(() => {
  const saved = localStorage.getItem(VIEW_MODE_KEY);
  if (saved === 'card' || saved === 'list') return saved;
  return null; // 让自动逻辑决定
});

// 根据数量自动决定
useEffect(() => {
  if (viewMode === null && list.length > 0) {
    setViewMode(list.length > 8 ? 'list' : 'card');
  }
}, [list.length, viewMode]);

// 保存用户选择
const handleViewModeChange = (mode: 'card' | 'list') => {
  setViewMode(mode);
  localStorage.setItem(VIEW_MODE_KEY, mode);
};
```

## 6. 组件结构

```
ServiceListPage.tsx
├── 页面头部
│   ├── 标题 + 描述
│   └── 操作按钮组
│       ├── Segmented (视图切换)
│       ├── 刷新按钮
│       └── 创建服务按钮
├── 统计卡片
├── 筛选区
└── 内容区
    ├── 卡片视图 (ServiceCardGrid)
    └── 列表视图 (ServiceTable)

ServiceProvisionPage.tsx
├── 页面顶部导航
│   └── 返回按钮
├── 页面标题
└── 编辑器区域
    ├── 左侧：表单编辑器
    │   ├── 项目选择器
    │   ├── 基本信息
    │   ├── 配置选项
    │   └── 操作按钮
    └── 右侧：预览区
```

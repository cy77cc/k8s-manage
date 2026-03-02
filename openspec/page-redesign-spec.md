# OpsPilot 页面重设计方案

## 1. 整体布局 (Layout)

### 1.1 布局结构

```
┌────────────────────────────────────────────────────────────┐
│  ┌──────┐  ┌──────────────────────────────────────────────┐│
│  │      │  │  Header (64px)                               ││
│  │      │  │  面包屑 + 搜索 + 快捷操作 + 用户菜单          ││
│  │  S   │  ├──────────────────────────────────────────────┤│
│  │  i   │  │                                              ││
│  │  d   │  │  Content Area                                ││
│  │  e   │  │  (padding: 32px)                             ││
│  │  b   │  │                                              ││
│  │  a   │  │  ┌────────────┐  ┌────────────┐             ││
│  │  r   │  │  │   Card     │  │   Card     │             ││
│  │      │  │  │            │  │            │             ││
│  │      │  │  └────────────┘  └────────────┘             ││
│  │      │  │                                              ││
│  └──────┘  └──────────────────────────────────────────────┘│
│   240px                  内容区                             │
└────────────────────────────────────────────────────────────┘
```

### 1.2 侧边栏 (Sidebar)

**当前问题**:
- 深色渐变背景过于厚重
- 图标和文字对比度不够
- 折叠状态不够优雅

**重设计方案**:

```css
/* 侧边栏容器 */
width: 240px;
background: #ffffff;
border-right: 1px solid #e9ecef;
box-shadow: 1px 0 3px 0 rgba(0, 0, 0, 0.05);

/* Logo 区域 */
height: 64px;
padding: 16px 20px;
border-bottom: 1px solid #e9ecef;
display: flex;
align-items: center;
gap: 12px;

/* Logo Icon */
width: 32px;
height: 32px;
background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);
border-radius: 8px;
display: flex;
align-items: center;
justify-content: center;

/* Logo Text */
font-size: 18px;
font-weight: 700;
color: #212529;
letter-spacing: -0.5px;

/* 菜单项 */
padding: 8px 12px;
margin: 4px 12px;
border-radius: 8px;
font-size: 14px;
color: #495057;
transition: all 150ms;
cursor: pointer;

/* 菜单项 Hover */
background: #f8f9fa;
color: #212529;

/* 菜单项 Active */
background: #eef2ff;
color: #4338ca;
font-weight: 500;

/* 菜单图标 */
font-size: 18px;
margin-right: 12px;
color: inherit;
```

### 1.3 顶部导航 (Header)

```css
height: 64px;
background: #ffffff;
border-bottom: 1px solid #e9ecef;
padding: 0 32px;
display: flex;
align-items: center;
justify-content: space-between;
position: sticky;
top: 0;
z-index: 1020;

/* 左侧区域 */
display: flex;
align-items: center;
gap: 16px;

/* 面包屑 */
font-size: 14px;
color: #6c757d;

/* 面包屑分隔符 */
margin: 0 8px;
color: #dee2e6;

/* 面包屑当前页 */
color: #212529;
font-weight: 500;

/* 右侧区域 */
display: flex;
align-items: center;
gap: 16px;

/* 搜索框 */
width: 280px;
background: #f8f9fa;
border: 1px solid #e9ecef;
border-radius: 8px;

/* 搜索框 Focus */
background: #ffffff;
border-color: #6366f1;
box-shadow: 0 0 0 4px rgba(99, 102, 241, 0.1);
```

## 2. Dashboard 重设计

### 2.1 页面结构

```
┌─────────────────────────────────────────────────────────────┐
│  主控台                                    [刷新] [设置]     │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │ 服务总数 │  │ 运行中   │  │ 部署次数 │  │ 告警数   │   │
│  │   24     │  │   22     │  │   156    │  │    3     │   │
│  │  +12%    │  │  91.7%   │  │  今日    │  │  活跃    │   │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  服务健康状态                              [查看全部] │   │
│  │  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │   │
│  │                                                     │   │
│  │  ✓ api-gateway        CPU: 45%   MEM: 62%  [→]    │   │
│  │  ✓ user-service       CPU: 32%   MEM: 48%  [→]    │   │
│  │  ⚠ payment-service    CPU: 78%   MEM: 85%  [→]    │   │
│  │  ✓ notification       CPU: 12%   MEM: 28%  [→]    │   │
│  │  ✓ order-service      CPU: 56%   MEM: 71%  [→]    │   │
│  │                                                     │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ┌──────────────────────┐  ┌──────────────────────────┐   │
│  │  最近部署            │  │  活跃告警                │   │
│  │  ─────────────────   │  │  ──────────────────────  │   │
│  │                      │  │                          │   │
│  │  ● v1.2.3            │  │  ⚠ High CPU usage        │   │
│  │    api-gateway       │  │    payment-service       │   │
│  │    2小时前 · 成功    │  │    5分钟前               │   │
│  │                      │  │                          │   │
│  │  ● v1.2.2            │  │  ⚠ Memory leak detected  │   │
│  │    user-service      │  │    user-service          │   │
│  │    1天前 · 成功      │  │    15分钟前              │   │
│  │                      │  │                          │   │
│  │  ✗ v1.2.1            │  │  ⚠ Slow response time    │   │
│  │    payment-service   │  │    api-gateway           │   │
│  │    2天前 · 失败      │  │    1小时前               │   │
│  │                      │  │                          │   │
│  └──────────────────────┘  └──────────────────────────┘   │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  资源使用趋势 (24小时)                              │   │
│  │  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │   │
│  │                                                     │   │
│  │  [折线图: CPU、内存、网络使用率]                    │   │
│  │                                                     │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 统计卡片设计

```css
/* 卡片容器 */
background: #ffffff;
border: 1px solid #e9ecef;
border-radius: 12px;
padding: 24px;
box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.08);
transition: all 250ms;

/* Hover 效果 */
box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
transform: translateY(-2px);

/* 图标 */
width: 48px;
height: 48px;
background: linear-gradient(135deg, #eef2ff 0%, #e0e7ff 100%);
border-radius: 12px;
display: flex;
align-items: center;
justify-content: center;
font-size: 24px;
color: #6366f1;
margin-bottom: 16px;

/* 标题 */
font-size: 13px;
color: #6c757d;
margin-bottom: 8px;
text-transform: uppercase;
letter-spacing: 0.5px;

/* 数值 */
font-size: 32px;
font-weight: 700;
color: #212529;
line-height: 1;
margin-bottom: 8px;

/* 副标题/趋势 */
font-size: 13px;
color: #10b981; /* 正向趋势 */
display: flex;
align-items: center;
gap: 4px;
```

### 2.3 服务健康列表

```css
/* 列表项 */
display: flex;
align-items: center;
justify-content: space-between;
padding: 16px;
border-radius: 8px;
transition: background 150ms;
cursor: pointer;

/* Hover */
background: #f8f9fa;

/* 左侧：状态 + 名称 */
display: flex;
align-items: center;
gap: 12px;

/* 状态图标 */
width: 8px;
height: 8px;
border-radius: 50%;
background: #10b981; /* 健康 */
background: #f59e0b; /* 警告 */
background: #ef4444; /* 错误 */

/* 服务名称 */
font-size: 14px;
font-weight: 500;
color: #212529;

/* 中间：指标 */
display: flex;
gap: 24px;

/* 指标项 */
font-size: 13px;
color: #6c757d;

/* 指标值 */
font-weight: 500;
color: #212529;

/* 右侧：操作 */
opacity: 0;
transition: opacity 150ms;

/* Hover 时显示 */
opacity: 1;
```

## 3. 服务列表页

### 3.1 页面结构

```
┌─────────────────────────────────────────────────────────────┐
│  服务管理                                                    │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  [搜索框]  [状态筛选▼]  [环境筛选▼]     [创建服务]  │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  服务列表                                           │   │
│  │  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │   │
│  │                                                     │   │
│  │  ┌─────────────────────────────────────────────┐   │   │
│  │  │ ✓ api-gateway                               │   │   │
│  │  │   生产环境 · v1.2.3 · 3个实例 · 运行中      │   │   │
│  │  │   CPU: 45% · MEM: 62% · 最后部署: 2小时前   │   │   │
│  │  │   [查看] [部署] [扩容] [日志] [...]         │   │   │
│  │  └─────────────────────────────────────────────┘   │   │
│  │                                                     │   │
│  │  ┌─────────────────────────────────────────────┐   │   │
│  │  │ ✓ user-service                              │   │   │
│  │  │   生产环境 · v1.2.2 · 5个实例 · 运行中      │   │   │
│  │  │   CPU: 32% · MEM: 48% · 最后部署: 1天前     │   │   │
│  │  │   [查看] [部署] [扩容] [日志] [...]         │   │   │
│  │  └─────────────────────────────────────────────┘   │   │
│  │                                                     │   │
│  │  ┌─────────────────────────────────────────────┐   │   │
│  │  │ ⚠ payment-service                           │   │   │
│  │  │   生产环境 · v1.2.1 · 3个实例 · 降级        │   │   │
│  │  │   CPU: 78% · MEM: 85% · 最后部署: 2天前     │   │   │
│  │  │   [查看] [部署] [扩容] [日志] [...]         │   │   │
│  │  └─────────────────────────────────────────────┘   │   │
│  │                                                     │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### 3.2 服务卡片设计

```css
/* 卡片容器 */
background: #ffffff;
border: 1px solid #e9ecef;
border-radius: 12px;
padding: 20px;
margin-bottom: 16px;
transition: all 250ms;
cursor: pointer;

/* Hover */
border-color: #c7d2fe;
box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
transform: translateY(-2px);

/* 顶部：状态 + 名称 */
display: flex;
align-items: center;
justify-content: space-between;
margin-bottom: 12px;

/* 服务名称 */
font-size: 16px;
font-weight: 600;
color: #212529;
display: flex;
align-items: center;
gap: 8px;

/* 状态徽章 */
display: inline-flex;
align-items: center;
gap: 4px;
padding: 4px 8px;
border-radius: 4px;
font-size: 12px;
font-weight: 500;

/* 中部：元数据 */
display: flex;
align-items: center;
gap: 16px;
margin-bottom: 12px;
font-size: 13px;
color: #6c757d;

/* 元数据项 */
display: flex;
align-items: center;
gap: 4px;

/* 底部：指标 */
display: flex;
align-items: center;
gap: 24px;
padding-top: 12px;
border-top: 1px solid #e9ecef;
font-size: 13px;

/* 指标项 */
display: flex;
align-items: center;
gap: 8px;

/* 指标标签 */
color: #6c757d;

/* 指标值 */
font-weight: 500;
color: #212529;

/* 操作按钮组 */
display: flex;
gap: 8px;
opacity: 0;
transition: opacity 150ms;

/* Hover 时显示 */
opacity: 1;
```

## 4. 服务详情页

### 4.1 页面结构

```
┌─────────────────────────────────────────────────────────────┐
│  ← 返回  api-gateway                                        │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  ✓ 运行中  v1.2.3  生产环境  3个实例                │   │
│  │  [部署] [扩容] [回滚] [重启] [日志] [配置] [...]    │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  实时指标 (最近 1 小时)                             │   │
│  │  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │   │
│  │                                                     │   │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐         │   │
│  │  │ CPU      │  │ 内存     │  │ 网络     │         │   │
│  │  │ 45%      │  │ 62%      │  │ 1.2MB/s  │         │   │
│  │  │ [图表]   │  │ [图表]   │  │ [图表]   │         │   │
│  │  └──────────┘  └──────────┘  └──────────┘         │   │
│  │                                                     │   │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐         │   │
│  │  │ 请求数   │  │ 错误率   │  │ 响应时间 │         │   │
│  │  │ 1.2k/s   │  │ 0.02%    │  │ 45ms     │         │   │
│  │  │ [图表]   │  │ [图表]   │  │ [图表]   │         │   │
│  │  └──────────┘  └──────────┘  └──────────┘         │   │
│  │                                                     │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ┌──────────────────────┐  ┌──────────────────────────┐   │
│  │  实例列表            │  │  部署历史                │   │
│  │  ─────────────────   │  │  ──────────────────────  │   │
│  │                      │  │                          │   │
│  │  ✓ pod-1             │  │  ● v1.2.3                │   │
│  │    CPU: 42%          │  │    2小时前 · 成功        │   │
│  │    MEM: 58%          │  │    by admin              │   │
│  │    [日志] [终端]     │  │    [查看] [回滚]         │   │
│  │                      │  │                          │   │
│  │  ✓ pod-2             │  │  ● v1.2.2                │   │
│  │    CPU: 45%          │  │    1天前 · 成功          │   │
│  │    MEM: 63%          │  │    by admin              │   │
│  │    [日志] [终端]     │  │    [查看] [回滚]         │   │
│  │                      │  │                          │   │
│  │  ✓ pod-3             │  │  ✗ v1.2.1                │   │
│  │    CPU: 48%          │  │    2天前 · 失败          │   │
│  │    MEM: 65%          │  │    by admin              │   │
│  │    [日志] [终端]     │  │    [查看详情]            │   │
│  │                      │  │                          │   │
│  └──────────────────────┘  └──────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### 4.2 Tab 导航

```css
/* Tab 容器 */
display: flex;
gap: 8px;
border-bottom: 1px solid #e9ecef;
margin-bottom: 24px;

/* Tab 项 */
padding: 12px 16px;
font-size: 14px;
font-weight: 500;
color: #6c757d;
border-bottom: 2px solid transparent;
transition: all 150ms;
cursor: pointer;

/* Tab Hover */
color: #212529;

/* Tab Active */
color: #6366f1;
border-bottom-color: #6366f1;
```

## 5. 部署流程页

### 5.1 步骤式部署

```
┌─────────────────────────────────────────────────────────────┐
│  部署服务: api-gateway                                       │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  ① 选择版本  →  ② 配置参数  →  ③ 确认部署  →  ④ 完成 │   │
│  │     ●              ○              ○              ○   │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  步骤 1: 选择版本                                   │   │
│  │  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │   │
│  │                                                     │   │
│  │  ○ v1.2.3 (最新)                                   │   │
│  │    2024-03-02 · by admin · feat: 新增支付功能      │   │
│  │                                                     │   │
│  │  ○ v1.2.2                                          │   │
│  │    2024-03-01 · by admin · fix: 修复登录bug        │   │
│  │                                                     │   │
│  │  ○ v1.2.1                                          │   │
│  │    2024-02-28 · by admin · refactor: 重构代码      │   │
│  │                                                     │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  [取消]                                          [下一步 →] │
└─────────────────────────────────────────────────────────────┘
```

### 5.2 部署进度

```
┌─────────────────────────────────────────────────────────────┐
│  正在部署: api-gateway v1.2.3                                │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  ✓ 拉取镜像                                         │   │
│  │  ✓ 创建容器                                         │   │
│  │  ⟳ 启动服务 (2/3)                                  │   │
│  │  ○ 健康检查                                         │   │
│  │  ○ 切换流量                                         │   │
│  │                                                     │   │
│  │  [进度条: 60%]                                      │   │
│  │                                                     │   │
│  │  [实时日志]                                         │   │
│  │  2024-03-02 10:30:15 INFO  Starting container...   │   │
│  │  2024-03-02 10:30:16 INFO  Container started       │   │
│  │  2024-03-02 10:30:17 INFO  Waiting for health...   │   │
│  │                                                     │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  [取消部署]                                                  │
└─────────────────────────────────────────────────────────────┘
```

## 6. 响应式设计

### 6.1 断点策略

```
Desktop (>= 1024px)
- 显示完整侧边栏
- 多列布局
- 完整的表格

Tablet (768px - 1023px)
- 可折叠侧边栏
- 两列布局
- 简化的表格

Mobile (< 768px)
- 隐藏侧边栏，使用底部导航
- 单列布局
- 卡片式列表
```

### 6.2 移动端优化

```css
/* 底部导航 (Mobile) */
position: fixed;
bottom: 0;
left: 0;
right: 0;
height: 64px;
background: #ffffff;
border-top: 1px solid #e9ecef;
display: flex;
justify-content: space-around;
align-items: center;
z-index: 1030;

/* 导航项 */
display: flex;
flex-direction: column;
align-items: center;
gap: 4px;
padding: 8px 16px;
font-size: 12px;
color: #6c757d;

/* 导航项 Active */
color: #6366f1;

/* 图标 */
font-size: 20px;
```


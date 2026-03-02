# Spec: ui-layout

**Status**: Draft
**Created**: 2026-03-02

---

## Overview

本规范定义了 OpsPilot PaaS 平台的布局系统，包括侧边栏、顶部导航、内容区和响应式适配。布局采用优化的左侧边栏设计，提供清晰的导航结构和良好的空间利用。

---

## Requirements

### REQ-1: 整体布局结构

系统 SHALL 采用左侧边栏 + 顶部导航 + 内容区的布局结构。

**布局规范**:
- 侧边栏宽度 SHALL 为 240px
- 侧边栏 SHALL 固定在左侧
- 顶部导航高度 SHALL 为 64px
- 顶部导航 SHALL 固定在顶部（sticky）
- 内容区 SHALL 占据剩余空间
- 内容区内边距 SHALL 为 32px

#### Scenario: 桌面布局

**WHEN** 用户在桌面设备（>= 1024px）上访问系统
**THEN** SHALL 显示完整的侧边栏（240px 宽）
**AND** SHALL 显示固定的顶部导航（64px 高）
**AND** 内容区 SHALL 占据剩余空间
**AND** 内容区左边距 SHALL 为 240px

---

### REQ-2: 侧边栏设计

系统 SHALL 提供浅色主题的侧边栏，包含 Logo、导航菜单和用户信息。

**样式规范**:
- 背景色 SHALL 为白色 (#ffffff)
- 右边框 SHALL 为 1px solid #e9ecef
- 阴影 SHALL 为 1px 0 3px 0 rgba(0, 0, 0, 0.05)
- Z-Index SHALL 为 1030

**Logo 区域**:
- 高度 SHALL 为 64px
- 内边距 SHALL 为 16px 20px
- 底部边框 SHALL 为 1px solid #e9ecef
- Logo 图标尺寸 SHALL 为 32px
- Logo 图标圆角 SHALL 为 8px
- Logo 图标背景 SHALL 为渐变色（Indigo 500 到 Purple 500）
- Logo 文字字体 SHALL 为 18px, Bold (700)
- Logo 文字颜色 SHALL 为 Gray 900

**菜单项**:
- 内边距 SHALL 为 8px 12px
- 外边距 SHALL 为 4px 12px
- 圆角 SHALL 为 8px
- 字体 SHALL 为 14px
- 默认颜色 SHALL 为 Gray 700
- 图标尺寸 SHALL 为 18px
- 图标与文字间距 SHALL 为 12px
- 悬停背景色 SHALL 为 Gray 100
- 悬停文字色 SHALL 为 Gray 900
- 激活背景色 SHALL 为 Indigo 50
- 激活文字色 SHALL 为 Indigo 700
- 激活字重 SHALL 为 Medium (500)
- 过渡动画 SHALL 为 150ms

#### Scenario: 菜单项交互

**WHEN** 用户悬停在菜单项上
**THEN** 背景色 SHALL 变为 Gray 100
**AND** 文字色 SHALL 变为 Gray 900
**AND** 过渡 SHALL 平滑进行

**WHEN** 用户点击菜单项
**THEN** 该菜单项 SHALL 变为激活状态
**AND** 背景色 SHALL 为 Indigo 50
**AND** 文字色 SHALL 为 Indigo 700
**AND** 字重 SHALL 为 Medium

---

### REQ-3: 顶部导航设计

系统 SHALL 提供固定的顶部导航，包含面包屑、搜索框和快捷操作。

**样式规范**:
- 高度 SHALL 为 64px
- 背景色 SHALL 为白色 (#ffffff)
- 底部边框 SHALL 为 1px solid #e9ecef
- 内边距 SHALL 为 0 32px
- Position SHALL 为 sticky
- Top SHALL 为 0
- Z-Index SHALL 为 1020

**面包屑**:
- 字体 SHALL 为 14px
- 默认颜色 SHALL 为 Gray 500
- 当前页颜色 SHALL 为 Gray 900
- 当前页字重 SHALL 为 Medium (500)
- 分隔符 SHALL 为 "/"
- 分隔符颜色 SHALL 为 Gray 300
- 分隔符间距 SHALL 为 8px

**搜索框**:
- 宽度 SHALL 为 280px
- 背景色 SHALL 为 Gray 100
- 边框 SHALL 为 1px solid #e9ecef
- 圆角 SHALL 为 8px
- Focus 时背景色 SHALL 为白色
- Focus 时边框色 SHALL 为 Indigo 500
- Focus 时 SHALL 显示外发光

**快捷操作**:
- 图标按钮尺寸 SHALL 为 40px
- 图标尺寸 SHALL 为 18px
- 按钮间距 SHALL 为 8px
- 悬停背景色 SHALL 为 Gray 100
- 圆角 SHALL 为 8px

#### Scenario: 面包屑导航

**WHEN** 用户查看服务详情页
**THEN** 面包屑 SHALL 显示 "首页 / 服务管理 / api-gateway"
**AND** "首页"和"服务管理" SHALL 为可点击链接
**AND** "api-gateway" SHALL 为当前页，不可点击
**AND** 当前页文字 SHALL 为 Gray 900 色

---

### REQ-4: 内容区设计

系统 SHALL 提供清晰的内容区，包含页面标题和内容。

**样式规范**:
- 背景色 SHALL 为 Gray 50 (#fafbfc)
- 内边距 SHALL 为 32px
- 最小高度 SHALL 为 calc(100vh - 64px)

**页面标题区**:
- 标题字体 SHALL 为 24px, Bold (700)
- 标题颜色 SHALL 为 Gray 900
- 标题与内容间距 SHALL 为 24px
- 标题区 SHALL 支持右侧操作按钮

**内容区**:
- 卡片间距 SHALL 为 24px
- 区块间距 SHALL 为 32px

#### Scenario: 页面内容区

**WHEN** 用户查看任意页面
**THEN** 内容区背景色 SHALL 为 Gray 50
**AND** 内容区内边距 SHALL 为 32px
**AND** 页面标题 SHALL 在顶部显示
**AND** 页面内容 SHALL 在标题下方 24px 处开始

---

### REQ-5: 响应式布局

系统 SHALL 支持响应式布局，适配桌面、平板和移动设备。

**桌面 (>= 1024px)**:
- SHALL 显示完整侧边栏（240px）
- SHALL 显示完整顶部导航
- SHALL 使用多列布局

**平板 (768px - 1023px)**:
- SHALL 显示可折叠侧边栏
- 侧边栏默认 SHALL 折叠，仅显示图标
- 点击图标 SHALL 展开侧边栏
- SHALL 使用两列布局

**移动 (< 768px)**:
- SHALL 隐藏侧边栏
- SHALL 显示底部导航栏
- SHALL 使用单列布局
- 顶部导航 SHALL 简化，仅显示标题和菜单按钮

#### Scenario: 平板设备布局

**WHEN** 用户在平板设备（768px - 1023px）上访问系统
**THEN** 侧边栏 SHALL 默认折叠，仅显示图标
**AND** 侧边栏宽度 SHALL 为 64px
**AND** 点击菜单图标时，侧边栏 SHALL 展开到 240px
**AND** 点击遮罩层时，侧边栏 SHALL 折叠

#### Scenario: 移动设备布局

**WHEN** 用户在移动设备（< 768px）上访问系统
**THEN** 侧边栏 SHALL 完全隐藏
**AND** SHALL 显示底部导航栏（64px 高）
**AND** 底部导航 SHALL 包含 4-5 个主要导航项
**AND** 顶部导航 SHALL 仅显示标题和菜单按钮
**AND** 点击菜单按钮 SHALL 打开抽屉式侧边栏

---

### REQ-6: 底部导航（移动端）

系统 SHALL 在移动端提供底部导航栏。

**样式规范**:
- Position SHALL 为 fixed
- Bottom SHALL 为 0
- 高度 SHALL 为 64px
- 背景色 SHALL 为白色
- 顶部边框 SHALL 为 1px solid #e9ecef
- Z-Index SHALL 为 1030

**导航项**:
- SHALL 包含 4-5 个主要导航项
- 导航项 SHALL 均匀分布
- 图标尺寸 SHALL 为 20px
- 文字字体 SHALL 为 12px
- 默认颜色 SHALL 为 Gray 500
- 激活颜色 SHALL 为 Indigo 500
- 导航项 SHALL 垂直排列（图标在上，文字在下）
- 图标与文字间距 SHALL 为 4px

#### Scenario: 移动端底部导航

**WHEN** 用户在移动设备上使用系统
**THEN** 底部 SHALL 显示固定的导航栏
**AND** 导航栏 SHALL 包含"首页"、"服务"、"部署"、"监控"、"更多"
**AND** 当前页导航项 SHALL 高亮显示（Indigo 500）
**AND** 点击导航项 SHALL 切换到对应页面

---

### REQ-7: 抽屉式侧边栏（移动端）

系统 SHALL 在移动端提供抽屉式侧边栏。

**样式规范**:
- 宽度 SHALL 为 280px
- 背景色 SHALL 为白色
- 阴影 SHALL 为 2xl
- 从左侧滑入
- 动画持续时间 SHALL 为 250ms

**遮罩层**:
- 背景色 SHALL 为 rgba(0, 0, 0, 0.45)
- SHALL 有 4px 的模糊效果
- 点击遮罩层 SHALL 关闭抽屉

#### Scenario: 打开移动端菜单

**WHEN** 用户在移动设备上点击菜单按钮
**THEN** 抽屉式侧边栏 SHALL 从左侧滑入
**AND** 遮罩层 SHALL 淡入显示
**AND** 侧边栏 SHALL 显示完整的导航菜单
**AND** 点击遮罩层 SHALL 关闭侧边栏

---

### REQ-8: 滚动行为

系统 SHALL 定义统一的滚动行为。

**顶部导航**:
- SHALL 固定在顶部（sticky）
- 滚动时 SHALL 保持可见
- 滚动时 SHALL 添加阴影效果

**侧边栏**:
- SHALL 固定在左侧
- 内容过长时 SHALL 支持垂直滚动
- 滚动条 SHALL 使用自定义样式（细滚动条）

**内容区**:
- SHALL 支持垂直滚动
- 滚动条 SHALL 使用自定义样式

#### Scenario: 页面滚动

**WHEN** 用户向下滚动页面
**THEN** 顶部导航 SHALL 保持固定在顶部
**AND** 顶部导航 SHALL 显示阴影效果
**AND** 侧边栏 SHALL 保持固定在左侧
**AND** 内容区 SHALL 正常滚动

---

## Implementation Notes

### 布局组件结构

```typescript
// src/components/Layout/AppLayout.tsx
<Layout>
  <Sider width={240} className="app-sider">
    <Logo />
    <Menu />
    <UserInfo />
  </Sider>
  <Layout>
    <Header className="app-header">
      <Breadcrumb />
      <Search />
      <Actions />
    </Header>
    <Content className="app-content">
      {children}
    </Content>
  </Layout>
</Layout>
```

### 响应式实现

```typescript
// 使用 Ant Design 的 Grid 断点
import { Grid } from 'antd';

const { useBreakpoint } = Grid;

const AppLayout = () => {
  const screens = useBreakpoint();

  const isMobile = !screens.md; // < 768px
  const isTablet = screens.md && !screens.lg; // 768px - 1023px
  const isDesktop = screens.lg; // >= 1024px

  // 根据断点渲染不同布局
};
```

---

## References

- [Ant Design Layout](https://ant.design/components/layout)
- [Ant Design Grid](https://ant.design/components/grid)


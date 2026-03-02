# Spec: ui-component-library

**Status**: Draft
**Created**: 2026-03-02

---

## Overview

本规范定义了 OpsPilot PaaS 平台的组件库，基于 Ant Design 6.3.0 进行深度定制。组件库采用极简主义设计风格，提供一致、现代、易用的 UI 组件。

---

## Requirements

### REQ-1: Button 组件

系统 SHALL 提供 Button 组件，支持多种变体、尺寸和状态。

**变体**:
- Primary: 主要操作按钮
- Default: 次要操作按钮
- Ghost: 幽灵按钮
- Link: 链接按钮
- Danger: 危险操作按钮

**尺寸**:
- Small: 32px 高度
- Middle: 40px 高度（默认）
- Large: 48px 高度

**样式规范**:
- 圆角 SHALL 为 8px
- 字重 SHALL 为 Medium (500)
- 内边距 SHALL 为 0 16px
- Primary 按钮背景色 SHALL 为 Indigo 500 (#6366f1)
- 悬停时背景色 SHALL 变为 Indigo 600 (#4f46e5)
- 激活时背景色 SHALL 变为 Indigo 700 (#4338ca)
- 过渡动画 SHALL 为 150ms

#### Scenario: 主要操作按钮

**WHEN** 用户看到主要操作按钮（如"创建服务"）
**THEN** 按钮 SHALL 使用 Primary 变体
**AND** 高度 SHALL 为 40px
**AND** 背景色 SHALL 为 Indigo 500
**AND** 悬停时背景色 SHALL 平滑过渡到 Indigo 600

---

### REQ-2: Input 组件

系统 SHALL 提供 Input 组件，支持多种类型和状态。

**类型**:
- Text: 文本输入
- Password: 密码输入
- TextArea: 多行文本输入
- Search: 搜索输入

**尺寸**:
- Small: 32px 高度
- Middle: 40px 高度（默认）
- Large: 48px 高度

**样式规范**:
- 圆角 SHALL 为 8px
- 边框 SHALL 为 1px solid #dee2e6
- 内边距 SHALL 为 0 12px
- Focus 时边框色 SHALL 为 Indigo 500
- Focus 时 SHALL 显示 4px 的外发光 (rgba(99, 102, 241, 0.1))
- Error 时边框色 SHALL 为 Error 色 (#ef4444)
- 过渡动画 SHALL 为 150ms

#### Scenario: 输入框获得焦点

**WHEN** 用户点击输入框
**THEN** 边框色 SHALL 变为 Indigo 500
**AND** SHALL 显示 4px 的蓝色外发光
**AND** 过渡 SHALL 平滑进行

---

### REQ-3: Card 组件

系统 SHALL 提供 Card 组件，用于内容分组。

**变体**:
- Bordered: 带边框（默认）
- Borderless: 无边框
- Hoverable: 可悬停

**样式规范**:
- 圆角 SHALL 为 12px
- 边框 SHALL 为 1px solid #e9ecef
- 内边距 SHALL 为 24px
- 阴影 SHALL 为 md (0 1px 3px 0 rgba(0, 0, 0, 0.08))
- Hoverable 卡片悬停时阴影 SHALL 变为 lg
- Hoverable 卡片悬停时 SHALL 向上移动 2px
- 过渡动画 SHALL 为 250ms

#### Scenario: 可悬停卡片

**WHEN** 用户悬停在可悬停卡片上
**THEN** 卡片 SHALL 向上移动 2px
**AND** 阴影 SHALL 从 md 变为 lg
**AND** 过渡 SHALL 平滑进行

---

### REQ-4: Table 组件

系统 SHALL 提供 Table 组件，用于数据展示。

**样式规范**:
- 表头背景色 SHALL 为 Gray 100 (#f8f9fa)
- 表头字体 SHALL 为 13px, Semibold (600)
- 表头文字 SHALL 大写，字间距 0.5px
- 行高 SHALL 为 56px（cellPaddingBlock: 16px）
- 边框色 SHALL 为 Gray 200 (#e9ecef)
- 悬停行背景色 SHALL 为 Gray 100 (#f8f9fa)
- 选中行背景色 SHALL 为 Indigo 50 (#eef2ff)

**功能**:
- SHALL 支持排序
- SHALL 支持筛选
- SHALL 支持行选择
- SHALL 支持固定列
- SHALL 支持虚拟滚动（大数据量）

#### Scenario: 表格行悬停

**WHEN** 用户悬停在表格行上
**THEN** 行背景色 SHALL 变为 Gray 100
**AND** 内联操作按钮 SHALL 显示
**AND** 过渡 SHALL 为 150ms

---

### REQ-5: Modal 组件

系统 SHALL 提供 Modal 组件，用于模态对话框。

**样式规范**:
- 圆角 SHALL 为 12px
- 阴影 SHALL 为 xl
- 背景遮罩 SHALL 为 rgba(0, 0, 0, 0.45)
- 背景遮罩 SHALL 有 4px 的模糊效果
- 头部内边距 SHALL 为 24px 24px 16px
- 内容内边距 SHALL 为 24px
- 底部内边距 SHALL 为 16px 24px
- 头部和内容之间 SHALL 有 1px 的分隔线
- 内容和底部之间 SHALL 有 1px 的分隔线

**动画**:
- 进入动画 SHALL 为 250ms
- 进入时 SHALL 从 95% 缩放到 100%
- 进入时 SHALL 从 -20px 向上移动到 0
- 进入时 SHALL 从透明度 0 到 1

#### Scenario: 打开模态框

**WHEN** 用户打开模态框
**THEN** 背景遮罩 SHALL 淡入显示
**AND** 模态框 SHALL 从 95% 缩放到 100%
**AND** 模态框 SHALL 向上滑入
**AND** 动画 SHALL 为 250ms

---

### REQ-6: Form 组件

系统 SHALL 提供 Form 组件，用于表单输入。

**布局**:
- Vertical: 垂直布局（默认）
- Horizontal: 水平布局
- Inline: 内联布局

**样式规范**:
- Form Item 间距 SHALL 为 24px
- Label 字体 SHALL 为 14px, Medium (500)
- Label 颜色 SHALL 为 Gray 700 (#495057)
- Label 与输入框间距 SHALL 为 8px
- Required 标记颜色 SHALL 为 Error 色
- Help 文字字体 SHALL 为 13px
- Help 文字颜色 SHALL 为 Gray 500
- Error 文字字体 SHALL 为 13px
- Error 文字颜色 SHALL 为 Error 色

**验证**:
- SHALL 支持实时验证（失去焦点时）
- SHALL 支持提交验证
- SHALL 在字段下方显示错误信息
- SHALL 显示错误图标

#### Scenario: 表单验证错误

**WHEN** 用户提交表单但验证失败
**THEN** 错误字段 SHALL 显示红色边框
**AND** 错误字段下方 SHALL 显示错误信息
**AND** 错误信息 SHALL 包含错误图标
**AND** 页面 SHALL 滚动到第一个错误字段

---

### REQ-7: Tag 组件

系统 SHALL 提供 Tag 组件，用于标签和状态指示。

**颜色**:
- Default: Gray 色系
- Blue: Indigo 色系
- Green: Success 色系
- Orange: Warning 色系
- Red: Error 色系

**样式规范**:
- 圆角 SHALL 为 4px
- 内边距 SHALL 为 4px 8px
- 字体 SHALL 为 12px, Medium (500)
- 边框 SHALL 为 1px solid
- Default 背景色 SHALL 为 Gray 100
- Default 边框色 SHALL 为 Gray 300
- Default 文字色 SHALL 为 Gray 700

#### Scenario: 服务状态标签

**WHEN** 用户看到服务状态标签
**THEN** "运行中"状态 SHALL 使用 Green 色系
**AND** 背景色 SHALL 为 #d1fae5
**AND** 边框色 SHALL 为 #6ee7b7
**AND** 文字色 SHALL 为 #047857

---

### REQ-8: Notification 组件

系统 SHALL 提供 Notification 组件，用于通知消息。

**类型**:
- Success: 成功通知
- Error: 错误通知
- Warning: 警告通知
- Info: 信息通知

**样式规范**:
- 圆角 SHALL 为 8px
- 边框 SHALL 为 1px solid #e9ecef
- 左侧色条 SHALL 为 4px 宽
- 内边距 SHALL 为 16px
- 阴影 SHALL 为 lg
- 最小宽度 SHALL 为 384px
- 最大宽度 SHALL 为 480px
- 标题字体 SHALL 为 14px, Semibold (600)
- 描述字体 SHALL 为 13px
- 描述颜色 SHALL 为 Gray 500

#### Scenario: 成功通知

**WHEN** 系统显示成功通知
**THEN** 左侧色条 SHALL 为 Success 色 (#10b981)
**AND** 通知 SHALL 从右侧滑入
**AND** 通知 SHALL 在 4.5 秒后自动关闭

---

### REQ-9: Skeleton 组件

系统 SHALL 提供 Skeleton 组件，用于加载占位。

**样式规范**:
- 背景色 SHALL 为 Gray 100 (#f8f9fa)
- 动画 SHALL 为渐变扫过效果
- 动画持续时间 SHALL 为 1.5s
- 圆角 SHALL 与实际内容一致

#### Scenario: 页面加载

**WHEN** 页面正在加载数据
**THEN** SHALL 显示骨架屏
**AND** 骨架屏 SHALL 模拟实际内容的布局
**AND** SHALL 显示渐变扫过动画

---

### REQ-10: Empty 组件

系统 SHALL 提供 Empty 组件，用于空状态展示。

**样式规范**:
- 容器内边距 SHALL 为 48px 24px
- 图标尺寸 SHALL 为 64px
- 图标颜色 SHALL 为 Gray 300
- 图标与标题间距 SHALL 为 16px
- 标题字体 SHALL 为 16px, Semibold (600)
- 标题颜色 SHALL 为 Gray 700
- 描述字体 SHALL 为 14px
- 描述颜色 SHALL 为 Gray 500
- 描述最大宽度 SHALL 为 400px
- 描述与操作按钮间距 SHALL 为 24px

#### Scenario: 空列表

**WHEN** 用户查看空的服务列表
**THEN** SHALL 显示空状态组件
**AND** SHALL 显示相关图标
**AND** SHALL 显示"暂无服务"标题
**AND** SHALL 显示"创建第一个服务"操作按钮

---

## Implementation Notes

### Ant Design 主题配置

组件库 SHALL 通过 Ant Design 主题配置实现：

```typescript
// src/theme/antd-theme.ts
import { ThemeConfig } from 'antd';

export const antdTheme: ThemeConfig = {
  token: {
    colorPrimary: '#6366f1',
    borderRadius: 8,
    fontSize: 14,
    lineHeight: 1.5,
  },
  components: {
    Button: {
      controlHeight: 40,
      borderRadius: 8,
      fontWeight: 500,
    },
    Input: {
      controlHeight: 40,
      borderRadius: 8,
    },
    Card: {
      borderRadius: 12,
      boxShadow: '0 1px 3px 0 rgba(0, 0, 0, 0.08)',
    },
    Table: {
      rowHoverBg: '#f8f9fa',
      headerBg: '#f8f9fa',
      cellPaddingBlock: 16,
    },
    Modal: {
      borderRadius: 12,
    },
  },
};
```

---

## References

- [Ant Design Components](https://ant.design/components/overview)
- [Ant Design Theming](https://ant.design/docs/react/customize-theme)


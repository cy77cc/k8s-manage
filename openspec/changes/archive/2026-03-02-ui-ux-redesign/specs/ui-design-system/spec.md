# Spec: ui-design-system

**Status**: Draft
**Created**: 2026-03-02

---

## Overview

本规范定义了 OpsPilot PaaS 平台的设计系统，包括色彩、排版、间距、圆角、阴影、动画等基础设计元素。设计系统采用极简主义风格，灵感来源于 Vercel、Linear 和 Stripe，旨在提供清晰、现代、一致的用户界面。

---

## Requirements

### REQ-1: 色彩系统

系统 SHALL 定义完整的色彩系统，包括主色、中性色和语义色。

**主色调 (Primary Colors)**:
- 系统 SHALL 使用 Indigo 色系作为主色调
- 主色 SHALL 为 Indigo 500 (#6366f1)
- 悬停状态 SHALL 使用 Indigo 600 (#4f46e5)
- 激活状态 SHALL 使用 Indigo 700 (#4338ca)
- 系统 SHALL 提供 Indigo 50, 100, 200, 500, 600, 700, 900 共 7 个色阶

**中性色 (Neutral Colors)**:
- 系统 SHALL 使用 Gray 色系作为中性色
- 系统 SHALL 提供 Gray 50 (#fafbfc) 到 Gray 900 (#212529) 共 8 个色阶
- Gray 50-100 SHALL 用于背景色
- Gray 200-300 SHALL 用于边框和分隔线
- Gray 500-900 SHALL 用于文本

**语义色 (Semantic Colors)**:
- Success SHALL 使用 #10b981 (绿色)
- Warning SHALL 使用 #f59e0b (橙色)
- Error SHALL 使用 #ef4444 (红色)
- Info SHALL 使用 #3b82f6 (蓝色)

**使用规则**:
- 主色 SHALL 仅用于主要操作按钮、链接、选中状态
- 中性色 SHALL 用于文本、边框、背景、图标
- 语义色 SHALL 仅用于状态指示，不得用于装饰

#### Scenario: 按钮使用主色

**WHEN** 用户看到主要操作按钮（如"创建服务"、"部署"）
**THEN** 按钮背景色 SHALL 为 Indigo 500 (#6366f1)
**AND** 悬停时背景色 SHALL 变为 Indigo 600 (#4f46e5)
**AND** 激活时背景色 SHALL 变为 Indigo 700 (#4338ca)

#### Scenario: 状态标签使用语义色

**WHEN** 用户看到服务状态标签
**THEN** "运行中"状态 SHALL 使用 Success 色 (#10b981)
**AND** "警告"状态 SHALL 使用 Warning 色 (#f59e0b)
**AND** "失败"状态 SHALL 使用 Error 色 (#ef4444)

---

### REQ-2: 排版系统

系统 SHALL 定义完整的排版系统，包括字体家族、字号、行高和字重。

**字体家族**:
- 主字体 SHALL 使用系统字体栈: `-apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif`
- 等宽字体 SHALL 使用: `'SF Mono', 'Monaco', 'Cascadia Code', 'Consolas', monospace`
- 等宽字体 SHALL 仅用于代码、终端、日志等场景

**字号层级**:
- Display: 32px / 2rem - 用于页面标题
- H1: 24px / 1.5rem - 用于一级标题
- H2: 20px / 1.25rem - 用于二级标题
- H3: 18px / 1.125rem - 用于三级标题
- Body Large: 16px / 1rem - 用于大号正文
- Body: 14px / 0.875rem - 用于正文（默认）
- Body Small: 13px / 0.8125rem - 用于小号正文
- Caption: 12px / 0.75rem - 用于辅助文字

**行高**:
- Tight (1.25) SHALL 用于标题
- Normal (1.5) SHALL 用于正文（默认）
- Relaxed (1.75) SHALL 用于长文本

**字重**:
- Regular (400) SHALL 用于正文
- Medium (500) SHALL 用于强调
- Semibold (600) SHALL 用于小标题
- Bold (700) SHALL 用于标题

#### Scenario: 页面标题使用 Display 字号

**WHEN** 用户进入任意页面
**THEN** 页面主标题 SHALL 使用 Display 字号 (32px)
**AND** 字重 SHALL 为 Bold (700)
**AND** 行高 SHALL 为 Tight (1.25)
**AND** 颜色 SHALL 为 Gray 900 (#212529)

#### Scenario: 正文使用 Body 字号

**WHEN** 用户阅读页面内容
**THEN** 正文 SHALL 使用 Body 字号 (14px)
**AND** 字重 SHALL 为 Regular (400)
**AND** 行高 SHALL 为 Normal (1.5)
**AND** 颜色 SHALL 为 Gray 700 (#495057)

---

### REQ-3: 间距系统

系统 SHALL 定义基于 8px 基准的间距系统。

**间距值**:
- xs: 4px / 0.25rem
- sm: 8px / 0.5rem
- md: 16px / 1rem
- lg: 24px / 1.5rem
- xl: 32px / 2rem
- 2xl: 48px / 3rem
- 3xl: 64px / 4rem

**使用规则**:
- 组件内部 SHALL 使用 sm (8px)
- 组件之间 SHALL 使用 md (16px) 或 lg (24px)
- 页面边距 SHALL 使用 xl (32px)
- 区块间距 SHALL 使用 2xl (48px)

#### Scenario: 卡片内边距

**WHEN** 用户看到卡片组件
**THEN** 卡片内边距 SHALL 为 lg (24px)
**AND** 卡片标题与内容间距 SHALL 为 md (16px)
**AND** 卡片之间间距 SHALL 为 lg (24px)

#### Scenario: 页面内容区边距

**WHEN** 用户查看页面内容区
**THEN** 内容区左右边距 SHALL 为 xl (32px)
**AND** 内容区顶部边距 SHALL 为 xl (32px)
**AND** 内容区底部边距 SHALL 为 xl (32px)

---

### REQ-4: 圆角系统

系统 SHALL 定义统一的圆角系统。

**圆角值**:
- none: 0px - 无圆角
- sm: 4px - 小圆角（标签、徽章）
- md: 8px - 中等圆角（按钮、输入框）
- lg: 12px - 大圆角（卡片）
- xl: 16px - 超大圆角（模态框）
- full: 9999px - 完全圆形（头像）

**使用规则**:
- 按钮和输入框 SHALL 使用 md (8px)
- 卡片 SHALL 使用 lg (12px)
- 模态框 SHALL 使用 xl (16px)
- 标签和徽章 SHALL 使用 sm (4px)

#### Scenario: 按钮圆角

**WHEN** 用户看到按钮
**THEN** 按钮圆角 SHALL 为 md (8px)
**AND** 所有按钮 SHALL 使用相同的圆角值

#### Scenario: 卡片圆角

**WHEN** 用户看到卡片
**THEN** 卡片圆角 SHALL 为 lg (12px)
**AND** 所有卡片 SHALL 使用相同的圆角值

---

### REQ-5: 阴影系统

系统 SHALL 定义 5 个层级的阴影系统。

**阴影值**:
- sm: `0 1px 2px 0 rgba(0, 0, 0, 0.05)` - 微妙阴影（悬停状态）
- md: `0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px -1px rgba(0, 0, 0, 0.1)` - 标准阴影（卡片）
- lg: `0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -2px rgba(0, 0, 0, 0.1)` - 中等阴影（下拉菜单）
- xl: `0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -4px rgba(0, 0, 0, 0.1)` - 大阴影（模态框）
- 2xl: `0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 8px 10px -6px rgba(0, 0, 0, 0.1)` - 超大阴影（抽屉）

**使用规则**:
- 卡片默认 SHALL 使用 md 阴影
- 卡片悬停 SHALL 使用 lg 阴影
- 下拉菜单 SHALL 使用 lg 阴影
- 模态框 SHALL 使用 xl 阴影
- 抽屉 SHALL 使用 2xl 阴影

#### Scenario: 卡片阴影

**WHEN** 用户看到卡片
**THEN** 卡片默认阴影 SHALL 为 md
**AND** 当用户悬停在可悬停卡片上时，阴影 SHALL 变为 lg
**AND** 阴影过渡 SHALL 使用 250ms 的动画

---

### REQ-6: 动画系统

系统 SHALL 定义统一的动画系统，包括缓动函数和持续时间。

**缓动函数**:
- standard: `cubic-bezier(0.4, 0.0, 0.2, 1)` - 标准缓动（大多数动画）
- decelerate: `cubic-bezier(0.0, 0.0, 0.2, 1)` - 减速缓动（进入动画）
- accelerate: `cubic-bezier(0.4, 0.0, 1, 1)` - 加速缓动（退出动画）
- bounce: `cubic-bezier(0.68, -0.55, 0.265, 1.55)` - 弹性缓动（特殊效果）

**持续时间**:
- fast: 150ms - 快速动画（悬停、焦点）
- normal: 250ms - 标准动画（默认）
- slow: 350ms - 慢速动画（页面切换）
- slower: 500ms - 更慢动画（复杂动画）

**使用规则**:
- 悬停和焦点状态 SHALL 使用 fast (150ms)
- 组件动画 SHALL 使用 normal (250ms)
- 页面切换 SHALL 使用 slow (350ms)
- 复杂动画 SHALL 使用 slower (500ms)
- 默认 SHALL 使用 standard 缓动函数

#### Scenario: 按钮悬停动画

**WHEN** 用户悬停在按钮上
**THEN** 背景色过渡 SHALL 使用 fast (150ms) 持续时间
**AND** 缓动函数 SHALL 为 standard
**AND** 阴影过渡 SHALL 使用相同的持续时间和缓动函数

#### Scenario: 页面切换动画

**WHEN** 用户切换页面
**THEN** 页面淡入动画 SHALL 使用 slow (350ms) 持续时间
**AND** 缓动函数 SHALL 为 decelerate
**AND** 页面淡出动画 SHALL 使用 accelerate 缓动函数

---

### REQ-7: 图标系统

系统 SHALL 定义统一的图标系统。

**图标库**:
- 系统 SHALL 使用 Ant Design Icons
- 图标 SHALL 保持一致的视觉风格

**图标尺寸**:
- xs: 12px - 极小图标（内联文本）
- sm: 14px - 小图标（按钮）
- md: 16px - 中等图标（默认）
- lg: 20px - 大图标（标题）
- xl: 24px - 超大图标（页面图标）
- 2xl: 32px - 特大图标（空状态）

**使用规则**:
- 按钮图标 SHALL 使用 sm (14px)
- 列表和表格图标 SHALL 使用 md (16px)
- 页面标题图标 SHALL 使用 lg (20px)
- 空状态图标 SHALL 使用 2xl (32px)

#### Scenario: 按钮图标尺寸

**WHEN** 用户看到带图标的按钮
**THEN** 图标尺寸 SHALL 为 sm (14px)
**AND** 图标与文字间距 SHALL 为 xs (4px)
**AND** 图标颜色 SHALL 与文字颜色一致

---

### REQ-8: 断点系统

系统 SHALL 定义响应式断点系统。

**断点值**:
- xs: 0px - 手机（竖屏）
- sm: 640px - 手机（横屏）
- md: 768px - 平板
- lg: 1024px - 笔记本
- xl: 1280px - 桌面
- 2xl: 1536px - 大屏幕

**使用规则**:
- 移动优先设计 SHALL 从 xs 开始
- 布局 SHALL 在每个断点进行适配
- 关键内容 SHALL 在所有断点可见

#### Scenario: 响应式布局

**WHEN** 用户在不同设备上访问系统
**THEN** 在 xs-sm (0-640px) 时，SHALL 显示单列布局和底部导航
**AND** 在 md (768px+) 时，SHALL 显示两列布局和可折叠侧边栏
**AND** 在 lg (1024px+) 时，SHALL 显示完整侧边栏和多列布局

---

### REQ-9: Z-Index 层级

系统 SHALL 定义统一的 Z-Index 层级系统。

**层级值**:
- dropdown: 1000 - 下拉菜单
- sticky: 1020 - 固定元素
- fixed: 1030 - 固定导航
- modal-backdrop: 1040 - 模态框背景
- modal: 1050 - 模态框
- popover: 1060 - 弹出框
- tooltip: 1070 - 工具提示
- notification: 1080 - 通知

**使用规则**:
- 组件 SHALL 使用预定义的 Z-Index 值
- 不得随意使用自定义 Z-Index 值
- 层级 SHALL 保持一致性

#### Scenario: 模态框层级

**WHEN** 用户打开模态框
**THEN** 模态框背景 Z-Index SHALL 为 1040
**AND** 模态框内容 Z-Index SHALL 为 1050
**AND** 模态框 SHALL 覆盖所有其他内容（除通知外）

---

## Implementation Notes

### Tailwind CSS 配置

设计系统 SHALL 通过 Tailwind CSS 配置实现：

```javascript
// tailwind.config.js
export default {
  theme: {
    extend: {
      colors: {
        primary: {
          50: '#eef2ff',
          100: '#e0e7ff',
          200: '#c7d2fe',
          500: '#6366f1',
          600: '#4f46e5',
          700: '#4338ca',
          900: '#312e81',
        },
        gray: {
          50: '#fafbfc',
          100: '#f8f9fa',
          200: '#e9ecef',
          300: '#dee2e6',
          400: '#ced4da',
          500: '#6c757d',
          700: '#495057',
          900: '#212529',
        },
      },
      spacing: {
        xs: '4px',
        sm: '8px',
        md: '16px',
        lg: '24px',
        xl: '32px',
        '2xl': '48px',
        '3xl': '64px',
      },
      borderRadius: {
        sm: '4px',
        md: '8px',
        lg: '12px',
        xl: '16px',
      },
      boxShadow: {
        sm: '0 1px 2px 0 rgba(0, 0, 0, 0.05)',
        md: '0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px -1px rgba(0, 0, 0, 0.1)',
        lg: '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -2px rgba(0, 0, 0, 0.1)',
        xl: '0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -4px rgba(0, 0, 0, 0.1)',
        '2xl': '0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 8px 10px -6px rgba(0, 0, 0, 0.1)',
      },
    },
  },
};
```

### Ant Design 主题配置

设计系统 SHALL 通过 Ant Design 主题配置实现：

```typescript
// src/theme/antd-theme.ts
import { ThemeConfig } from 'antd';

export const antdTheme: ThemeConfig = {
  token: {
    colorPrimary: '#6366f1',
    colorSuccess: '#10b981',
    colorWarning: '#f59e0b',
    colorError: '#ef4444',
    colorInfo: '#3b82f6',
    borderRadius: 8,
    fontSize: 14,
    lineHeight: 1.5,
  },
};
```

### 设计 Token 文件

设计系统 SHALL 提供 TypeScript 类型定义：

```typescript
// src/design-system/tokens.ts
export const colors = {
  primary: {
    50: '#eef2ff',
    100: '#e0e7ff',
    200: '#c7d2fe',
    500: '#6366f1',
    600: '#4f46e5',
    700: '#4338ca',
    900: '#312e81',
  },
  // ... 其他颜色
};

export const spacing = {
  xs: '4px',
  sm: '8px',
  md: '16px',
  lg: '24px',
  xl: '32px',
  '2xl': '48px',
  '3xl': '64px',
};

// ... 其他 token
```

---

## References

- [Tailwind CSS Documentation](https://tailwindcss.com/)
- [Ant Design Theming](https://ant.design/docs/react/customize-theme)
- [Material Design Color System](https://material.io/design/color)
- [Vercel Design System](https://vercel.com/design)


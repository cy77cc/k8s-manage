# Spec: ui-animations

**Status**: Draft
**Created**: 2026-03-02

---

## Overview

本规范定义了 OpsPilot PaaS 平台的动画系统。动画系统提供流畅、自然的视觉反馈，提升用户体验。

---

## Requirements

### REQ-1: 页面切换动画

系统 SHALL 在页面切换时提供动画效果。

**动画类型**: 淡入淡出 (Fade)
**持续时间**: 350ms
**缓动函数**: decelerate

#### Scenario: 页面切换

**WHEN** 用户从服务列表跳转到服务详情
**THEN** 服务列表 SHALL 淡出
**AND** 服务详情 SHALL 淡入
**AND** 动画 SHALL 为 350ms

---

### REQ-2: 组件动画

系统 SHALL 为组件提供动画效果。

**卡片悬停**: 向上移动 2px + 阴影变化，250ms
**按钮点击**: 缩放到 98%，150ms
**列表项展开**: 高度动画，250ms

#### Scenario: 卡片悬停

**WHEN** 用户悬停在卡片上
**THEN** 卡片 SHALL 向上移动 2px
**AND** 阴影 SHALL 从 md 变为 lg
**AND** 动画 SHALL 为 250ms

---

### REQ-3: 加载动画

系统 SHALL 提供加载动画。

**骨架屏**: 渐变扫过效果，1.5s 循环
**Spinner**: 旋转动画，1s 循环
**进度条**: 宽度动画，根据进度

#### Scenario: 骨架屏动画

**WHEN** 页面正在加载
**THEN** SHALL 显示骨架屏
**AND** SHALL 显示渐变扫过动画
**AND** 动画 SHALL 持续 1.5s 并循环

---

## Implementation Notes

使用 Framer Motion 实现动画。

---

## References

- [Framer Motion](https://www.framer.com/motion/)


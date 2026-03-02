# OpsPilot 设计系统规范

## 间距系统 (Spacing)

采用 8px 基准的间距系统：

```
xs    4px   0.25rem  ← 极小间距（图标与文字）
sm    8px   0.5rem   ← 小间距（组件内部）
md    16px  1rem     ← 中等间距（默认）
lg    24px  1.5rem   ← 大间距（组件之间）
xl    32px  2rem     ← 超大间距（区块之间）
2xl   48px  3rem     ← 页面级间距
3xl   64px  4rem     ← 大区块间距
```

### 使用规则

- **组件内部**: 使用 sm (8px)
- **组件之间**: 使用 md (16px) 或 lg (24px)
- **页面边距**: 使用 xl (32px)
- **区块间距**: 使用 2xl (48px)

## 圆角系统 (Border Radius)

```
none    0px      ← 无圆角
sm      4px      ← 小圆角（标签、徽章）
md      8px      ← 中等圆角（按钮、输入框）
lg      12px     ← 大圆角（卡片）
xl      16px     ← 超大圆角（模态框）
full    9999px   ← 完全圆形（头像）
```

## 阴影系统 (Shadows)

```css
/* 微妙阴影 - 用于悬停状态 */
shadow-sm: 0 1px 2px 0 rgba(0, 0, 0, 0.05);

/* 标准阴影 - 用于卡片 */
shadow-md: 0 1px 3px 0 rgba(0, 0, 0, 0.1),
           0 1px 2px -1px rgba(0, 0, 0, 0.1);

/* 中等阴影 - 用于下拉菜单 */
shadow-lg: 0 4px 6px -1px rgba(0, 0, 0, 0.1),
           0 2px 4px -2px rgba(0, 0, 0, 0.1);

/* 大阴影 - 用于模态框 */
shadow-xl: 0 10px 15px -3px rgba(0, 0, 0, 0.1),
           0 4px 6px -4px rgba(0, 0, 0, 0.1);

/* 超大阴影 - 用于抽屉 */
shadow-2xl: 0 20px 25px -5px rgba(0, 0, 0, 0.1),
            0 8px 10px -6px rgba(0, 0, 0, 0.1);
```

## 边框系统 (Borders)

```
细边框    1px solid #dee2e6  ← 默认边框
中边框    2px solid #dee2e6  ← 强调边框
粗边框    4px solid #6366f1  ← 焦点边框
```

## 动画系统 (Animations)

### 缓动函数 (Easing)

```css
/* 标准缓动 - 用于大多数动画 */
ease-standard: cubic-bezier(0.4, 0.0, 0.2, 1);

/* 减速缓动 - 用于进入动画 */
ease-decelerate: cubic-bezier(0.0, 0.0, 0.2, 1);

/* 加速缓动 - 用于退出动画 */
ease-accelerate: cubic-bezier(0.4, 0.0, 1, 1);

/* 弹性缓动 - 用于特殊效果 */
ease-bounce: cubic-bezier(0.68, -0.55, 0.265, 1.55);
```

### 持续时间 (Duration)

```
fast      150ms  ← 快速动画（悬停、焦点）
normal    250ms  ← 标准动画（默认）
slow      350ms  ← 慢速动画（页面切换）
slower    500ms  ← 更慢动画（复杂动画）
```

### 常用动画

```css
/* 淡入淡出 */
@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

/* 滑入 */
@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 缩放 */
@keyframes scaleIn {
  from {
    opacity: 0;
    transform: scale(0.95);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}
```

## 图标系统 (Icons)

使用 Ant Design Icons，统一尺寸：

```
xs    12px  ← 极小图标（内联文本）
sm    14px  ← 小图标（按钮）
md    16px  ← 中等图标（默认）
lg    20px  ← 大图标（标题）
xl    24px  ← 超大图标（页面图标）
2xl   32px  ← 特大图标（空状态）
```

## Z-Index 层级

```
dropdown    1000  ← 下拉菜单
sticky      1020  ← 固定元素
fixed       1030  ← 固定导航
modal-backdrop 1040 ← 模态框背景
modal       1050  ← 模态框
popover     1060  ← 弹出框
tooltip     1070  ← 工具提示
notification 1080 ← 通知
```

## 断点系统 (Breakpoints)

```
xs    0px      ← 手机（竖屏）
sm    640px    ← 手机（横屏）
md    768px    ← 平板
lg    1024px   ← 笔记本
xl    1280px   ← 桌面
2xl   1536px   ← 大屏幕
```

## 栅格系统 (Grid)

```
列数: 12列
间距: 24px (lg), 16px (md), 8px (sm)
容器最大宽度: 1280px
```


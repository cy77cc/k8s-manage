# OpsPilot UI/UX 重设计 - 快速参考

## 📋 文档索引

| 文档 | 描述 | 用途 |
|------|------|------|
| [UI_UX_REDESIGN_PROPOSAL.md](./UI_UX_REDESIGN_PROPOSAL.md) | 执行摘要和概览 | 了解整体方案 |
| [design-system-spec.md](./design-system-spec.md) | 设计系统规范 | 查阅色彩、排版、间距等 |
| [component-library-spec.md](./component-library-spec.md) | 组件库规范 | 查阅组件设计规范 |
| [page-redesign-spec.md](./page-redesign-spec.md) | 页面重设计方案 | 查阅页面布局和设计 |
| [implementation-plan.md](./implementation-plan.md) | 实施计划 | 查阅开发计划和时间表 |

---

## 🎨 设计系统速查

### 色彩

```css
/* 主色 */
--primary-500: #6366f1;
--primary-600: #4f46e5;

/* 中性色 */
--gray-50: #fafbfc;
--gray-100: #f8f9fa;
--gray-500: #6c757d;
--gray-900: #212529;

/* 语义色 */
--success: #10b981;
--warning: #f59e0b;
--error: #ef4444;
--info: #3b82f6;
```

### 间距

```css
--spacing-xs: 4px;
--spacing-sm: 8px;
--spacing-md: 16px;
--spacing-lg: 24px;
--spacing-xl: 32px;
--spacing-2xl: 48px;
```

### 圆角

```css
--radius-sm: 4px;
--radius-md: 8px;
--radius-lg: 12px;
--radius-xl: 16px;
```

### 阴影

```css
--shadow-sm: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
--shadow-md: 0 1px 3px 0 rgba(0, 0, 0, 0.1);
--shadow-lg: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
```

---

## 🧩 组件速查

### Button

```tsx
// Primary
<Button type="primary">创建服务</Button>

// Secondary
<Button>取消</Button>

// Danger
<Button danger>删除</Button>

// Sizes
<Button size="small">小按钮</Button>
<Button size="middle">中按钮</Button>
<Button size="large">大按钮</Button>
```

### Input

```tsx
// Standard
<Input placeholder="请输入..." />

// With Icon
<Input prefix={<SearchOutlined />} placeholder="搜索..." />

// Password
<Input.Password placeholder="请输入密码" />
```

### Card

```tsx
// Standard
<Card title="标题">内容</Card>

// Hoverable
<Card hoverable>内容</Card>

// No Border
<Card bordered={false}>内容</Card>
```

---

## 📐 布局速查

### 页面结构

```
┌────────────────────────────────────────┐
│  ┌──────┐  ┌──────────────────────────┐│
│  │      │  │  Header (64px)           ││
│  │  S   │  ├──────────────────────────┤│
│  │  i   │  │                          ││
│  │  d   │  │  Content (padding: 32px) ││
│  │  e   │  │                          ││
│  │  b   │  │                          ││
│  │  a   │  │                          ││
│  │  r   │  │                          ││
│  │      │  │                          ││
│  └──────┘  └──────────────────────────┘│
│   240px                                 │
└────────────────────────────────────────┘
```

### 响应式断点

```
xs:   0px    (手机竖屏)
sm:   640px  (手机横屏)
md:   768px  (平板)
lg:   1024px (笔记本)
xl:   1280px (桌面)
2xl:  1536px (大屏幕)
```

---

## ⌨️ 快捷键速查

### 全局快捷键

```
Cmd/Ctrl + K    打开命令面板
/               聚焦搜索
Esc             关闭模态框/面板
```

### 导航快捷键

```
g + h           回到首页
g + s           服务列表
g + d           部署管理
g + m           监控告警
g + h           主机管理
```

### 列表快捷键

```
j               下一项
k               上一项
Enter           打开/选择
Space           选中/取消选中
```

---

## 🚀 性能目标

| 指标 | 目标值 |
|------|--------|
| 首屏加载时间 | < 2秒 |
| 页面切换时间 | < 300ms |
| Lighthouse 性能评分 | > 90 |
| 打包体积减小 | 30% |

---

## 📅 时间表

| 阶段 | 时长 | 内容 |
|------|------|------|
| 阶段 1 | 2周 | 设计系统建立 |
| 阶段 2 | 3周 | 组件库重构 |
| 阶段 3 | 4周 | 核心页面重设计 |
| 阶段 4 | 2周 | 交互优化 |
| 阶段 5 | 2周 | 性能优化 |
| 阶段 6 | 2周 | 测试和修复 |
| **总计** | **15周** | **约 3.5 个月** |

---

## ✅ 检查清单

### 设计系统

- [ ] 色彩系统定义完成
- [ ] 排版系统定义完成
- [ ] 间距系统定义完成
- [ ] 圆角系统定义完成
- [ ] 阴影系统定义完成
- [ ] 动画系统定义完成

### 组件库

- [ ] Button 组件重构完成
- [ ] Input 组件重构完成
- [ ] Card 组件重构完成
- [ ] Table 组件重构完成
- [ ] Modal 组件重构完成
- [ ] Form 组件重构完成

### 核心页面

- [ ] Dashboard 重设计完成
- [ ] 服务列表重设计完成
- [ ] 服务详情重设计完成
- [ ] 部署管理重设计完成
- [ ] 主机管理重设计完成
- [ ] 监控页面重设计完成

### 交互优化

- [ ] 命令面板实现完成
- [ ] 键盘快捷键实现完成
- [ ] 页面切换动画完成
- [ ] 组件动画完成

### 性能优化

- [ ] 虚拟滚动实现完成
- [ ] 代码分割完成
- [ ] 资源预加载完成
- [ ] 乐观更新完成

### 测试

- [ ] 浏览器兼容性测试通过
- [ ] 响应式测试通过
- [ ] 功能测试通过
- [ ] 性能测试通过

---

## 🔗 相关链接

### 设计参考

- [Vercel Dashboard](https://vercel.com/dashboard)
- [Linear](https://linear.app/)
- [Stripe Dashboard](https://dashboard.stripe.com/)
- [Tailwind UI](https://tailwindui.com/)

### 技术文档

- [Ant Design](https://ant.design/)
- [Tailwind CSS](https://tailwindcss.com/)
- [Framer Motion](https://www.framer.com/motion/)
- [cmdk](https://cmdk.paco.me/)
- [react-window](https://react-window.vercel.app/)

---

## 📞 联系方式

如有问题或建议，请联系：

- **设计团队**: design@opspilot.com
- **开发团队**: dev@opspilot.com
- **项目经理**: pm@opspilot.com

---

**最后更新**: 2026-03-02

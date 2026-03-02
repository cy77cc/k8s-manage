# OpsPilot UI/UX 重设计实现计划

## 1. 项目概览

### 1.1 项目范围

**包含内容**:
- 设计系统建立（色彩、排版、间距等）
- 组件库重构（基于 Ant Design 深度定制）
- 核心页面重设计（Dashboard、服务管理、部署管理等）
- 交互优化（命令面板、键盘快捷键、动画等）
- 性能优化（虚拟滚动、代码分割、预加载等）
- 响应式适配（桌面、平板、移动端）

**不包含内容**:
- 后端 API 变更
- 业务逻辑重构
- 数据库结构调整

### 1.2 技术栈

**保持不变**:
- React 19.2.0
- TypeScript
- Vite
- React Router
- Ant Design 6.3.0 (深度定制)
- Tailwind CSS 3.4.19

**新增依赖**:
```json
{
  "cmdk": "^1.0.0",           // 命令面板
  "react-window": "^1.8.10",  // 虚拟滚动
  "framer-motion": "^11.15.0" // 已有，用于动画
}
```

### 1.3 时间估算

```
阶段 1: 设计系统建立        2周
阶段 2: 组件库重构          3周
阶段 3: 核心页面重设计      4周
阶段 4: 交互优化            2周
阶段 5: 性能优化            2周
阶段 6: 测试和修复          2周
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
总计                        15周 (约 3.5 个月)
```

---

## 2. 阶段 1: 设计系统建立 (2周)

### 2.1 目标

建立统一的设计系统，包括色彩、排版、间距、阴影等基础规范。

### 2.2 任务清单

#### Week 1: 设计 Token 定义

- [ ] 创建设计 Token 文件 `src/design-system/tokens.ts`
- [ ] 定义色彩系统（主色、中性色、语义色）
- [ ] 定义排版系统（字体、字号、行高、字重）
- [ ] 定义间距系统（8px 基准）
- [ ] 定义圆角系统
- [ ] 定义阴影系统
- [ ] 定义动画系统（缓动函数、持续时间）
- [ ] 定义断点系统（响应式）

#### Week 2: Tailwind 配置和 Ant Design 主题

- [ ] 配置 Tailwind CSS (`tailwind.config.js`)
  - 扩展色彩
  - 扩展间距
  - 扩展字体
  - 扩展阴影
- [ ] 配置 Ant Design 主题 (`src/theme/antd-theme.ts`)
  - 全局 token
  - 组件 token
- [ ] 创建全局样式文件 (`src/styles/global.css`)
- [ ] 更新 `index.css` 应用新的设计系统
- [ ] 创建设计系统文档 (`docs/design-system.md`)

### 2.3 交付物

- `src/design-system/tokens.ts` - 设计 Token 定义
- `tailwind.config.js` - Tailwind 配置
- `src/theme/antd-theme.ts` - Ant Design 主题配置
- `src/styles/global.css` - 全局样式
- `docs/design-system.md` - 设计系统文档

### 2.4 验收标准

- [ ] 所有设计 Token 已定义并导出
- [ ] Tailwind 配置正确，可以使用自定义类名
- [ ] Ant Design 主题应用成功，组件样式符合设计规范
- [ ] 全局样式生效，页面基础样式正确
- [ ] 设计系统文档完整，包含所有规范和示例

---

## 3. 阶段 2: 组件库重构 (3周)

### 3.1 目标

基于新的设计系统，重构核心组件，建立统一的组件库。

### 3.2 任务清单

#### Week 1: 基础组件

- [ ] Button 组件重构
  - 更新样式（圆角、阴影、动画）
  - 添加新的变体
  - 优化交互状态
- [ ] Input 组件重构
  - 更新样式
  - 优化 focus 状态
  - 添加错误状态样式
- [ ] Card 组件重构
  - 更新样式（圆角、阴影、边框）
  - 添加 hover 效果
  - 优化内边距
- [ ] Tag 组件重构
  - 更新色彩方案
  - 优化尺寸
- [ ] Modal 组件重构
  - 更新样式
  - 优化动画
  - 添加背景模糊效果

#### Week 2: 数据展示组件

- [ ] Table 组件重构
  - 增加行高
  - 优化 hover 状态
  - 添加内联操作
  - 实现虚拟滚动（大数据量）
- [ ] List 组件重构
  - 更新样式
  - 优化间距
- [ ] Descriptions 组件重构
  - 更新样式
  - 优化布局
- [ ] Empty 组件重构
  - 重新设计空状态
  - 添加插图
- [ ] Skeleton 组件重构
  - 更新动画
  - 优化样式

#### Week 3: 反馈组件和布局组件

- [ ] Notification 组件重构
  - 更新样式
  - 优化动画
  - 添加左侧色条
- [ ] Message 组件重构
  - 更新样式
  - 优化位置
- [ ] Progress 组件重构
  - 更新色彩
  - 优化动画
- [ ] Spin 组件重构
  - 更新样式
- [ ] Layout 组件重构
  - 重新设计侧边栏
  - 重新设计顶部导航
  - 优化响应式布局

### 3.3 交付物

- 重构后的组件文件（覆盖原有组件样式）
- 组件 Storybook 文档（可选）
- 组件使用文档

### 3.4 验收标准

- [ ] 所有组件样式符合设计规范
- [ ] 组件交互流畅，动画自然
- [ ] 组件在不同尺寸下表现正常
- [ ] 组件可访问性良好（键盘导航、ARIA 属性）
- [ ] 组件文档完整

---

## 4. 阶段 3: 核心页面重设计 (4周)

### 4.1 目标

基于新的设计系统和组件库，重设计核心页面。

### 4.2 任务清单

#### Week 1: 布局和 Dashboard

- [ ] 重构 AppLayout 组件
  - 重新设计侧边栏（浅色主题）
  - 重新设计顶部导航
  - 优化响应式布局
  - 添加移动端底部导航
- [ ] 重构 Dashboard 页面
  - 重新设计统计卡片
  - 重新设计服务健康列表
  - 添加资源使用趋势图表
  - 优化布局和间距

#### Week 2: 服务管理

- [ ] 重构服务列表页 (`ServiceListPage`)
  - 重新设计筛选栏
  - 重新设计服务卡片
  - 添加快捷操作
  - 优化加载状态
- [ ] 重构服务详情页 (`ServiceDetailPage`)
  - 重新设计页面头部
  - 重新设计指标展示
  - 重新设计实例列表
  - 重新设计部署历史
  - 添加 Tab 导航

#### Week 3: 部署管理

- [ ] 重构部署列表页 (`DeploymentPage`)
  - 重新设计部署卡片
  - 添加状态筛选
  - 优化时间线展示
- [ ] 创建部署流程页
  - 设计步骤式部署界面
  - 添加版本选择
  - 添加参数配置
  - 添加部署进度展示
  - 添加实时日志

#### Week 4: 其他核心页面

- [ ] 重构主机列表页 (`HostListPage`)
  - 重新设计主机卡片
  - 优化筛选和搜索
- [ ] 重构主机详情页 (`HostDetailPage`)
  - 重新设计指标展示
  - 优化布局
- [ ] 重构监控页面 (`MonitorPage`)
  - 重新设计告警列表
  - 优化图表展示
- [ ] 重构配置中心页面
  - 重新设计配置列表
  - 优化编辑器界面

### 4.3 交付物

- 重构后的页面组件
- 页面截图和演示视频
- 页面文档

### 4.4 验收标准

- [ ] 所有页面样式符合设计规范
- [ ] 页面布局合理，信息层次清晰
- [ ] 页面交互流畅，动画自然
- [ ] 页面在不同尺寸下表现正常
- [ ] 页面加载速度快，性能良好

---

## 5. 阶段 4: 交互优化 (2周)

### 5.1 目标

添加高级交互功能，提升用户体验。

### 5.2 任务清单

#### Week 1: 命令面板和键盘快捷键

- [ ] 实现命令面板 (`CommandPalette`)
  - 集成 `cmdk` 库
  - 添加全局快捷键 (Cmd+K / Ctrl+K)
  - 实现导航命令
  - 实现搜索命令
  - 实现操作命令（创建服务、部署等）
  - 添加最近使用记录
- [ ] 实现键盘快捷键系统
  - 定义快捷键映射
  - 实现快捷键监听
  - 添加快捷键提示
  - 实现常用快捷键：
    - `j/k`: 列表上下移动
    - `/`: 聚焦搜索
    - `g+h`: 回到首页
    - `g+s`: 服务列表
    - `g+d`: 部署管理
    - `Esc`: 关闭模态框/面板

#### Week 2: 动画和微交互

- [ ] 添加页面切换动画
  - 使用 Framer Motion
  - 实现淡入淡出效果
  - 实现滑动效果
- [ ] 添加组件动画
  - 卡片 hover 动画
  - 按钮点击动画
  - 列表项展开/折叠动画
  - 加载动画
- [ ] 添加微交互
  - 按钮点击反馈
  - 表单验证反馈
  - 操作成功/失败反馈
  - 拖拽反馈

### 5.3 交付物

- `CommandPalette` 组件
- 键盘快捷键系统
- 动画配置文件
- 交互文档

### 5.4 验收标准

- [ ] 命令面板功能完整，响应快速
- [ ] 键盘快捷键工作正常，无冲突
- [ ] 动画流畅自然，不卡顿
- [ ] 微交互细腻，提升用户体验
- [ ] 交互文档完整，包含所有快捷键说明

---

## 6. 阶段 5: 性能优化 (2周)

### 6.1 目标

优化应用性能，提升加载速度和响应速度。

### 6.2 任务清单

#### Week 1: 代码优化

- [ ] 实现虚拟滚动
  - 使用 `react-window` 优化长列表
  - 应用到服务列表、主机列表、日志列表等
- [ ] 实现代码分割
  - 路由级别代码分割
  - 组件级别懒加载
  - 动态导入大型依赖
- [ ] 优化图片和资源
  - 压缩图片
  - 使用 WebP 格式
  - 实现图片懒加载
- [ ] 优化打包配置
  - 配置 Vite 打包优化
  - 启用 Tree Shaking
  - 启用代码压缩

#### Week 2: 数据优化

- [ ] 实现数据预加载
  - 预加载关键数据
  - 预加载下一页数据
- [ ] 实现乐观更新
  - 操作立即反馈
  - 后台同步数据
- [ ] 优化 API 调用
  - 合并重复请求
  - 实现请求缓存
  - 实现请求去抖
- [ ] 实现骨架屏
  - 替换 loading spinner
  - 提升感知性能

### 6.3 交付物

- 优化后的代码
- 性能测试报告
- 优化文档

### 6.4 验收标准

- [ ] 首屏加载时间 < 2秒
- [ ] 页面切换时间 < 300ms
- [ ] 长列表滚动流畅，无卡顿
- [ ] 打包体积减小 30% 以上
- [ ] Lighthouse 性能评分 > 90

---

## 7. 阶段 6: 测试和修复 (2周)

### 7.1 目标

全面测试应用，修复 bug，确保质量。

### 7.2 任务清单

#### Week 1: 功能测试

- [ ] 浏览器兼容性测试
  - Chrome
  - Firefox
  - Safari
  - Edge
- [ ] 响应式测试
  - 桌面 (1920x1080, 1366x768)
  - 平板 (768x1024)
  - 手机 (375x667, 414x896)
- [ ] 功能测试
  - 所有页面功能正常
  - 所有交互正常
  - 所有动画正常
- [ ] 性能测试
  - 加载速度测试
  - 内存使用测试
  - CPU 使用测试

#### Week 2: Bug 修复和优化

- [ ] 修复测试中发现的 bug
- [ ] 优化性能问题
- [ ] 优化样式细节
- [ ] 优化交互细节
- [ ] 更新文档
- [ ] 准备发布

### 7.3 交付物

- 测试报告
- Bug 修复记录
- 最终版本代码
- 发布文档

### 7.4 验收标准

- [ ] 所有已知 bug 已修复
- [ ] 所有功能测试通过
- [ ] 所有性能指标达标
- [ ] 代码质量良好，无明显技术债
- [ ] 文档完整，可以交付

---

## 8. 技术实现细节

### 8.1 设计 Token 实现

```typescript
// src/design-system/tokens.ts

export const colors = {
  // Primary
  primary: {
    50: '#eef2ff',
    100: '#e0e7ff',
    200: '#c7d2fe',
    500: '#6366f1',
    600: '#4f46e5',
    700: '#4338ca',
    900: '#312e81',
  },

  // Neutral
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

  // Semantic
  success: '#10b981',
  warning: '#f59e0b',
  error: '#ef4444',
  info: '#3b82f6',
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

export const borderRadius = {
  none: '0',
  sm: '4px',
  md: '8px',
  lg: '12px',
  xl: '16px',
  full: '9999px',
};

export const shadows = {
  sm: '0 1px 2px 0 rgba(0, 0, 0, 0.05)',
  md: '0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px -1px rgba(0, 0, 0, 0.1)',
  lg: '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -2px rgba(0, 0, 0, 0.1)',
  xl: '0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -4px rgba(0, 0, 0, 0.1)',
  '2xl': '0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 8px 10px -6px rgba(0, 0, 0, 0.1)',
};

export const animations = {
  duration: {
    fast: '150ms',
    normal: '250ms',
    slow: '350ms',
    slower: '500ms',
  },
  easing: {
    standard: 'cubic-bezier(0.4, 0.0, 0.2, 1)',
    decelerate: 'cubic-bezier(0.0, 0.0, 0.2, 1)',
    accelerate: 'cubic-bezier(0.4, 0.0, 1, 1)',
    bounce: 'cubic-bezier(0.68, -0.55, 0.265, 1.55)',
  },
};
```

### 8.2 Tailwind 配置

```javascript
// tailwind.config.js

import { colors, spacing, borderRadius, shadows } from './src/design-system/tokens';

export default {
  content: [
    './index.html',
    './src/**/*.{js,ts,jsx,tsx}',
  ],
  theme: {
    extend: {
      colors,
      spacing,
      borderRadius,
      boxShadow: shadows,
      fontFamily: {
        sans: [
          '-apple-system',
          'BlinkMacSystemFont',
          'Segoe UI',
          'PingFang SC',
          'Hiragino Sans GB',
          'Microsoft YaHei',
          'sans-serif',
        ],
        mono: [
          'SF Mono',
          'Monaco',
          'Cascadia Code',
          'Consolas',
          'monospace',
        ],
      },
    },
  },
  plugins: [],
};
```

### 8.3 Ant Design 主题配置

```typescript
// src/theme/antd-theme.ts

import { ThemeConfig } from 'antd';
import { colors } from '../design-system/tokens';

export const antdTheme: ThemeConfig = {
  token: {
    colorPrimary: colors.primary[500],
    colorSuccess: colors.success,
    colorWarning: colors.warning,
    colorError: colors.error,
    colorInfo: colors.info,

    borderRadius: 8,
    fontSize: 14,
    lineHeight: 1.5,

    padding: 16,
    paddingLG: 24,
    paddingXL: 32,

    lineWidth: 1,
    lineType: 'solid',

    boxShadow: '0 1px 3px 0 rgba(0, 0, 0, 0.08)',
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
      rowHoverBg: colors.gray[100],
      headerBg: colors.gray[100],
      cellPaddingBlock: 16,
      borderColor: colors.gray[200],
    },

    Modal: {
      borderRadius: 12,
    },
  },
};
```

### 8.4 命令面板实现

```typescript
// src/components/CommandPalette/CommandPalette.tsx

import { useEffect, useState } from 'react';
import { Command } from 'cmdk';
import { useNavigate } from 'react-router-dom';
import './command-palette.css';

export const CommandPalette = () => {
  const [open, setOpen] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (e.key === 'k' && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setOpen((open) => !open);
      }
    };

    document.addEventListener('keydown', down);
    return () => document.removeEventListener('keydown', down);
  }, []);

  const handleSelect = (callback: () => void) => {
    callback();
    setOpen(false);
  };

  return (
    <Command.Dialog open={open} onOpenChange={setOpen} label="全局命令">
      <Command.Input placeholder="搜索或执行命令..." />
      <Command.List>
        <Command.Empty>未找到结果</Command.Empty>

        <Command.Group heading="导航">
          <Command.Item onSelect={() => handleSelect(() => navigate('/'))}>
            <span>主控台</span>
          </Command.Item>
          <Command.Item onSelect={() => handleSelect(() => navigate('/services'))}>
            <span>服务列表</span>
          </Command.Item>
          <Command.Item onSelect={() => handleSelect(() => navigate('/deployment'))}>
            <span>部署管理</span>
          </Command.Item>
          <Command.Item onSelect={() => handleSelect(() => navigate('/hosts'))}>
            <span>主机管理</span>
          </Command.Item>
          <Command.Item onSelect={() => handleSelect(() => navigate('/monitor'))}>
            <span>监控告警</span>
          </Command.Item>
        </Command.Group>

        <Command.Group heading="操作">
          <Command.Item onSelect={() => handleSelect(() => navigate('/services/provision'))}>
            <span>创建新服务</span>
          </Command.Item>
          <Command.Item onSelect={() => handleSelect(() => navigate('/deployment'))}>
            <span>部署服务</span>
          </Command.Item>
          <Command.Item onSelect={() => handleSelect(() => navigate('/hosts/onboarding'))}>
            <span>添加主机</span>
          </Command.Item>
        </Command.Group>
      </Command.List>
    </Command.Dialog>
  );
};
```

---

## 9. 风险和挑战

### 9.1 技术风险

**风险**: Ant Design 深度定制可能导致升级困难
**缓解**:
- 使用 Ant Design 的主题系统，避免直接修改组件源码
- 记录所有定制内容
- 定期关注 Ant Design 更新

**风险**: 性能优化可能引入新的 bug
**缓解**:
- 充分测试
- 渐进式优化
- 保留回滚方案

### 9.2 项目风险

**风险**: 时间估算不准确，项目延期
**缓解**:
- 预留 buffer 时间
- 定期评估进度
- 及时调整计划

**风险**: 设计和开发不一致
**缓解**:
- 建立设计系统
- 定期 review
- 使用设计 Token

---

## 10. 成功指标

### 10.1 用户体验指标

- [ ] 用户满意度提升 30%
- [ ] 任务完成时间减少 20%
- [ ] 错误率降低 50%

### 10.2 性能指标

- [ ] 首屏加载时间 < 2秒
- [ ] 页面切换时间 < 300ms
- [ ] Lighthouse 性能评分 > 90
- [ ] 打包体积减小 30%

### 10.3 代码质量指标

- [ ] 代码覆盖率 > 80%
- [ ] 无严重 bug
- [ ] 技术债务减少 50%

---

## 11. 后续计划

### 11.1 持续优化

- 收集用户反馈
- 持续优化性能
- 持续优化交互
- 持续优化样式

### 11.2 功能扩展

- 添加深色模式
- 添加更多主题
- 添加更多动画效果
- 添加更多快捷键

### 11.3 文档完善

- 完善设计系统文档
- 完善组件文档
- 完善开发文档
- 完善用户文档


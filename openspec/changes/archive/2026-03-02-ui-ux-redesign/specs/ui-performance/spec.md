# Spec: ui-performance

**Status**: Draft
**Created**: 2026-03-02

---

## Overview

本规范定义了 OpsPilot PaaS 平台的性能优化要求。性能优化包括代码分割、虚拟滚动、资源预加载、乐观更新等技术，旨在提升应用的加载速度和响应速度。

---

## Requirements

### REQ-1: 首屏加载性能

系统 SHALL 优化首屏加载性能。

**性能目标**:
- 首屏加载时间 SHALL < 2秒
- First Contentful Paint (FCP) SHALL < 1.5秒
- Largest Contentful Paint (LCP) SHALL < 2.5秒
- Time to Interactive (TTI) SHALL < 3秒

**优化措施**:
- SHALL 使用代码分割
- SHALL 使用资源预加载
- SHALL 使用图片懒加载
- SHALL 压缩和优化资源

#### Scenario: 首屏加载

**WHEN** 用户首次访问系统
**THEN** 首屏 SHALL 在 2 秒内加载完成
**AND** 用户 SHALL 能看到主要内容
**AND** 用户 SHALL 能进行交互

---

### REQ-2: 代码分割

系统 SHALL 使用代码分割优化打包体积。

**分割策略**:
- SHALL 按路由分割代码
- SHALL 按组件懒加载
- SHALL 分离第三方库
- SHALL 分离公共代码

**实现规则**:
- 每个路由 SHALL 是独立的 chunk
- 大型组件 SHALL 懒加载
- Ant Design SHALL 按需加载
- 图表库 SHALL 懒加载

#### Scenario: 路由级代码分割

**WHEN** 用户访问服务列表页
**THEN** SHALL 仅加载服务列表相关的代码
**AND** 其他页面的代码 SHALL 不加载
**AND** 切换到其他页面时 SHALL 动态加载对应代码

---

### REQ-3: 虚拟滚动

系统 SHALL 在长列表中使用虚拟滚动。

**适用场景**:
- 服务列表（> 50 项）
- 主机列表（> 50 项）
- 日志查看器（> 100 行）
- 部署历史（> 50 项）

**实现规则**:
- SHALL 使用 react-window 库
- SHALL 仅渲染可见区域的项
- SHALL 支持动态高度（如需要）
- 滚动 SHALL 流畅，无卡顿

#### Scenario: 服务列表虚拟滚动

**WHEN** 服务列表有 200 个服务
**THEN** SHALL 仅渲染可见区域的约 10-15 个服务
**AND** 滚动时 SHALL 动态渲染新的服务
**AND** 滚动 SHALL 流畅，无卡顿

---

### REQ-4: 图片优化

系统 SHALL 优化图片加载。

**优化措施**:
- SHALL 使用 WebP 格式（如果浏览器支持）
- SHALL 压缩图片
- SHALL 使用懒加载
- SHALL 使用响应式图片

**实现规则**:
- 图片 SHALL 在进入视口时加载
- 图片 SHALL 有占位符
- 图片 SHALL 有加载失败处理

#### Scenario: 图片懒加载

**WHEN** 页面包含多张图片
**THEN** 仅可见区域的图片 SHALL 加载
**AND** 滚动时 SHALL 加载新的图片
**AND** 加载前 SHALL 显示占位符

---

### REQ-5: 数据预加载

系统 SHALL 预加载关键数据。

**预加载策略**:
- SHALL 预加载下一页数据
- SHALL 预加载关联数据
- SHALL 预加载常用数据

**实现规则**:
- 预加载 SHALL 在空闲时进行
- 预加载 SHALL 不阻塞主要操作
- 预加载的数据 SHALL 缓存

#### Scenario: 预加载下一页

**WHEN** 用户在服务列表第 1 页
**THEN** 系统 SHALL 在后台预加载第 2 页数据
**AND** 用户点击"下一页"时 SHALL 立即显示
**AND** 不需要等待加载

---

### REQ-6: 乐观更新

系统 SHALL 使用乐观更新提升响应速度。

**适用场景**:
- 点赞/收藏
- 状态切换
- 简单的编辑操作

**实现规则**:
- SHALL 立即更新 UI
- SHALL 在后台发送请求
- 请求失败时 SHALL 回滚 UI
- SHALL 显示错误提示

#### Scenario: 乐观更新

**WHEN** 用户点击"启动服务"按钮
**THEN** 服务状态 SHALL 立即变为"启动中"
**AND** 按钮 SHALL 立即禁用
**AND** 后台 SHALL 发送启动请求
**AND** 如果请求失败，SHALL 回滚状态并显示错误

---

### REQ-7: 请求优化

系统 SHALL 优化 API 请求。

**优化措施**:
- SHALL 合并重复请求
- SHALL 缓存请求结果
- SHALL 使用请求去抖
- SHALL 取消未完成的请求

**实现规则**:
- 相同的请求 SHALL 仅发送一次
- 请求结果 SHALL 缓存 5 分钟（可配置）
- 搜索请求 SHALL 去抖 300ms
- 页面切换时 SHALL 取消未完成的请求

#### Scenario: 请求缓存

**WHEN** 用户访问服务详情页
**THEN** 系统 SHALL 发送请求获取服务信息
**AND** 请求结果 SHALL 缓存
**AND** 5 分钟内再次访问 SHALL 使用缓存
**AND** 不需要重新请求

---

### REQ-8: 骨架屏

系统 SHALL 使用骨架屏替代 loading spinner。

**适用场景**:
- 页面加载
- 列表加载
- 卡片加载

**实现规则**:
- 骨架屏 SHALL 模拟实际内容的布局
- SHALL 显示渐变扫过动画
- 加载完成后 SHALL 平滑过渡到实际内容

#### Scenario: 页面骨架屏

**WHEN** 用户访问服务列表页
**THEN** SHALL 显示服务列表骨架屏
**AND** 骨架屏 SHALL 模拟服务卡片的布局
**AND** 数据加载完成后 SHALL 平滑过渡

---

### REQ-9: 打包优化

系统 SHALL 优化打包配置。

**优化措施**:
- SHALL 启用 Tree Shaking
- SHALL 启用代码压缩
- SHALL 启用 Gzip 压缩
- SHALL 分析打包体积

**性能目标**:
- 初始 bundle 大小 SHALL < 200KB (gzipped)
- 总 bundle 大小 SHALL < 1MB (gzipped)
- 第三方库 SHALL 单独打包

#### Scenario: 打包体积

**WHEN** 执行生产构建
**THEN** 初始 bundle SHALL < 200KB (gzipped)
**AND** 总 bundle SHALL < 1MB (gzipped)
**AND** SHALL 生成打包分析报告

---

### REQ-10: 性能监控

系统 SHALL 监控性能指标。

**监控指标**:
- 首屏加载时间
- 页面切换时间
- API 请求时间
- 内存使用
- 错误率

**实现规则**:
- SHALL 使用 Performance API
- SHALL 上报性能数据（可选）
- SHALL 在开发环境显示性能警告

#### Scenario: 性能监控

**WHEN** 页面加载完成
**THEN** 系统 SHALL 记录性能指标
**AND** 如果首屏加载时间 > 3秒，SHALL 显示警告
**AND** 性能数据 SHALL 上报到监控系统（可选）

---

### REQ-11: Lighthouse 评分

系统 SHALL 达到 Lighthouse 性能评分目标。

**评分目标**:
- Performance SHALL > 90
- Accessibility SHALL > 90
- Best Practices SHALL > 90
- SEO SHALL > 80

#### Scenario: Lighthouse 评分

**WHEN** 运行 Lighthouse 测试
**THEN** Performance 评分 SHALL > 90
**AND** Accessibility 评分 SHALL > 90
**AND** Best Practices 评分 SHALL > 90

---

### REQ-12: 响应式性能

系统 SHALL 在不同设备上保持良好性能。

**性能目标**:
- 桌面: 首屏 < 2秒
- 平板: 首屏 < 2.5秒
- 移动: 首屏 < 3秒

**优化措施**:
- SHALL 根据设备加载不同大小的资源
- SHALL 在移动端禁用非必要动画
- SHALL 在移动端简化布局

#### Scenario: 移动端性能

**WHEN** 用户在移动设备上访问系统
**THEN** 首屏 SHALL 在 3 秒内加载完成
**AND** SHALL 加载移动端优化的资源
**AND** 非必要动画 SHALL 禁用

---

## Implementation Notes

### 代码分割

```typescript
// 路由级代码分割
import { lazy } from 'react';

const Dashboard = lazy(() => import('./pages/Dashboard'));
const Services = lazy(() => import('./pages/Services'));
const Deployment = lazy(() => import('./pages/Deployment'));

// 组件懒加载
const MonacoEditor = lazy(() => import('@monaco-editor/react'));
```

### 虚拟滚动

```typescript
// 使用 react-window
import { FixedSizeList } from 'react-window';

const ServiceList = ({ services }) => {
  return (
    <FixedSizeList
      height={600}
      itemCount={services.length}
      itemSize={120}
      width="100%"
    >
      {({ index, style }) => (
        <div style={style}>
          <ServiceCard service={services[index]} />
        </div>
      )}
    </FixedSizeList>
  );
};
```

### 请求缓存

```typescript
// 使用 React Query
import { useQuery } from '@tanstack/react-query';

const useService = (id: string) => {
  return useQuery({
    queryKey: ['service', id],
    queryFn: () => Api.services.getById(id),
    staleTime: 5 * 60 * 1000, // 5 分钟
  });
};
```

### 性能监控

```typescript
// 使用 Performance API
const measurePageLoad = () => {
  const perfData = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;

  const metrics = {
    dns: perfData.domainLookupEnd - perfData.domainLookupStart,
    tcp: perfData.connectEnd - perfData.connectStart,
    request: perfData.responseStart - perfData.requestStart,
    response: perfData.responseEnd - perfData.responseStart,
    dom: perfData.domContentLoadedEventEnd - perfData.domContentLoadedEventStart,
    load: perfData.loadEventEnd - perfData.loadEventStart,
    total: perfData.loadEventEnd - perfData.fetchStart,
  };

  console.log('Performance Metrics:', metrics);

  // 上报到监控系统
  // reportMetrics(metrics);
};
```

---

## References

- [Web Vitals](https://web.dev/vitals/)
- [React Performance Optimization](https://react.dev/learn/render-and-commit)
- [Vite Performance](https://vitejs.dev/guide/performance.html)
- [react-window](https://react-window.vercel.app/)


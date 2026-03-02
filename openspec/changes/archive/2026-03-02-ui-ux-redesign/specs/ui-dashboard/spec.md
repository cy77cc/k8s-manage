# Spec: ui-dashboard

**Status**: Draft
**Created**: 2026-03-02

---

## Overview

本规范定义了 OpsPilot PaaS 平台的 Dashboard（主控台）页面设计。Dashboard 提供系统全局视图，包括关键指标、服务健康状态、最近部署和活跃告警等信息。

---

## Requirements

### REQ-1: 页面结构

Dashboard SHALL 包含页面标题、统计卡片、服务健康状态、最近活动和资源趋势。

**布局**:
- 页面标题 SHALL 在顶部
- 统计卡片 SHALL 在第一行，4 列布局
- 服务健康状态 SHALL 在第二行，全宽
- 最近部署和活跃告警 SHALL 在第三行，2 列布局
- 资源使用趋势 SHALL 在第四行，全宽

#### Scenario: Dashboard 页面加载

**WHEN** 用户访问 Dashboard
**THEN** SHALL 显示页面标题"主控台"
**AND** SHALL 显示 4 个统计卡片
**AND** SHALL 显示服务健康状态列表
**AND** SHALL 显示最近部署和活跃告警
**AND** SHALL 显示资源使用趋势图表

---

### REQ-2: 统计卡片

Dashboard SHALL 显示 4 个关键指标的统计卡片。

**指标**:
1. 服务总数
2. 运行中服务数
3. 今日部署次数
4. 活跃告警数

**卡片样式**:
- 背景色 SHALL 为白色
- 圆角 SHALL 为 12px
- 内边距 SHALL 为 24px
- 阴影 SHALL 为 md
- 卡片间距 SHALL 为 16px
- 悬停时 SHALL 向上移动 2px
- 悬停时阴影 SHALL 变为 lg

**卡片内容**:
- 图标尺寸 SHALL 为 48px
- 图标背景 SHALL 为渐变色（Indigo 50 到 Indigo 100）
- 图标颜色 SHALL 为 Indigo 500
- 图标圆角 SHALL 为 12px
- 图标与标题间距 SHALL 为 16px
- 标题字体 SHALL 为 13px, Semibold (600)
- 标题颜色 SHALL 为 Gray 500
- 标题 SHALL 大写，字间距 0.5px
- 数值字体 SHALL 为 32px, Bold (700)
- 数值颜色 SHALL 为 Gray 900
- 副标题字体 SHALL 为 13px
- 副标题颜色 SHALL 为 Success 色（正向趋势）或 Error 色（负向趋势）

#### Scenario: 统计卡片显示

**WHEN** 用户查看统计卡片
**THEN** "服务总数"卡片 SHALL 显示当前服务总数
**AND** SHALL 显示与上周相比的增长百分比
**AND** 增长百分比 SHALL 使用 Success 色（正增长）或 Error 色（负增长）
**AND** 悬停时卡片 SHALL 向上移动 2px

---

### REQ-3: 服务健康状态

Dashboard SHALL 显示服务健康状态列表，展示前 5 个服务的实时指标。

**列表样式**:
- 背景色 SHALL 为白色
- 圆角 SHALL 为 12px
- 内边距 SHALL 为 24px
- 阴影 SHALL 为 md

**列表项**:
- 内边距 SHALL 为 16px
- 圆角 SHALL 为 8px
- 悬停背景色 SHALL 为 Gray 100
- 列表项间距 SHALL 为 8px

**列表项内容**:
- 状态图标尺寸 SHALL 为 8px
- 状态图标为圆形
- 健康状态 SHALL 为 Success 色
- 警告状态 SHALL 为 Warning 色
- 错误状态 SHALL 为 Error 色
- 服务名称字体 SHALL 为 14px, Medium (500)
- 服务名称颜色 SHALL 为 Gray 900
- 指标字体 SHALL 为 13px
- 指标标签颜色 SHALL 为 Gray 500
- 指标值颜色 SHALL 为 Gray 900
- 指标值字重 SHALL 为 Medium (500)
- 操作按钮 SHALL 在悬停时显示

#### Scenario: 服务健康状态列表

**WHEN** 用户查看服务健康状态
**THEN** SHALL 显示前 5 个服务
**AND** 每个服务 SHALL 显示状态图标、名称、CPU 使用率、内存使用率
**AND** 悬停时 SHALL 显示"查看详情"按钮
**AND** 点击服务 SHALL 跳转到服务详情页

---

### REQ-4: 最近部署

Dashboard SHALL 显示最近 3 次部署记录。

**卡片样式**:
- 背景色 SHALL 为白色
- 圆角 SHALL 为 12px
- 内边距 SHALL 为 24px
- 阴影 SHALL 为 md

**部署记录**:
- 记录间距 SHALL 为 16px
- 状态图标尺寸 SHALL 为 8px
- 成功状态 SHALL 为 Success 色
- 失败状态 SHALL 为 Error 色
- 版本号字体 SHALL 为 14px, Medium (500)
- 版本号颜色 SHALL 为 Gray 900
- 服务名称字体 SHALL 为 13px
- 服务名称颜色 SHALL 为 Gray 500
- 时间字体 SHALL 为 13px
- 时间颜色 SHALL 为 Gray 500

#### Scenario: 最近部署列表

**WHEN** 用户查看最近部署
**THEN** SHALL 显示最近 3 次部署
**AND** 每条记录 SHALL 显示状态、版本号、服务名称、时间、结果
**AND** 点击记录 SHALL 跳转到部署详情

---

### REQ-5: 活跃告警

Dashboard SHALL 显示最近 3 条活跃告警。

**卡片样式**:
- 背景色 SHALL 为白色
- 圆角 SHALL 为 12px
- 内边距 SHALL 为 24px
- 阴影 SHALL 为 md

**告警记录**:
- 记录间距 SHALL 为 16px
- 严重程度图标尺寸 SHALL 为 16px
- Critical SHALL 为 Error 色
- Warning SHALL 为 Warning 色
- Info SHALL 为 Info 色
- 告警标题字体 SHALL 为 14px, Medium (500)
- 告警标题颜色 SHALL 为 Gray 900
- 来源字体 SHALL 为 13px
- 来源颜色 SHALL 为 Gray 500
- 时间字体 SHALL 为 13px
- 时间颜色 SHALL 为 Gray 500

#### Scenario: 活跃告警列表

**WHEN** 用户查看活跃告警
**THEN** SHALL 显示最近 3 条告警
**AND** 每条告警 SHALL 显示严重程度、标题、来源、时间
**AND** 点击告警 SHALL 跳转到告警详情

---

### REQ-6: 资源使用趋势

Dashboard SHALL 显示最近 24 小时的资源使用趋势图表。

**卡片样式**:
- 背景色 SHALL 为白色
- 圆角 SHALL 为 12px
- 内边距 SHALL 为 24px
- 阴影 SHALL 为 md

**图表**:
- 图表类型 SHALL 为折线图
- SHALL 显示 CPU、内存、网络三条曲线
- CPU 曲线颜色 SHALL 为 Indigo 500
- 内存曲线颜色 SHALL 为 Success 色
- 网络曲线颜色 SHALL 为 Warning 色
- 曲线宽度 SHALL 为 2px
- 曲线 SHALL 平滑
- X 轴 SHALL 显示时间（24 小时）
- Y 轴 SHALL 显示百分比（0-100%）
- SHALL 显示图例
- 悬停时 SHALL 显示 Tooltip

#### Scenario: 资源使用趋势图表

**WHEN** 用户查看资源使用趋势
**THEN** SHALL 显示最近 24 小时的折线图
**AND** SHALL 显示 CPU、内存、网络三条曲线
**AND** 悬停在图表上 SHALL 显示具体数值
**AND** 图例 SHALL 可点击切换显示/隐藏曲线

---

### REQ-7: 刷新功能

Dashboard SHALL 提供手动刷新功能。

**刷新按钮**:
- SHALL 位于页面标题右侧
- 图标 SHALL 为刷新图标
- 点击时 SHALL 重新加载所有数据
- 加载时 SHALL 显示 loading 状态
- 加载时图标 SHALL 旋转

**自动刷新**:
- Dashboard SHALL 每 30 秒自动刷新一次
- 用户离开页面时 SHALL 停止自动刷新
- 用户返回页面时 SHALL 恢复自动刷新

#### Scenario: 手动刷新

**WHEN** 用户点击刷新按钮
**THEN** 按钮 SHALL 显示 loading 状态
**AND** 刷新图标 SHALL 旋转
**AND** SHALL 重新加载所有数据
**AND** 加载完成后 SHALL 恢复正常状态

---

### REQ-8: 空状态

Dashboard SHALL 在无数据时显示空状态。

**空状态样式**:
- SHALL 显示在对应区域中央
- 图标尺寸 SHALL 为 64px
- 图标颜色 SHALL 为 Gray 300
- 提示文字字体 SHALL 为 14px
- 提示文字颜色 SHALL 为 Gray 500

#### Scenario: 无服务时的空状态

**WHEN** 系统中没有任何服务
**THEN** 服务健康状态区域 SHALL 显示空状态
**AND** SHALL 显示"暂无服务"提示
**AND** SHALL 显示"创建第一个服务"按钮

---

### REQ-9: 响应式适配

Dashboard SHALL 支持响应式布局。

**桌面 (>= 1024px)**:
- 统计卡片 SHALL 为 4 列布局
- 最近部署和活跃告警 SHALL 为 2 列布局

**平板 (768px - 1023px)**:
- 统计卡片 SHALL 为 2 列布局
- 最近部署和活跃告警 SHALL 为 1 列布局

**移动 (< 768px)**:
- 统计卡片 SHALL 为 1 列布局
- 最近部署和活跃告警 SHALL 为 1 列布局
- 资源使用趋势图表 SHALL 适配小屏幕

#### Scenario: 移动端 Dashboard

**WHEN** 用户在移动设备上访问 Dashboard
**THEN** 统计卡片 SHALL 垂直排列
**AND** 每个卡片 SHALL 占据全宽
**AND** 服务健康状态列表 SHALL 适配小屏幕
**AND** 图表 SHALL 适配小屏幕，支持横向滚动

---

## Implementation Notes

### 数据加载

```typescript
// Dashboard 数据加载
const Dashboard = () => {
  const [loading, setLoading] = useState(true);
  const [data, setData] = useState<DashboardData>();

  const loadData = async () => {
    setLoading(true);
    try {
      const [services, deployments, alerts, metrics] = await Promise.all([
        Api.services.getList(),
        Api.deployment.getRecent(),
        Api.monitoring.getAlerts(),
        Api.monitoring.getMetrics(),
      ]);
      setData({ services, deployments, alerts, metrics });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
    const interval = setInterval(loadData, 30000); // 30秒自动刷新
    return () => clearInterval(interval);
  }, []);
};
```

### 图表库

使用 Ant Design Charts 或 Recharts 实现资源使用趋势图表。

---

## References

- [Ant Design Charts](https://charts.ant.design/)
- [Recharts](https://recharts.org/)


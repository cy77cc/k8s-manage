# Spec: ui-service-management

**Status**: Draft
**Created**: 2026-03-02

---

## Overview

本规范定义了 OpsPilot PaaS 平台的服务管理页面，包括服务列表和服务详情。服务管理是 PaaS 平台的核心功能，提供服务的创建、查看、编辑、部署、扩缩容等操作。

---

## Requirements

### REQ-1: 服务列表页面结构

服务列表页 SHALL 包含筛选栏、搜索框、服务卡片列表和分页。

**布局**:
- 筛选栏 SHALL 在顶部
- 服务卡片 SHALL 使用卡片式布局（非表格）
- 卡片间距 SHALL 为 16px
- 分页 SHALL 在底部

#### Scenario: 服务列表页面加载

**WHEN** 用户访问服务列表页
**THEN** SHALL 显示页面标题"服务管理"
**AND** SHALL 显示筛选栏和搜索框
**AND** SHALL 显示服务卡片列表
**AND** SHALL 显示"创建服务"按钮

---

### REQ-2: 筛选和搜索

服务列表 SHALL 提供筛选和搜索功能。

**筛选项**:
- 状态筛选（全部、运行中、已停止、降级、错误）
- 环境筛选（全部、开发、测试、预发布、生产）
- 项目筛选（如果有多项目）

**搜索**:
- 搜索框宽度 SHALL 为 280px
- SHALL 支持按服务名称搜索
- SHALL 支持实时搜索（去抖 300ms）
- 搜索框 SHALL 有搜索图标

**筛选栏样式**:
- 背景色 SHALL 为白色
- 圆角 SHALL 为 12px
- 内边距 SHALL 为 16px 24px
- 阴影 SHALL 为 md
- 筛选项间距 SHALL 为 16px

#### Scenario: 筛选服务

**WHEN** 用户选择"运行中"状态筛选
**THEN** 列表 SHALL 仅显示运行中的服务
**AND** 筛选按钮 SHALL 显示激活状态
**AND** URL SHALL 更新为包含筛选参数

---

### REQ-3: 服务卡片

服务列表 SHALL 使用卡片式布局展示服务。

**卡片样式**:
- 背景色 SHALL 为白色
- 圆角 SHALL 为 12px
- 内边距 SHALL 为 20px
- 阴影 SHALL 为 md
- 悬停时边框色 SHALL 变为 Indigo 200
- 悬停时阴影 SHALL 变为 lg
- 悬停时 SHALL 向上移动 2px
- 过渡动画 SHALL 为 250ms

**卡片内容**:
- 顶部：状态图标 + 服务名称 + 快捷操作
- 中部：环境、版本、实例数等元数据
- 底部：CPU、内存等实时指标

**服务名称**:
- 字体 SHALL 为 16px, Semibold (600)
- 颜色 SHALL 为 Gray 900
- SHALL 与状态图标在同一行

**状态图标**:
- 尺寸 SHALL 为 8px
- 为圆形
- 运行中 SHALL 为 Success 色
- 已停止 SHALL 为 Gray 400
- 降级 SHALL 为 Warning 色
- 错误 SHALL 为 Error 色

**元数据**:
- 字体 SHALL 为 13px
- 标签颜色 SHALL 为 Gray 500
- 值颜色 SHALL 为 Gray 900
- 项间距 SHALL 为 16px
- SHALL 使用图标 + 文字形式

**指标**:
- SHALL 显示在底部
- 上方 SHALL 有 1px 分隔线
- 指标间距 SHALL 为 24px
- 标签字体 SHALL 为 13px
- 标签颜色 SHALL 为 Gray 500
- 值字体 SHALL 为 13px, Medium (500)
- 值颜色 SHALL 为 Gray 900

**操作按钮**:
- SHALL 在悬停时显示
- 按钮尺寸 SHALL 为 32px
- 按钮间距 SHALL 为 8px
- SHALL 包含：查看、部署、扩容、日志、更多

#### Scenario: 服务卡片显示

**WHEN** 用户查看服务列表
**THEN** 每个服务 SHALL 显示为卡片
**AND** 卡片 SHALL 显示状态、名称、环境、版本、实例数
**AND** 卡片底部 SHALL 显示 CPU、内存使用率
**AND** 悬停时 SHALL 显示操作按钮

---

### REQ-4: 批量操作

服务列表 SHALL 支持批量操作。

**选择**:
- 每个卡片左上角 SHALL 有复选框
- 复选框 SHALL 在悬停或选中时显示
- 顶部 SHALL 有全选复选框

**批量操作栏**:
- 选中服务后 SHALL 显示批量操作栏
- 操作栏 SHALL 固定在底部
- 操作栏背景色 SHALL 为 Indigo 500
- 操作栏文字色 SHALL 为白色
- SHALL 显示已选中数量
- SHALL 提供批量启动、停止、删除等操作

#### Scenario: 批量删除服务

**WHEN** 用户选中 3 个服务并点击批量删除
**THEN** SHALL 显示确认对话框
**AND** 对话框 SHALL 列出将要删除的服务
**AND** 确认后 SHALL 执行删除操作
**AND** 删除成功后 SHALL 显示成功通知

---

### REQ-5: 服务详情页面结构

服务详情页 SHALL 包含页面头部、Tab 导航和内容区。

**页面头部**:
- SHALL 包含返回按钮、服务名称、状态、版本、环境、实例数
- SHALL 包含快捷操作按钮（部署、扩容、回滚、重启、日志、配置等）
- 背景色 SHALL 为白色
- 圆角 SHALL 为 12px
- 内边距 SHALL 为 24px
- 阴影 SHALL 为 md

**Tab 导航**:
- SHALL 包含：概览、配置、日志、监控、事件
- Tab 样式参考设计系统
- 激活 Tab SHALL 有底部色条（Indigo 500）

#### Scenario: 服务详情页面加载

**WHEN** 用户点击服务卡片
**THEN** SHALL 跳转到n**AND** SHALL 显示服务名称和状态
**AND** SHALL 显示快捷操作按钮
**AND** SHALL 默认显示"概览" Tab

---

### REQ-6: 概览 Tab

概览 Tab SHALL 显示服务的关键信息和实时指标。

**实时指标**:
- SHALL 显示 6 个指标卡片：CPU、内存、网络、请求数、错误率、响应时间
- 卡片布局 SHALL 为 3 列
- 每个卡片 SHALL 包含：标题、当前值、迷你图表
- 迷你图表 SHALL 显示最近 1 小时的趋势

**实例列表**:
- SHALL 显示所有实例
- 每个实例 SHALL 显示：名称、状态、CPU、内存、启动时间
- SHALL 提供查看日志、进入终端等操作

**部署历史**:
- SHALL 显示最近 5 次部署
- 每条记录 SHALL 显示：版本、时间、操作人、结果
- SHALL 提供查看详情、回滚等操作

#### Scenario: 查看服务概览

**WHEN** 用户查看服务概览
**THEN** SHALL 显示 6 个实时指标卡片
**AND** SHALL 显示实例列表
**AND** SHALL 显示部署历史
**AND** 指10 秒自动刷新

---

### REQ-7: 配置 Tab

配置 Tab SHALL 显示和编辑服务配置。

**配置项**:
- 基本信息（名称、描述、标签）
- 镜像配置（镜像地址、版本、拉取策略）
- 资源配置（CPU、内存限制和请求）
- 环境变量
- 端口配置
- 健康检查
- 持久化存储

**编辑模式**:
- SHALL 提供"编辑"按钮
- 点击编辑 SHALL 进入编辑模式
- 编辑模式 SHALL 显示表单
- SHALL 提供"保存"和"取消"按钮
- 保存后 SHALL 显示确认对话框（是否立即部署）

#### Scenario: 编辑服务配置

**WHEN** 用户点击"编辑"按钮
**THEN** SHALL 进入编辑模式
**AND** SHALL 显示配置表单
**AND** 用户修改配置后点击"保存"
**THEN** SHALL 显示确认对话框
**AND** 用户确认后 SHALL 保存配置并触发部署

---

### REQ-8: 日志 Tab

日志 Tab SHALL 显示服务日志。

**日志查看器**:
- SHALL 使用等宽字体
- 背景色 SHALL 为 Gray 900
- 文字色 SHALL 为 Gray 100
- 行高 SHALL 为 1.5
- SHALL 支持虚拟滚动（大量日志）
- SHALL 支持日志级别高亮（ERROR 红色、WARN 黄色、INFO 蓝色）

**日志筛选**:
- SHALL 支持按实例筛选
- SHALL 支持按日志级别筛选
- SHALL 支持按时间范围筛选
- SHALL 支持关键词搜索

**日志操作**:
- SHALL 提供"实时日志"开关
- SHALL 提供"下载日志"按钮
- SHALL 提供"清空"按钮

#### Scenario: 查看实时日志

**WHEN** 用户打开"实时日志"开关
**THEN** SHALL 建立 WebSocket 连接
**AND** SHALL 实时显示新日志
**AND** SHALL 自动滚动到底部
**AND** 关闭开关时 SHALL 断开连接

---

### REQ-9: 监控 Tab

监控 Tab SHALL 显示服务监控图表。

**图表**:
- SHALL 显示 CPU、内存、网络、请求数、错误率、响应时间等图表
- 图表类型 SHALL 为折线图或面积图
- SHALL 支持时间范围选择（1小时、6小时、24小时、7天、30天）
- SHALL 支持图表缩放和拖拽

**告警规则**:
- SHALL 显示已配置的告警规则
- SHALL 提供添加、编辑、删除告警规则

#### Scenario: 查看监控图表

**WHEN** 用户查看监控 Tab
**THEN** SHALL 显示多个监控图表
**AND** 默认时间范围 SHALL 为最近 1 小时
**AND** 用户可以切换时间范围
**AND** 图表 SHALL 根据时间范围更新

---

### REQ-10: 响应式适配

服务管理页面 SHALL 支持响应式布局。

**桌面 (>= 1024px)**:
- 服务卡片 SHALL 为 2 列布局
- 详情页指标卡片 SHALL 为 3 列布局

**平板 (768px - 1023px)**:
- 服务卡片 SHALL 为 1 列布局
- 详情页指标卡片 SHALL 为 2 列布局

**移动 (< 768px)**:
- 服务卡片 SHALL 为 1 列布局
- 详情页指标卡片 SHALL 为 1 列布局
- 操作按钮 SHALL 简化为下拉菜单

#### Scenario: 移动端服务列表

**WHEN** 用户在移动设备上查看服务列表
**THEN** 服务卡片 SHALL 垂直排列
**AND** 每个卡片 SHALL 占据全宽
**AND** 操作按钮 SHALL 合并为"更多"菜单

---

## Implementation Notes

### 服务卡片组件

```typescript
// ServiceCard.tsx
interface ServiceCardProps {
  service: Service;
  onView: (id: string) => void;
  onDeploy: (id: string) => void;
  onScale: (id: string) => void;
}

const ServiceCard: React.FC<ServiceCardProps> = ({ service, onView, onDeploy, onScale }) => {
  const [hovered, setHovered] = useState(false);

  return (
    <Card
      hoverable
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      onClick={() => onView(service.id)}
    >
      {/* 卡片内容 */}
      {hovered && (
        <div className="actions">
          <Button icon={<EyeOutlined />} onClick={(e) => { e.stopPropagation(); onView(service.id); }} />
          <Button icon={<RocketOutlined />} onClick={(e) => { e.stopPropagation(); onDeploy(service.id); }} />
          <Button icon={<ExpandOutlined />} onClick={(e) => { e.stopPropagation(); onScale(service.id); }} />
        </div>
      )}
    </Card>
  );
};
```

---

## References

- [Ant Design Card](https://ant.design/components/card)
- [Ant Design Tabs](https://ant.design/components/tabs)


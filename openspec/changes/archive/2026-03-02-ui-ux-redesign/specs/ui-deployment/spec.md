# Spec: ui-deployment

**Status**: Draft
**Created**: 2026-03-02

---

## Overview

本规范定义了 OpsPilot PaaS 平台的部署管理页面，包括部署列表、部署流程和部署详情。部署管理是 PaaS 平台的核心功能，提供服务的版本部署、回滚、进度跟踪等操作。

---

## Requirements

### REQ-1: 部署列表页面结构

部署列表页 SHALL 包含筛选栏、部署记录列表和分页。

**布局**:
- 筛选栏 SHALL 在顶部
- 部署记录 SHALL 使用时间线布局
- 记录间距 SHALL 为 16px
- 分页 SHALL 在底部

#### Scenario: 部署列表页面加载

**WHEN** 用户访问部署列表页
**THEN** SHALL 显示页面标题"部署管理"
**AND** SHALL 显示筛选栏
**AND** SHALL 显示部署记录时间线
**AND** SHALL 显示"新建部署"按钮

---

### REQ-2: 筛选功能

部署列表 SHALL 提供筛选功能。

**筛选项**:
- 服务筛选（全部、或选择特定服务）
- 状态筛选（全部、进行中、成功、失败、已回滚）
- 环境筛选（全部、开发、测试、预发布、生产）
- 时间范围（今天、最近7天、最近30天、自定义）

**筛选栏样式**:
- 背景色 SHALL 为白色
- 圆角 SHALL 为 12px
- 内边距 SHALL 为 16px 24px
- 阴影 SHALL 为 md

#### Scenario: 筛选部署记录

**WHEN** 用户选择"失败"状态筛选
**THEN** 列表 SHALL 仅显示失败的部署
**AND** 筛选按钮 SHALL 显示激活状态

---

### REQ-3: 部署记录时间线

部署列表 SHALL 使用时间线布局展示部署记录。

**时间线样式**:
- 时间线颜色 SHALL 为 Gray 300
- 时间线宽度 SHALL 为 2px
- 节点尺寸 SHALL 为 12px
- 成功节点 SHALL 为 Success 色
- 失败节点 SHALL 为 Error 色
- 进行中节点 SHALL 为 Indigo 500，带动画

**记录卡片**:
- 背景色 SHALL 为白色
- 圆角 SHALL 为 12px
- 内边距 SHALL 为 20px
- 阴影 SHALL 为 md
- 悬停时阴影 SHALL 变为 lg

**卡片内容**:
- 顶部：服务名称 + 版本号 + 状态标签
- 中部：环境、操作人、开始时间、持续时间
- 底部：操作按钮（查看详情、查看日志、回滚等）

**状态标签**:
- 进行中 SHALL 为 Indigo 色系
- 成功 SHALL 为 Success 色系
- 失败 SHALL 为 Error 色系
- 已回滚 SHALL 为 Warning 色系

#### Scenario: 部署记录显示

**WHEN** 用户查看部署列表
**THEN** 记录 SHALL 按时间倒序排列
**AND** 每条记录 SHALL 显示服务、版本、状态、时间等信息
**AND** 进行中的部署 SHALL 显示进度条
**AND** 点击记录 SHALL 展开详情

---

### REQ-4: 部署流程 - 步骤式界面

新建部署 SHALL 使用步骤式界面。

**步骤**:
1. 选择服务和环境
2. 选择版本
3. 配置参数
4. 确认部署

**步骤指示器**:
- SHALL 显示在顶部
- 当前步骤 SHALL 高亮（Indigo 500）
- 已完成步骤 SHALL 显示勾选图标
- 未完成步骤 SHALL 为灰色
- 步骤间 SHALL 有连接线

**步骤内容区**:
- 背景色 SHALL 为白色
- 圆角 SHALL 为 12px
- 内边距 SHALL 为 32px
- 阴影 SHALL 为 md

**底部操作栏**:
- SHALL 固定在底部
- 背景色 SHALL 为白色
- 顶部边框 SHALL 为 1px solid #e9ecef
- SHALL 包含"上一步"、"下一步"、"取消"按钮

#### Scenario: 步骤式部署流程

**WHEN** 用户点击"新建部署"
**THEN** SHALL 显示步骤 1：选择服务和环境
**AND** 用户选择后点击"下一步"
**THEN** SHALL 进入步骤 2：选择版本
**AND** 步骤指示器 SHALL 更新
**AND** 用户可以点击"上一步"返回

---

### REQ-5: 步骤 1 - 选择服务和环境

步骤 1 SHALL 允许用户选择要部署的服务和目标环境。

**服务选择**:
- SHALL 显示服务列表（卡片或下拉）
- 每个服务 SHALL 显示名称、当前版本、状态
- 选中的服务 SHALL 高亮显示

**环境选择**:
- SHALL 显示环境选项（开发、测试、预发布、生产）
- 环境 SHALL 使用单选按钮
- 生产环境 SHALL 有警告提示

#### Scenario: 选择服务和环境

**WHEN** 用户在步骤 1 选择服务"api-gateway"和环境"生产"
**THEN** 服务卡片 SHALL 高亮显示
**AND** 生产环境 SHALL 显示警告提示
**AND** "下一步"按钮 SHALL 可用

---

### REQ-6: 步骤 2 - 选择版本

步骤 2 SHALL 允许用户选择要部署的版本。

**版本列表**:
- SHALL 显示最近 10 个版本
- 每个版本 SHALL 显示：版本号、提交信息、作者、时间
- 最新版本 SHALL 有"最新"标签
- 当前运行版本 SHALL 有"当前"标签
- 版本 SHALL 使用单选按钮

**版本详情**:
- 选中版本后 SHALL 显示详细信息
- SHALL 显示完整的提交信息
- SHALL 显示文件变更列表（可选）

#### Scenario: 选择版本

**WHEN** 用户选择版本"v1.2.3"
**THEN** 版本 SHALL 高亮显示
**AND** SHALL 显示该版本的详细信息
**AND** "下一步"按钮 SHALL 可用

---

### REQ-7: 步骤 3 - 配置参数

步骤 3 SHALL 允许用户配置部署参数。

**配置项**:
- 实例数量（滑块或输入框）
- 资源限制（CPU、内存）
- 环境变量（可选）
- 部署策略（滚动更新、蓝绿部署、金丝雀）
- 健康检查配置

**表单样式**:
- SHALL 使用垂直表单布局
- Label SHALL 清晰明确
- SHALL 提供默认值
- SHALL 提供帮助文本

#### Scenario: 配置部署参数

**WHEN** 用户配置实例数量为 3
**THEN** 滑块 SHALL 更新到 3
**AND** SHALL 显示预估资源使用量
**AND** "下一步"按钮 SHALL 可用

---

### REQ-8: 步骤 4 - 确认部署

步骤 4 SHALL 显示部署摘要并确认。

**摘要内容**:
- 服务名称和环境
- 版本信息
- 配置参数
- 预估影响（实例数变化、资源变化）

**确认操作**:
- SHALL 显示"确认部署"按钮
- 按钮 SHALL 为 Primary 样式
- 点击后 SHALL 显示二次确认对话框（生产环境）
- 确认后 SHALL 开始部署并跳转到部署进度页

#### Scenario: 确认部署

**WHEN** 用户在步骤 4 点击"确认部署"
**THEN** SHALL 显示部署摘要
**AND** 用户确认后 SHALL 开始部署
**AND** SHALL 跳转到部署进度页

---

### REQ-9: 部署进度页

部署进行时 SHALL 显示实时进度。

**进度显示**:
- SHALL 显示进度条（0-100%）
- SHALL 显示当前步骤
- SHALL 显示步骤列表（拉取镜像、创建容器、启动服务、健康检查、切换流量）
- 已完成步骤 SHALL 显示勾选图标
- 进行中步骤 SHALL 显示加载图标
- 未开始步骤 SHALL 为灰色

**实时日志**:
- SHALL 显示部署日志
- 日志 SHALL 实时更新
- 日志 SHALL 使用等宽字体
- 日志 SHALL 支持滚动

**操作**:
- SHALL 提供"取消部署"按钮
- 取消时 SHALL 显示确认对话框
- 部署完成后 SHALL 显示"查看服务"按钮

#### Scenario: 查看部署进度

**WHEN** 部署正在进行
**THEN** SHALL 显示实时进度
**AND** 进度条 SHALL 更新
**AND** 当前步骤 SHALL 高亮显示
**AND** 日志 SHALL 实时滚动

---

### REQ-10: 部署详情页

部署详情页 SHALL 显示部署的完整信息。

**基本信息**:
- 服务名称、版本、环境
- 状态、开始时间、结束时间、持续时间
- 操作人

**部署步骤**:
- SHALL 显示所有步骤及其状态
- 每个步骤 SHALL 显示开始时间和持续时间
- 失败步骤 SHALL 显示错误信息

**部署日志**:
- SHALL 显示完整的部署日志
- SHALL 支持下载日志

**操作**:
- 成功的部署 SHALL 提供"回滚"按钮
- 失败的部署 SHALL 提供"重试"按钮

#### Scenario: 查看部署详情

**WHEN** 用户点击部署记录
**THEN** SHALL 显示部署详情页
**AND** SHALL 显示基本信息、步骤、日志
**AND** 如果部署成功，SHALL 显示"回滚"按钮

---

### REQ-11: 回滚功能

系统 SHALL 支持一键回滚到之前的版本。

**回滚操作**:
- SHALL 显示确认对话框
- 对话框 SHALL 说明将回滚到哪个版本
- 确认后 SHALL 创建新的部署记录
- 回滚 SHALL 使用相同的部署流程

#### Scenario: 回滚部署

**WHEN** 用户点击"回滚"按钮
**THEN** SHALL 显示确认对话框
**AND** 对话框 SHALL 显示目标版本信息
**AND** 用户确认后 SHALL 开始回滚
**AND** SHALL 创建新的部署记录（类型为"回滚"）

---

### REQ-12: 响应式适配

部署管理页面 SHALL 支持响应式布局。

**移动端**:
- 时间线 SHALL 简化为垂直列表
- 步骤指示器 SHALL 简化为进度条
- 表单 SHALL 适配小屏幕

#### Scenario: 移动端部署列表

**WHEN** 用户在移动设备上查看部署列表
**THEN** 记录 SHALL 垂直排列
**AND** 时间线 SHALL 简化
**AND** 操作按钮 SHALL 合并为下拉菜单

---

## Implementation Notes

### 部署状态机

```typescript
enum DeploymentStatus {
  PENDING = 'pending',
  PULLING = 'pulling',
  CREATING = 'creating',
  STARTING = 'starting',
  HEALTH_CHECK = 'health_check',
  SWITCHING = 'switching',
  SUCCESS = 'success',
  FAILED = 'failed',
  CANCELLED = 'cancelled',
  ROLLED_BACK = 'rolled_back',
}
```

### WebSocket 实时更新

```typescript
// 部署进度实时更新
const useDeploymentProgress = (deploymentId: string) => {
  const [progress, setProgress] = useState<DeploymentProgress>();

  useEffect(() => {
    const ws = new WebSocket(`/api/v1/deployment/${deploymentId}/progress`);

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      setProgress(data);
    };

    return () => ws.close();
  }, [deploymentId]);

  return progress;
};
```

---

## References

- [Ant Design Steps](https://ant.design/components/steps)
- [Ant Design Timeline](https://ant.design/components/timeline)


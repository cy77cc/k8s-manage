# 组件库使用文档

## 概述

本文档介绍了 OpsPilot 平台重构后的组件库使用方法。所有组件都基于 Ant Design，并通过统一的主题配置实现了一致的视觉风格。

## AI Chat V2

当前 AI 入口已经统一到 `web/src/pages/AIChat/`。旧的 `web/src/components/AI/*` 组件和旧命令中心页面已移除，不再作为可复用组件来源。

### 页面结构

- `ChatPage`
- `ConversationSidebar`
- `ChatMain`
- `useChatSession`
- `useSSEConnection`
- `useConfirmation`
- `useAIChatShortcuts`

### 设计约束

- AI 页面是页面级组合，不再提供旧的全局悬浮助手组件
- `/ai` 路由直接渲染 `AIChatPage`
- 工具轨迹、审批卡片、推荐动作统一在 `ChatMain` 内展示
- 会话数据通过 `/api/v1/ai/sessions*` 和 SSE `done.session` 同步

### 使用方式

```tsx
import AIChatPage from '../src/pages/AIChat/ChatPage';

export default function AIPageRoute() {
  return <AIChatPage />;
}
```

### 关键类型

AI 页面内部类型集中在 `web/src/pages/AIChat/types.ts`：

- `AIChatSession`
- `AIChatMessage`
- `AIChatToolTrace`
- `AIChatAskRequest`
- `AIChatDonePayload`

### 不再使用的旧组件

以下组件已从代码库移除，不应继续引用：

- `ChatInterface`
- `GlobalAIAssistant`
- `CommandPanel`
- `AnalysisPanel`
- `RecommendationPanel`

## 设计原则

- **一致性**: 所有组件遵循统一的设计规范
- **简洁性**: 界面简洁清晰，避免过度装饰
- **可访问性**: 支持键盘导航和屏幕阅读器
- **响应式**: 适配不同屏幕尺寸

## 基础组件

### Button 按钮

用于触发操作的交互元素。

**样式规范**
- 圆角: 8px
- 高度: 40px (default), 48px (large), 32px (small)
- 字重: 500 (medium)

**使用示例**

```tsx
import { Button } from 'antd';
import { PlusOutlined } from '@ant-design/icons';

// 主按钮
<Button type="primary">主要操作</Button>

// 带图标的按钮
<Button type="primary" icon={<PlusOutlined />}>新建</Button>

// 危险操作按钮
<Button danger>删除</Button>

// 不同尺寸
<Button size="large">大按钮</Button>
<Button>默认按钮</Button>
<Button size="small">小按钮</Button>
```

### Input 输入框

用于接收用户输入的文本。

**样式规范**
- 圆角: 8px
- 高度: 40px (default), 48px (large), 32px (small)
- Focus: 2px outline in primary-500

**使用示例**

```tsx
import { Input, Form } from 'antd';
import { SearchOutlined } from '@ant-design/icons';

// 基础输入框
<Input placeholder="请输入内容" />

// 带图标
<Input prefix={<SearchOutlined />} placeholder="搜索" />

// 表单中使用（带验证）
<Form.Item
  label="用户名"
  name="username"
  rules={[{ required: true, message: '请输入用户名' }]}
>
  <Input />
</Form.Item>
```

### Card 卡片

用于组织和展示内容的容器。

**样式规范**
- 圆角: 12px
- 阴影: 0 1px 3px rgba(0,0,0,0.08)
- 内边距: 24px

**使用示例**

```tsx
import { Card, Button } from 'antd';

// 基础卡片
<Card>
  <p>卡片内容</p>
</Card>

// 带标题和操作
<Card
  title="卡片标题"
  extra={<Button type="link">更多</Button>}
>
  <p>卡片内容</p>
</Card>

// 可悬停卡片
<Card hoverable>
  <p>鼠标悬停有提升效果</p>
</Card>
```

### Tag 标签

用于标记和分类。

**样式规范**
- 圆角: 4px
- 语义色: success, warning, error, processing

**使用示例**

```tsx
import { Tag } from 'antd';
import { CheckCircleOutlined } from '@ant-design/icons';

// 语义色标签
<Tag color="success">成功</Tag>
<Tag color="warning">警告</Tag>
<Tag color="error">错误</Tag>
<Tag color="processing">进行中</Tag>

// 带图标
<Tag color="success" icon={<CheckCircleOutlined />}>
  运行中
</Tag>

// 可关闭
<Tag closable onClose={() => console.log('关闭')}>
  可关闭标签
</Tag>
```

### Modal 对话框

用于显示重要信息或需要用户确认的操作。

**样式规范**
- 圆角: 12px
- 阴影: 0 20px 25px -5px rgba(0,0,0,0.1)
- 遮罩: rgba(0,0,0,0.45)

**使用示例**

```tsx
import { Modal, Button } from 'antd';
import { useState } from 'react';

const [visible, setVisible] = useState(false);

// 基础对话框
<Modal
  title="对话框标题"
  open={visible}
  onOk={() => setVisible(false)}
  onCancel={() => setVisible(false)}
>
  <p>对话框内容</p>
</Modal>

// 确认对话框
<Button onClick={(odal.confirm({
    title: '确认删除',
    content: '删除后无法恢复，确定要删除吗？',
    onOk: () => console.log('确认'),
  });
}}>
  删除
</Button>

// 信息提示
Modal.info({
  title: '提示',
  content: '这是一条信息'
});
```

## 数据展示组件

### Table 表格

用于展示结构化数据。

**样式规范**
- 行高: 单元格内边距 16px
- Hover: #f8f9fa
- 表头背景: #f8f9fa
- 选中行: #eef2ff

**使用示例**

```tsx
import { Table, Button, Space, Tooltip } from 'antd';
import { EditOutlined, DeleteOutlined } from '@ant-design/icons';

const columns = [
  {
    title: '名称',
    dataIndex: 'name',
    key: 'name',
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    render: (status) => (
      <Tag color={status === 'active' ? 'success' : 'default'}>
        {status}
      </Tag>
    ),
  },
  {
    title: '操作',
    key: 'action',
    render: (_, record) => (
      <Space size="small">
        <Tooltip title="编辑">
          <Button type="text" icon={<EditOutlined />} />
        </Tooltip>
        <Tooltip title="删除">
          <Button type="text" danger icon={<DeleteOutlined />} />
        </Tooltip>
      </Space>
    ),
  },
];

<Table
  columns={columns}
  dataSource={data}
  rowKey="id"
  pagination={{ pageSize: 10 }}
/>
```

### List 列表

用于展示连续的内容。

**使用示例**

```tsx
import { List, Avatar, Button } from 'antd';
import { UserOutlined } from '@ant-design/icons';

<List
  dataSource={data}
  renderItem={(item) => (
    <List.Item
      actions={[
        <Button type="link">查看</Button>,
        <Button type="link">编辑</Button>,
      ]}
    >
      <List.Item.Meta
        avatar={<Avatar icon={<UserOutlined />} />}
        title={item.title}
        description={item.description}
      />
    </List.Item>
  )}
/>
```

### Descriptions 描述列表

用于展示详细信息。

**使用示例**

```tsx
import { Descriptions, Tag } from 'antd';

<Descriptions bordered column={2}>
  <Descriptions.Item label="服务名称">
    user-service
  </Descriptions.Item>
  <Descriptions.Item label="状态">
    <Tag color="success">运行中</Tag>
  </Descriptions.Item>
  <Descriptions.Item label="实例数">3</Descriptions.Item>
  <Descriptions.Item label="CPU 使用率">45%</Descriptions.Item>
</Descriptions>
```

### Empty 空状态

用于展示无数据状态。

**使用示例**

```tsx
import { Empty, Button } from 'antd';

// 默认空状态
<Empty description="暂无数据" />

// 自定义空状态
<Empty description="还没有创建任何服务">
  <Button type="primary">创建服务</Button>
</Empty>

// 简单空状态
<Empty image={Empty.PRESENTED_IMAGE_SIMPLE} />
```

### Skeleton 骨架屏

用于展示加载状态。

**使用示例**

```tsx
import { Skeleton, Card } from 'antd';

// 基础骨架屏
<Skeleton active />

// 带头像的骨架屏
<Skeleton active avatar paragraph={{ rows: 2 }} />

// 包裹内容
<Skeleton loading={loading} active>
  <Card>实际内容</Card>
</Skeleton>
```

## 反馈组件

### Notification 通知

用于显示全局通知消息。

**样式规范**
- 宽度: 384px
- 圆角: 8px
- 位置: topRight

**使用示例**

```tsx
import { notification, Button } from 'antd';

const [api, contextHolder] = notification.useNotification();

// 成功通知
api.success({
  message: '操作成功',
  description: '您的操作已成功完成',
});

// 错误通知
api.error({
  message: '操作失败',
  description: '请稍后重试',
});

// 在组件中使用
<>
  {contextHolder}
  <Button onClick={() => api.info({ message: '提示' })}>
    显示通知
  </Button>
</>
```

### Message 消息

用于显示轻量级反馈。

**使用示例**

```tsx
import { message, Button } from 'antd';

const [messageApi, contextHolder] = message.useMessage();

// 成功消息
messageApi.success('操作成功');

// 错误消息
messageApi.error('操作失败');

// 加载消息
messageApi.loading('加载中...');
```

### Progress 进度条

用于展示操作进度。

**样式规范**
- 默认颜色: #6366f1
- 剩余颜色: #e9ecef

**使用示例**

```tsx
import { Progress } from 'antd';

// 线形进度条
<Progress percent={30} />
<Progress percent={100} status="success" />
<Progress percent={70} status="exception" />

// 圆形进度条
<Progress type="circle" percent={75} />

// 仪表盘进度条
<Progress type="dashboard" percent={75} />
```

### Spin 加载

用于展示加载状态。

**使用示例**

```tsx
import { Spin, Card } from 'antd';

// 基础加载
<Spin />

// 不同尺寸
<Spin size="small" />
<Spin size="large" />

// 容器加载
<Spin spinning={loading} tip="加载中...">
  <Card>内容</Card>
</Spin>
```

## 布局组件

### Layout 布局

用于页面整体布局。

**样式规范**
- Header 高度: 64px
- Sider 宽度: 240px (展开), 80px (折叠)
- Body 背景: #fafbfc

**使用示例**

```tsx
import { Layout, Menu } from 'antd';
import { HomeOutlined, SettingOutlined } from '@ant-design/icons';

const { Header, Sider, Content } = Layout;

<Layout style={{ minHeight: '100vh' }}>
  <Sider theme="light" collapsible>
    <div className="logo" />
    <Menu
      mode="inline"
      items={[
        { key: '1', icon: <HomeOutlined />, label: '首页' },
        { key: '2', icon: <SettingOutlined />, label: '设置' },
      ]}
    />
  </Sider>
  <Layout>
    <Header className="bg-white">
      顶部导航
    </Header>
    <Content className="m-6">
      <Card>内容区域</Card>
    </Content>
  </Layout>
</Layout>
```

## 最佳实践

### 1. 使用语义化的颜色

```tsx
// ✓ 好的做法
<Tag color="success">运行中</Tag>
<Tag color="error">失败</Tag>

// ✗ 避免
<Tag color="green">运行中</Tag>
<Tag color="red">失败</Tag>
```

### 2. 提供操作反馈

```tsx
// ✓ 好的做法
const handleDelete = async () => {
  const hide = message.loading('删除中...');
  try {
    await deleteService(id);
    hide();
    message.success('删除成功');
  } catch (error) {
    hide();
    message.error('删除失败');
  }
};
```

### 3. 使用 Tooltip 提供额外信息

```tsx
// ✓ 好的做法
<Tooltip title="编辑服务">
  <Button type="text" icon={<EditOutlined />} />
</Tooltip>
```

### 4. 危险操作需要确认

```tsx
// ✓ 好的做法
<Popconfirm
  title="确定要删除吗？"
  onConfirm={handleDelete}
>
  <Button danger>删除</Button>
</Popconfirm>
```

### 5. 使用骨架屏优化加载体验

```tsx
// ✓ 好的做法
<Skeleton loading={loading} active>
  <Card>{content}</Card>
</Skeleton>

// ✗ 避免
{loading ? <Spin /> : <Card>{content}</Card>}
```

## 响应式设计

使用 Ant Design 的 Grid 系统实现响应式布局：

```tsx
import { Row, Col } from 'antd';

<Row gutter={[16, 16]}>
  <Col xs={24} sm={12} md={8} lg={6}>
    <Card>卡片 1</Card>
  </Col>
  <Col xs={24} sm={12} md={8} lg={6}>
    <Card>卡片 2</Card>
  </Col>
  <Col xs={24} sm={12} md={8} lg={6}>
    <Card>卡片 3</Card>
  </Col>
  <Col xs={24} sm={12} md={8} lg={6}>
    <Card>卡片 4</Card>
  </Col>
</Row>
```

## 主题定制

如需修改主题，编辑 `src/theme/antd-theme.ts`：

```typescript
export const antdTheme: ThemeConfig = {
  token: {
    colorPrimary: '#6366f1',  // 修改主色
    borderRadius: 8,          // 修改圆角
    // ... 其他配置
  },
  components: {
    Button: {
      controlHeight: 40,      // 修改按钮高度
      // ... 其他配置
    },
  },
};
```

## 参考资源

- [Ant Design 官方文档](https://ant.design/)
- [设计系统文档](./design-system.md)
- [组件展示页面](../test/component-showcase.tsx)
- [数据展示组件展示](../test/data-display-showcase.tsx)
- [反馈布局组件展示](../test/feedback-layout-showcase.tsx)

## 更新日志

### 2026-03-02 - Phase 2 完成

- ✓ 重构了所有基础组件样式
- ✓ 重构了数据展示组件
- ✓ 重构了反馈和布局组件
- ✓ 统一了视觉风格和交互规范
- ✓ 创建了完整的组件展示页面

# OpsPilot 组件库规范

## 1. 按钮 (Button)

### 变体 (Variants)

```typescript
// Primary - 主要操作
<Button type="primary">创建服务</Button>

// Secondary - 次要操作
<Button>取消</Button>

// Ghost - 幽灵按钮
<Button type="ghost">查看详情</Button>

// Link - 链接按钮
<Button type="link">了解更多</Button>

// Danger - 危险操作
<Button danger>删除</Button>
```

### 尺寸 (Sizes)

```
small   32px height  ← 紧凑场景
middle  40px height  ← 默认尺寸
large   48px height  ← 强调操作
```

### 状态

- **Default**: 默认状态
- **Hover**: 悬停状态（背景色加深）
- **Active**: 激活状态（背景色更深）
- **Disabled**: 禁用状态（灰色，不可点击）
- **Loading**: 加载状态（显示加载图标）

### 设计规范

```css
/* Primary Button */
background: #6366f1;
color: #ffffff;
border: none;
border-radius: 8px;
padding: 0 16px;
font-size: 14px;
font-weight: 500;
box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
transition: all 150ms cubic-bezier(0.4, 0.0, 0.2, 1);

/* Hover */
background: #4f46e5;
box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1);

/* Active */
background: #4338ca;
transform: translateY(1px);
```

## 2. 输入框 (Input)

### 变体

```typescript
// 标准输入框
<Input placeholder="请输入..." />

// 带前缀图标
<Input prefix={<SearchOutlined />} placeholder="搜索..." />

// 带后缀图标
<Input suffix={<CloseCircleOutlined />} />

// 密码输入框
<Input.Password placeholder="请输入密码" />

// 文本域
<Input.TextArea rows={4} placeholder="请输入描述..." />
```

### 尺寸

```
small   32px height
middle  40px height  ← 默认
large   48px height
```

### 状态

- **Default**: 默认状态
- **Focus**: 聚焦状态（蓝色边框）
- **Error**: 错误状态（红色边框）
- **Disabled**: 禁用状态
- **Read-only**: 只读状态

### 设计规范

```css
background: #ffffff;
border: 1px solid #dee2e6;
border-radius: 8px;
padding: 0 12px;
font-size: 14px;
color: #212529;
transition: all 150ms cubic-bezier(0.4, 0.0, 0.2, 1);

/* Focus */
border-color: #6366f1;
box-shadow: 0 0 0 4px rgba(99, 102, 241, 0.1);
outline: none;

/* Error */
border-color: #ef4444;
box-shadow: 0 0 0 4px rgba(239, 68, 68, 0.1);
```

## 3. 卡片 (Card)

### 变体

```typescript
// 标准卡片
<Card title="服务概览">
  <p>内容...</p>
</Card>

// 无边框卡片
<Card bordered={false}>
  <p>内容...</p>
</Card>

// 可悬停卡片
<Card hoverable>
  <p>内容...</p>
</Card>
```

### 设计规范

```css
background: #ffffff;
border: 1px solid #e9ecef;
border-radius: 12px;
padding: 24px;
box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.08);
transition: all 250ms cubic-bezier(0.4, 0.0, 0.2, 1);

/* Hover (if hoverable) */
box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
transform: translateY(-2px);

/* Card Header */
border-bottom: 1px solid #e9ecef;
padding-bottom: 16px;
margin-bottom: 16px;
font-size: 16px;
font-weight: 600;
color: #212529;
```

## 4. 表格 (Table)

### 设计规范

```css
/* Table Container */
background: #ffffff;
border: 1px solid #e9ecef;
border-radius: 12px;
overflow: hidden;

/* Table Header */
background: #f8f9fa;
border-bottom: 1px solid #e9ecef;
padding: 16px;
font-size: 13px;
font-weight: 600;
color: #495057;
text-transform: uppercase;
letter-spacing: 0.5px;

/* Table Row */
border-bottom: 1px solid #e9ecef;
padding: 16px;
font-size: 14px;
color: #212529;
transition: background 150ms;

/* Row Hover */
background: #f8f9fa;

/* Row Selected */
background: #eef2ff;

/* Cell */
padding: 16px;
vertical-align: middle;
```

### 交互增强

- **排序**: 点击列头排序，显示排序图标
- **筛选**: 列头筛选按钮，下拉筛选面板
- **选择**: 复选框选择行，支持全选
- **展开**: 可展开行显示详细信息
- **固定列**: 支持固定左右列
- **虚拟滚动**: 大数据量使用虚拟滚动

## 5. 表单 (Form)

### 布局

```typescript
// 垂直布局（默认）
<Form layout="vertical">
  <Form.Item label="服务名称" name="name">
    <Input />
  </Form.Item>
</Form>

// 水平布局
<Form layout="horizontal" labelCol={{ span: 4 }}>
  <Form.Item label="服务名称" name="name">
    <Input />
  </Form.Item>
</Form>

// 内联布局
<Form layout="inline">
  <Form.Item label="搜索" name="search">
    <Input />
  </Form.Item>
</Form>
```

### 设计规范

```css
/* Form Item */
margin-bottom: 24px;

/* Label */
font-size: 14px;
font-weight: 500;
color: #495057;
margin-bottom: 8px;
display: block;

/* Required Mark */
color: #ef4444;
margin-right: 4px;

/* Help Text */
font-size: 13px;
color: #6c757d;
margin-top: 4px;

/* Error Message */
font-size: 13px;
color: #ef4444;
margin-top: 4px;
display: flex;
align-items: center;
gap: 4px;
```

### 验证规则

- **实时验证**: 失去焦点时验证
- **提交验证**: 提交时验证所有字段
- **错误提示**: 字段下方显示错误信息
- **成功提示**: 可选的成功状态图标

## 6. 模态框 (Modal)

### 变体

```typescript
// 标准模态框
<Modal title="创建服务" open={open} onOk={handleOk} onCancel={handleCancel}>
  <p>内容...</p>
</Modal>

// 确认对话框
<Modal title="确认删除" open={open} okText="删除" okButtonProps={{ danger: true }}>
  <p>确定要删除这个服务吗？此操作不可恢复。</p>
</Modal>

// 全屏模态框
<Modal title="编辑配置" open={open} width="100%" style={{ top: 0 }}>
  <MonacoEditor />
</Modal>
```

### 设计规范

```css
/* Modal Mask */
background: rgba(0, 0, 0, 0.45);
backdrop-filter: blur(4px);

/* Modal Container */
background: #ffffff;
border-radius: 12px;
box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
max-width: 520px;
width: 90%;

/* Modal Header */
padding: 24px 24px 16px;
border-bottom: 1px solid #e9ecef;
font-size: 18px;
font-weight: 600;
color: #212529;

/* Modal Body */
padding: 24px;
font-size: 14px;
color: #495057;
line-height: 1.5;

/* Modal Footer */
padding: 16px 24px;
border-top: 1px solid #e9ecef;
display: flex;
justify-content: flex-end;
gap: 8px;
```

### 动画

```css
/* 进入动画 */
@keyframes modalFadeIn {
  from {
    opacity: 0;
    transform: scale(0.95) translateY(-20px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

animation: modalFadeIn 250ms cubic-bezier(0.4, 0.0, 0.2, 1);
```

## 7. 标签 (Tag)

### 变体

```typescript
// 默认标签
<Tag>默认</Tag>

// 彩色标签
<Tag color="blue">运行中</Tag>
<Tag color="green">成功</Tag>
<Tag color="orange">警告</Tag>
<Tag color="red">失败</Tag>

// 可关闭标签
<Tag closable onClose={handleClose}>可关闭</Tag>
```

### 设计规范

```css
display: inline-flex;
align-items: center;
gap: 4px;
padding: 4px 8px;
border-radius: 4px;
font-size: 12px;
font-weight: 500;
line-height: 1;
border: 1px solid;

/* Default */
background: #f8f9fa;
border-color: #dee2e6;
color: #495057;

/* Blue */
background: #eef2ff;
border-color: #c7d2fe;
color: #4338ca;

/* Green */
background: #d1fae5;
border-color: #6ee7b7;
color: #047857;

/* Orange */
background: #fed7aa;
border-color: #fdba74;
color: #c2410c;

/* Red */
background: #fee2e2;
border-color: #fca5a5;
color: #b91c1c;
```

## 8. 通知 (Notification)

### 类型

```typescript
// 成功通知
notification.success({
  message: '操作成功',
  description: '服务已成功创建',
});

// 错误通知
notification.error({
  message: '操作失败',
  description: '服务创建失败，请重试',
});

// 警告通知
notification.warning({
  message: '注意',
  description: '服务资源使用率过高',
});

// 信息通知
notification.info({
  message: '提示',
  description: '系统将在 5 分钟后维护',
});
```

### 设计规范

```css
background: #ffffff;
border: 1px solid #e9ecef;
border-left: 4px solid;
border-radius: 8px;
padding: 16px;
box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
min-width: 384px;
max-width: 480px;

/* Success */
border-left-color: #10b981;

/* Error */
border-left-color: #ef4444;

/* Warning */
border-left-color: #f59e0b;

/* Info */
border-left-color: #3b82f6;

/* Title */
font-size: 14px;
font-weight: 600;
color: #212529;
margin-bottom: 4px;

/* Description */
font-size: 13px;
color: #6c757d;
line-height: 1.5;
```

## 9. 加载状态 (Loading)

### 骨架屏 (Skeleton)

```typescript
<Skeleton active paragraph={{ rows: 4 }} />
```

### 加载指示器 (Spinner)

```typescript
<Spin size="small" />
<Spin size="default" />
<Spin size="large" />
```

### 进度条 (Progress)

```typescript
<Progress percent={60} />
<Progress percent={100} status="success" />
<Progress percent={50} status="exception" />
```

## 10. 空状态 (Empty State)

### 设计规范

```css
display: flex;
flex-direction: column;
align-items: center;
justify-content: center;
padding: 48px 24px;
text-align: center;

/* Icon */
font-size: 64px;
color: #dee2e6;
margin-bottom: 16px;

/* Title */
font-size: 16px;
font-weight: 600;
color: #495057;
margin-bottom: 8px;

/* Description */
font-size: 14px;
color: #6c757d;
margin-bottom: 24px;
max-width: 400px;

/* Action Button */
/* 使用 Primary Button 样式 */
```


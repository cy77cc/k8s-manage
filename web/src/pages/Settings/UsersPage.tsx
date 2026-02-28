import React from 'react';
import { Button, Card, Form, Input, Modal, Select, Space, Table, Tag, message, Drawer, Empty } from 'antd';
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons';
import { Api } from '../../api';
import type { User } from '../../api/modules/rbac';
import { ApiRequestError } from '../../api/api';
import AccessDeniedPage from '../../components/Auth/AccessDeniedPage';

const UsersPage: React.FC = () => {
  const [loading, setLoading] = React.useState(false);
  const [list, setList] = React.useState<User[]>([]);
  const [open, setOpen] = React.useState(false);
  const [query, setQuery] = React.useState('');
  const [active, setActive] = React.useState<User | null>(null);
  const [accessDenied, setAccessDenied] = React.useState(false);
  const [form] = Form.useForm();

  const load = React.useCallback(async () => {
    setLoading(true);
    try {
      const res = await Api.rbac.getUserList({ page: 1, pageSize: 200 });
      setList(res.data.list || []);
      setAccessDenied(false);
    } catch (err) {
      if (err instanceof ApiRequestError && (err.statusCode === 403 || err.businessCode === 2004)) {
        setAccessDenied(true);
        return;
      }
      message.error(err instanceof Error ? err.message : '加载用户失败');
    } finally {
      setLoading(false);
    }
  }, []);

  React.useEffect(() => { load(); }, [load]);

  const create = async () => {
    const v = await form.validateFields();
    const startedAt = performance.now();
    await Api.rbac.createUser({
      username: v.username,
      name: v.name || v.username,
      email: v.email,
      password: v.password,
      roles: v.roles,
      status: v.status || 'active',
    });
    void Api.rbac.recordMigrationEvent({
      eventType: 'governance_task',
      action: 'user.create',
      status: 'success',
      durationMs: Math.round(performance.now() - startedAt),
    }).catch(() => undefined);
    message.success('创建成功');
    setOpen(false);
    form.resetFields();
    load();
  };

  const remove = async (id: string) => {
    Modal.confirm({
      title: '删除用户',
      content: '本次操作将影响 1 个用户账号，确认继续？',
      okText: '确认删除',
      okButtonProps: { danger: true },
      onOk: async () => {
        const startedAt = performance.now();
        await Api.rbac.deleteUser(id);
        void Api.rbac.recordMigrationEvent({
          eventType: 'governance_task',
          action: 'user.delete',
          status: 'success',
          durationMs: Math.round(performance.now() - startedAt),
        }).catch(() => undefined);
        message.success('删除成功');
        load();
      },
    });
  };

  const filtered = list.filter((item) => {
    const q = query.trim().toLowerCase();
    if (!q) return true;
    return [item.username, item.name, item.email].some((v) => String(v || '').toLowerCase().includes(q));
  });

  if (accessDenied) {
    return <AccessDeniedPage />;
  }

  return (
    <Card
      title="用户管理"
      extra={
        <Space>
          <Input
            allowClear
            aria-label="搜索用户"
            placeholder="搜索用户名/姓名/邮箱"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            style={{ width: 260 }}
          />
          <Button className="governance-action-btn" icon={<ReloadOutlined />} onClick={load} loading={loading}>刷新</Button>
          <Button className="governance-action-btn" type="primary" icon={<PlusOutlined />} onClick={() => setOpen(true)}>新增用户</Button>
        </Space>
      }
    >
      <Table
        rowKey="id"
        loading={loading}
        locale={{ emptyText: <Empty description="暂无用户数据" /> }}
        dataSource={filtered}
        onRow={(record) => ({
          onClick: () => setActive(record),
          onKeyDown: (event) => {
            if (event.key === 'Enter' || event.key === ' ') {
              event.preventDefault();
              setActive(record);
            }
          },
          tabIndex: 0,
          role: 'button',
          className: 'governance-interactive-row',
          'aria-label': `查看用户 ${record.username} 详情`,
        })}
        columns={[
          { title: '用户名', dataIndex: 'username' },
          { title: '姓名', dataIndex: 'name' },
          { title: '邮箱', dataIndex: 'email' },
          { title: '角色', dataIndex: 'roles', render: (roles: string[]) => <>{roles?.map((r) => <Tag key={r}>{r}</Tag>)}</> },
          { title: '状态', dataIndex: 'status', render: (v: string) => <Tag color={v === 'active' ? 'success' : 'default'}>{v}</Tag> },
          {
            title: '操作',
            render: (_: unknown, row: User) => (
              <Button
                className="governance-action-btn"
                danger
                type="link"
                aria-label={`删除用户 ${row.username}`}
                onClick={(event) => {
                  event.stopPropagation();
                  remove(row.id);
                }}
              >
                删除
              </Button>
            ),
          },
        ]}
      />

      <Modal title="新增用户" open={open} onCancel={() => setOpen(false)} onOk={create}>
        <Form form={form} layout="vertical" initialValues={{ roles: ['viewer'], status: 'active' }}>
          <Form.Item name="username" label="用户名" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="name" label="姓名"><Input /></Form.Item>
          <Form.Item name="email" label="邮箱" rules={[{ required: true, type: 'email' }]}><Input /></Form.Item>
          <Form.Item name="password" label="密码" rules={[{ required: true, min: 6 }]}><Input.Password /></Form.Item>
          <Form.Item name="roles" label="角色" rules={[{ required: true }]}><Select mode="multiple" options={[{ value: 'admin' }, { value: 'operator' }, { value: 'viewer' }]} /></Form.Item>
          <Form.Item name="status" label="状态"><Select options={[{ value: 'active' }, { value: 'disabled' }]} /></Form.Item>
        </Form>
      </Modal>

      <Drawer title="用户详情" open={Boolean(active)} onClose={() => setActive(null)} width={420}>
        {!active ? null : (
          <Space direction="vertical" size={12} style={{ width: '100%' }}>
            <div><strong>用户名：</strong>{active.username}</div>
            <div><strong>姓名：</strong>{active.name}</div>
            <div><strong>邮箱：</strong>{active.email}</div>
            <div>
              <strong>角色：</strong>{' '}
              {(active.roles || []).map((r) => <Tag key={r}>{r}</Tag>)}
            </div>
            <div><strong>状态：</strong>{active.status}</div>
          </Space>
        )}
      </Drawer>
    </Card>
  );
};

export default UsersPage;

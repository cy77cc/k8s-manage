import React from 'react';
import { Button, Card, Form, Input, Modal, Select, Space, Table, Tag, message } from 'antd';
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons';
import { Api } from '../../api';
import type { User } from '../../api/modules/rbac';

const UsersPage: React.FC = () => {
  const [loading, setLoading] = React.useState(false);
  const [list, setList] = React.useState<User[]>([]);
  const [open, setOpen] = React.useState(false);
  const [form] = Form.useForm();

  const load = React.useCallback(async () => {
    setLoading(true);
    try {
      const res = await Api.rbac.getUserList({ page: 1, pageSize: 200 });
      setList(res.data.list || []);
    } catch (err) {
      message.error(err instanceof Error ? err.message : '加载用户失败');
    } finally {
      setLoading(false);
    }
  }, []);

  React.useEffect(() => { load(); }, [load]);

  const create = async () => {
    const v = await form.validateFields();
    await Api.rbac.createUser({
      username: v.username,
      name: v.name || v.username,
      email: v.email,
      password: v.password,
      roles: v.roles,
      status: v.status || 'active',
    });
    message.success('创建成功');
    setOpen(false);
    form.resetFields();
    load();
  };

  const remove = async (id: string) => {
    await Api.rbac.deleteUser(id);
    message.success('删除成功');
    load();
  };

  return (
    <Card
      title="用户管理"
      extra={
        <Space>
          <Button icon={<ReloadOutlined />} onClick={load} loading={loading}>刷新</Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setOpen(true)}>新增用户</Button>
        </Space>
      }
    >
      <Table
        rowKey="id"
        loading={loading}
        dataSource={list}
        columns={[
          { title: '用户名', dataIndex: 'username' },
          { title: '姓名', dataIndex: 'name' },
          { title: '邮箱', dataIndex: 'email' },
          { title: '角色', dataIndex: 'roles', render: (roles: string[]) => <>{roles?.map((r) => <Tag key={r}>{r}</Tag>)}</> },
          { title: '状态', dataIndex: 'status', render: (v: string) => <Tag color={v === 'active' ? 'success' : 'default'}>{v}</Tag> },
          {
            title: '操作',
            render: (_: unknown, row: User) => <Button danger type="link" onClick={() => remove(row.id)}>删除</Button>,
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
    </Card>
  );
};

export default UsersPage;

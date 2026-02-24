import React from 'react';
import { Button, Card, Form, Input, Modal, Select, Space, Table, Tag, message } from 'antd';
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons';
import { Api } from '../../api';
import type { Permission, Role } from '../../api/modules/rbac';

const RolesPage: React.FC = () => {
  const [loading, setLoading] = React.useState(false);
  const [roles, setRoles] = React.useState<Role[]>([]);
  const [permissions, setPermissions] = React.useState<Permission[]>([]);
  const [open, setOpen] = React.useState(false);
  const [form] = Form.useForm();

  const load = React.useCallback(async () => {
    setLoading(true);
    try {
      const [roleRes, permRes] = await Promise.all([
        Api.rbac.getRoleList({ page: 1, pageSize: 200 }),
        Api.rbac.getPermissionList({ page: 1, pageSize: 500 }),
      ]);
      setRoles(roleRes.data.list || []);
      setPermissions(permRes.data.list || []);
    } catch (err) {
      message.error(err instanceof Error ? err.message : '加载角色失败');
    } finally {
      setLoading(false);
    }
  }, []);

  React.useEffect(() => { load(); }, [load]);

  const create = async () => {
    const v = await form.validateFields();
    await Api.rbac.createRole({ name: v.name, description: v.description, permissions: v.permissions || [] });
    message.success('创建角色成功');
    setOpen(false);
    form.resetFields();
    load();
  };

  return (
    <Card
      title="角色管理"
      extra={
        <Space>
          <Button icon={<ReloadOutlined />} onClick={load} loading={loading}>刷新</Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setOpen(true)}>新增角色</Button>
        </Space>
      }
    >
      <Table
        rowKey="id"
        loading={loading}
        dataSource={roles}
        columns={[
          { title: '角色', dataIndex: 'name' },
          { title: '描述', dataIndex: 'description' },
          {
            title: '权限',
            dataIndex: 'permissions',
            render: (perms: string[]) => (
              <Space wrap>
                {(perms || []).slice(0, 6).map((p) => <Tag key={p}>{p}</Tag>)}
                {(perms || []).length > 6 ? <Tag>+{(perms || []).length - 6}</Tag> : null}
              </Space>
            ),
          },
        ]}
      />

      <Modal title="新增角色" open={open} onCancel={() => setOpen(false)} onOk={create} width={680}>
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="角色名" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="description" label="描述"><Input /></Form.Item>
          <Form.Item name="permissions" label="权限">
            <Select mode="multiple" options={permissions.map((p) => ({ value: p.code, label: `${p.code} (${p.name})` }))} />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  );
};

export default RolesPage;

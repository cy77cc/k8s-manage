import React from 'react';
import { Button, Card, Form, Input, Modal, Select, Space, Table, Tag, message, Drawer, Empty } from 'antd';
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons';
import { Api } from '../../api';
import type { Permission, Role } from '../../api/modules/rbac';
import { ApiRequestError } from '../../api/api';
import AccessDeniedPage from '../../components/Auth/AccessDeniedPage';

const RolesPage: React.FC = () => {
  const [loading, setLoading] = React.useState(false);
  const [roles, setRoles] = React.useState<Role[]>([]);
  const [permissions, setPermissions] = React.useState<Permission[]>([]);
  const [open, setOpen] = React.useState(false);
  const [query, setQuery] = React.useState('');
  const [active, setActive] = React.useState<Role | null>(null);
  const [accessDenied, setAccessDenied] = React.useState(false);
  const [detailForm] = Form.useForm();
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
      setAccessDenied(false);
    } catch (err) {
      if (err instanceof ApiRequestError && (err.statusCode === 403 || err.businessCode === 2004)) {
        setAccessDenied(true);
        return;
      }
      message.error(err instanceof Error ? err.message : '加载角色失败');
    } finally {
      setLoading(false);
    }
  }, []);

  React.useEffect(() => { load(); }, [load]);

  const create = async () => {
    const v = await form.validateFields();
    const startedAt = performance.now();
    await Api.rbac.createRole({ name: v.name, description: v.description, permissions: v.permissions || [] });
    void Api.rbac.recordMigrationEvent({
      eventType: 'governance_task',
      action: 'role.create',
      status: 'success',
      durationMs: Math.round(performance.now() - startedAt),
    }).catch(() => undefined);
    message.success('创建角色成功');
    setOpen(false);
    form.resetFields();
    load();
  };

  const filtered = roles.filter((item) => {
    const q = query.trim().toLowerCase();
    if (!q) return true;
    return [item.name, item.description].some((v) => String(v || '').toLowerCase().includes(q));
  });

  React.useEffect(() => {
    if (active) {
      detailForm.setFieldsValue({
        name: active.name,
        description: active.description,
        permissions: active.permissions || [],
      });
    } else {
      detailForm.resetFields();
    }
  }, [active, detailForm]);

  const saveRolePermissions = async () => {
    if (!active) return;
    const values = await detailForm.validateFields();
    const nextPermissions: string[] = values.permissions || [];
    const removedCount = (active.permissions || []).filter((p) => !nextPermissions.includes(p)).length;
    Modal.confirm({
      title: '确认更新角色权限',
      content: `本次将回收 ${removedCount} 项权限，影响对象数量：1 个角色。是否继续？`,
      okText: '确认更新',
      onOk: async () => {
        const startedAt = performance.now();
        await Api.rbac.updateRole(active.id, {
          name: values.name,
          description: values.description,
          permissions: nextPermissions,
        });
        void Api.rbac.recordMigrationEvent({
          eventType: 'governance_task',
          action: 'role.update_permissions',
          status: 'success',
          durationMs: Math.round(performance.now() - startedAt),
        }).catch(() => undefined);
        message.success('角色权限更新成功');
        setActive(null);
        load();
      },
    });
  };

  if (accessDenied) {
    return <AccessDeniedPage />;
  }

  return (
    <Card
      title="角色管理"
      extra={
        <Space>
          <Input
            allowClear
            aria-label="搜索角色"
            placeholder="搜索角色名/描述"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            style={{ width: 260 }}
          />
          <Button className="governance-action-btn" icon={<ReloadOutlined />} onClick={load} loading={loading}>刷新</Button>
          <Button className="governance-action-btn" type="primary" icon={<PlusOutlined />} onClick={() => setOpen(true)}>新增角色</Button>
        </Space>
      }
    >
      <Table
        rowKey="id"
        loading={loading}
        locale={{ emptyText: <Empty description="暂无角色数据" /> }}
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
          'aria-label': `查看角色 ${record.name} 详情`,
        })}
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

      <Drawer
        title="角色详情与权限编辑"
        open={Boolean(active)}
        onClose={() => setActive(null)}
        width={520}
        extra={<Button className="governance-action-btn" type="primary" onClick={saveRolePermissions}>保存变更</Button>}
      >
        <Form form={detailForm} layout="vertical">
          <Form.Item name="name" label="角色名" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="description" label="描述"><Input /></Form.Item>
          <Form.Item name="permissions" label="权限">
            <Select
              mode="multiple"
              options={permissions.map((p) => ({ value: p.code, label: `${p.code} (${p.name})` }))}
            />
          </Form.Item>
        </Form>
      </Drawer>
    </Card>
  );
};

export default RolesPage;

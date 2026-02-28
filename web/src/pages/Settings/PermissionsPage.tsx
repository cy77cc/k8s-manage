import React from 'react';
import { Card, Table, Tag, Input, Empty, Drawer, Space } from 'antd';
import { Api } from '../../api';
import type { Permission } from '../../api/modules/rbac';
import { ApiRequestError } from '../../api/api';
import AccessDeniedPage from '../../components/Auth/AccessDeniedPage';

const PermissionsPage: React.FC = () => {
  const [list, setList] = React.useState<Permission[]>([]);
  const [loading, setLoading] = React.useState(false);
  const [query, setQuery] = React.useState('');
  const [accessDenied, setAccessDenied] = React.useState(false);
  const [active, setActive] = React.useState<Permission | null>(null);

  React.useEffect(() => {
    const load = async () => {
      setLoading(true);
      try {
        const res = await Api.rbac.getPermissionList({ page: 1, pageSize: 1000 });
        setList(res.data.list || []);
        setAccessDenied(false);
      } catch (err) {
        if (err instanceof ApiRequestError && (err.statusCode === 403 || err.businessCode === 2004)) {
          setAccessDenied(true);
          return;
        }
      } finally {
        setLoading(false);
      }
    };
    load();
  }, []);

  const filtered = list.filter((item) => {
    const q = query.trim().toLowerCase();
    if (!q) return true;
    return [item.code, item.name, item.description, item.category]
      .some((v) => String(v || '').toLowerCase().includes(q));
  });

  if (accessDenied) {
    return <AccessDeniedPage />;
  }

  return (
    <Card
      title="权限列表"
      extra={
        <Input
          allowClear
          aria-label="搜索权限"
          placeholder="搜索权限 code/名称/分类"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          style={{ width: 280 }}
        />
      }
    >
      <Table
        rowKey="id"
        loading={loading}
        locale={{ emptyText: <Empty description="暂无权限数据" /> }}
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
          'aria-label': `查看权限 ${record.code} 详情`,
        })}
        columns={[
          { title: 'Code', dataIndex: 'code', render: (v: string) => <code>{v}</code> },
          { title: '名称', dataIndex: 'name' },
          { title: '分类', dataIndex: 'category', render: (v: string) => <Tag>{v}</Tag> },
          { title: '描述', dataIndex: 'description' },
        ]}
      />

      <Drawer title="权限详情" open={Boolean(active)} onClose={() => setActive(null)} width={420}>
        {!active ? null : (
          <Space direction="vertical" size={12} style={{ width: '100%' }}>
            <div><strong>Code：</strong><code>{active.code}</code></div>
            <div><strong>名称：</strong>{active.name}</div>
            <div><strong>分类：</strong>{active.category}</div>
            <div><strong>描述：</strong>{active.description || '-'}</div>
          </Space>
        )}
      </Drawer>
    </Card>
  );
};

export default PermissionsPage;

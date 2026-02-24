import React from 'react';
import { Card, Table, Tag } from 'antd';
import { Api } from '../../api';
import type { Permission } from '../../api/modules/rbac';

const PermissionsPage: React.FC = () => {
  const [list, setList] = React.useState<Permission[]>([]);
  const [loading, setLoading] = React.useState(false);

  React.useEffect(() => {
    const load = async () => {
      setLoading(true);
      try {
        const res = await Api.rbac.getPermissionList({ page: 1, pageSize: 1000 });
        setList(res.data.list || []);
      } finally {
        setLoading(false);
      }
    };
    load();
  }, []);

  return (
    <Card title="权限列表">
      <Table
        rowKey="id"
        loading={loading}
        dataSource={list}
        columns={[
          { title: 'Code', dataIndex: 'code', render: (v: string) => <code>{v}</code> },
          { title: '名称', dataIndex: 'name' },
          { title: '分类', dataIndex: 'category', render: (v: string) => <Tag>{v}</Tag> },
          { title: '描述', dataIndex: 'description' },
        ]}
      />
    </Card>
  );
};

export default PermissionsPage;

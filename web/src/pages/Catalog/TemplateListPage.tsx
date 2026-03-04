import React, { useCallback, useEffect, useState } from 'react';
import { Button, Card, Popconfirm, Select, Space, Table, Tag, message } from 'antd';
import { useNavigate } from 'react-router-dom';
import { Api } from '../../api';
import type { CatalogTemplate, CatalogTemplateStatus } from '../../api/modules/catalog';

const TemplateListPage: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [status, setStatus] = useState<CatalogTemplateStatus | 'all'>('all');
  const [list, setList] = useState<CatalogTemplate[]>([]);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const resp = await Api.catalog.listTemplates({
        mine: true,
        status: status === 'all' ? undefined : status,
      });
      setList(resp.data.list || []);
    } catch (err) {
      message.error(err instanceof Error ? err.message : '加载模板失败');
    } finally {
      setLoading(false);
    }
  }, [status]);

  useEffect(() => {
    void load();
  }, [load]);

  const removeTemplate = async (id: number) => {
    try {
      await Api.catalog.deleteTemplate(id);
      message.success('删除成功');
      void load();
    } catch (err) {
      message.error(err instanceof Error ? err.message : '删除失败');
    }
  };

  return (
    <div className="p-6 space-y-4">
      <Card>
        <Space>
          <Button type="primary" onClick={() => navigate('/catalog/templates/create')}>创建模板</Button>
          <Select
            value={status}
            style={{ width: 220 }}
            onChange={(value) => setStatus(value)}
            options={[
              { label: '全部状态', value: 'all' },
              { label: '草稿', value: 'draft' },
              { label: '待审核', value: 'pending_review' },
              { label: '已发布', value: 'published' },
              { label: '已驳回', value: 'rejected' },
            ]}
          />
        </Space>
      </Card>

      <Card title="我的模板">
        <Table
          loading={loading}
          rowKey="id"
          dataSource={list}
          columns={[
            { title: '名称', dataIndex: 'display_name' },
            { title: '标识', dataIndex: 'name' },
            { title: '状态', dataIndex: 'status', render: (value) => <Tag>{value}</Tag> },
            { title: '版本', dataIndex: 'version' },
            {
              title: '操作',
              render: (_, row: CatalogTemplate) => (
                <Space>
                  <Button size="small" onClick={() => navigate(`/catalog/templates/${row.id}/edit`)}>编辑</Button>
                  <Popconfirm title="确认删除模板？" onConfirm={() => removeTemplate(row.id)}>
                    <Button size="small" danger>删除</Button>
                  </Popconfirm>
                </Space>
              ),
            },
          ]}
        />
      </Card>
    </div>
  );
};

export default TemplateListPage;

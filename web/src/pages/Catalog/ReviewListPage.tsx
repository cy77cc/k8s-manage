import React, { useCallback, useEffect, useState } from 'react';
import { Button, Card, Input, Modal, Space, Table, Tag, message } from 'antd';
import { Api } from '../../api';
import type { CatalogTemplate } from '../../api/modules/catalog';

const ReviewListPage: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [list, setList] = useState<CatalogTemplate[]>([]);
  const [rejecting, setRejecting] = useState<CatalogTemplate | null>(null);
  const [rejectReason, setRejectReason] = useState('');

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const resp = await Api.catalog.listTemplates({ status: 'pending_review' });
      setList(resp.data.list || []);
    } catch (err) {
      message.error(err instanceof Error ? err.message : '加载审核列表失败');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  const approve = async (id: number) => {
    try {
      await Api.catalog.publishTemplate(id);
      message.success('发布成功');
      void load();
    } catch (err) {
      message.error(err instanceof Error ? err.message : '发布失败');
    }
  };

  const reject = async () => {
    if (!rejecting) {
      return;
    }
    try {
      await Api.catalog.rejectTemplate(rejecting.id, rejectReason);
      message.success('已驳回模板');
      setRejecting(null);
      setRejectReason('');
      void load();
    } catch (err) {
      message.error(err instanceof Error ? err.message : '驳回失败');
    }
  };

  return (
    <div className="p-6 space-y-4">
      <Card title="待审核模板">
        <Table
          loading={loading}
          rowKey="id"
          dataSource={list}
          columns={[
            { title: '模板名称', dataIndex: 'display_name' },
            { title: '提交者', dataIndex: 'owner_id' },
            { title: '提交时间', dataIndex: 'updated_at' },
            { title: '状态', dataIndex: 'status', render: (value) => <Tag color="orange">{value}</Tag> },
            {
              title: '操作',
              render: (_, row: CatalogTemplate) => (
                <Space>
                  <Button type="primary" size="small" onClick={() => approve(row.id)}>发布</Button>
                  <Button danger size="small" onClick={() => setRejecting(row)}>驳回</Button>
                </Space>
              ),
            },
          ]}
        />
      </Card>

      <Modal
        title="驳回模板"
        open={Boolean(rejecting)}
        onCancel={() => setRejecting(null)}
        onOk={reject}
      >
        <Input.TextArea
          rows={4}
          value={rejectReason}
          onChange={(e) => setRejectReason(e.target.value)}
          placeholder="请输入驳回原因"
        />
      </Modal>
    </div>
  );
};

export default ReviewListPage;

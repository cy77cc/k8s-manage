import React, { useEffect, useState } from 'react';
import { Button, Card, Form, Input, Modal, Select, Table, message } from 'antd';
import { Api } from '../../api';
import type { CMDBAsset } from '../../api/modules/cmdb';

const CMDBPage: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [assets, setAssets] = useState<CMDBAsset[]>([]);
  const [open, setOpen] = useState(false);
  const [form] = Form.useForm();

  const load = async () => {
    setLoading(true);
    try {
      const res = await Api.cmdb.listAssets();
      setAssets(res.data || []);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const create = async () => {
    const values = await form.validateFields();
    await Api.cmdb.createAsset({
      assetType: values.assetType,
      name: values.name,
      owner: values.owner,
      source: values.source,
    });
    message.success('资产创建成功');
    setOpen(false);
    form.resetFields();
    load();
  };

  return (
    <Card
      title="CMDB 资产台账"
      extra={<Button type="primary" onClick={() => setOpen(true)}>新增资产</Button>}
    >
      <Table
        rowKey="id"
        loading={loading}
        dataSource={assets}
        columns={[
          { title: 'ID', dataIndex: 'id', width: 80 },
          { title: '名称', dataIndex: 'name' },
          { title: '类型', dataIndex: 'assetType' },
          { title: '来源', dataIndex: 'source' },
          { title: '状态', dataIndex: 'status' },
          { title: 'Owner', dataIndex: 'owner' },
        ]}
      />
      <Modal title="新增资产" open={open} onCancel={() => setOpen(false)} onOk={create}>
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="资产名称" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="assetType" label="资产类型" rules={[{ required: true }]}><Select options={[{ value: 'host' }, { value: 'service' }, { value: 'cluster' }, { value: 'custom' }]} /></Form.Item>
          <Form.Item name="source" label="来源" initialValue="manual"><Select options={[{ value: 'manual' }, { value: 'system' }]} /></Form.Item>
          <Form.Item name="owner" label="负责人"><Input /></Form.Item>
        </Form>
      </Modal>
    </Card>
  );
};

export default CMDBPage;

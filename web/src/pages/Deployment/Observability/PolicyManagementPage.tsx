import React, { useState, useEffect } from 'react';
import { Card, Button, Space, Table, Modal, Form, Input, Select, message, Tabs } from 'antd';
import { PlusOutlined, ReloadOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';

interface Policy {
  id: string;
  service_id: string;
  service_name: string;
  type: 'traffic' | 'resilience' | 'access' | 'slo';
  name: string;
  config: any;
  enabled: boolean;
  created_at: string;
}

const PolicyManagementPage: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [policies, setPolicies] = useState<Policy[]>([]);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingPolicy, setEditingPolicy] = useState<Policy | null>(null);
  const [form] = Form.useForm();

  const load = async () => {
    setLoading(true);
    try {
      // Mock data - replace with actual API call
      const mockPolicies: Policy[] = [
        {
          id: '1',
          service_id: '1',
          service_name: 'api-gateway',
          type: 'traffic',
          name: 'Rate Limiting',
          config: { rate: 1000, per: 'minute' },
          enabled: true,
          created_at: new Date().toISOString(),
        },
        {
          id: '2',
          service_id: '1',
          service_name: 'api-gateway',
          type: 'resilience',
          name: 'Circuit Breaker',
          config: { threshold: 5, timeout: 30 },
          enabled: true,
          created_at: new Date().toISOString(),
        },
      ];
      setPolicies(mockPolicies);
    } catch (err) {
      message.error('加载策略失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const handleCreate = () => {
    setEditingPolicy(null);
    form.resetFields();
    setModalVisible(true);
  };

  const handleEdit = (policy: Policy) => {
    setEditingPolicy(policy);
    form.setFieldsValue({
      service_id: policy.service_id,
      type: policy.type,
      name: policy.name,
      config: JSON.stringify(policy.config, null, 2),
    });
    setModalVisible(true);
  };

  const handleDelete = async (id: string) => {
    Modal.confirm({
      title: '确认删除',
      content: '确定要删除此策略吗？',
      okText: '删除',
      okButtonProps: { danger: true },
      cancelText: '取消',
      onOk: async () => {
        try {
          // API call to delete policy
          message.success('策略已删除');
          load();
        } catch (err) {
          message.error('删除失败');
        }
      },
    });
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      let config = {};
      try {
        config = JSON.parse(values.config);
      } catch {
        message.error('配置 JSON 格式错误');
        return;
      }

      // API call to create/update policy
      message.success(editingPolicy ? '策略已更新' : '策略已创建');
      setModalVisible(false);
      load();
    } catch (err) {
      // Validation failed
    }
  };

  const columns: ColumnsType<Policy> = [
    {
      title: '服务',
      dataIndex: 'service_name',
      key: 'service_name',
    },
    {
      title: '策略类型',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => {
        const labels: Record<string, string> = {
          traffic: '流量策略',
          resilience: '弹性策略',
          access: '访问策略',
          slo: 'SLO 策略',
        };
        return labels[type] || type;
      },
    },
    {
      title: '策略名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '状态',
      dataIndex: 'enabled',
      key: 'enabled',
      render: (enabled: boolean) => (
        <span className={enabled ? 'text-green-600' : 'text-gray-400'}>
          {enabled ? '已启用' : '已禁用'}
        </span>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (time: string) => new Date(time).toLocaleString(),
    },
    {
      title: '操作',
      key: 'actions',
      render: (_: any, record: Policy) => (
        <Space>
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)}>
            编辑
          </Button>
          <Button
            type="link"
            size="small"
            danger
            icon={<DeleteOutlined />}
            onClick={() => handleDelete(record.id)}
          >
            删除
          </Button>
        </Space>
      ),
    },
  ];

  const groupedPolicies = policies.reduce((acc, policy) => {
    if (!acc[policy.service_name]) {
      acc[policy.service_name] = [];
    }
    acc[policy.service_name].push(policy);
    return acc;
  }, {} as Record<string, Policy[]>);

  return (
    <div className="space-y-6">
      {/* Page header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-gray-900">策略管理</h1>
          <p className="text-sm text-gray-500 mt-1">配置服务的流量、弹性、访问和 SLO 策略</p>
        </div>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={load} loading={loading}>
            刷新
          </Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
            创建策略
          </Button>
        </Space>
      </div>

      {/* Policies grouped by service */}
      <Card>
        <Tabs
          items={Object.entries(groupedPolicies).map(([serviceName, servicePolicies]) => ({
            key: serviceName,
            label: `${serviceName} (${servicePolicies.length})`,
            children: (
              <Table
                dataSource={servicePolicies}
                columns={columns}
                rowKey="id"
                pagination={false}
              />
            ),
          }))}
        />
      </Card>

      {/* Create/Edit modal */}
      <Modal
        title={editingPolicy ? '编辑策略' : '创建策略'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => {
          setModalVisible(false);
          setEditingPolicy(null);
        }}
        width={720}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            name="service_id"
            label="服务"
            rules={[{ required: true, message: '请选择服务' }]}
          >
            <Select
              placeholder="选择服务"
              options={[
                { value: '1', label: 'api-gateway' },
                { value: '2', label: 'user-service' },
                { value: '3', label: 'order-service' },
              ]}
            />
          </Form.Item>
          <Form.Item
            name="type"
            label="策略类型"
            rules={[{ required: true, message: '请选择策略类型' }]}
          >
            <Select
              placeholder="选择策略类型"
              options={[
                { value: 'traffic', label: '流量策略' },
                { value: 'resilience', label: '弹性策略' },
                { value: 'access', label: '访问策略' },
                { value: 'slo', label: 'SLO 策略' },
              ]}
            />
          </Form.Item>
          <Form.Item
            name="name"
            label="策略名称"
            rules={[{ required: true, message: '请输入策略名称' }]}
          >
            <Input placeholder="例如: Rate Limiting" />
          </Form.Item>
          <Form.Item
            name="config"
            label="配置 (JSON)"
            rules={[{ required: true, message: '请输入配置' }]}
          >
            <Input.TextArea
              rows={10}
              placeholder={'{\n  "rate": 1000,\n  "per": "minute"\n}'}
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default PolicyManagementPage;

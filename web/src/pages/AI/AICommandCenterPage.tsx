import React from 'react';
import { Button, Card, Form, Input, InputNumber, Select, Space, Typography, message } from 'antd';
import { Api } from '../../api';
import CommandPanel from '../../components/AI/CommandPanel';

const { Title, Text } = Typography;

const AICommandCenterPage: React.FC = () => {
  const [form] = Form.useForm();

  const bootstrapByAIEntry = async () => {
    const v = await form.validateFields();
    const payload: any = {
      name: String(v.name),
      runtime_type: String(v.runtime_type),
      package_version: String(v.package_version),
      env: String(v.env || 'staging'),
    };
    if (payload.runtime_type === 'k8s') {
      payload.control_plane_host_id = Number(v.control_plane_host_id || 0);
      payload.worker_host_ids = [];
    } else {
      payload.node_ids = [Number(v.node_id || 0)];
    }
    const resp = await Api.deployment.startEnvironmentBootstrap(payload);
    message.success(`AI入口已发起部署任务: ${resp.data.job_id}`);
  };

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Card>
        <Title level={4} style={{ marginBottom: 0 }}>AI 命令中心</Title>
        <Text type="secondary">通过命令预览、确认执行、历史回放，完成跨域运维操作。</Text>
      </Card>
      <Card title="部署快速动作（与 Deployment 页面共用治理链）">
        <Form form={form} layout="inline" initialValues={{ runtime_type: 'k8s', package_version: 'v0.1.0', env: 'staging' }}>
          <Form.Item name="name" rules={[{ required: true }]}>
            <Input placeholder="任务名称" />
          </Form.Item>
          <Form.Item name="runtime_type" rules={[{ required: true }]}>
            <Select style={{ width: 120 }} options={[{ value: 'k8s' }, { value: 'compose' }]} />
          </Form.Item>
          <Form.Item name="package_version" rules={[{ required: true }]}>
            <Input placeholder="版本" />
          </Form.Item>
          <Form.Item name="env">
            <Select style={{ width: 120 }} options={[{ value: 'staging' }, { value: 'production' }]} />
          </Form.Item>
          <Form.Item shouldUpdate noStyle>
            {({ getFieldValue }) => (getFieldValue('runtime_type') === 'k8s' ? (
              <Form.Item name="control_plane_host_id" rules={[{ required: true }]}>
                <InputNumber min={1} placeholder="控制平面HostID" />
              </Form.Item>
            ) : (
              <Form.Item name="node_id" rules={[{ required: true }]}>
                <InputNumber min={1} placeholder="Compose HostID" />
              </Form.Item>
            ))}
          </Form.Item>
          <Button type="primary" onClick={() => void bootstrapByAIEntry()}>发起环境部署</Button>
        </Form>
      </Card>
      <CommandPanel scene="scene:ai" />
    </Space>
  );
};

export default AICommandCenterPage;

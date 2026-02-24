import React from 'react';
import { Button, Card, Form, Input, InputNumber, Select, Space, message } from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { Api } from '../../api';

const ServiceProvisionPage: React.FC = () => {
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const [loading, setLoading] = React.useState(false);

  const onFinish = async (values: any) => {
    setLoading(true);
    try {
      await Api.services.create({
        name: values.name,
        env: values.env,
        owner: values.owner,
        image: values.image,
        replicas: values.replicas,
        cpuLimit: values.cpuLimit,
        memLimit: values.memLimit,
        tags: values.tags || [],
        config: values.config || '',
      });
      message.success('服务创建成功');
      navigate('/services');
    } catch (err) {
      message.error(err instanceof Error ? err.message : '创建失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card
      title="创建服务"
      extra={<Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/services')}>返回</Button>}
    >
      <Form form={form} layout="vertical" onFinish={onFinish} initialValues={{ env: 'staging', replicas: 1, cpuLimit: 500, memLimit: 512 }}>
        <Form.Item label="名称" name="name" rules={[{ required: true, message: '请输入服务名称' }]}>
          <Input placeholder="例如 user-service" />
        </Form.Item>
        <Form.Item label="镜像" name="image" rules={[{ required: true, message: '请输入镜像' }]}>
          <Input placeholder="ghcr.io/acme/user-service:v1" />
        </Form.Item>
        <Space style={{ width: '100%' }} align="start">
          <Form.Item label="环境" name="env" rules={[{ required: true }]} style={{ minWidth: 180 }}>
            <Select options={[{ value: 'production', label: 'production' }, { value: 'staging', label: 'staging' }, { value: 'development', label: 'development' }]} />
          </Form.Item>
          <Form.Item label="负责人" name="owner" rules={[{ required: true }]} style={{ minWidth: 180 }}>
            <Input placeholder="ops" />
          </Form.Item>
          <Form.Item label="副本" name="replicas" style={{ minWidth: 120 }}>
            <InputNumber min={1} />
          </Form.Item>
        </Space>
        <Space style={{ width: '100%' }} align="start">
          <Form.Item label="CPU(m)" name="cpuLimit" style={{ minWidth: 120 }}><InputNumber min={100} /></Form.Item>
          <Form.Item label="内存(MB)" name="memLimit" style={{ minWidth: 140 }}><InputNumber min={128} /></Form.Item>
          <Form.Item label="标签" name="tags" style={{ minWidth: 280 }}><Select mode="tags" /></Form.Item>
        </Space>
        <Form.Item label="配置" name="config">
          <Input.TextArea rows={8} placeholder="service:\n  port: 8080" />
        </Form.Item>
        <Button type="primary" htmlType="submit" loading={loading}>创建</Button>
      </Form>
    </Card>
  );
};

export default ServiceProvisionPage;

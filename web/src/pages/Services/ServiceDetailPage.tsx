import React from 'react';
import { Button, Card, Descriptions, Space, Table, Tag, message } from 'antd';
import { ArrowLeftOutlined, RollbackOutlined, CloudUploadOutlined, ReloadOutlined } from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import { Api } from '../../api';
import type { ServiceEvent, ServiceItem } from '../../api/modules/services';

const ServiceDetailPage: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const [loading, setLoading] = React.useState(false);
  const [service, setService] = React.useState<ServiceItem | null>(null);
  const [events, setEvents] = React.useState<ServiceEvent[]>([]);

  const load = React.useCallback(async () => {
    if (!id) return;
    setLoading(true);
    try {
      const [detail, eventRes] = await Promise.all([
        Api.services.getDetail(id),
        Api.services.getEvents(id),
      ]);
      setService(detail.data);
      setEvents(eventRes.data.list || []);
    } catch (err) {
      message.error(err instanceof Error ? err.message : '加载服务详情失败');
    } finally {
      setLoading(false);
    }
  }, [id]);

  React.useEffect(() => {
    load();
  }, [load]);

  const deploy = async () => {
    if (!id) return;
    await Api.services.deploy(id, { source: 'ui' });
    message.success('部署已触发');
    load();
  };

  const rollback = async () => {
    if (!id) return;
    await Api.services.rollback(id);
    message.success('回滚已触发');
    load();
  };

  return (
    <div className="space-y-4">
      <Space>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/services')}>返回</Button>
        <Button icon={<ReloadOutlined />} onClick={load} loading={loading}>刷新</Button>
        <Button type="primary" icon={<CloudUploadOutlined />} onClick={deploy}>部署</Button>
        <Button icon={<RollbackOutlined />} onClick={rollback}>回滚</Button>
      </Space>

      <Card title="服务详情" loading={loading}>
        {service && (
          <Descriptions bordered column={2}>
            <Descriptions.Item label="名称">{service.name}</Descriptions.Item>
            <Descriptions.Item label="环境"><Tag>{service.env}</Tag></Descriptions.Item>
            <Descriptions.Item label="状态"><Tag color={service.status === 'running' ? 'success' : 'warning'}>{service.status}</Tag></Descriptions.Item>
            <Descriptions.Item label="负责人">{service.owner}</Descriptions.Item>
            <Descriptions.Item label="镜像" span={2}><code>{service.image}</code></Descriptions.Item>
            <Descriptions.Item label="副本">{service.replicas}</Descriptions.Item>
            <Descriptions.Item label="CPU">{service.cpuLimit}m</Descriptions.Item>
            <Descriptions.Item label="内存">{service.memLimit}MB</Descriptions.Item>
            <Descriptions.Item label="标签" span={2}>{service.tags.join(', ') || '-'}</Descriptions.Item>
          </Descriptions>
        )}
      </Card>

      <Card title="事件">
        <Table
          rowKey="id"
          dataSource={events}
          columns={[
            { title: '时间', dataIndex: 'createdAt', render: (v: string) => (v ? new Date(v).toLocaleString() : '-') },
            { title: '类型', dataIndex: 'type', render: (v: string) => <Tag>{v}</Tag> },
            { title: '级别', dataIndex: 'level', render: (v: string) => <Tag color={v === 'warning' ? 'warning' : v === 'error' ? 'error' : 'blue'}>{v}</Tag> },
            { title: '内容', dataIndex: 'message' },
          ]}
          pagination={false}
        />
      </Card>
    </div>
  );
};

export default ServiceDetailPage;

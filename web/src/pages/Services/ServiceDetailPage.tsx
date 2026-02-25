import React from 'react';
import { Button, Card, Descriptions, Form, Input, Select, Space, Table, Tag, message } from 'antd';
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
  const [helmOutput, setHelmOutput] = React.useState('');
  const [helmLoading, setHelmLoading] = React.useState(false);
  const [deployTarget, setDeployTarget] = React.useState<'k8s' | 'compose' | 'helm'>('k8s');
  const [form] = Form.useForm();

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
      if (detail.data.runtimeType) {
        setDeployTarget(detail.data.runtimeType);
      }
    } catch (err) {
      message.error(err instanceof Error ? err.message : '加载服务详情失败');
    } finally {
      setLoading(false);
    }
  }, [id]);

  React.useEffect(() => {
    void load();
  }, [load]);

  const deploy = async () => {
    if (!id) return;
    await Api.services.deploy(id, { deploy_target: deployTarget, cluster_id: Number(localStorage.getItem('clusterId') || 1) });
    message.success('部署已触发');
    await load();
  };

  const rollback = async () => {
    if (!id) return;
    await Api.services.rollback(id);
    message.success('回滚已触发');
    await load();
  };

  const helmImport = async () => {
    if (!id) return;
    const v = await form.validateFields();
    await Api.services.helmImport({
      service_id: Number(id),
      chart_name: v.chart_name,
      chart_version: v.chart_version,
      chart_ref: v.chart_ref,
      values_yaml: v.values_yaml,
    });
    message.success('Helm chart 已导入');
  };

  const helmRender = async () => {
    const v = await form.validateFields(['chart_name', 'chart_ref']);
    setHelmLoading(true);
    try {
      const res = await Api.services.helmRender({ chart_name: v.chart_name, chart_ref: v.chart_ref, values_yaml: form.getFieldValue('values_yaml') || '' });
      setHelmOutput(res.data.rendered_yaml || '');
      message.success('Helm 渲染完成');
    } catch (err) {
      message.error(err instanceof Error ? err.message : 'Helm 渲染失败');
    } finally {
      setHelmLoading(false);
    }
  };

  return (
    <div className="space-y-4">
      <Space>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/services')}>返回</Button>
        <Button icon={<ReloadOutlined />} onClick={load} loading={loading}>刷新</Button>
        <Select value={deployTarget} onChange={(v) => setDeployTarget(v)} style={{ width: 120 }} options={[{ value: 'k8s' }, { value: 'compose' }, { value: 'helm' }]} />
        <Button type="primary" icon={<CloudUploadOutlined />} onClick={deploy}>部署</Button>
        <Button icon={<RollbackOutlined />} onClick={rollback}>回滚</Button>
      </Space>

      <Card title="服务详情" loading={loading}>
        {service && (
          <Descriptions bordered column={2}>
            <Descriptions.Item label="名称">{service.name}</Descriptions.Item>
            <Descriptions.Item label="环境"><Tag>{service.env}</Tag></Descriptions.Item>
            <Descriptions.Item label="运行时"><Tag color="blue">{service.runtimeType}</Tag></Descriptions.Item>
            <Descriptions.Item label="配置模式"><Tag>{service.configMode}</Tag></Descriptions.Item>
            <Descriptions.Item label="状态"><Tag color={service.status === 'running' ? 'success' : 'warning'}>{service.status}</Tag></Descriptions.Item>
            <Descriptions.Item label="负责人">{service.owner}</Descriptions.Item>
            <Descriptions.Item label="项目/团队">{service.projectId} / {service.teamId}</Descriptions.Item>
            <Descriptions.Item label="服务分类">{service.serviceKind}</Descriptions.Item>
            <Descriptions.Item label="标签" span={2}>{(service.labels || []).map((x) => `${x.key}=${x.value}`).join(', ') || '-'}</Descriptions.Item>
            <Descriptions.Item label="渲染产物" span={2}><pre style={{ maxHeight: 220, overflow: 'auto' }}>{service.renderedYaml || service.customYaml || '-'}</pre></Descriptions.Item>
          </Descriptions>
        )}
      </Card>

      <Card title="Helm 导入与渲染">
        <Form form={form} layout="vertical" initialValues={{ chart_name: service?.name || 'release', chart_version: '', chart_ref: '', values_yaml: '' }}>
          <Form.Item name="chart_name" label="Chart Name" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="chart_version" label="Chart Version"><Input /></Form.Item>
          <Form.Item name="chart_ref" label="Chart Ref/Path" rules={[{ required: true }]}><Input placeholder="./charts/my-app 或 oci://..." /></Form.Item>
          <Form.Item name="values_yaml" label="Values YAML"><Input.TextArea rows={8} /></Form.Item>
          <Space>
            <Button onClick={helmImport}>导入</Button>
            <Button loading={helmLoading} onClick={helmRender}>渲染</Button>
            {id ? <Button onClick={() => Api.services.deployHelm(id).then(() => message.success('Helm 部署状态已提交'))}>部署 Helm</Button> : null}
          </Space>
        </Form>
        <pre style={{ marginTop: 12, maxHeight: 280, overflow: 'auto' }}>{helmOutput || '# 暂无 Helm 渲染输出'}</pre>
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

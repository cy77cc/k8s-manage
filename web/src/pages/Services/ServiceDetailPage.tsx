import React from 'react';
import { Alert, Button, Card, Col, Descriptions, Form, Input, InputNumber, Row, Select, Space, Table, Tabs, Tag, message } from 'antd';
import { ArrowLeftOutlined, CloudUploadOutlined, ReloadOutlined, SaveOutlined } from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import { Api } from '../../api';
import type { ServiceEvent, ServiceItem, ServiceReleaseRecord, ServiceRevision, TemplateVar, VariableValueSet } from '../../api/modules/services';

const ServiceDetailPage: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const [loading, setLoading] = React.useState(false);
  const [service, setService] = React.useState<ServiceItem | null>(null);
  const [events, setEvents] = React.useState<ServiceEvent[]>([]);
  const [revisions, setRevisions] = React.useState<ServiceRevision[]>([]);
  const [releases, setReleases] = React.useState<ServiceReleaseRecord[]>([]);
  const [varSchema, setVarSchema] = React.useState<TemplateVar[]>([]);
  const [varSet, setVarSet] = React.useState<VariableValueSet | null>(null);
  const [previewYAML, setPreviewYAML] = React.useState('');
  const [previewWarnings, setPreviewWarnings] = React.useState<Array<{ level: string; code: string; message: string }>>([]);
  const [deploying, setDeploying] = React.useState(false);
  const [targetForm] = Form.useForm();
  const [varForm] = Form.useForm();

  const env = Form.useWatch('env', varForm) || 'staging';

  const load = React.useCallback(async () => {
    if (!id) return;
    setLoading(true);
    try {
      const [detail, eventRes, revRes, relRes, schemaRes] = await Promise.all([
        Api.services.getDetail(id),
        Api.services.getEvents(id),
        Api.services.listRevisions(id),
        Api.services.listReleases(id),
        Api.services.getVariableSchema(id),
      ]);
      setService(detail.data);
      setEvents(eventRes.data.list || []);
      setRevisions(revRes.data.list || []);
      setReleases(relRes.data.list || []);
      setVarSchema(schemaRes.data.vars || []);
      targetForm.setFieldsValue({
        cluster_id: Number(localStorage.getItem('clusterId') || 1),
        namespace: 'default',
        deploy_target: detail.data.runtimeType || 'k8s',
      });
      varForm.setFieldValue('env', detail.data.env || 'staging');
    } catch (err) {
      message.error(err instanceof Error ? err.message : '加载服务详情失败');
    } finally {
      setLoading(false);
    }
  }, [id, targetForm, varForm]);

  const loadVarSet = React.useCallback(async () => {
    if (!id) return;
    try {
      const resp = await Api.services.getVariableValues(id, env);
      setVarSet(resp.data);
      const values: Record<string, any> = { env };
      Object.entries(resp.data.values || {}).forEach(([k, v]) => { values[`var_${k}`] = v; });
      varForm.setFieldsValue(values);
    } catch (err) {
      message.error(err instanceof Error ? err.message : '加载变量集失败');
    }
  }, [id, env, varForm]);

  React.useEffect(() => {
    void load();
  }, [load]);

  React.useEffect(() => {
    if (!service || !id) return;
    void loadVarSet();
  }, [service, id, loadVarSet]);

  const saveDeployTarget = async () => {
    if (!id) return;
    const v = await targetForm.validateFields();
    await Api.services.upsertDeployTarget(id, {
      cluster_id: Number(v.cluster_id),
      namespace: v.namespace,
      deploy_target: v.deploy_target,
      policy: {},
    });
    message.success('默认部署目标已保存');
  };

  const saveVarValues = async () => {
    if (!id) return;
    const values = await varForm.validateFields();
    const envVal = values.env;
    const vars: Record<string, string> = {};
    Object.keys(values).forEach((k) => {
      if (k.startsWith('var_') && String(values[k] || '').trim() !== '') {
        vars[k.replace(/^var_/, '')] = String(values[k]);
      }
    });
    const resp = await Api.services.upsertVariableValues(id, {
      env: envVal,
      values: vars,
      secret_keys: [],
    });
    setVarSet(resp.data);
    message.success('变量集已保存');
  };

  const doDeployPreview = async () => {
    if (!id) return;
    const target = await targetForm.validateFields();
    const values = varForm.getFieldsValue(true);
    const vars: Record<string, string> = {};
    Object.keys(values).forEach((k) => {
      if (k.startsWith('var_') && String(values[k] || '').trim() !== '') {
        vars[k.replace(/^var_/, '')] = String(values[k]);
      }
    });
    const resp = await Api.services.deployPreview(id, {
      env: values.env || service?.env,
      cluster_id: Number(target.cluster_id),
      namespace: target.namespace,
      deploy_target: target.deploy_target,
      variables: vars,
    });
    setPreviewYAML(resp.data.resolved_yaml || '');
    setPreviewWarnings(resp.data.warnings || []);
  };

  const deploy = async () => {
    if (!id) return;
    setDeploying(true);
    try {
      const target = await targetForm.validateFields();
      const values = varForm.getFieldsValue(true);
      const vars: Record<string, string> = {};
      Object.keys(values).forEach((k) => {
        if (k.startsWith('var_') && String(values[k] || '').trim() !== '') {
          vars[k.replace(/^var_/, '')] = String(values[k]);
        }
      });
      const resp = await Api.services.deploy(id, {
        env: values.env || service?.env,
        cluster_id: Number(target.cluster_id),
        namespace: target.namespace,
        deploy_target: target.deploy_target,
        variables: vars,
      });
      message.success(`部署已触发，release #${resp.data.release_record_id}`);
      await load();
    } catch (err) {
      message.error(err instanceof Error ? err.message : '部署失败');
    } finally {
      setDeploying(false);
    }
  };

  const createRevision = async () => {
    if (!id || !service) return;
    await Api.services.createRevision(id, {
      config_mode: service.configMode,
      render_target: (service.runtimeType === 'helm' ? 'k8s' : service.runtimeType) as 'k8s' | 'compose',
      standard_config: service.standardConfig,
      custom_yaml: service.customYaml,
      variable_schema: varSchema,
    });
    message.success('新 revision 已创建');
    await load();
  };

  return (
    <div className="space-y-4">
      <Space>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/services')}>返回</Button>
        <Button icon={<ReloadOutlined />} onClick={() => void load()} loading={loading}>刷新</Button>
        <Button icon={<SaveOutlined />} onClick={createRevision}>创建 Revision</Button>
        <Button icon={<CloudUploadOutlined />} type="primary" loading={deploying} onClick={deploy}>Deploy</Button>
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
            <Descriptions.Item label="模板引擎">{service.templateEngineVersion || 'v1'}</Descriptions.Item>
            <Descriptions.Item label="最新Revision">{service.lastRevisionId || '-'}</Descriptions.Item>
            <Descriptions.Item label="标签" span={2}>{(service.labels || []).map((x) => `${x.key}=${x.value}`).join(', ') || '-'}</Descriptions.Item>
          </Descriptions>
        )}
      </Card>

      <Row gutter={16}>
        <Col span={10}>
          <Card title="部署目标">
            <Form form={targetForm} layout="vertical">
              <Form.Item name="cluster_id" label="Cluster ID" rules={[{ required: true }]}><InputNumber min={1} style={{ width: '100%' }} /></Form.Item>
              <Form.Item name="namespace" label="Namespace" rules={[{ required: true }]}><Input /></Form.Item>
              <Form.Item name="deploy_target" label="Deploy Target"><Select options={[{ value: 'k8s' }, { value: 'compose' }, { value: 'helm' }]} /></Form.Item>
              <Space>
                <Button onClick={saveDeployTarget}>保存默认目标</Button>
                <Button onClick={doDeployPreview}>Deploy Preview</Button>
              </Space>
            </Form>
          </Card>

          <Card title="环境变量集" style={{ marginTop: 16 }}>
            <Form form={varForm} layout="vertical">
              <Form.Item name="env" label="环境"><Select options={[{ value: 'development' }, { value: 'staging' }, { value: 'production' }]} /></Form.Item>
              {varSchema.map((v) => (
                <Form.Item key={v.name} name={`var_${v.name}`} label={`${v.name}${v.required ? ' *' : ''}`}>
                  <Input placeholder={v.default || ''} />
                </Form.Item>
              ))}
              <Button onClick={saveVarValues}>保存变量集</Button>
            </Form>
            {varSet?.updated_at ? <Alert type="info" style={{ marginTop: 8 }} message={`最近更新时间: ${new Date(varSet.updated_at).toLocaleString()}`} /> : null}
          </Card>
        </Col>
        <Col span={14}>
          <Card title="Deploy Preview / 渲染输出">
            {previewWarnings.length > 0 ? (
              <Space direction="vertical" style={{ width: '100%', marginBottom: 8 }}>
                {previewWarnings.map((w, idx) => <Alert key={`${w.code}-${idx}`} type={w.level === 'error' ? 'error' : 'warning'} showIcon message={w.message} />)}
              </Space>
            ) : null}
            <pre style={{ maxHeight: 460, overflow: 'auto' }}>{previewYAML || service?.renderedYaml || service?.customYaml || '# 暂无输出'}</pre>
          </Card>
        </Col>
      </Row>

      <Card title="版本与发布">
        <Tabs items={[
          {
            key: 'revisions',
            label: `Revisions (${revisions.length})`,
            children: (
              <Table
                rowKey="id"
                dataSource={revisions}
                pagination={false}
                columns={[
                  { title: '#', dataIndex: 'revision_no' },
                  { title: '模式', dataIndex: 'config_mode', render: (v: string) => <Tag>{v}</Tag> },
                  { title: '目标', dataIndex: 'render_target', render: (v: string) => <Tag color="blue">{v}</Tag> },
                  { title: '创建人', dataIndex: 'created_by' },
                  { title: '时间', dataIndex: 'created_at', render: (v: string) => new Date(v).toLocaleString() },
                ]}
              />
            ),
          },
          {
            key: 'releases',
            label: `Releases (${releases.length})`,
            children: (
              <Table
                rowKey="id"
                dataSource={releases}
                pagination={false}
                columns={[
                  { title: '#', dataIndex: 'id' },
                  { title: '环境', dataIndex: 'env', render: (v: string) => <Tag>{v}</Tag> },
                  { title: '目标', dataIndex: 'deploy_target', render: (v: string) => <Tag color="blue">{v}</Tag> },
                  { title: '集群/命名空间', render: (_: any, r: ServiceReleaseRecord) => `${r.cluster_id} / ${r.namespace}` },
                  { title: '状态', dataIndex: 'status', render: (v: string) => <Tag color={v === 'succeeded' ? 'success' : v === 'failed' ? 'error' : 'processing'}>{v}</Tag> },
                  { title: '时间', dataIndex: 'created_at', render: (v: string) => new Date(v).toLocaleString() },
                ]}
              />
            ),
          },
          {
            key: 'events',
            label: `Events (${events.length})`,
            children: (
              <Table
                rowKey="id"
                dataSource={events}
                pagination={false}
                columns={[
                  { title: '时间', dataIndex: 'createdAt', render: (v: string) => (v ? new Date(v).toLocaleString() : '-') },
                  { title: '类型', dataIndex: 'type', render: (v: string) => <Tag>{v}</Tag> },
                  { title: '级别', dataIndex: 'level', render: (v: string) => <Tag color={v === 'warning' ? 'warning' : v === 'error' ? 'error' : 'blue'}>{v}</Tag> },
                  { title: '内容', dataIndex: 'message' },
                ]}
              />
            ),
          },
        ]} />
      </Card>
    </div>
  );
};

export default ServiceDetailPage;

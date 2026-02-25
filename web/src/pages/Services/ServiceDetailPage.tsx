import React from 'react';
import {
  Alert,
  Button,
  Card,
  Col,
  Descriptions,
  Form,
  Input,
  InputNumber,
  Modal,
  Row,
  Select,
  Space,
  Table,
  Tabs,
  Tag,
  message,
} from 'antd';
import { ArrowLeftOutlined, CloudUploadOutlined, EditOutlined, ReloadOutlined, SaveOutlined } from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import { Api } from '../../api';
import type {
  LabelKV,
  ServiceCreateParams,
  ServiceEvent,
  ServiceItem,
  ServiceReleaseRecord,
  ServiceRevision,
  StandardServiceConfig,
  TemplateVar,
  VariableValueSet,
} from '../../api/modules/services';

type ServiceEditFormValues = {
  name: string;
  env: string;
  owner: string;
  service_kind: string;
  service_type: 'stateless' | 'stateful';
  runtime_type: 'k8s' | 'compose' | 'helm';
  config_mode: 'standard' | 'custom';
  render_target: 'k8s' | 'compose' | 'helm';
  status: string;
  labels_text?: string;
  standard_config_text?: string;
  custom_yaml?: string;
};

const parseLabelText = (text: string): LabelKV[] => {
  const rows = String(text || '').split('\n').map((x) => x.trim()).filter(Boolean);
  const list: LabelKV[] = [];
  rows.forEach((row) => {
    const idx = row.indexOf('=');
    if (idx <= 0) {
      return;
    }
    const key = row.slice(0, idx).trim();
    const value = row.slice(idx + 1).trim();
    if (key) {
      list.push({ key, value });
    }
  });
  return list;
};

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
  const [editOpen, setEditOpen] = React.useState(false);
  const [editSaving, setEditSaving] = React.useState(false);
  const [targetForm] = Form.useForm();
  const [varForm] = Form.useForm();
  const [editForm] = Form.useForm<ServiceEditFormValues>();

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

  const openEdit = () => {
    if (!service) return;
    editForm.setFieldsValue({
      name: service.name,
      env: service.env,
      owner: service.owner,
      service_kind: service.serviceKind,
      service_type: service.serviceType || 'stateless',
      runtime_type: service.runtimeType,
      config_mode: service.configMode,
      render_target: service.renderTarget || (service.runtimeType === 'helm' ? 'k8s' : service.runtimeType),
      status: service.status,
      labels_text: (service.labels || []).map((x) => `${x.key}=${x.value}`).join('\n'),
      standard_config_text: service.standardConfig ? JSON.stringify(service.standardConfig, null, 2) : '{\n  "image": "",\n  "replicas": 1,\n  "ports": [],\n  "envs": []\n}',
      custom_yaml: service.customYaml || '',
    });
    setEditOpen(true);
  };

  const saveServiceEdit = async () => {
    if (!id || !service) return;
    const values = await editForm.validateFields();

    let standardConfig: StandardServiceConfig | undefined;
    if (values.config_mode === 'standard') {
      try {
        standardConfig = JSON.parse(values.standard_config_text || '{}') as StandardServiceConfig;
      } catch {
        message.error('标准配置 JSON 格式错误');
        return;
      }
    }

    const payload: Partial<ServiceCreateParams> = {
      name: values.name,
      env: values.env,
      owner: values.owner,
      service_kind: values.service_kind,
      service_type: values.service_type,
      runtime_type: values.runtime_type,
      config_mode: values.config_mode,
      render_target: (values.render_target === 'helm' ? 'k8s' : values.render_target) as 'k8s' | 'compose',
      status: values.status,
      labels: parseLabelText(values.labels_text || ''),
      standard_config: values.config_mode === 'standard' ? standardConfig : undefined,
      custom_yaml: values.config_mode === 'custom' ? (values.custom_yaml || '') : '',
    };

    setEditSaving(true);
    try {
      await Api.services.update(id, payload);
      message.success('服务配置已更新');
      setEditOpen(false);
      await load();
      Modal.confirm({
        title: '是否基于最新配置创建 Revision？',
        okText: '创建 Revision',
        cancelText: '稍后再说',
        onOk: () => createRevision(),
      });
    } catch (err) {
      message.error(err instanceof Error ? err.message : '保存服务配置失败');
    } finally {
      setEditSaving(false);
    }
  };

  return (
    <div className="space-y-4">
      <Space>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/services')}>返回</Button>
        <Button icon={<ReloadOutlined />} onClick={() => void load()} loading={loading}>刷新</Button>
        <Button icon={<EditOutlined />} onClick={openEdit}>编辑服务配置</Button>
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

      <Modal
        title="编辑服务配置"
        open={editOpen}
        onCancel={() => setEditOpen(false)}
        onOk={() => void saveServiceEdit()}
        okText="保存"
        confirmLoading={editSaving}
        width={860}
      >
        <Form form={editForm} layout="vertical">
          <Row gutter={12}>
            <Col span={12}>
              <Form.Item name="name" label="服务名" rules={[{ required: true, message: '请输入服务名' }]}>
                <Input />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="env" label="环境" rules={[{ required: true }]}>
                <Select options={[{ value: 'development' }, { value: 'staging' }, { value: 'production' }]} />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={12}>
            <Col span={12}>
              <Form.Item name="owner" label="负责人" rules={[{ required: true, message: '请输入负责人' }]}>
                <Input />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="status" label="状态" rules={[{ required: true }]}>
                <Select options={[{ value: 'draft' }, { value: 'running' }, { value: 'stopped' }, { value: 'error' }]} />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={12}>
            <Col span={8}>
              <Form.Item name="service_kind" label="服务分类" rules={[{ required: true }]}>
                <Input placeholder="web/worker/job" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="service_type" label="服务类型" rules={[{ required: true }]}>
                <Select options={[{ value: 'stateless' }, { value: 'stateful' }]} />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="runtime_type" label="运行时" rules={[{ required: true }]}>
                <Select options={[{ value: 'k8s' }, { value: 'compose' }, { value: 'helm' }]} />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={12}>
            <Col span={8}>
              <Form.Item name="config_mode" label="配置模式" rules={[{ required: true }]}>
                <Select options={[{ value: 'standard' }, { value: 'custom' }]} />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="render_target" label="渲染目标" rules={[{ required: true }]}>
                <Select options={[{ value: 'k8s' }, { value: 'compose' }, { value: 'helm' }]} />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item name="labels_text" label="标签（每行 key=value）">
            <Input.TextArea rows={3} placeholder={'app=api\nteam=platform'} />
          </Form.Item>

          <Form.Item noStyle shouldUpdate={(prev, next) => prev.config_mode !== next.config_mode}>
            {({ getFieldValue }) => (
              getFieldValue('config_mode') === 'standard' ? (
                <Form.Item name="standard_config_text" label="标准配置（JSON）" rules={[{ required: true, message: '请输入标准配置 JSON' }]}> 
                  <Input.TextArea rows={10} />
                </Form.Item>
              ) : (
                <Form.Item name="custom_yaml" label="自定义 YAML" rules={[{ required: true, message: '请输入 YAML 配置' }]}> 
                  <Input.TextArea rows={10} />
                </Form.Item>
              )
            )}
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default ServiceDetailPage;

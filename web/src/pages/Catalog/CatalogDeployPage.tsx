import React, { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Alert,
  Button,
  Card,
  Form,
  Input,
  InputNumber,
  Radio,
  Select,
  Space,
  Spin,
  Switch,
  Typography,
  message,
} from 'antd';
import { useNavigate, useParams } from 'react-router-dom';
import { Api } from '../../api';
import type { CatalogTemplate, CatalogVariableSchema } from '../../api/modules/catalog';

const { Title, Paragraph } = Typography;

type VariableFormValue = string | number | boolean;

const buildVariableInput = (schema: CatalogVariableSchema) => {
  switch (schema.type) {
    case 'number':
      return <InputNumber className="w-full" />;
    case 'password':
      return <Input.Password />;
    case 'boolean':
      return <Switch />;
    case 'select':
      return <Select options={(schema.options || []).map((item) => ({ label: item, value: item }))} />;
    case 'textarea':
      return <Input.TextArea autoSize={{ minRows: 3, maxRows: 6 }} />;
    default:
      return <Input />;
  }
};

const CatalogDeployPage: React.FC = () => {
  const params = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const [template, setTemplate] = useState<CatalogTemplate | null>(null);
  const [loading, setLoading] = useState(false);
  const [previewing, setPreviewing] = useState(false);
  const [deploying, setDeploying] = useState(false);
  const [previewYAML, setPreviewYAML] = useState('');
  const [unresolved, setUnresolved] = useState<string[]>([]);

  const load = useCallback(async () => {
    if (!params.id) {
      return;
    }
    setLoading(true);
    try {
      const resp = await Api.catalog.getTemplate(Number(params.id));
      setTemplate(resp.data);

      const initialVars = (resp.data.variables_schema || []).reduce<Record<string, VariableFormValue>>((acc, item) => {
        if (item.default !== undefined) {
          acc[item.name] = item.default as VariableFormValue;
        }
        return acc;
      }, {});
      form.setFieldsValue({
        target: resp.data.compose_template ? 'compose' : 'k8s',
        service_name: `${resp.data.name}-instance`,
        project_id: Number(localStorage.getItem('projectId') || '0') || undefined,
        environment: 'staging',
        variables: initialVars,
      });
    } catch (err) {
      message.error(err instanceof Error ? err.message : '加载部署信息失败');
    } finally {
      setLoading(false);
    }
  }, [params.id, form]);

  useEffect(() => {
    void load();
  }, [load]);

  const variableSchema = useMemo(() => template?.variables_schema || [], [template]);

  const handlePreview = async () => {
    if (!template) {
      return;
    }
    const values = await form.validateFields();
    setPreviewing(true);
    try {
      const resp = await Api.catalog.preview({
        template_id: template.id,
        target: values.target,
        variables: values.variables || {},
      });
      setPreviewYAML(resp.data.rendered_yaml || '');
      setUnresolved(resp.data.unresolved_vars || []);
      if ((resp.data.unresolved_vars || []).length > 0) {
        message.warning('仍有变量未填充，请检查后部署');
      }
    } catch (err) {
      message.error(err instanceof Error ? err.message : '预览失败');
    } finally {
      setPreviewing(false);
    }
  };

  const handleDeploy = async () => {
    if (!template) {
      return;
    }
    const values = await form.validateFields();
    setDeploying(true);
    try {
      const resp = await Api.catalog.deploy({
        template_id: template.id,
        target: values.target,
        project_id: Number(values.project_id),
        team_id: values.team_id ? Number(values.team_id) : undefined,
        service_name: values.service_name,
        namespace: values.namespace,
        cluster_id: values.cluster_id ? Number(values.cluster_id) : undefined,
        environment: values.environment,
        variables: values.variables || {},
        deploy_now: true,
      });
      message.success(`部署成功，服务 ID: ${resp.data.service_id}`);
      navigate(`/services/${resp.data.service_id}`);
    } catch (err) {
      message.error(err instanceof Error ? err.message : '部署失败');
    } finally {
      setDeploying(false);
    }
  };

  return (
    <div className="p-6 space-y-4">
      <Title level={3} className="!mb-1">模板部署</Title>
      <Paragraph className="!mb-0 text-gray-500">根据变量定义生成部署参数，先预览再部署。</Paragraph>

      <Spin spinning={loading}>
        {!template ? null : (
          <Form layout="vertical" form={form}>
            <Card title="部署目标" className="mb-4">
              <Form.Item name="target" label="部署目标" rules={[{ required: true }]}>
                <Radio.Group>
                  <Radio.Button value="k8s">K8s 集群</Radio.Button>
                  <Radio.Button value="compose">Compose 环境</Radio.Button>
                </Radio.Group>
              </Form.Item>
              <Space size={16} wrap>
                <Form.Item name="project_id" label="Project ID" rules={[{ required: true }]}>
                  <Input placeholder="例如 1" style={{ width: 160 }} />
                </Form.Item>
                <Form.Item name="team_id" label="Team ID">
                  <Input placeholder="可选" style={{ width: 160 }} />
                </Form.Item>
                <Form.Item name="cluster_id" label="Cluster ID">
                  <Input placeholder="K8s 可选" style={{ width: 160 }} />
                </Form.Item>
                <Form.Item name="namespace" label="Namespace">
                  <Input placeholder="例如 default" style={{ width: 200 }} />
                </Form.Item>
                <Form.Item name="environment" label="环境">
                  <Select style={{ width: 160 }} options={[
                    { label: '开发', value: 'dev' },
                    { label: '测试', value: 'staging' },
                    { label: '生产', value: 'production' },
                  ]} />
                </Form.Item>
              </Space>
              <Form.Item name="service_name" label="服务名称" rules={[{ required: true }]}>
                <Input placeholder="请输入服务名称" />
              </Form.Item>
            </Card>

            <Card title="变量配置" className="mb-4">
              {variableSchema.map((schema) => (
                <Form.Item
                  key={schema.name}
                  name={['variables', schema.name]}
                  label={`${schema.name} (${schema.type})`}
                  rules={schema.required ? [{ required: true, message: '必填字段' }] : undefined}
                  extra={schema.description || ''}
                  valuePropName={schema.type === 'boolean' ? 'checked' : 'value'}
                >
                  {buildVariableInput(schema)}
                </Form.Item>
              ))}
            </Card>

            <Card title="YAML 预览" className="mb-4">
              {unresolved.length > 0 && (
                <Alert
                  type="warning"
                  showIcon
                  className="mb-3"
                  message={`未填充变量: ${unresolved.join(', ')}`}
                />
              )}
              <Input.TextArea
                value={previewYAML}
                readOnly
                autoSize={{ minRows: 12, maxRows: 24 }}
                placeholder="点击“预览 YAML”生成渲染结果"
              />
            </Card>

            <Space>
              <Button onClick={handlePreview} loading={previewing}>预览 YAML</Button>
              <Button type="primary" onClick={handleDeploy} loading={deploying}>确认部署</Button>
            </Space>
          </Form>
        )}
      </Spin>
    </div>
  );
};

export default CatalogDeployPage;

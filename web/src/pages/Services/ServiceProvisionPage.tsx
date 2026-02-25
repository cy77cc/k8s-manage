import React from 'react';
import { Button, Card, Col, Form, Input, InputNumber, Row, Select, Space, Typography, message } from 'antd';
import { ArrowLeftOutlined, SwapOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { Api } from '../../api';
import type { LabelKV, StandardServiceConfig } from '../../api/modules/services';

const { Text } = Typography;

const ServiceProvisionPage: React.FC = () => {
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const [loading, setLoading] = React.useState(false);
  const [previewing, setPreviewing] = React.useState(false);
  const [previewYAML, setPreviewYAML] = React.useState('');
  const [diagnostics, setDiagnostics] = React.useState<Array<{ level: string; code: string; message: string }>>([]);

  const mode = Form.useWatch('config_mode', form) || 'standard';
  const target = Form.useWatch('render_target', form) || 'k8s';
  const watchName = Form.useWatch('name', form);
  const watchImage = Form.useWatch('image', form);
  const watchReplicas = Form.useWatch('replicas', form);
  const watchServicePort = Form.useWatch('service_port', form);
  const watchContainerPort = Form.useWatch('container_port', form);
  const watchCPU = Form.useWatch('cpu', form);
  const watchMemory = Form.useWatch('memory', form);
  const watchCustomYAML = Form.useWatch('custom_yaml', form);

  const toLabels = (raw: string[]): LabelKV[] =>
    (raw || [])
      .map((x) => x.trim())
      .filter(Boolean)
      .map((pair) => {
        const [key, ...rest] = pair.split('=');
        return { key: key.trim(), value: rest.join('=').trim() };
      })
      .filter((x) => x.key);

  const buildStandardConfig = (values: any): StandardServiceConfig => ({
    image: values.image,
    replicas: values.replicas || 1,
    ports: [{
      name: 'http',
      protocol: 'TCP',
      container_port: values.container_port || 8080,
      service_port: values.service_port || 80,
    }],
    envs: (values.envs || []).map((x: string) => {
      const [k, ...r] = x.split('=');
      return { key: k.trim(), value: r.join('=').trim() };
    }).filter((x: any) => x.key),
    resources: {
      cpu: values.cpu || '500m',
      memory: values.memory || '512Mi',
    },
  });

  const refreshPreview = async () => {
    const values = form.getFieldsValue(true);
    if (!values.name) {
      return;
    }
    try {
      setPreviewing(true);
      const res = await Api.services.preview({
        mode: values.config_mode,
        target: values.render_target,
        service_name: values.name,
        service_type: values.service_type,
        standard_config: values.config_mode === 'standard' ? buildStandardConfig(values) : undefined,
        custom_yaml: values.config_mode === 'custom' ? values.custom_yaml : undefined,
      });
      setPreviewYAML(res.data.rendered_yaml || '');
      setDiagnostics(res.data.diagnostics || []);
    } catch (err) {
      setPreviewYAML('');
      message.error(err instanceof Error ? err.message : '预览失败');
    } finally {
      setPreviewing(false);
    }
  };

  React.useEffect(() => {
    const timer = setTimeout(() => {
      void refreshPreview();
    }, 350);
    return () => clearTimeout(timer);
  }, [mode, target, watchName, watchImage, watchReplicas, watchServicePort, watchContainerPort, watchCPU, watchMemory, watchCustomYAML]);

  const transformToCustom = async () => {
    const values = await form.validateFields(['name', 'service_type', 'render_target', 'image']);
    const all = form.getFieldsValue(true);
    const res = await Api.services.transform({
      standard_config: buildStandardConfig(all),
      target: values.render_target,
      service_name: values.name,
      service_type: values.service_type,
    });
    form.setFieldsValue({ config_mode: 'custom', custom_yaml: res.data.custom_yaml });
    setPreviewYAML(res.data.custom_yaml);
    message.success('已转换为自定义 YAML');
  };

  const onFinish = async (values: any) => {
    setLoading(true);
    try {
      await Api.services.create({
        project_id: Number(values.project_id || localStorage.getItem('projectId') || 1),
        team_id: Number(values.team_id || localStorage.getItem('teamId') || 1),
        name: values.name,
        env: values.env,
        owner: values.owner,
        runtime_type: values.runtime_type,
        config_mode: values.config_mode,
        service_kind: values.service_kind,
        service_type: values.service_type,
        render_target: values.render_target,
        labels: toLabels(values.labels || []),
        standard_config: values.config_mode === 'standard' ? buildStandardConfig(values) : undefined,
        custom_yaml: values.config_mode === 'custom' ? values.custom_yaml : undefined,
        source_template_version: 'v1',
        status: 'draft',
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
    <Card title="创建服务" extra={<Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/services')}>返回</Button>}>
      <Form
        form={form}
        layout="vertical"
        onFinish={onFinish}
        initialValues={{
          project_id: Number(localStorage.getItem('projectId') || 1),
          team_id: Number(localStorage.getItem('teamId') || 1),
          env: 'staging',
          owner: 'system',
          runtime_type: 'k8s',
          config_mode: 'standard',
          service_kind: 'web',
          service_type: 'stateless',
          render_target: 'k8s',
          replicas: 1,
          service_port: 80,
          container_port: 8080,
          cpu: '500m',
          memory: '512Mi',
        }}
      >
        <Row gutter={16}>
          <Col span={12}><Form.Item label="项目ID" name="project_id" rules={[{ required: true }]}><InputNumber min={1} style={{ width: '100%' }} /></Form.Item></Col>
          <Col span={12}><Form.Item label="团队ID" name="team_id" rules={[{ required: true }]}><InputNumber min={1} style={{ width: '100%' }} /></Form.Item></Col>
        </Row>
        <Row gutter={16}>
          <Col span={12}><Form.Item label="服务名" name="name" rules={[{ required: true }]}><Input placeholder="user-service" /></Form.Item></Col>
          <Col span={12}><Form.Item label="负责人" name="owner" rules={[{ required: true }]}><Input /></Form.Item></Col>
        </Row>
        <Row gutter={16}>
          <Col span={6}><Form.Item label="环境" name="env"><Select options={[{ value: 'development' }, { value: 'staging' }, { value: 'production' }]} /></Form.Item></Col>
          <Col span={6}><Form.Item label="运行时" name="runtime_type"><Select options={[{ value: 'k8s' }, { value: 'compose' }, { value: 'helm' }]} /></Form.Item></Col>
          <Col span={6}><Form.Item label="配置模式" name="config_mode"><Select options={[{ value: 'standard', label: '通用配置' }, { value: 'custom', label: '自定义配置' }]} /></Form.Item></Col>
          <Col span={6}><Form.Item label="渲染目标" name="render_target"><Select options={[{ value: 'k8s' }, { value: 'compose' }]} /></Form.Item></Col>
        </Row>
        <Row gutter={16}>
          <Col span={8}><Form.Item label="服务分类" name="service_kind"><Input placeholder="web/backend/job" /></Form.Item></Col>
          <Col span={8}><Form.Item label="服务类型" name="service_type"><Select options={[{ value: 'stateless' }, { value: 'stateful' }]} /></Form.Item></Col>
          <Col span={8}><Form.Item label="标签(key=value)" name="labels"><Select mode="tags" placeholder="app=user,tier=backend" /></Form.Item></Col>
        </Row>

        {mode === 'standard' ? (
          <>
            <Row gutter={16}>
              <Col span={12}><Form.Item label="镜像" name="image" rules={[{ required: true }]}><Input placeholder="ghcr.io/org/app:v1" /></Form.Item></Col>
              <Col span={6}><Form.Item label="副本" name="replicas"><InputNumber min={1} style={{ width: '100%' }} /></Form.Item></Col>
              <Col span={6}><Form.Item label="环境变量(KEY=VALUE)" name="envs"><Select mode="tags" /></Form.Item></Col>
            </Row>
            <Row gutter={16}>
              <Col span={6}><Form.Item label="Service Port" name="service_port"><InputNumber min={1} style={{ width: '100%' }} /></Form.Item></Col>
              <Col span={6}><Form.Item label="Container Port" name="container_port"><InputNumber min={1} style={{ width: '100%' }} /></Form.Item></Col>
              <Col span={6}><Form.Item label="CPU" name="cpu"><Input placeholder="500m" /></Form.Item></Col>
              <Col span={6}><Form.Item label="Memory" name="memory"><Input placeholder="512Mi" /></Form.Item></Col>
            </Row>
            <Button icon={<SwapOutlined />} onClick={transformToCustom}>通用配置转自定义 YAML</Button>
          </>
        ) : (
          <Form.Item label="自定义 YAML" name="custom_yaml" rules={[{ required: true, message: '请输入 YAML' }]}>
            <Input.TextArea rows={14} placeholder="apiVersion: apps/v1\nkind: Deployment\n..." />
          </Form.Item>
        )}

        <Card size="small" title={previewing ? '实时预览（渲染中）' : '实时预览（渲染结果）'} style={{ marginTop: 16 }}>
          <pre style={{ maxHeight: 320, overflow: 'auto', marginBottom: 8 }}>{previewYAML || '# 暂无预览'}</pre>
          {diagnostics.length > 0 ? (
            <Space direction="vertical" size={2}>
              {diagnostics.map((d, idx) => (
                <Text key={`${d.code}-${idx}`} type={d.level === 'error' ? 'danger' : 'secondary'}>
                  [{d.level}] {d.message}
                </Text>
              ))}
            </Space>
          ) : null}
        </Card>

        <div style={{ marginTop: 16 }}>
          <Button type="primary" htmlType="submit" loading={loading}>创建服务</Button>
        </div>
      </Form>
    </Card>
  );
};

export default ServiceProvisionPage;

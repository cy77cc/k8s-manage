import React, { useEffect, useState } from 'react';
import { Alert, Button, Card, Form, Input, Select, Space, Spin, Typography, message } from 'antd';
import Editor from '@monaco-editor/react';
import { useNavigate, useParams } from 'react-router-dom';
import { Api } from '../../api';
import type { CatalogTemplate, TemplateCreatePayload } from '../../api/modules/catalog';

const { Title } = Typography;

const TemplateEditPage: React.FC = () => {
  const params = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [form] = Form.useForm<TemplateCreatePayload>();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [template, setTemplate] = useState<CatalogTemplate | null>(null);

  useEffect(() => {
    if (!params.id) {
      return;
    }
    setLoading(true);
    void Api.catalog.getTemplate(Number(params.id))
      .then((resp) => {
        setTemplate(resp.data);
        form.setFieldsValue({
          name: resp.data.name,
          display_name: resp.data.display_name,
          description: resp.data.description,
          icon: resp.data.icon,
          category_id: resp.data.category_id,
          version: resp.data.version,
          visibility: resp.data.visibility,
          k8s_template: resp.data.k8s_template,
          compose_template: resp.data.compose_template,
          variables_schema: resp.data.variables_schema,
          readme: resp.data.readme,
          tags: resp.data.tags,
        });
      })
      .catch((err) => message.error(err instanceof Error ? err.message : '加载模板失败'))
      .finally(() => setLoading(false));
  }, [params.id, form]);

  const save = async () => {
    if (!template) {
      return;
    }
    try {
      const values = await form.validateFields();
      setSaving(true);
      await Api.catalog.updateTemplate(template.id, values);
      message.success('模板更新成功');
      navigate('/catalog/my-templates');
    } catch (err) {
      if (err instanceof Error) {
        message.error(err.message);
      }
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="p-6 space-y-4">
      <Title level={3}>编辑模板</Title>
      <Spin spinning={loading}>
        {!template ? (
          <Card><Alert type="warning" showIcon message="未找到模板" /></Card>
        ) : (
          <Card>
            <Form form={form} layout="vertical">
              <Form.Item name="display_name" label="显示名称" rules={[{ required: true }]}><Input /></Form.Item>
              <Form.Item name="description" label="描述"><Input.TextArea rows={3} /></Form.Item>
              <Form.Item name="visibility" label="可见性">
                <Select options={[{ label: '私有', value: 'private' }, { label: '公开', value: 'public' }]} />
              </Form.Item>
              <Form.Item name="k8s_template" label="K8s 模板"><Editor height="280px" defaultLanguage="yaml" value={form.getFieldValue('k8s_template')} /></Form.Item>
              <Form.Item name="compose_template" label="Compose 模板"><Editor height="220px" defaultLanguage="yaml" value={form.getFieldValue('compose_template')} /></Form.Item>
              <Form.Item name="readme" label="使用说明"><Input.TextArea rows={10} /></Form.Item>
            </Form>
            <Space>
              <Button onClick={() => navigate('/catalog/my-templates')}>取消</Button>
              <Button type="primary" loading={saving} onClick={save}>保存</Button>
            </Space>
          </Card>
        )}
      </Spin>
    </div>
  );
};

export default TemplateEditPage;

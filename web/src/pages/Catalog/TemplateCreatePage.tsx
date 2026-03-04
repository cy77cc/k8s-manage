import React, { useEffect, useMemo, useState } from 'react';
import { Button, Card, Form, Input, Select, Space, Steps, Table, Typography, message } from 'antd';
import Editor from '@monaco-editor/react';
import { useNavigate } from 'react-router-dom';
import { Api } from '../../api';
import type { CatalogCategory, CatalogVariableSchema, TemplateCreatePayload } from '../../api/modules/catalog';

const { Title } = Typography;

const defaultVariable: CatalogVariableSchema = {
  name: 'example_var',
  type: 'string',
  required: false,
  description: '',
};

const TemplateCreatePage: React.FC = () => {
  const navigate = useNavigate();
  const [form] = Form.useForm<TemplateCreatePayload>();
  const [submitting, setSubmitting] = useState(false);
  const [step, setStep] = useState(0);
  const [categories, setCategories] = useState<CatalogCategory[]>([]);
  const [variables, setVariables] = useState<CatalogVariableSchema[]>([defaultVariable]);

  useEffect(() => {
    void Api.catalog.listCategories().then((resp) => setCategories(resp.data.list || []));
  }, []);

  const steps = useMemo(() => [
    { title: '基本信息' },
    { title: 'K8s 模板' },
    { title: 'Compose 模板' },
    { title: '变量定义' },
    { title: '使用说明' },
  ], []);

  const detectVariables = () => {
    const k8s = form.getFieldValue('k8s_template') || '';
    const compose = form.getFieldValue('compose_template') || '';
    const content = `${k8s}\n${compose}`;
    const matches = content.match(/\{\{\s*([a-zA-Z_][a-zA-Z0-9_.-]*)/g) || [];
    const names = Array.from(new Set(matches.map((item) => item.replace(/\{\{|\s/g, ''))));
    if (names.length === 0) {
      return;
    }
    setVariables((current) => {
      const exists = new Set(current.map((item) => item.name));
      const next = [...current];
      for (const name of names) {
        if (!exists.has(name)) {
          next.push({ name, type: 'string', required: false, description: '' });
        }
      }
      return next;
    });
  };

  const save = async () => {
    try {
      const payload = await form.validateFields();
      setSubmitting(true);
      await Api.catalog.createTemplate({
        ...payload,
        variables_schema: variables,
        tags: payload.tags || [],
        visibility: payload.visibility || 'private',
      });
      message.success('模板创建成功');
      navigate('/catalog/my-templates');
    } catch (err) {
      if (err instanceof Error) {
        message.error(err.message);
      }
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="p-6 space-y-4">
      <Title level={3}>创建模板</Title>
      <Card><Steps current={step} items={steps} /></Card>

      <Form layout="vertical" form={form} initialValues={{ visibility: 'private', version: '1.0.0', tags: [] }}>
        {step === 0 && (
          <Card title="Step 1: 基本信息" className="mb-4">
            <Form.Item name="name" label="模板标识" rules={[{ required: true }]}><Input /></Form.Item>
            <Form.Item name="display_name" label="显示名称" rules={[{ required: true }]}><Input /></Form.Item>
            <Form.Item name="description" label="描述"><Input.TextArea rows={3} /></Form.Item>
            <Form.Item name="icon" label="图标"><Input /></Form.Item>
            <Form.Item name="category_id" label="分类" rules={[{ required: true }]}>
              <Select options={categories.map((item) => ({ label: item.display_name, value: item.id }))} />
            </Form.Item>
            <Form.Item name="visibility" label="可见性">
              <Select options={[{ label: '私有', value: 'private' }, { label: '公开', value: 'public' }]} />
            </Form.Item>
            <Form.Item name="tags" label="标签">
              <Select mode="tags" tokenSeparators={[',']} />
            </Form.Item>
          </Card>
        )}

        {step === 1 && (
          <Card title="Step 2: K8s 模板编辑器" className="mb-4" extra={<Button onClick={detectVariables}>自动检测变量</Button>}>
            <Form.Item name="k8s_template" rules={[{ required: true, message: '请填写 K8s 模板' }]}>
              <Editor height="360px" defaultLanguage="yaml" />
            </Form.Item>
          </Card>
        )}

        {step === 2 && (
          <Card title="Step 3: Compose 模板编辑器 (可选)" className="mb-4" extra={<Button onClick={detectVariables}>自动检测变量</Button>}>
            <Form.Item name="compose_template">
              <Editor height="300px" defaultLanguage="yaml" />
            </Form.Item>
          </Card>
        )}

        {step === 3 && (
          <Card title="Step 4: 变量定义" className="mb-4">
            <Table
              pagination={false}
              rowKey={(row) => row.name}
              dataSource={variables}
              columns={[
                {
                  title: '变量名',
                  dataIndex: 'name',
                  render: (_, row, index) => (
                    <Input
                      value={row.name}
                      onChange={(e) => {
                        const next = [...variables];
                        next[index] = { ...next[index], name: e.target.value };
                        setVariables(next);
                      }}
                    />
                  ),
                },
                {
                  title: '类型',
                  dataIndex: 'type',
                  render: (_, row, index) => (
                    <Select
                      value={row.type}
                      options={[
                        { label: 'string', value: 'string' },
                        { label: 'number', value: 'number' },
                        { label: 'password', value: 'password' },
                        { label: 'boolean', value: 'boolean' },
                        { label: 'select', value: 'select' },
                        { label: 'textarea', value: 'textarea' },
                      ]}
                      onChange={(value) => {
                        const next = [...variables];
                        next[index] = { ...next[index], type: value };
                        setVariables(next);
                      }}
                    />
                  ),
                },
                {
                  title: '默认值',
                  dataIndex: 'default',
                  render: (_, row, index) => (
                    <Input
                      value={String(row.default ?? '')}
                      onChange={(e) => {
                        const next = [...variables];
                        next[index] = { ...next[index], default: e.target.value };
                        setVariables(next);
                      }}
                    />
                  ),
                },
                {
                  title: '必填',
                  dataIndex: 'required',
                  render: (_, row, index) => (
                    <Select
                      value={row.required ? 'yes' : 'no'}
                      options={[{ label: '是', value: 'yes' }, { label: '否', value: 'no' }]}
                      onChange={(value) => {
                        const next = [...variables];
                        next[index] = { ...next[index], required: value === 'yes' };
                        setVariables(next);
                      }}
                    />
                  ),
                },
              ]}
            />
            <Button
              className="mt-3"
              onClick={() => setVariables([...variables, { ...defaultVariable, name: `var_${variables.length + 1}` }])}
            >
              添加变量
            </Button>
          </Card>
        )}

        {step === 4 && (
          <Card title="Step 5: 使用说明 (Markdown)" className="mb-4">
            <Form.Item name="readme"><Input.TextArea rows={12} /></Form.Item>
          </Card>
        )}
      </Form>

      <Space>
        <Button disabled={step === 0} onClick={() => setStep(step - 1)}>上一步</Button>
        <Button disabled={step === steps.length - 1} onClick={() => setStep(step + 1)}>下一步</Button>
        <Button type="primary" loading={submitting} onClick={save}>保存草稿</Button>
      </Space>
    </div>
  );
};

export default TemplateCreatePage;

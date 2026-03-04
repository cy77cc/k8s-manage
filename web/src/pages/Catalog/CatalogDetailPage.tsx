import React, { useCallback, useEffect, useState } from 'react';
import { Button, Card, Descriptions, Empty, Spin, Table, Tag, Typography, message } from 'antd';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { useNavigate, useParams } from 'react-router-dom';
import { Api } from '../../api';
import type { CatalogTemplate } from '../../api/modules/catalog';

const { Title, Paragraph } = Typography;

const CatalogDetailPage: React.FC = () => {
  const params = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [template, setTemplate] = useState<CatalogTemplate | null>(null);

  const load = useCallback(async () => {
    if (!params.id) {
      return;
    }
    setLoading(true);
    try {
      const resp = await Api.catalog.getTemplate(Number(params.id));
      setTemplate(resp.data);
    } catch (err) {
      message.error(err instanceof Error ? err.message : '加载模板详情失败');
    } finally {
      setLoading(false);
    }
  }, [params.id]);

  useEffect(() => {
    void load();
  }, [load]);

  return (
    <div className="p-6 space-y-4">
      <Spin spinning={loading}>
        {!template ? (
          <Card><Empty description="模板不存在" /></Card>
        ) : (
          <>
            <Card>
              <div className="flex items-start justify-between gap-4">
                <div>
                  <Title level={3} className="!mb-1">{template.display_name}</Title>
                  <Paragraph className="!mb-2 text-gray-500">{template.description || '暂无描述'}</Paragraph>
                  <div className="flex flex-wrap gap-2">
                    {template.tags.map((tag) => <Tag key={tag}>{tag}</Tag>)}
                  </div>
                </div>
                <Button type="primary" onClick={() => navigate(`/catalog/${template.id}/deploy`)}>
                  部署
                </Button>
              </div>
            </Card>

            <Card title="模板信息">
              <Descriptions column={2}>
                <Descriptions.Item label="模板名称">{template.name}</Descriptions.Item>
                <Descriptions.Item label="版本">{template.version}</Descriptions.Item>
                <Descriptions.Item label="状态">{template.status}</Descriptions.Item>
                <Descriptions.Item label="部署次数">{template.deploy_count}</Descriptions.Item>
                <Descriptions.Item label="可见性">{template.visibility}</Descriptions.Item>
                <Descriptions.Item label="分类 ID">{template.category_id}</Descriptions.Item>
              </Descriptions>
            </Card>

            <Card title="变量定义">
              <Table
                pagination={false}
                rowKey="name"
                dataSource={template.variables_schema}
                columns={[
                  { title: '变量名', dataIndex: 'name' },
                  { title: '类型', dataIndex: 'type' },
                  { title: '默认值', dataIndex: 'default', render: (value) => String(value ?? '') },
                  { title: '必填', dataIndex: 'required', render: (value: boolean) => (value ? '是' : '否') },
                  { title: '说明', dataIndex: 'description' },
                ]}
              />
            </Card>

            <Card title="使用说明">
              <div className="prose max-w-none">
                <ReactMarkdown remarkPlugins={[remarkGfm]}>{template.readme || '暂无说明'}</ReactMarkdown>
              </div>
            </Card>
          </>
        )}
      </Spin>
    </div>
  );
};

export default CatalogDetailPage;

import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Card, Col, Empty, Input, List, Row, Segmented, Spin, Tag, Typography, message } from 'antd';
import { useNavigate } from 'react-router-dom';
import { Api } from '../../api';
import type { CatalogCategory, CatalogTemplate } from '../../api/modules/catalog';

const { Title, Paragraph } = Typography;

const CatalogListPage: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [categories, setCategories] = useState<CatalogCategory[]>([]);
  const [templates, setTemplates] = useState<CatalogTemplate[]>([]);
  const [search, setSearch] = useState('');
  const [activeCategory, setActiveCategory] = useState<number | 'all'>('all');

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const [categoryResp, templateResp] = await Promise.all([
        Api.catalog.listCategories(),
        Api.catalog.listTemplates({ status: 'published' }),
      ]);
      setCategories(categoryResp.data.list || []);
      setTemplates(templateResp.data.list || []);
    } catch (err) {
      message.error(err instanceof Error ? err.message : '加载服务目录失败');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  const categoryOptions = useMemo(() => {
    return [
      { label: '全部', value: 'all' as const },
      ...categories.map((item) => ({ label: item.display_name, value: item.id })),
    ];
  }, [categories]);

  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase();
    return templates.filter((item) => {
      const hitCategory = activeCategory === 'all' || item.category_id === activeCategory;
      if (!hitCategory) {
        return false;
      }
      if (!q) {
        return true;
      }
      return (
        item.display_name.toLowerCase().includes(q) ||
        item.name.toLowerCase().includes(q) ||
        (item.description || '').toLowerCase().includes(q) ||
        item.tags.some((tag) => tag.toLowerCase().includes(q))
      );
    });
  }, [templates, activeCategory, search]);

  return (
    <div className="p-6 space-y-4">
      <div>
        <Title level={3} className="!mb-1">服务目录</Title>
        <Paragraph className="!mb-0 text-gray-500">浏览和部署已发布的服务模板</Paragraph>
      </div>

      <Card>
        <Row gutter={[16, 16]} align="middle">
          <Col xs={24} md={12}>
            <Input.Search
              placeholder="搜索模板名称、描述或标签"
              allowClear
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
          </Col>
          <Col xs={24} md={12}>
            <Segmented
              block
              options={categoryOptions}
              value={activeCategory}
              onChange={(value) => setActiveCategory(value as number | 'all')}
            />
          </Col>
        </Row>
      </Card>

      <Spin spinning={loading}>
        {filtered.length === 0 ? (
          <Card>
            <Empty description="暂无服务模板" />
          </Card>
        ) : (
          <List
            grid={{ gutter: 16, xs: 1, sm: 1, md: 2, lg: 3 }}
            dataSource={filtered}
            renderItem={(item) => (
              <List.Item>
                <Card
                  hoverable
                  onClick={() => navigate(`/catalog/${item.id}`)}
                  title={item.display_name}
                  extra={<Tag color="blue">部署 {item.deploy_count}</Tag>}
                >
                  <Paragraph className="line-clamp-2 !mb-3" ellipsis={{ rows: 2 }}>
                    {item.description || '暂无描述'}
                  </Paragraph>
                  <div className="flex flex-wrap gap-2">
                    {item.tags.map((tag) => (
                      <Tag key={tag}>{tag}</Tag>
                    ))}
                  </div>
                </Card>
              </List.Item>
            )}
          />
        )}
      </Spin>
    </div>
  );
};

export default CatalogListPage;

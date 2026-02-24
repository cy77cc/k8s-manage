import React from 'react';
import { Button, Card, Col, Input, Row, Select, Space, Statistic, Table, Tag, message } from 'antd';
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { Api } from '../../api';
import type { ServiceItem } from '../../api/modules/services';

const ServiceListPage: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = React.useState(false);
  const [list, setList] = React.useState<ServiceItem[]>([]);
  const [search, setSearch] = React.useState('');
  const [env, setEnv] = React.useState<string>('all');

  const load = React.useCallback(async () => {
    setLoading(true);
    try {
      const res = await Api.services.getList({ page: 1, pageSize: 100 });
      setList(res.data.list || []);
    } catch (err) {
      message.error(err instanceof Error ? err.message : '加载服务失败');
    } finally {
      setLoading(false);
    }
  }, []);

  React.useEffect(() => {
    load();
    const handler = () => load();
    window.addEventListener('project:changed', handler as EventListener);
    return () => window.removeEventListener('project:changed', handler as EventListener);
  }, [load]);

  const filtered = list.filter((item) => {
    const hitSearch = item.name.toLowerCase().includes(search.toLowerCase()) || item.owner.toLowerCase().includes(search.toLowerCase());
    const hitEnv = env === 'all' || item.env === env;
    return hitSearch && hitEnv;
  });

  const running = filtered.filter((x) => x.status === 'running').length;
  const deploying = filtered.filter((x) => x.status === 'deploying' || x.status === 'syncing').length;
  const error = filtered.filter((x) => x.status === 'error').length;

  return (
    <div className="space-y-4">
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={8}><Card><Statistic title="运行中" value={running} suffix={`/ ${filtered.length}`} /></Card></Col>
        <Col xs={24} sm={8}><Card><Statistic title="部署中" value={deploying} /></Card></Col>
        <Col xs={24} sm={8}><Card><Statistic title="异常" value={error} /></Card></Col>
      </Row>

      <Card
        title="服务管理"
        extra={
          <Space>
            <Input placeholder="搜索服务/负责人" value={search} onChange={(e) => setSearch(e.target.value)} style={{ width: 220 }} />
            <Select
              value={env}
              style={{ width: 150 }}
              options={[
                { value: 'all', label: '全部环境' },
                { value: 'production', label: '生产' },
                { value: 'staging', label: '预发' },
                { value: 'development', label: '开发' },
              ]}
              onChange={setEnv}
            />
            <Button icon={<ReloadOutlined />} onClick={load} loading={loading}>刷新</Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/services/provision')}>创建服务</Button>
          </Space>
        }
      >
        <Table
          rowKey="id"
          loading={loading}
          dataSource={filtered}
          columns={[
            { title: '名称', dataIndex: 'name', render: (v: string, r: ServiceItem) => <a onClick={() => navigate(`/services/${r.id}`)}>{v}</a> },
            { title: '环境', dataIndex: 'env', render: (v: string) => <Tag>{v}</Tag> },
            { title: '状态', dataIndex: 'status', render: (v: string) => <Tag color={v === 'running' ? 'success' : v === 'deploying' ? 'processing' : v === 'error' ? 'error' : 'default'}>{v}</Tag> },
            { title: '负责人', dataIndex: 'owner' },
            { title: '镜像', dataIndex: 'image' },
            { title: '副本', dataIndex: 'replicas' },
            { title: 'CPU', dataIndex: 'cpuLimit', render: (v: number) => `${v}m` },
            { title: '内存', dataIndex: 'memLimit', render: (v: number) => `${v}MB` },
          ]}
        />
      </Card>
    </div>
  );
};

export default ServiceListPage;

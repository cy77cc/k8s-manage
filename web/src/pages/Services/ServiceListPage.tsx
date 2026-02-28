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
  const [query, setQuery] = React.useState('');
  const [env, setEnv] = React.useState<string>('all');
  const [runtime, setRuntime] = React.useState<string>('all');
  const [teamId, setTeamId] = React.useState<string>('');
  const [labelSelector, setLabelSelector] = React.useState('');

  const load = React.useCallback(async () => {
    setLoading(true);
    try {
      const res = await Api.services.getList({
        page: 1,
        pageSize: 100,
        env,
        runtimeType: runtime as any,
        teamId,
        labelSelector,
        q: query,
      });
      setList(res.data.list || []);
    } catch (err) {
      message.error(err instanceof Error ? err.message : '加载服务失败');
    } finally {
      setLoading(false);
    }
  }, [env, runtime, teamId, labelSelector, query]);

  React.useEffect(() => {
    void load();
  }, [load]);

  const running = list.filter((x) => x.status === 'running').length;
  const deploying = list.filter((x) => x.status === 'deploying' || x.status === 'syncing').length;
  const draft = list.filter((x) => x.status === 'draft').length;

  return (
    <div className="space-y-4">
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={8}><Card><Statistic title="运行中" value={running} suffix={`/ ${list.length}`} /></Card></Col>
        <Col xs={24} sm={8}><Card><Statistic title="部署中" value={deploying} /></Card></Col>
        <Col xs={24} sm={8}><Card><Statistic title="草稿" value={draft} /></Card></Col>
      </Row>

      <Card
        title="服务管理"
        extra={(
          <Space wrap>
            <Input placeholder="搜索服务/负责人" value={query} onChange={(e) => setQuery(e.target.value)} style={{ width: 200 }} />
            <Input placeholder="团队ID" value={teamId} onChange={(e) => setTeamId(e.target.value)} style={{ width: 100 }} />
            <Select
              value={runtime}
              style={{ width: 120 }}
              options={[{ value: 'all', label: '全部运行时' }, { value: 'k8s' }, { value: 'compose' }, { value: 'helm' }]}
              onChange={setRuntime}
            />
            <Select
              value={env}
              style={{ width: 120 }}
              options={[{ value: 'all', label: '全部环境' }, { value: 'development' }, { value: 'staging' }, { value: 'production' }]}
              onChange={setEnv}
            />
            <Input placeholder="标签选择器 app=user,tier=backend" value={labelSelector} onChange={(e) => setLabelSelector(e.target.value)} style={{ width: 220 }} />
            <Button icon={<ReloadOutlined />} onClick={load} loading={loading}>刷新</Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/services/provision')}>创建服务</Button>
          </Space>
        )}
      >
        <Table
          rowKey="id"
          loading={loading}
          dataSource={list}
          columns={[
            { title: '名称', dataIndex: 'name', render: (v: string, r: ServiceItem) => <a onClick={() => navigate(`/services/${r.id}`)}>{v}</a> },
            { title: '环境', dataIndex: 'env', render: (v: string) => <Tag>{v}</Tag> },
            { title: '运行时', dataIndex: 'runtimeType', render: (v: string) => <Tag color="blue">{v}</Tag> },
            { title: '配置', dataIndex: 'configMode', render: (v: string) => <Tag>{v}</Tag> },
            { title: '状态', dataIndex: 'status', render: (v: string) => <Tag color={v === 'running' ? 'success' : v === 'deploying' ? 'processing' : v === 'error' ? 'error' : 'default'}>{v}</Tag> },
            { title: '团队', dataIndex: 'teamId' },
            { title: '负责人', dataIndex: 'owner' },
            {
              title: '标签',
              dataIndex: 'labels',
              render: (labels: Array<{ key: string; value: string }>) => (
                <Space size={[4, 4]} wrap>
                  {(labels || []).slice(0, 3).map((l) => <Tag key={`${l.key}:${l.value}`}>{l.key}={l.value}</Tag>)}
                  {(labels || []).length > 3 ? <Tag>+{(labels || []).length - 3}</Tag> : null}
                </Space>
              ),
            },
          ]}
        />
      </Card>
    </div>
  );
};

export default ServiceListPage;

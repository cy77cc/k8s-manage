import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Card, Col, Empty, Input, Progress, Row, Select, Skeleton, Space, Statistic, Table, Tag, Button } from 'antd';
import { AlertOutlined, CloudOutlined, DesktopOutlined, ReloadOutlined, ScheduleOutlined } from '@ant-design/icons';
import { Api } from '../../api';
import { useVisibilityRefresh } from '../../hooks/useVisibilityRefresh';

interface DashboardState {
  hostTotal: number;
  hostOnline: number;
  jobTotal: number;
  jobRunning: number;
  clusterTotal: number;
  alertTotal: number;
  topHosts: Array<{ id: string; name: string; ip: string; status: string; cpu: number; memory: number }>;
  recentAlerts: Array<{ id: string; message: string; severity: string; createdAt: string }>;
  recentFailedReleases: number;
}

const Dashboard: React.FC = () => {
  const [loading, setLoading] = useState(true);
  const [query, setQuery] = useState('');
  const [severity, setSeverity] = useState<string>('all');
  const [state, setState] = useState<DashboardState>({
    hostTotal: 0,
    hostOnline: 0,
    jobTotal: 0,
    jobRunning: 0,
    clusterTotal: 0,
    alertTotal: 0,
    topHosts: [],
    recentAlerts: [],
    recentFailedReleases: 0,
  });

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const [hosts, jobs, clusters, alerts, releases] = await Promise.all([
        Api.hosts.getHostList({ page: 1, pageSize: 100 }),
        Api.tasks.getTaskList({ page: 1, pageSize: 100 }),
        Api.kubernetes.getClusterList({ page: 1, pageSize: 20 }),
        Api.monitoring.getAlertList({ page: 1, pageSize: 20 }),
        Api.deployment.getReleases(),
      ]);

      const hostList = hosts.data.list || [];
      const jobList = jobs.data.list || [];
      const clusterList = clusters.data.list || [];
      const alertList = alerts.data.list || [];
      const releaseList = releases.data.list || [];
      const recentFailedReleases = releaseList.filter((r) => r.status === 'failed').length;

      setState({
        hostTotal: hostList.length,
        hostOnline: hostList.filter((h) => h.status === 'online').length,
        jobTotal: jobList.length,
        jobRunning: jobList.filter((j) => j.status === 'running').length,
        clusterTotal: clusterList.length,
        alertTotal: alertList.filter((a) => a.status === 'firing').length,
        topHosts: hostList.slice(0, 10).map((h) => ({ id: String(h.id), name: h.name, ip: h.ip, status: h.status, cpu: h.cpu ?? 0, memory: h.memory ?? 0 })),
        recentAlerts: alertList.slice(0, 20).map((a) => ({ id: String(a.id), message: a.title || a.source || '告警事件', severity: a.severity, createdAt: a.createdAt })),
        recentFailedReleases,
      });
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void load();
    const handler = () => void load();
    window.addEventListener('project:changed', handler as EventListener);
    return () => window.removeEventListener('project:changed', handler as EventListener);
  }, [load]);

  useVisibilityRefresh(() => void load(), 30000, [load]);

  const widgets = useMemo(() => [
    { key: 'host-health', title: '主机健康', value: `${state.hostOnline}/${state.hostTotal}`, extra: `${Math.round((state.hostOnline / Math.max(1, state.hostTotal)) * 100)}%` },
    { key: 'task-success', title: '任务成功率', value: `${state.jobTotal - state.jobRunning}/${Math.max(1, state.jobTotal)}`, extra: `${Math.round(((state.jobTotal - state.jobRunning) / Math.max(1, state.jobTotal)) * 100)}%` },
    { key: 'release-frequency', title: '最近失败发布', value: state.recentFailedReleases, extra: '24h' },
    { key: 'alert-trend', title: '活跃告警', value: state.alertTotal, extra: 'active' },
  ], [state]);

  const filteredAlerts = useMemo(() => {
    return state.recentAlerts.filter((alert) => {
      const matchQuery = query.trim() ? alert.message.toLowerCase().includes(query.trim().toLowerCase()) : true;
      const matchSeverity = severity === 'all' ? true : alert.severity === severity;
      return matchQuery && matchSeverity;
    });
  }, [query, severity, state.recentAlerts]);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold text-slate-100">主控台</h1>
        <Button icon={<ReloadOutlined />} onClick={() => void load()} loading={loading}>刷新</Button>
      </div>

      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}><Card><Statistic title="主机总数" value={state.hostTotal} prefix={<DesktopOutlined />} /></Card></Col>
        <Col xs={24} sm={12} lg={6}><Card><Statistic title="在线主机" value={state.hostOnline} prefix={<CloudOutlined />} /></Card></Col>
        <Col xs={24} sm={12} lg={6}><Card><Statistic title="任务总数" value={state.jobTotal} prefix={<ScheduleOutlined />} /></Card></Col>
        <Col xs={24} sm={12} lg={6}><Card><Statistic title="活跃告警" value={state.alertTotal} prefix={<AlertOutlined />} /></Card></Col>
      </Row>

      <Row gutter={[16, 16]}>
        <Col xs={24} xl={16}>
          <Card title="监控概览" extra={<Tag color={state.alertTotal > 0 ? 'error' : 'success'}>{state.alertTotal > 0 ? '需要关注' : '健康'}</Tag>}>
            {loading ? (
              <Skeleton active paragraph={{ rows: 8 }} />
            ) : (
              <Row gutter={[12, 12]}>
                {widgets.map((w) => (
                  <Col xs={24} sm={12} key={w.key}>
                    <Card size="small" title={w.title}>
                      <div className="text-lg font-semibold">{w.value}</div>
                      <div className="text-slate-400">{w.extra}</div>
                    </Card>
                  </Col>
                ))}
                <Col span={24}>
                  <Card size="small" title="资源压力 Top 10">
                    <Table
                      rowKey="id"
                      pagination={{ pageSize: 5 }}
                      dataSource={state.topHosts}
                      locale={{ emptyText: <Empty description="暂无主机数据" /> }}
                      columns={[
                        { title: '主机', dataIndex: 'name', sorter: (a, b) => a.name.localeCompare(b.name) },
                        { title: 'IP', dataIndex: 'ip' },
                        { title: '状态', dataIndex: 'status', filters: [{ text: 'online', value: 'online' }, { text: 'offline', value: 'offline' }], onFilter: (v, r) => r.status === v, render: (v: string) => <Tag color={v === 'online' ? 'success' : 'default'}>{v}</Tag> },
                        { title: 'CPU', dataIndex: 'cpu', sorter: (a, b) => a.cpu - b.cpu, render: (v: number) => <Progress percent={Math.min(100, Math.round(v))} size="small" /> },
                      ]}
                    />
                  </Card>
                </Col>
              </Row>
            )}
          </Card>
        </Col>

        <Col xs={24} xl={8}>
          <Card
            title="告警处理队列"
            extra={(
              <Space>
                <Input allowClear placeholder="搜索告警" value={query} onChange={(e) => setQuery(e.target.value)} style={{ width: 140 }} />
                <Select value={severity} style={{ width: 120 }} onChange={setSeverity} options={[{ value: 'all', label: '全部级别' }, { value: 'critical', label: 'critical' }, { value: 'warning', label: 'warning' }, { value: 'info', label: 'info' }]} />
              </Space>
            )}
          >
            {loading ? (
              <Skeleton active paragraph={{ rows: 10 }} />
            ) : (
              <Table
                rowKey="id"
                dataSource={filteredAlerts}
                pagination={{ pageSize: 6 }}
                locale={{ emptyText: <Empty description="暂无告警数据" /> }}
                columns={[
                  { title: '告警', dataIndex: 'message', sorter: (a, b) => a.message.localeCompare(b.message) },
                  { title: '级别', dataIndex: 'severity', filters: [{ text: 'critical', value: 'critical' }, { text: 'warning', value: 'warning' }, { text: 'info', value: 'info' }], onFilter: (v, r) => r.severity === v, render: (v: string) => <Tag color={v === 'critical' ? 'error' : v === 'warning' ? 'warning' : 'blue'}>{v}</Tag> },
                ]}
              />
            )}
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]}>
        <Col span={24}>
          <Card title="服务列表（监控联动）">
            {loading ? <Skeleton active paragraph={{ rows: 6 }} /> : (
              <Table
                rowKey="id"
                dataSource={state.topHosts}
                pagination={{ pageSize: 5 }}
                locale={{ emptyText: <Empty description="暂无服务数据" /> }}
                columns={[
                  { title: '服务/主机', dataIndex: 'name' },
                  { title: '状态', dataIndex: 'status', render: (v: string) => <Tag color={v === 'online' ? 'success' : 'error'}>{v}</Tag> },
                  { title: 'CPU', dataIndex: 'cpu', sorter: (a, b) => a.cpu - b.cpu },
                  { title: '内存', dataIndex: 'memory', sorter: (a, b) => a.memory - b.memory },
                ]}
              />
            )}
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Dashboard;

import React, { useEffect, useMemo, useState } from 'react';
import { Card, Col, List, Progress, Row, Space, Statistic, Table, Tag, Button } from 'antd';
import { AlertOutlined, CloudOutlined, DesktopOutlined, ReloadOutlined, ScheduleOutlined } from '@ant-design/icons';
import { Api } from '../../api';

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

  const load = async () => {
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
        topHosts: hostList.slice(0, 5).map((h) => ({ id: String(h.id), name: h.name, ip: h.ip, status: h.status, cpu: h.cpu ?? 0, memory: h.memory ?? 0 })),
        recentAlerts: alertList.slice(0, 6).map((a) => ({ id: String(a.id), message: a.title || a.source || '告警事件', severity: a.severity, createdAt: a.createdAt })),
        recentFailedReleases,
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
    const handler = () => load();
    window.addEventListener('project:changed', handler as EventListener);
    return () => window.removeEventListener('project:changed', handler as EventListener);
  }, []);

  const widgets = useMemo(() => [
    { key: 'host-health', title: '主机健康', value: `${state.hostOnline}/${state.hostTotal}`, extra: `${Math.round((state.hostOnline / Math.max(1, state.hostTotal)) * 100)}%` },
    { key: 'task-success', title: '任务成功率', value: `${state.jobTotal - state.jobRunning}/${Math.max(1, state.jobTotal)}`, extra: `${Math.round(((state.jobTotal - state.jobRunning) / Math.max(1, state.jobTotal)) * 100)}%` },
    { key: 'release-frequency', title: '最近失败发布', value: state.recentFailedReleases, extra: '24h' },
    { key: 'alert-trend', title: '告警趋势', value: state.alertTotal, extra: 'active' },
    { key: 'k8s-capacity', title: 'K8s 容量', value: state.clusterTotal, extra: 'clusters' },
    { key: 'service-slo', title: '服务 SLO', value: '99.90%', extra: '目标 99.95%' },
    { key: 'error-rate', title: '错误率', value: `${state.alertTotal}%`, extra: '估算' },
    { key: 'top-resources', title: 'Top 资源消耗', value: state.topHosts[0]?.name || '-', extra: `${state.topHosts[0]?.cpu || 0}% CPU` },
  ], [state]);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-bold text-white">主控台</h1>
        <Button icon={<ReloadOutlined />} onClick={load} loading={loading}>刷新</Button>
      </div>

      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}><Card><Statistic title="主机总数" value={state.hostTotal} prefix={<DesktopOutlined />} /></Card></Col>
        <Col xs={24} sm={12} lg={6}><Card><Statistic title="在线主机" value={state.hostOnline} prefix={<CloudOutlined />} /></Card></Col>
        <Col xs={24} sm={12} lg={6}><Card><Statistic title="任务总数" value={state.jobTotal} prefix={<ScheduleOutlined />} /></Card></Col>
        <Col xs={24} sm={12} lg={6}><Card><Statistic title="活跃告警" value={state.alertTotal} prefix={<AlertOutlined />} /></Card></Col>
      </Row>

      <Row gutter={[16, 16]}>
        {widgets.map((w) => (
          <Col xs={24} sm={12} md={8} lg={6} key={w.key}>
            <Card size="small" title={w.title}>
              <div className="text-lg font-bold">{w.value}</div>
              <div className="text-gray-500">{w.extra}</div>
            </Card>
          </Col>
        ))}
      </Row>

      <Row gutter={[16, 16]}>
        <Col xs={24} lg={14}>
          <Card title="主机概览">
            <Table
              rowKey="id"
              pagination={false}
              dataSource={state.topHosts}
              columns={[
                { title: '主机', dataIndex: 'name' },
                { title: 'IP', dataIndex: 'ip' },
                { title: '状态', dataIndex: 'status', render: (v: string) => <Tag color={v === 'online' ? 'success' : 'default'}>{v}</Tag> },
                { title: 'CPU', dataIndex: 'cpu', render: (v: number) => <Progress percent={Math.min(100, v)} size="small" /> },
              ]}
            />
          </Card>
        </Col>
        <Col xs={24} lg={10}>
          <Card title="最近告警">
            <List
              dataSource={state.recentAlerts}
              renderItem={(item) => (
                <List.Item>
                  <div className="w-full flex justify-between">
                    <span>{item.message}</span>
                    <Tag color={item.severity === 'critical' ? 'error' : item.severity === 'warning' ? 'warning' : 'blue'}>{item.severity}</Tag>
                  </div>
                </List.Item>
              )}
            />
          </Card>
        </Col>
      </Row>

      <Card title="运行摘要">
        <Space wrap>
          <Tag color="blue">在线主机: {state.hostOnline}</Tag>
          <Tag color="geekblue">集群数: {state.clusterTotal}</Tag>
          <Tag color={state.recentFailedReleases > 0 ? 'error' : 'success'}>失败发布: {state.recentFailedReleases}</Tag>
        </Space>
      </Card>
    </div>
  );
};

export default Dashboard;

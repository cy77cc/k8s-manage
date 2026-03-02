import React, { useEffect, useMemo, useState } from 'react';
import { Card, Col, List, Progress, Row, Statistic, Table, Tag, Button, Empty } from 'antd';
import {
  AlertOutlined,
  CloudOutlined,
  DesktopOutlined,
  ReloadOutlined,
  ScheduleOutlined,
  ArrowUpOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
} from '@ant-design/icons';
import { Api } from '../../api';
import { StaggerList, StaggerItem } from '../../components/Motion';

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
        topHosts: hostList.slice(0, 5).map((h) => ({
          id: String(h.id),
          name: h.name,
          ip: h.ip,
          status: h.status,
          cpu: h.cpu ?? 0,
          memory: h.memory ?? 0,
        })),
        recentAlerts: alertList.slice(0, 6).map((a) => ({
          id: String(a.id),
          message: a.title || a.source || '告警事件',
          severity: a.severity,
          createdAt: a.createdAt,
        })),
        recentFailedReleases,
      });
    } finally {
      setLoading(false);
    }
  };

  // 3.1.11 自动刷新（30秒）
  useEffect(() => {
    load();
    const handler = () => load();
    window.addEventListener('project:changed', handler as EventListener);

    // 30秒自动刷新
    const interval = setInterval(load, 30000);

    return () => {
      window.removeEventListener('project:changed', handler as EventListener);
      clearInterval(interval);
    };
  }, []);

  // 计算健康率
  const healthRate = useMemo(() => {
    if (state.hostTotal === 0) return 0;
    return Math.round((state.hostOnline / state.hostTotal) * 100);
  }, [state.hostOnline, state.hostTotal]);

  // 计算任务成功率
  const taskSuccessRate = useMemo(() => {
    if (state.jobTotal === 0) return 0;
    return Math.round(((state.jobTotal - state.jobRunning) / state.jobTotal) * 100);
  }, [state.jobTotal, state.jobRunning]);

  return (
    <div className="space-y-6">
      {/* 页面头部 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-gray-900">主控台</h1>
          <p className="text-sm text-gray-500 mt-1">实时监控系统运行状态</p>
        </div>
        <Button
          type="primary"
          icon={<ReloadOutlined />}
          onClick={load}
          loading={loading}
        >
          刷新数据
        </Button>
      </div>

      {/* 3.1.8 统计卡片 - 主要指标 */}
      <StaggerList staggerDelay={0.05}>
        <Row gutter={[16, 16]}>
          <Col xs={24} sm={12} lg={6}>
            <StaggerItem>
              <Card className="hover:shadow-lg transition-shadow">
                <Statistic
                  title={<span className="text-gray-600">主机总数</span>}
                  value={state.hostTotal}
                  prefix={<DesktopOutlined className="text-primary-500" />}
                  valueStyle={{ color: '#495057', fontSize: '28px', fontWeight: 600 }}
                />
                <div className="mt-2 flex items-center text-sm">
                  <span className="text-success flex items-center">
                    <ArrowUpOutlined className="mr-1" />
                    {state.hostOnline} 在线
                  </span>
                  <span className="text-gray-400 mx-2">|</span>
                  <span className="text-gray-500">
                    {state.hostTotal - state.hostOnline} 离线
                  </span>
                </div>
              </Card>
            </StaggerItem>
          </Col>

          <Col xs={24} sm={12} lg={6}>
            <StaggerItem>
              <Card className="hover:shadow-lg transition-shadow">
                <Statistic
                  title={<span className="text-gray-600">主机健康率</span>}
                  value={healthRate}
                  suffix="%"
                  prefix={
                    healthRate >= 90 ? (
                      <CheckCircleOutlined className="text-success" />
                    ) : (
                      <CloseCircleOutlined className="text-error" />
                    )
                  }
                  valueStyle={{
                    color: healthRate >= 90 ? '#10b981' : '#ef4444',
                    fontSize: '28px',
                    fontWeight: 600,
                  }}
                />
                <Progress
                  percent={healthRate}
                  strokeColor={healthRate >= 90 ? '#10b981' : '#ef4444'}
                  showInfo={false}
                  className="mt-2"
                />
              </Card>
            </StaggerItem>
          </Col>

          <Col xs={24} sm={12} lg={6}>
            <StaggerItem>
              <Card className="hover:shadow-lg transition-shadow">
                <Statistic
                  title={<span className="text-gray-600">任务总数</span>}
                  value={state.jobTotal}
                  prefix={<ScheduleOutlined className="text-primary-500" />}
                  valueStyle={{ color: '#495057', fontSize: '28px', fontWeight: 600 }}
                />
                <div className="mt-2 flex items-center text-sm">
                  <span className="text-primary-600 flex items-center">
                    {state.jobRunning} 运行中
                  </span>
                  <span className="text-gray-400 mx-2">|</span>
                  <span className="text-gray-500">
                    成功率 {taskSuccessRate}%
                  </span>
                </div>
              </Card>
            </StaggerItem>
          </Col>

          <Col xs={24} sm={12} lg={6}>
            <StaggerItem>
              <Card className="hover:shadow-lg transition-shadow">
                <Statistic
                  title={<span className="text-gray-600">活跃告警</span>}
                  value={state.alertTotal}
                  prefix={<AlertOutlined className="text-warning" />}
                  valueStyle={{
                    color: state.alertTotal > 0 ? '#f59e0b' : '#10b981',
                    fontSize: '28px',
                    fontWeight: 600,
                  }}
                />
                <div className="mt-2 text-sm">
                  {state.alertTotal > 0 ? (
                    <span className="text-warning">需要关注</span>
                  ) : (
                    <span className="text-success">系统正常</span>
                  )}
                </div>
              </Card>
            </StaggerItem>
          </Col>
        </Row>
      </StaggerList>

      {/* 次要指标 */}
      <StaggerList staggerDelay={0.05}>
        <Row gutter={[16, 16]}>
          <Col xs={24} sm={12} md={8} lg={6}>
            <StaggerItem>
              <Card size="small" className="hover:shadow-md transition-shadow">
                <div className="flex items-center justify-between">
                  <div>
                    <div className="text-sm text-gray-600 mb-1">K8s 集群</div>
                    <div className="text-2xl font-semibold text-gray-900">
                      {state.clusterTotal}
                    </div>
                  </div>
                  <CloudOutlined className="text-3xl text-primary-500 opacity-20" />
                </div>
              </Card>
            </StaggerItem>
          </Col>

          <Col xs={24} sm={12} md={8} lg={6}>
            <StaggerItem>
              <Card size="small" className="hover:shadow-md transition-shadow">
                <div className="flex items-center justify-between">
                  <div>
                    <div className="text-sm text-gray-600 mb-1">失败发布</div>
                    <div className="text-2xl font-semibold text-gray-900">
                      {state.recentFailedReleases}
                    </div>
                    <div className="text-xs text-gray-500 mt-1">最近 24h</div>
                  </div>
                  <CloseCircleOutlined
                    className={`text-3xl opacity-20 ${
                      state.recentFailedReleases > 0 ? 'text-error' : 'text-success'
                    }`}
                  />
                </div>
              </Card>
            </StaggerItem>
          </Col>

          <Col xs={24} sm={12} md={8} lg={6}>
            <StaggerItem>
              <Card size="small" className="hover:shadow-md transition-shadow">
                <div className="flex items-center justify-between">
                  <div>
                    <div className="text-sm text-gray-600 mb-1">服务 SLO</div>
                    <div className="text-2xl font-semibold text-gray-900">99.90%</div>
                    <div className="text-xs text-gray-500 mt-1">目标 99.95%</div>
                  </div>
                  <CheckCircleOutlined className="text-3xl text-success opacity-20" />
                </div>
              </Card>
            </StaggerItem>
          </Col>

          <Col xs={24} sm={12} md={8} lg={6}>
            <StaggerItem>
              <Card size="small" className="hover:shadow-md transition-shadow">
                <div className="flex items-center justify-between">
                  <div>
                    <div className="text-sm text-gray-600 mb-1">Top 资源</div>
                    <div className="text-base font-medium text-gray-900 truncate">
                      {state.topHosts[0]?.name || '-'}
                    </div>
                    <div className="text-xs text-gray-500 mt-1">
                      CPU {state.topHosts[0]?.cpu || 0}%
                    </div>
                  </div>
                  <DesktopOutlined className="text-3xl text-primary-500 opacity-20" />
                </div>
              </Card>
            </StaggerItem>
          </Col>
        </Row>
      </StaggerList>

      {/* 3.1.9 服务健康状态列表 */}
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={14}>
          <Card
            title={<span className="text-base font-semibold">主机概览</span>}
            extra={<Button type="link">查看全部</Button>}
            className="h-full"
          >
            {state.topHosts.length > 0 ? (
              <Table
                rowKey="id"
                pagination={false}
                dataSource={state.topHosts}
                size="small"
                columns={[
                  {
                    title: '主机名称',
                    dataIndex: 'name',
                    key: 'name',
                    render: (text) => (
                      <span className="font-medium text-gray-900">{text}</span>
                    ),
                  },
                  {
                    title: 'IP 地址',
                    dataIndex: 'ip',
                    key: 'ip',
                    render: (text) => (
                      <span className="text-gray-600 font-mono text-sm">{text}</span>
                    ),
                  },
                  {
                    title: '状态',
                    dataIndex: 'status',
                    key: 'status',
                    render: (v: string) => (
                      <Tag
                        color={v === 'online' ? 'success' : 'default'}
                        icon={
                          v === 'online' ? (
                            <CheckCircleOutlined />
                          ) : (
                            <CloseCircleOutlined />
                          )
                        }
                      >
                        {v === 'online' ? '在线' : '离线'}
                      </Tag>
                    ),
                  },
                  {
                    title: 'CPU 使用率',
                    dataIndex: 'cpu',
                    key: 'cpu',
                    render: (v: number) => (
                      <div className="w-32">
                        <Progress
                          percent={Math.min(100, v)}
                          size="small"
                          strokeColor={v > 80 ? '#ef4444' : v > 60 ? '#f59e0b' : '#10b981'}
                        />
                      </div>
                    ),
                  },
                ]}
              />
            ) : (
              <Empty description="暂无主机数据" />
            )}
          </Card>
        </Col>

        <Col xs={24} lg={10}>
          <Card
            title={<span className="text-base font-semibold">最近告警</span>}
            extra={<Button type="link">查看全部</Button>}
            className="h-full"
          >
            {state.recentAlerts.length > 0 ? (
              <List
                dataSource={state.recentAlerts}
                renderItem={(item) => (
                  <List.Item className="hover:bg-gray-50 transition-colors px-2 rounded">
                    <div className="w-full flex items-center justify-between">
                      <span className="text-gray-700 flex-1 truncate">{item.message}</span>
                      <Tag
                        color={
                          item.severity === 'critical'
                            ? 'error'
                            : item.severity === 'warning'
                            ? 'warning'
                            : 'processing'
                        }
                        className="ml-2"
                      >
                        {item.severity}
                      </Tag>
                    </div>
                  </List.Item>
                )}
              />
            ) : (
              <Empty description="暂无告警" image={Empty.PRESENTED_IMAGE_SIMPLE} />
            )}
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Dashboard;

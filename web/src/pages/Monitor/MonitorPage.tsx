import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Card, Col, Empty, Input, Row, Select, Skeleton, Statistic, Table, Tabs, Tag, Progress, Button, Space } from 'antd';
import { AlertOutlined, ReloadOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import { Api } from '../../api';
import type { Alert, AlertRule, MetricData } from '../../api/modules/monitoring';
import { useVisibilityRefresh } from '../../hooks/useVisibilityRefresh';

const MonitorPage: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [rules, setRules] = useState<AlertRule[]>([]);
  const [cpuMetrics, setCpuMetrics] = useState<MetricData[]>([]);
  const [memMetrics, setMemMetrics] = useState<MetricData[]>([]);
  const [query, setQuery] = useState('');
  const [severity, setSeverity] = useState<string>('all');

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const end = new Date().toISOString();
      const start = dayjs().subtract(24, 'hour').toDate().toISOString();
      const [alertRes, ruleRes, cpuRes, memRes] = await Promise.all([
        Api.monitoring.getAlertList({ page: 1, pageSize: 100 }),
        Api.monitoring.getAlertRuleList({ page: 1, pageSize: 100 }),
        Api.monitoring.getMetrics({ metric: 'cpu_usage', startTime: start, endTime: end }),
        Api.monitoring.getMetrics({ metric: 'memory_usage', startTime: start, endTime: end }),
      ]);
      setAlerts(alertRes.data.list || []);
      setRules(ruleRes.data.list || []);
      setCpuMetrics(cpuRes.data?.series || []);
      setMemMetrics(memRes.data?.series || []);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  useVisibilityRefresh(() => void load(), 15000, [load]);

  const filteredAlerts = useMemo(() => {
    return alerts.filter((item) => {
      const text = `${item.title || ''} ${item.message || ''} ${item.source || ''}`.toLowerCase();
      const matchQuery = query.trim() ? text.includes(query.trim().toLowerCase()) : true;
      const matchSeverity = severity === 'all' ? true : item.severity === severity;
      return matchQuery && matchSeverity;
    });
  }, [alerts, query, severity]);

  const firingCount = useMemo(() => filteredAlerts.filter((a) => a.status === 'firing').length, [filteredAlerts]);
  const criticalCount = useMemo(() => filteredAlerts.filter((a) => a.severity === 'critical' && a.status === 'firing').length, [filteredAlerts]);
  const cpuAvg = useMemo(() => (cpuMetrics.length ? cpuMetrics.reduce((s, i) => s + Number(i.value), 0) / cpuMetrics.length : 0), [cpuMetrics]);
  const memAvg = useMemo(() => (memMetrics.length ? memMetrics.reduce((s, i) => s + Number(i.value), 0) / memMetrics.length : 0), [memMetrics]);

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <Space wrap>
          <Input allowClear placeholder="搜索告警内容/来源" value={query} onChange={(e) => setQuery(e.target.value)} style={{ width: 220 }} />
          <Select value={severity} style={{ width: 140 }} onChange={setSeverity} options={[{ value: 'all', label: '全部级别' }, { value: 'critical', label: 'critical' }, { value: 'warning', label: 'warning' }, { value: 'info', label: 'info' }]} />
        </Space>
        <Button icon={<ReloadOutlined />} loading={loading} onClick={() => void load()}>刷新</Button>
      </div>

      <Row gutter={[16, 16]}>
        <Col xs={24} sm={8}><Card><Statistic title="活跃告警" value={firingCount} prefix={<AlertOutlined />} /></Card></Col>
        <Col xs={24} sm={8}><Card><Statistic title="严重告警" value={criticalCount} /></Card></Col>
        <Col xs={24} sm={8}><Card><Statistic title="告警规则" value={rules.length} /></Card></Col>
      </Row>

      <Row gutter={[16, 16]}>
        <Col xs={24} md={12}>
          <Card title="CPU 平均使用率">{loading ? <Skeleton active paragraph={{ rows: 2 }} /> : <Progress percent={Math.round(cpuAvg)} />}</Card>
        </Col>
        <Col xs={24} md={12}>
          <Card title="内存平均使用率">{loading ? <Skeleton active paragraph={{ rows: 2 }} /> : <Progress percent={Math.round(memAvg)} strokeColor="var(--color-brand-500)" />}</Card>
        </Col>
      </Row>

      <Tabs
        items={[
          {
            key: 'alerts',
            label: '告警历史',
            children: (
              <Table
                rowKey="id"
                loading={loading}
                dataSource={filteredAlerts}
                locale={{ emptyText: <Empty description="暂无告警数据" /> }}
                columns={[
                  { title: '消息', dataIndex: 'title', sorter: (a, b) => String(a.title || a.message || '').localeCompare(String(b.title || b.message || '')), render: (_: string, r: any) => r.message || r.title || '-' },
                  { title: '级别', dataIndex: 'severity', filters: [{ text: 'critical', value: 'critical' }, { text: 'warning', value: 'warning' }, { text: 'info', value: 'info' }], onFilter: (v, r) => r.severity === v, render: (v: string) => <Tag color={v === 'critical' ? 'error' : v === 'warning' ? 'warning' : 'blue'}>{v}</Tag> },
                  { title: '来源', dataIndex: 'source', render: (_: string, r: any) => r.metric || r.source || '-' },
                  { title: '状态', dataIndex: 'status', filters: [{ text: 'firing', value: 'firing' }, { text: 'resolved', value: 'resolved' }], onFilter: (v, r) => r.status === v, render: (v: string) => <Tag color={v === 'firing' ? 'error' : 'success'}>{v}</Tag> },
                  { title: '时间', dataIndex: 'createdAt', sorter: (a, b) => new Date(a.createdAt || 0).getTime() - new Date(b.createdAt || 0).getTime(), render: (v: string) => (v ? new Date(v).toLocaleString() : '-') },
                ]}
              />
            ),
          },
          {
            key: 'rules',
            label: '告警规则',
            children: (
              <Table
                rowKey="id"
                loading={loading}
                dataSource={rules}
                locale={{ emptyText: <Empty description="暂无告警规则" /> }}
                columns={[
                  { title: '名称', dataIndex: 'name', sorter: (a, b) => String(a.name || '').localeCompare(String(b.name || '')) },
                  { title: '指标', dataIndex: 'condition', render: (_: string, r: any) => `${r.metric} ${r.operator} ${r.threshold}` },
                  { title: '级别', dataIndex: 'severity', filters: [{ text: 'critical', value: 'critical' }, { text: 'warning', value: 'warning' }, { text: 'info', value: 'info' }], onFilter: (v, r) => r.severity === v, render: (v: string) => <Tag>{v}</Tag> },
                  { title: '启用', dataIndex: 'enabled', filters: [{ text: '启用', value: true }, { text: '禁用', value: false }], onFilter: (v, r) => r.enabled === v, render: (v: boolean) => <Tag color={v ? 'success' : 'default'}>{v ? '启用' : '禁用'}</Tag> },
                ]}
              />
            ),
          },
        ]}
      />
    </div>
  );
};

export default MonitorPage;

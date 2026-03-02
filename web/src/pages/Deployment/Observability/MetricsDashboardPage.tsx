import React, { useState, useEffect } from 'react';
import { Card, Row, Col, Statistic, Select, Space, Button } from 'antd';
import {
  ReloadOutlined,
  RiseOutlined,
  FallOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
} from '@ant-design/icons';
import { Line } from '@ant-design/charts';

const MetricsDashboardPage: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [timeRange, setTimeRange] = useState<'daily' | 'weekly' | 'monthly'>('daily');
  const [envFilter, setEnvFilter] = useState<string>('all');

  const load = async () => {
    setLoading(true);
    try {
      // Mock data loading
      await new Promise((resolve) => setTimeout(resolve, 500));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, [timeRange, envFilter]);

  // Mock data for deployment frequency
  const deploymentFrequencyData = [
    { date: '2024-01', count: 45 },
    { date: '2024-02', count: 52 },
    { date: '2024-03', count: 48 },
    { date: '2024-04', count: 61 },
    { date: '2024-05', count: 55 },
    { date: '2024-06', count: 67 },
  ];

  // Mock data for success rate
  const successRateData = [
    { date: '2024-01', rate: 92 },
    { date: '2024-02', rate: 94 },
    { date: '2024-03', rate: 91 },
    { date: '2024-04', rate: 95 },
    { date: '2024-05', rate: 93 },
    { date: '2024-06', rate: 96 },
  ];

  const deploymentFrequencyConfig = {
    data: deploymentFrequencyData,
    xField: 'date',
    yField: 'count',
    point: {
      size: 5,
      shape: 'diamond',
    },
    label: {
      style: {
        fill: '#aaa',
      },
    },
  };

  const successRateConfig = {
    data: successRateData,
    xField: 'date',
    yField: 'rate',
    point: {
      size: 5,
      shape: 'circle',
    },
    yAxis: {
      min: 80,
      max: 100,
    },
    color: '#10b981',
  };

  return (
    <div className="space-y-6">
      {/* Page header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-gray-900">指标仪表板</h1>
          <p className="text-sm text-gray-500 mt-1">查看部署相关的关键指标和趋势</p>
        </div>
        <Space>
          <Select
            value={timeRange}
            style={{ width: 120 }}
            options={[
              { value: 'daily', label: '每日' },
              { value: 'weekly', label: '每周' },
              { value: 'monthly', label: '每月' },
            ]}
            onChange={setTimeRange}
          />
          <Select
            value={envFilter}
            style={{ width: 140 }}
            options={[
              { value: 'all', label: '全部环境' },
              { value: 'production', label: '生产环境' },
              { value: 'staging', label: '预发布' },
              { value: 'development', label: '开发环境' },
            ]}
            onChange={setEnvFilter}
          />
          <Button icon={<ReloadOutlined />} onClick={load} loading={loading}>
            刷新
          </Button>
        </Space>
      </div>

      {/* Key metrics */}
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="部署频率"
              value={67}
              suffix="次/月"
              prefix={<RiseOutlined className="text-success" />}
              valueStyle={{ color: '#10b981' }}
            />
            <div className="text-xs text-gray-500 mt-2">较上月 +12%</div>
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="成功率"
              value={96}
              suffix="%"
              prefix={<CheckCircleOutlined className="text-success" />}
              valueStyle={{ color: '#10b981' }}
            />
            <div className="text-xs text-gray-500 mt-2">较上月 +3%</div>
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="平均部署时长"
              value={8.5}
              suffix="分钟"
              prefix={<ClockCircleOutlined className="text-primary" />}
              valueStyle={{ color: '#6366f1' }}
            />
            <div className="text-xs text-gray-500 mt-2">较上月 -15%</div>
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="审批响应时间"
              value={2.3}
              suffix="小时"
              prefix={<ClockCircleOutlined className="text-warning" />}
              valueStyle={{ color: '#f59e0b' }}
            />
            <div className="text-xs text-gray-500 mt-2">较上月 +5%</div>
          </Card>
        </Col>
      </Row>

      {/* Deployment frequency chart */}
      <Card title="部署频率趋势">
        <Line {...deploymentFrequencyConfig} />
      </Card>

      {/* Success rate chart */}
      <Card title="成功率趋势">
        <Line {...successRateConfig} />
      </Card>

      {/* Environment comparison */}
      <Card title="环境对比">
        <Row gutter={[16, 16]}>
          <Col xs={24} md={8}>
            <Card className="bg-blue-50">
              <Statistic
                title="开发环境"
                value={156}
                suffix="次部署"
                valueStyle={{ color: '#3b82f6' }}
              />
              <div className="text-xs text-gray-600 mt-2">成功率: 98%</div>
            </Card>
          </Col>
          <Col xs={24} md={8}>
            <Card className="bg-orange-50">
              <Statistic
                title="预发布环境"
                value={89}
                suffix="次部署"
                valueStyle={{ color: '#f59e0b' }}
              />
              <div className="text-xs text-gray-600 mt-2">成功率: 95%</div>
            </Card>
          </Col>
          <Col xs={24} md={8}>
            <Card className="bg-red-50">
              <Statistic
                title="生产环境"
                value={67}
                suffix="次部署"
                valueStyle={{ color: '#ef4444' }}
              />
              <div className="text-xs text-gray-600 mt-2">成功率: 96%</div>
            </Card>
          </Col>
        </Row>
      </Card>
    </div>
  );
};

export default MetricsDashboardPage;

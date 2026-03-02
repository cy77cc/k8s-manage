import React, { useState, useEffect } from 'react';
import { Card, Row, Col, Alert, Button, Space, Tag, Collapse, Empty } from 'antd';
import {
  ReloadOutlined,
  WarningOutlined,
  CheckCircleOutlined,
  BulbOutlined,
  LineChartOutlined,
} from '@ant-design/icons';

const { Panel } = Collapse;

interface RiskFinding {
  severity: 'high' | 'medium' | 'low';
  category: string;
  message: string;
  recommendation: string;
}

interface Anomaly {
  timestamp: string;
  metric: string;
  value: number;
  threshold: number;
  description: string;
}

interface Suggestion {
  type: string;
  title: string;
  description: string;
  impact: string;
}

const AIOpsInsightsPage: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [riskFindings, setRiskFindings] = useState<RiskFinding[]>([]);
  const [anomalies, setAnomalies] = useState<Anomaly[]>([]);
  const [suggestions, setSuggestions] = useState<Suggestion[]>([]);

  const load = async () => {
    setLoading(true);
    try {
      // Mock data - replace with actual API call
      const mockRisks: RiskFinding[] = [
        {
          severity: 'high',
          category: 'Configuration',
          message: '检测到生产环境缺少资源限制配置',
          recommendation: '建议为所有生产环境容器配置 CPU 和内存限制',
        },
        {
          severity: 'medium',
          category: 'Performance',
          message: 'API Gateway 响应时间持续增长',
          recommendation: '考虑增加副本数或优化查询性能',
        },
        {
          severity: 'low',
          category: 'Security',
          message: '发现使用了过时的基础镜像版本',
          recommendation: '更新到最新的安全补丁版本',
        },
      ];

      const mockAnomalies: Anomaly[] = [
        {
          timestamp: new Date(Date.now() - 3600000).toISOString(),
          metric: 'error_rate',
          value: 5.2,
          threshold: 2.0,
          description: 'user-service 错误率异常升高',
        },
        {
          timestamp: new Date(Date.now() - 7200000).toISOString(),
          metric: 'response_time',
          value: 850,
          threshold: 500,
          description: 'order-service 响应时间超过阈值',
        },
      ];

      const mockSuggestions: Suggestion[] = [
        {
          type: 'optimization',
          title: '优化部署策略',
          description: '基于历史数据分析，建议在低峰时段（凌晨 2-4 点）进行生产环境部署',
          impact: '可降低部署风险 30%',
        },
        {
          type: 'scaling',
          title: '自动扩缩容建议',
          description: 'api-gateway 在工作日 9-11 点和 14-16 点流量高峰期建议增加副本数',
          impact: '可提升服务可用性 15%',
        },
        {
          type: 'resource',
          title: '资源优化',
          description: 'user-service 实际资源使用率仅为配置的 40%，建议调整资源配额',
          impact: '可节省资源成本 25%',
        },
      ];

      setRiskFindings(mockRisks);
      setAnomalies(mockAnomalies);
      setSuggestions(mockSuggestions);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'high':
        return 'error';
      case 'medium':
        return 'warning';
      case 'low':
        return 'default';
      default:
        return 'default';
    }
  };

  const getSeverityIcon = (severity: string) => {
    switch (severity) {
      case 'high':
        return <WarningOutlined style={{ color: '#ef4444' }} />;
      case 'medium':
        return <WarningOutlined style={{ color: '#f59e0b' }} />;
      case 'low':
        return <CheckCircleOutlined style={{ color: '#6c757d' }} />;
      default:
        return null;
    }
  };

  return (
    <div className="space-y-6">
      {/* Page header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-gray-900">AIOps 洞察</h1>
          <p className="text-sm text-gray-500 mt-1">AI 驱动的风险评估、异常检测和优化建议</p>
        </div>
        <Button icon={<ReloadOutlined />} onClick={load} loading={loading}>
          刷新
        </Button>
      </div>

      {/* Risk findings */}
      <Card title={<span><WarningOutlined className="mr-2" />风险发现</span>}>
        {riskFindings.length === 0 ? (
          <Empty description="暂无风险发现" />
        ) : (
          <div className="space-y-3">
            {riskFindings.map((risk, index) => (
              <Alert
                key={index}
                message={
                  <Space>
                    {getSeverityIcon(risk.severity)}
                    <Tag color={getSeverityColor(risk.severity)}>
                      {risk.severity.toUpperCase()}
                    </Tag>
                    <span className="font-semibold">{risk.category}</span>
                  </Space>
                }
                description={
                  <div className="mt-2">
                    <div className="mb-2">{risk.message}</div>
                    <div className="text-sm text-gray-600">
                      <strong>建议:</strong> {risk.recommendation}
                    </div>
                  </div>
                }
                type={risk.severity === 'high' ? 'error' : risk.severity === 'medium' ? 'warning' : 'info'}
                showIcon={false}
              />
            ))}
          </div>
        )}
      </Card>

      {/* Anomaly detection */}
      <Card title={<span><LineChartOutlined className="mr-2" />异常检测</span>}>
        {anomalies.length === 0 ? (
          <Empty description="暂无异常检测" />
        ) : (
          <Collapse>
            {anomalies.map((anomaly, index) => (
              <Panel
                key={index}
                header={
                  <Space>
                    <WarningOutlined style={{ color: '#f59e0b' }} />
                    <span>{anomaly.description}</span>
                    <Tag color="orange">{new Date(anomaly.timestamp).toLocaleString()}</Tag>
                  </Space>
                }
              >
                <div className="space-y-2">
                  <div>
                    <span className="text-sm font-semibold">指标: </span>
                    <span>{anomaly.metric}</span>
                  </div>
                  <div>
                    <span className="text-sm font-semibold">当前值: </span>
                    <span className="text-red-600 font-semibold">{anomaly.value}</span>
                  </div>
                  <div>
                    <span className="text-sm font-semibold">阈值: </span>
                    <span>{anomaly.threshold}</span>
                  </div>
                  <div>
                    <span className="text-sm font-semibold">时间: </span>
                    <span>{new Date(anomaly.timestamp).toLocaleString()}</span>
                  </div>
                </div>
              </Panel>
            ))}
          </Collapse>
        )}
      </Card>

      {/* Optimization suggestions */}
      <Card title={<span><BulbOutlined className="mr-2" />优化建议</span>}>
        {suggestions.length === 0 ? (
          <Empty description="暂无优化建议" />
        ) : (
          <Row gutter={[16, 16]}>
            {suggestions.map((suggestion, index) => (
              <Col xs={24} lg={8} key={index}>
                <Card className="h-full hover:shadow-lg transition-shadow">
                  <div className="space-y-3">
                    <div className="flex items-center justify-between">
                      <Tag color="blue">{suggestion.type}</Tag>
                      <BulbOutlined className="text-2xl text-yellow-500" />
                    </div>
                    <div>
                      <div className="font-semibold text-lg mb-2">{suggestion.title}</div>
                      <div className="text-sm text-gray-600 mb-3">{suggestion.description}</div>
                      <div className="text-xs text-green-600 font-semibold">
                        预期影响: {suggestion.impact}
                      </div>
                    </div>
                  </div>
                </Card>
              </Col>
            ))}
          </Row>
        )}
      </Card>
    </div>
  );
};

export default AIOpsInsightsPage;

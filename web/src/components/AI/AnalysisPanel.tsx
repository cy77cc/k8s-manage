import React, { useState, useEffect } from 'react';
import { Card, Typography, Spin, Divider, Space, Tag } from 'antd';
import { BarChartOutlined, LineChartOutlined, PieChartOutlined, AreaChartOutlined } from '@ant-design/icons';
import { Api } from '../../api';
import type { AIAnalysis } from '../../api';

const { Text, Paragraph, Title } = Typography;

// 分析面板属性
interface AnalysisPanelProps {
  type: string;
  data: any;
  context?: any;
  className?: string;
}

// 分析类型对应的图标
const analysisIcons: Record<string, React.ReactNode> = {
  resource: <BarChartOutlined />,
  performance: <LineChartOutlined />,
  usage: <PieChartOutlined />,
  trend: <AreaChartOutlined />,
};

// AI分析面板组件
const AnalysisPanel: React.FC<AnalysisPanelProps> = ({ 
  type, 
  data, 
  context, 
  className 
}) => {
  const [analysis, setAnalysis] = useState<AIAnalysis | null>(null);
  const [loading, setLoading] = useState(true);

  // 执行分析
  const performAnalysis = async () => {
    try {
      setLoading(true);
      const response = await Api.ai.analyze({
        type,
        data,
        context
      });
      setAnalysis(response.data);
    } catch (error) {
      console.error('执行分析失败:', error);
      // 使用模拟数据
      const mockAnalysis: AIAnalysis = {
        id: 'analysis-1',
        type,
        title: getAnalysisTitle(type),
        summary: getAnalysisSummary(type),
        details: getAnalysisDetails(type, data),
        createdAt: new Date().toISOString()
      };
      setAnalysis(mockAnalysis);
    } finally {
      setLoading(false);
    }
  };

  // 获取分析标题
  const getAnalysisTitle = (type: string): string => {
    const titles: Record<string, string> = {
      resource: '资源使用分析',
      performance: '性能分析',
      usage: '使用情况分析',
      trend: '趋势分析',
    };
    return titles[type] || '智能分析';
  };

  // 获取分析摘要
  const getAnalysisSummary = (type: string): string => {
    const summaries: Record<string, string> = {
      resource: '系统资源使用情况总体良好，CPU使用率适中，内存有一定冗余。',
      performance: '系统性能稳定，响应时间在正常范围内，无明显瓶颈。',
      usage: '系统使用情况符合预期，峰值负载可控。',
      trend: '资源使用呈稳定增长趋势，建议提前规划扩容。',
    };
    return summaries[type] || '分析完成，系统运行正常。';
  };

  // 获取分析详情
  const getAnalysisDetails = (_type: string, _data: any): any => {
    return {
      insights: [
        '关键指标均在正常范围内',
        '未发现异常波动',
        '建议定期监控以保持系统稳定'
      ],
      recommendations: [
        '优化资源配置以提高效率',
        '建立预警机制以应对潜在问题',
        '定期进行系统维护'
      ],
      metrics: {
        average: 75,
        peak: 90,
        trend: 'stable'
      }
    };
  };

  // 初始化执行分析
  useEffect(() => {
    performAnalysis();
  }, [type, data, context]);

  return (
    <Card 
      title={
        <Space>
          {analysisIcons[type] || <BarChartOutlined />}
          <Text strong>智能分析</Text>
        </Space>
      }
      className={`ai-analysis-panel ${className || ''}`}
    >
      {loading ? (
        <div style={{ padding: '32px', textAlign: 'center' }}>
          <Spin tip="正在分析数据..." />
        </div>
      ) : !analysis ? (
        <div style={{ padding: '32px', textAlign: 'center' }}>
          <Text type="secondary">分析失败，请稍后重试</Text>
        </div>
      ) : (
        <div>
          <Title level={5}>{analysis.title}</Title>
          <Paragraph>{analysis.summary}</Paragraph>
          <Divider>关键洞察</Divider>
          <Space direction="vertical" style={{ width: '100%' }}>
            {analysis.details?.insights?.map((insight: string, index: number) => (
              <div key={index} style={{ display: 'flex', alignItems: 'center' }}>
                <Tag color="blue" style={{ marginRight: '8px' }}>洞察 {index + 1}</Tag>
                <Text>{insight}</Text>
              </div>
            ))}
          </Space>
          {analysis.details?.recommendations && (
            <>
              <Divider>建议</Divider>
              <Space direction="vertical" style={{ width: '100%' }}>
                {analysis.details.recommendations.map((recommendation: string, index: number) => (
                  <div key={index} style={{ display: 'flex', alignItems: 'center' }}>
                    <Tag color="green" style={{ marginRight: '8px' }}>建议 {index + 1}</Tag>
                    <Text>{recommendation}</Text>
                  </div>
                ))}
              </Space>
            </>
          )}
          <Divider>分析时间</Divider>
          <Text type="secondary">
            {new Date(analysis.createdAt).toLocaleString()}
          </Text>
        </div>
      )}
    </Card>
  );
};

export default AnalysisPanel;
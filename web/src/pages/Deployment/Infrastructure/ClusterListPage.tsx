import React, { useState, useCallback, useEffect } from 'react';
import { Button, Card, Row, Col, Tag, Space, message, Statistic } from 'antd';
import { PlusOutlined, ReloadOutlined, ClusterOutlined, CheckCircleOutlined, ExclamationCircleOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { Api } from '../../../api';

interface Cluster {
  id: number;
  name: string;
  status: string;
  endpoint: string;
  management_mode: string;
  node_count?: number;
  created_at: string;
}

const ClusterListPage: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [clusters, setClusters] = useState<Cluster[]>([]);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const res = await Api.cluster.getClusters();
      setClusters(res.data.list || []);
    } catch (err) {
      message.error(err instanceof Error ? err.message : '加载集群列表失败');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  const getStatusColor = (status: string) => {
    const statusMap: Record<string, string> = {
      active: 'success',
      inactive: 'default',
      error: 'error',
      pending: 'processing',
    };
    return statusMap[status] || 'default';
  };

  const getStatusIcon = (status: string) => {
    return status === 'active' ? (
      <CheckCircleOutlined style={{ color: '#52c41a' }} />
    ) : (
      <ExclamationCircleOutlined style={{ color: '#faad14' }} />
    );
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-gray-900">集群管理</h1>
          <p className="text-sm text-gray-500 mt-1">管理 Kubernetes 集群</p>
        </div>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={load} loading={loading}>
            刷新
          </Button>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => navigate('/deployment/infrastructure/clusters/bootstrap')}
          >
            创建集群
          </Button>
        </Space>
      </div>

      <Row gutter={[16, 16]}>
        {clusters.map((cluster) => (
          <Col xs={24} sm={12} lg={8} key={cluster.id}>
            <Card
              hoverable
              className="h-full cursor-pointer transition-shadow hover:shadow-lg"
              onClick={() => navigate(`/deployment/infrastructure/clusters/${cluster.id}`)}
            >
              <Space direction="vertical" className="w-full" size="middle">
                <div className="flex items-center justify-between">
                  <Space>
                    <ClusterOutlined className="text-2xl text-blue-500" />
                    <span className="text-lg font-semibold">{cluster.name}</span>
                  </Space>
                  {getStatusIcon(cluster.status)}
                </div>

                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-gray-600">状态:</span>
                    <Tag color={getStatusColor(cluster.status)}>
                      {cluster.status}
                    </Tag>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-gray-600">管理模式:</span>
                    <Tag color="purple">{cluster.management_mode || 'platform'}</Tag>
                  </div>
                  {cluster.node_count !== undefined && (
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-gray-600">节点数:</span>
                      <span className="font-semibold">{cluster.node_count}</span>
                    </div>
                  )}
                </div>

                <div className="pt-2 border-t border-gray-200">
                  <div className="text-xs text-gray-500 truncate">
                    Endpoint: {cluster.endpoint || '-'}
                  </div>
                  <div className="text-xs text-gray-400 mt-1">
                    创建于: {new Date(cluster.created_at).toLocaleDateString()}
                  </div>
                </div>
              </Space>
            </Card>
          </Col>
        ))}
      </Row>

      {clusters.length === 0 && !loading && (
        <Card className="text-center py-12">
          <ClusterOutlined className="text-6xl text-gray-300 mb-4" />
          <p className="text-gray-500 mb-4">暂无集群</p>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => navigate('/deployment/infrastructure/clusters/bootstrap')}
          >
            创建第一个集群
          </Button>
        </Card>
      )}
    </div>
  );
};

export default ClusterListPage;

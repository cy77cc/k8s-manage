import React, { useState, useCallback, useEffect } from 'react';
import { Button, Card, Descriptions, Tag, Space, message, Tabs, Table, Empty } from 'antd';
import { ArrowLeftOutlined, ReloadOutlined, CheckCircleOutlined, CloseCircleOutlined } from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import { Api } from '../../../api';

interface ClusterDetail {
  id: number;
  name: string;
  status: string;
  endpoint: string;
  management_mode: string;
  node_count?: number;
  version?: string;
  created_at: string;
  updated_at: string;
}

interface ClusterNode {
  id: number;
  name: string;
  ip: string;
  role: string;
  status: string;
  cpu_cores?: number;
  memory_gb?: number;
}

const ClusterDetailPage: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const [loading, setLoading] = useState(false);
  const [cluster, setCluster] = useState<ClusterDetail | null>(null);
  const [nodes, setNodes] = useState<ClusterNode[]>([]);

  const load = useCallback(async () => {
    if (!id) return;
    setLoading(true);
    try {
      const res = await Api.cluster.getClusterDetail(Number(id));
      setCluster(res.data);
      // Load nodes if available
      if (res.data.node_count && res.data.node_count > 0) {
        const nodesRes = await Api.cluster.getClusterNodes(Number(id));
        setNodes(nodesRes.data.list || []);
      }
    } catch (err) {
      message.error(err instanceof Error ? err.message : '加载集群详情失败');
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    void load();
  }, [load]);

  const nodeColumns = [
    {
      title: '节点名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: 'IP 地址',
      dataIndex: 'ip',
      key: 'ip',
    },
    {
      title: '角色',
      dataIndex: 'role',
      key: 'role',
      render: (role: string) => (
        <Tag color={role === 'control-plane' ? 'blue' : 'green'}>
          {role === 'control-plane' ? 'Control Plane' : 'Worker'}
        </Tag>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Space>
          {status === 'ready' ? (
            <CheckCircleOutlined style={{ color: '#52c41a' }} />
          ) : (
            <CloseCircleOutlined style={{ color: '#ff4d4f' }} />
          )}
          <span>{status}</span>
        </Space>
      ),
    },
    {
      title: 'CPU',
      dataIndex: 'cpu_cores',
      key: 'cpu_cores',
      render: (cores?: number) => cores ? `${cores} cores` : '-',
    },
    {
      title: '内存',
      dataIndex: 'memory_gb',
      key: 'memory_gb',
      render: (memory?: number) => memory ? `${memory} GB` : '-',
    },
  ];

  if (!cluster && !loading) {
    return (
      <div className="space-y-6">
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/deployment/infrastructure/ers')}>
          返回列表
        </Button>
        <Card>
          <Empty description="未找到集群" />
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/deployment/infrastructure/clusters')}>
            返回列表
          </Button>
          <div>
            <div className="flex items-center gap-3">
              <h1 className="text-2xl font-semibold text-gray-900">
          {cluster?.name || `集群 #${id}`}
              </h1>
              {cluster && (
                <Tag color={cluster.status === 'active' ? 'success' : 'default'}>
                  {cluster.status}
                </Tag>
              )}
            </div>
            <p className="text-sm text-gray-500 mt-1">
              {cluster?.created_at && `创建于 ${new Date(cluster.created_at).toLocaleString()}`}
            </p>
          </div>
        </div>
        <Button icon={<ReloadOutlined />} onClick={load} loading={loading}>
          刷新
        </Button>
      </div>

      {cluster && (
        <Card title={<span className="text-base font-semibold">基本信息</span>}>
          <Descriptions bordered column={2}>
            <Descriptions.Item label="集群 ID">{cluster.id}</Descriptions.Item>
            <Descriptions.Item label="状态">
              <Tag color={cluster.status === 'active' ? 'success' : 'default'}>
                {cluster.status}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="名称">{cluster.name}</Descriptions.Item>
            <Descriptions.Item label="管理模式">
              <Tag color="purple">{cluster.management_mode || 'platform'}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="Endpoint" span={2}>
              {cluster.endpoint || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="版本">{cluster.version || '-'}</Descriptions.Item>
            <Descriptions.Item label="节点数">{cluster.node_count || 0}</Descriptions.Item>
            <Descriptions.Item label="创建时间">
              {new Date(cluster.created_at).toLocaleString()}
            </Descriptions.Item>
            <Descriptions.Item label="更新时间">
              {new Date(cluster.updated_at).toLocaleString()}
            </Descriptions.Item>
          </Descriptions>
        </Card>
      )}

      <Card title={<span className="text-base font-semibold">集群节点</span>}>
        {nodes.length > 0 ? (
          <Table
            columns={nodeColumns}
            dataSource={nodes}
            rowKey="id"
            pagination={false}
          />
        ) : (
          <Empty description="暂无节点信息" image={Empty.PRESENTED_IMAGE_SIMPLE} />
        )}
      </Card>

      <Card title={<span className="text-base font-semibold">部署历史</span>}>
        <Empty description="部署历史功能即将推出" image={Empty.PRESENTED_IMAGE_SIMPLE} />
      </Card>
    </div>
  );
};

export default ClusterDetailPage;

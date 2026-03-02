import React, { useState, useEffect, useMemo } from 'react';
import { Card, Button, Space, Select, Empty, message } from 'antd';
import { ReloadOutlined, FullscreenOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';

interface ServiceNode {
  id: string;
  name: string;
  environment: string;
  status: 'healthy' | 'degraded' | 'down';
  lastDeployment?: string;
}

const DeploymentTopologyPage: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [services, setServices] = useState<ServiceNode[]>([]);
  const [envFilter, setEnvFilter] = useState<string>('all');
  const [selectedNode, setSelectedNode] = useState<ServiceNode | null>(null);

  const load = async () => {
    setLoading(true);
    try {
      // Mock data - replace with actual API call
      const mockServices: ServiceNode[] = [
        {
          id: '1',
          name: 'api-gateway',
          environment: 'production',
          status: 'healthy',
          lastDeployment: new Date().toISOString(),
        },
        {
          id: '2',
          name: 'user-service',
          environment: 'production',
          status: 'healthy',
          lastDeployment: new Date(Date.now() - 86400000).toISOString(),
        },
        {
          id: '3',
          name: 'order-service',
          environment: 'staging',
          status: 'degraded',
          lastDeployment: new Date(Date.now() - 3600000).toISOString(),
        },
      ];
      setServices(mockServices);
    } catch (err) {
      message.error('加载拓扑数据失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
    // Poll for updates
    const interval = setInterval(load, 10000);
    return () => clearInterval(interval);
  }, []);

  const filteredServices = useMemo(() => {
    if (envFilter === 'all') return services;
    return services.filter((s) => s.environment === envFilter);
  }, [services, envFilter]);

  const groupedByEnv = useMemo(() => {
    const groups: Record<string, ServiceNode[]> = {};
    filteredServices.forEach((s) => {
      if (!groups[s.environment]) {
        groups[s.environment] = [];
      }
      groups[s.environment].push(s);
    });
    return groups;
  }, [filteredServices]);

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'healthy':
        return '#10b981';
      case 'degraded':
        return '#f59e0b';
      case 'down':
        return '#ef4444';
      default:
        return '#6c757d';
    }
  };

  return (
    <div className="space-y-6">
      {/* Page header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-gray-900">部署拓扑</h1>
          <p className="text-sm text-gray-500 mt-1">可视化服务部署状态和依赖关系</p>
        </div>
        <Space>
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
          <Button icon={<FullscreenOutlined />}>全屏</Button>
        </Space>
      </div>

      {/* Topology visualization */}
      {Object.keys(groupedByEnv).length === 0 ? (
        <Card>
          <Empty description="暂无服务数据" />
        </Card>
      ) : (
        <div className="space-y-6">
          {Object.entries(groupedByEnv).map(([env, envServices]) => (
            <Card key={env} title={`${env} 环境`}>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
                {envServices.map((service) => (
                  <div
                    key={service.id}
                    className="p-4 border-2 rounded-lg cursor-pointer transition-all hover:shadow-lg"
                    style={{
                      borderColor: getStatusColor(service.status),
                      backgroundColor: selectedNode?.id === service.id ? '#f0f9ff' : 'white',
                    }}
                    onClick={() => setSelectedNode(service)}
                  >
                    <div className="flex items-center justify-between mb-2">
                      <span className="font-semibold text-gray-900">{service.name}</span>
                      <div
                        className="w-3 h-3 rounded-full"
                        style={{ backgroundColor: getStatusColor(service.status) }}
                      />
                    </div>
                    <div className="text-xs text-gray-500">
                      <div>状态: {service.status}</div>
                      {service.lastDeployment && (
                        <div>最后部署: {new Date(service.lastDeployment).toLocaleString()}</div>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </Card>
          ))}
        </div>
      )}

      {/* Selected node details */}
      {selectedNode && (
        <Card title="服务详情">
          <div className="space-y-3">
            <div>
              <span className="text-sm font-semibold">服务名称: </span>
              <span>{selectedNode.name}</span>
            </div>
            <div>
              <span className="text-sm font-semibold">环境: </span>
              <span>{selectedNode.environment}</span>
            </div>
            <div>
              <span className="text-sm font-semibold">状态: </span>
              <span
                className="px-2 py-1 rounded text-white text-xs"
                style={{ backgroundColor: getStatusColor(selectedNode.status) }}
              >
                {selectedNode.status}
              </span>
            </div>
            {selectedNode.lastDeployment && (
              <div>
                <span className="text-sm font-semibold">最后部署: </span>
                <span>{new Date(selectedNode.lastDeployment).toLocaleString()}</span>
              </div>
            )}
            <div className="pt-3">
              <Button type="primary" onClick={() => navigate(`/services/${selectedNode.id}`)}>
                查看服务详情
              </Button>
            </div>
          </div>
        </Card>
      )}
    </div>
  );
};

export default DeploymentTopologyPage;

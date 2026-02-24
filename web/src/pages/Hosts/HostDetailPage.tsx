import React, { useEffect, useMemo, useState } from 'react';
import { Breadcrumb, Button, Card, Col, Descriptions, Modal, Row, Space, Statistic, Table, Tag, message } from 'antd';
import { ArrowLeftOutlined, ReloadOutlined } from '@ant-design/icons';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { Api } from '../../api';
import type { Host, HostAuditItem, HostMetricPoint } from '../../api/modules/hosts';

const HostDetailPage: React.FC = () => {
  const { id = '' } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [host, setHost] = useState<Host | null>(null);
  const [metrics, setMetrics] = useState<HostMetricPoint[]>([]);
  const [audits, setAudits] = useState<HostAuditItem[]>([]);

  const load = async () => {
    if (!id) return;
    setLoading(true);
    try {
      const [hostRes, metricRes, auditRes] = await Promise.all([
        Api.hosts.getHostDetail(id),
        Api.hosts.getHostMetrics(id),
        Api.hosts.getHostAudits(id),
      ]);
      setHost(hostRes.data);
      setMetrics(metricRes.data || []);
      setAudits(auditRes.data || []);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, [id]);

  const latest = useMemo(() => {
    if (metrics.length === 0) return { cpu: 0, memory: 0, disk: 0, network: 0 };
    return metrics[0];
  }, [metrics]);

  const runAction = (action: string, confirm = false) => {
    const exec = async () => {
      await Api.hosts.hostAction(id, action);
      message.success('操作已提交');
      await load();
    };
    if (!confirm) {
      exec();
      return;
    }
    Modal.confirm({
      title: `确认执行 ${action}`,
      onOk: exec,
    });
  };

  return (
    <div>
      <Breadcrumb className="mb-4">
        <Breadcrumb.Item><Link to="/hosts">主机管理</Link></Breadcrumb.Item>
        <Breadcrumb.Item>{host?.name || id}</Breadcrumb.Item>
      </Breadcrumb>

      <Card
        loading={loading}
        title={<Space><Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/hosts')}>返回</Button><span>{host?.name || '主机详情'}</span></Space>}
        extra={
          <Space>
            <Button icon={<ReloadOutlined />} onClick={load}>刷新</Button>
            <Button onClick={() => navigate(`/hosts/terminal/${id}`)}>终端</Button>
            <Button onClick={() => runAction('check')}>巡检</Button>
            <Button onClick={() => runAction('restart', true)}>重启</Button>
            <Button danger onClick={() => runAction('shutdown', true)}>关机</Button>
          </Space>
        }
      >
        <Row gutter={[16, 16]}>
          <Col span={6}><Card><Statistic title="CPU" value={latest.cpu} suffix="%" /></Card></Col>
          <Col span={6}><Card><Statistic title="内存" value={latest.memory} suffix="%" /></Card></Col>
          <Col span={6}><Card><Statistic title="磁盘" value={latest.disk} suffix="%" /></Card></Col>
          <Col span={6}><Card><Statistic title="网络" value={latest.network} suffix="Mbps" /></Card></Col>
        </Row>

        <div className="mt-4" />
        <Descriptions bordered size="small" column={2}>
          <Descriptions.Item label="名称">{host?.name}</Descriptions.Item>
          <Descriptions.Item label="状态"><Tag color={host?.status === 'online' ? 'success' : host?.status === 'maintenance' ? 'warning' : 'default'}>{host?.status || '-'}</Tag></Descriptions.Item>
          <Descriptions.Item label="IP">{host?.ip}</Descriptions.Item>
          <Descriptions.Item label="系统">{host?.os || '-'}</Descriptions.Item>
          <Descriptions.Item label="SSH">{host?.username || 'root'}:{host?.port || 22}</Descriptions.Item>
          <Descriptions.Item label="区域">{host?.region || '-'}</Descriptions.Item>
          <Descriptions.Item label="CPU 核数">{host?.cpu}</Descriptions.Item>
          <Descriptions.Item label="内存 MB">{host?.memory}</Descriptions.Item>
          <Descriptions.Item label="磁盘 GB">{host?.disk}</Descriptions.Item>
          <Descriptions.Item label="更新时间">{host?.lastActive ? new Date(host.lastActive).toLocaleString() : '-'}</Descriptions.Item>
        </Descriptions>

        <div className="mt-4" />
        <Card title="指标序列（最近60条）" size="small">
          <Table
            rowKey="id"
            pagination={{ pageSize: 10 }}
            dataSource={metrics}
            columns={[
              { title: '时间', dataIndex: 'createdAt', render: (v: string) => (v ? new Date(v).toLocaleString() : '-') },
              { title: 'CPU', dataIndex: 'cpu', render: (v: number) => `${v}%` },
              { title: '内存', dataIndex: 'memory', render: (v: number) => `${v}%` },
              { title: '磁盘', dataIndex: 'disk', render: (v: number) => `${v}%` },
              { title: '网络', dataIndex: 'network', render: (v: number) => `${v} Mbps` },
            ]}
          />
        </Card>

        <div className="mt-4" />
        <Card title="操作审计" size="small">
          <Table
            rowKey="id"
            pagination={{ pageSize: 10 }}
            dataSource={audits}
            columns={[
              { title: '时间', dataIndex: 'createdAt', render: (v: string) => (v ? new Date(v).toLocaleString() : '-') },
              { title: '动作', dataIndex: 'action' },
              { title: '操作人', dataIndex: 'operator' },
              { title: '详情', dataIndex: 'detail' },
            ]}
          />
        </Card>
      </Card>
    </div>
  );
};

export default HostDetailPage;

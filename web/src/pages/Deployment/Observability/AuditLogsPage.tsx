import React, { useState, useEffect } from 'react';
import { Card, Table, Button, Space, Input, Select, DatePicker, Tag, Modal, message } from 'antd';
import {
  ReloadOutlined,
  SearchOutlined,
  DownloadOutlined,
  EyeOutlined,
  FilterOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';

const { RangePicker } = DatePicker;

interface AuditLog {
  id: string;
  action: string;
  actor: string;
  actor_email?: string;
  resource_type: string;
  resource_id: string;
  detail: any;
  ip_address?: string;
  user_agent?: string;
  created_at: string;
}

const AuditLogsPage: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [actionFilter, setActionFilter] = useState<string>('all');
  const [actorFilter, setActorFilter] = useState<string>('');
  const [searchQuery, setSearchQuery] = useState('');
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs | null, dayjs.Dayjs | null]>([null, null]);
  const [detailModalVisible, setDetailModalVisible] = useState(false);
  const [selectedLog, setSelectedLog] = useState<AuditLog | null>(null);

  const load = async () => {
    setLoading(true);
    try {
      // Mock data - replace with actual API call
      const mockLogs: AuditLog[] = [
        {
          id: '1',
          action: 'release.create',
          actor: 'admin',
          actor_email: 'admin@example.com',
          resource_type: 'release',
          resource_id: '123',
          detail: { service_id: 1, target_id: 2, strategy: 'rolling' },
          ip_address: '192.168.1.100',
          user_agent: 'Mozilla/5.0',
          created_at: new Date().toISOString(),
        },
        {
          id: '2',
          action: 'release.approve',
          actor: 'reviewer',
          actor_email: 'reviewer@example.com',
          resource_type: 'release',
          resource_id: '123',
          detail: { comment: 'Approved for production' },
          ip_address: '192.168.1.101',
          created_at: new Date(Date.now() - 3600000).toISOString(),
        },
      ];
      setLogs(mockLogs);
    } catch (err) {
      message.error('加载审计日志失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const handleExport = (format: 'csv' | 'json') => {
    const filteredLogs = getFilteredLogs();
    if (format === 'json') {
      const dataStr = JSON.stringify(filteredLogs, null, 2);
      const dataBlob = new Blob([dataStr], { type: 'application/json' });
      const url = URL.createObjectURL(dataBlob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `audit-logs-${Date.now()}.json`;
      link.click();
    } else {
      // CSV export
      const headers = ['ID', 'Action', 'Actor', 'Resource Type', 'Resource ID', 'IP Address', 'Created At'];
      const csvContent = [
        headers.join(','),
        ...filteredLogs.map((log) =>
          [
            log.id,
            log.action,
            log.actor,
            log.resource_type,
            log.resource_id,
            log.ip_address || '',
            log.created_at,
          ].join(',')
        ),
      ].join('\n');
      const dataBlob = new Blob([csvContent], { type: 'text/csv' });
      const url = URL.createObjectURL(dataBlob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `audit-logs-${Date.now()}.csv`;
      link.click();
    }
    message.success(`已导出 ${format.toUpperCase()} 格式`);
  };

  const getFilteredLogs = () => {
    let filtered = logs;

    // Action filter
    if (actionFilter !== 'all') {
      filtered = filtered.filter((log) => log.action.startsWith(actionFilter));
    }

    // Actor filter
    if (actorFilter) {
      filtered = filtered.filter((log) => log.actor.toLowerCase().includes(actorFilter.toLowerCase()));
    }

    // Search filter
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(
        (log) =>
          log.action.toLowerCase().includes(query) ||
          log.actor.toLowerCase().includes(query) ||
          JSON.stringify(log.detail).toLowerCase().includes(query)
      );
    }

    // Date range filter
    if (dateRange[0] && dateRange[1]) {
      filtered = filtered.filter((log) => {
        const logDate = dayjs(log.created_at);
        return logDate.isAfter(dateRange[0]) && logDate.isBefore(dateRange[1]);
      });
    }

    return filtered;
  };

  const columns: ColumnsType<AuditLog> = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: '操作',
      dataIndex: 'action',
      key: 'action',
      render: (action: string) => {
        const colors: Record<string, string> = {
          create: 'blue',
          update: 'orange',
          delete: 'red',
          approve: 'green',
          reject: 'default',
        };
        const actionType = action.split('.')[1] || action;
        return <Tag color={colors[actionType] || 'default'}>{action}</Tag>;
      },
    },
    {
      title: '操作人',
      dataIndex: 'actor',
      key: 'actor',
      render: (actor: string, record: AuditLog) => (
        <div>
          <div>{actor}</div>
          {record.actor_email && <div className="text-xs text-gray-500">{record.actor_email}</div>}
        </div>
      ),
    },
    {
      title: '资源类型',
      dataIndex: 'resource_type',
      key: 'resource_type',
    },
    {
      title: '资源 ID',
      dataIndex: 'resource_id',
      key: 'resource_id',
      render: (id: string, record: AuditLog) => {
        if (record.resource_type === 'release') {
          return <a onClick={() => navigate(`/deployment/${id}`)}>#{id}</a>;
        }
        return id;
      },
    },
    {
      title: 'IP 地址',
      dataIndex: 'ip_address',
      key: 'ip_address',
      render: (ip: string) => ip || '-',
    },
    {
      title: '时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (time: string) => new Date(time).toLocaleString(),
    },
    {
      title: '操作',
      key: 'actions',
      width: 100,
      render: (_: any, record: AuditLog) => (
        <Button
          type="link"
          size="small"
          icon={<EyeOutlined />}
          onClick={() => {
            setSelectedLog(record);
            setDetailModalVisible(true);
          }}
        >
          详情
        </Button>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      {/* Page header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-gray-900">审计日志</h1>
          <p className="text-sm text-gray-500 mt-1">查看所有系统操作记录</p>
        </div>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={load} loading={loading}>
            刷新
          </Button>
          <Button icon={<DownloadOutlined />} onClick={() => handleExport('csv')}>
            导出 CSV
          </Button>
          <Button icon={<DownloadOutlined />} onClick={() => handleExport('json')}>
            导出 JSON
          </Button>
        </Space>
      </div>

      {/* Filters */}
      <Card>
        <Space direction="vertical" size="middle" className="w-full">
          <div className="flex flex-wrap gap-3">
            <Input
              placeholder="搜索操作、操作人或详情"
              prefix={<SearchOutlined className="text-gray-400" />}
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              style={{ width: 280 }}
              allowClear
            />
            <Select
              value={actionFilter}
              style={{ width: 160 }}
              placeholder="操作类型"
              options={[
                { value: 'all', label: '全部操作' },
                { value: 'release', label: '发布相关' },
                { value: 'target', label: '目标相关' },
                { value: 'cluster', label: '集群相关' },
                { value: 'credential', label: '凭证相关' },
              ]}
              onChange={setActionFilter}
            />
            <Input
              placeholder="操作人筛选"
              value={actorFilter}
              onChange={(e) => setActorFilter(e.target.value)}
              style={{ width: 160 }}
              allowClear
            />
            <RangePicker
              value={dateRange}
              onChange={(dates) => setDateRange(dates as [dayjs.Dayjs | null, dayjs.Dayjs | null])}
              style={{ width: 280 }}
            />
          </div>
        </Space>
      </Card>

      {/* Audit log table */}
      <Card>
        <Table
          dataSource={getFilteredLogs()}
          columns={columns}
          rowKey="id"
          loading={loading}
          pagination={{ pageSize: 20, showSizeChanger: true, showTotal: (total) => `共 ${total} 条` }}
        />
      </Card>

      {/* Detail modal */}
      <Modal
        title="审计日志详情"
        open={detailModalVisible}
        onCancel={() => {
          setDetailModalVisible(false);
          setSelectedLog(null);
        }}
        footer={[
          <Button key="close" onClick={() => setDetailModalVisible(false)}>
            关闭
          </Button>,
        ]}
        width={720}
      >
        {selectedLog && (
          <div className="space-y-4">
            <div>
              <div className="text-sm font-semibold mb-1">操作:</div>
              <Tag color="blue">{selectedLog.action}</Tag>
            </div>
            <div>
              <div className="text-sm font-semibold mb-1">操作人:</div>
              <div>{selectedLog.actor}</div>
              {selectedLog.actor_email && <div className="text-xs text-gray-500">{selectedLog.actor_email}</div>}
            </div>
            <div>
              <div className="text-sm font-semibold mb-1">资源:</div>
              <div>
                {selectedLog.resource_type} #{selectedLog.resource_id}
              </div>
            </div>
            {selectedLog.ip_address && (
              <div>
                <div className="text-sm font-semibold mb-1">IP 地址:</div>
                <div>{selectedLog.ip_address}</div>
              </div>
            )}
            {selectedLog.user_agent && (
              <div>
                <div className="text-sm font-semibold mb-1">User Agent:</div>
                <div className="text-xs text-gray-600">{selectedLog.user_agent}</div>
              </div>
            )}
            <div>
              <div className="text-sm font-semibold mb-1">时间:</div>
              <div>{new Date(selectedLog.created_at).toLocaleString()}</div>
            </div>
            <div>
              <div className="text-sm font-semibold mb-2">详细信息:</div>
              <pre className="bg-gray-900 text-gray-100 p-4 rounded overflow-auto max-h-96 text-xs">
                {JSON.stringify(selectedLog.detail, null, 2)}
              </pre>
            </div>
          </div>
        )}
      </Modal>
    </div>
  );
};

export default AuditLogsPage;

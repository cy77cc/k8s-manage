import React, { useEffect, useMemo, useState } from 'react';
import { Button, Card, Dropdown, Input, Modal, Select, Space, Table, Tag, message } from 'antd';
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons';
import { Api } from '../../api';
import type { Host } from '../../api/modules/hosts';
import { useNavigate } from 'react-router-dom';

const HostListPage: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [hosts, setHosts] = useState<Host[]>([]);
  const [search, setSearch] = useState('');
  const [status, setStatus] = useState('all');
  const [selected, setSelected] = useState<React.Key[]>([]);
  const [group, setGroup] = useState('');

  const load = async () => {
    setLoading(true);
    try {
      const res = await Api.hosts.getHostList({ page: 1, pageSize: 200, status: status === 'all' ? undefined : status, region: group || undefined });
      setHosts(res.data.list || []);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
    const handler = () => load();
    window.addEventListener('project:changed', handler as EventListener);
    return () => window.removeEventListener('project:changed', handler as EventListener);
  }, [status, group]);

  const filtered = useMemo(() => hosts.filter((h) => {
    const hitSearch = h.name.toLowerCase().includes(search.toLowerCase()) || h.ip.includes(search) || (h.region || '').toLowerCase().includes(search.toLowerCase());
    const hitStatus = status === 'all' || h.status === status;
    return hitSearch && hitStatus;
  }), [hosts, search, status]);

  const batchAction = async (action: string) => {
    if (selected.length === 0) {
      message.warning('请选择主机');
      return;
    }
    await Api.hosts.batchUpdate({
      hostIds: selected.map((id) => String(id)),
      action,
    });
    message.success('批量操作已执行');
    setSelected([]);
    load();
  };

  const quickAction = async (id: string, action: string) => {
    await Api.hosts.hostAction(id, action);
    message.success('操作成功');
    load();
  };

  const batchExec = async () => {
    if (selected.length === 0) {
      message.warning('请选择主机');
      return;
    }
    const command = window.prompt('请输入要批量执行的命令', 'hostname');
    if (!command) return;
    const res = await Api.hosts.batchExec(selected.map((id) => String(id)), command);
    message.success(`批量执行完成: ${Object.keys(res.data || {}).length} 台`);
  };

  return (
    <Card
      title="主机管理"
      extra={
        <Space>
          <Input placeholder="搜索名称/IP/区域" value={search} onChange={(e) => setSearch(e.target.value)} style={{ width: 220 }} />
          <Select value={status} onChange={setStatus} style={{ width: 130 }} options={[{ value: 'all', label: '全部状态' }, { value: 'online', label: 'online' }, { value: 'offline', label: 'offline' }, { value: 'maintenance', label: 'maintenance' }, { value: 'error', label: 'error' }]} />
          <Input placeholder="分组(可选)" value={group} onChange={(e) => setGroup(e.target.value)} style={{ width: 130 }} />
          <Button onClick={() => batchAction('maintenance')}>批量维护</Button>
          <Button onClick={() => batchAction('online')}>批量上线</Button>
          <Button onClick={batchExec}>批量SSH执行</Button>
          <Button icon={<ReloadOutlined />} onClick={load} loading={loading}>刷新</Button>
          <Dropdown
            menu={{
              items: [
                { key: 'onboarding', label: 'SSH接入（密码/密钥）' },
                { key: 'cloud', label: '云平台导入（阿里云/腾讯云）' },
                { key: 'virt', label: 'KVM虚拟化创建' },
                { key: 'keys', label: 'SSH密钥管理' },
              ],
              onClick: ({ key }) => {
                if (key === 'onboarding') navigate('/hosts/onboarding');
                if (key === 'cloud') navigate('/hosts/cloud-import');
                if (key === 'virt') navigate('/hosts/virtualization');
                if (key === 'keys') navigate('/hosts/keys');
              },
            }}
          >
            <Button type="primary" icon={<PlusOutlined />}>新增主机</Button>
          </Dropdown>
        </Space>
      }
    >
      <Table
        rowKey="id"
        rowSelection={{ selectedRowKeys: selected, onChange: setSelected }}
        loading={loading}
        dataSource={filtered}
        columns={[
          { title: '名称', dataIndex: 'name', render: (v: string, r: Host) => <a onClick={() => navigate(`/hosts/detail/${r.id}`)}>{v}</a> },
          { title: 'IP', dataIndex: 'ip' },
          { title: '状态', dataIndex: 'status', render: (v: string) => <Tag color={v === 'online' ? 'success' : v === 'error' ? 'error' : v === 'maintenance' ? 'warning' : 'default'}>{v}</Tag> },
          { title: '区域', dataIndex: 'region' },
          { title: 'CPU', dataIndex: 'cpu' },
          { title: '内存(MB)', dataIndex: 'memory' },
          { title: '磁盘(GB)', dataIndex: 'disk' },
          {
            title: '操作',
            render: (_: unknown, row: Host) => (
              <Space>
                <Button type="link" onClick={() => quickAction(row.id, 'check')}>检查</Button>
                <Button type="link" onClick={() => quickAction(row.id, 'restart')}>重启</Button>
                <Button type="link" onClick={async () => {
                  const command = window.prompt('请输入命令', 'uptime');
                  if (!command) return;
                  const res = await Api.hosts.sshExec(row.id, command);
                  Modal.info({ title: '执行结果', content: <pre>{res.data.stdout || res.data.stderr || ''}</pre>, width: 720 });
                }}>SSH执行</Button>
                <Button type="link" onClick={() => navigate(`/hosts/terminal/${row.id}`)}>终端</Button>
              </Space>
            ),
          },
        ]}
      />

    </Card>
  );
};

export default HostListPage;

import React from 'react';
import {
  Alert,
  Breadcrumb,
  Button,
  Card,
  Col,
  Input,
  Modal,
  Row,
  Space,
  Statistic,
  Tag,
  Typography,
} from 'antd';
import {
  ArrowLeftOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  CloseCircleOutlined,
  DisconnectOutlined,
  ReloadOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { Api } from '../../api';
import type { Host } from '../../api/modules/hosts';

type ConnectionStatus = 'connecting' | 'connected' | 'disconnected' | 'error';
type LineType = 'input' | 'output' | 'error' | 'system';

interface TerminalLine {
  id: number;
  type: LineType;
  content: string;
  timestamp: string;
}

const lineColors: Record<LineType, string> = {
  input: '#58d68d',
  output: '#c9d1d9',
  error: '#ff6b6b',
  system: '#8b949e',
};

const quickCommands = ['pwd', 'whoami', 'hostname', 'uptime', 'ls -la', 'df -h', 'free -m', 'top -bn1 | head -20'];

const HostTerminalPage: React.FC = () => {
  const navigate = useNavigate();
  const { id = '' } = useParams<{ id: string }>();

  const [loading, setLoading] = React.useState(false);
  const [executing, setExecuting] = React.useState(false);
  const [status, setStatus] = React.useState<ConnectionStatus>('connecting');
  const [host, setHost] = React.useState<Host | null>(null);
  const [command, setCommand] = React.useState('');
  const [lines, setLines] = React.useState<TerminalLine[]>([]);
  const [sessionSeconds, setSessionSeconds] = React.useState(0);
  const [lastLatencyMs, setLastLatencyMs] = React.useState<number>(0);

  const lineIdRef = React.useRef(1);
  const terminalRef = React.useRef<HTMLDivElement>(null);
  const historyRef = React.useRef<string[]>([]);
  const historyIndexRef = React.useRef<number>(-1);

  const pushLine = React.useCallback((type: LineType, content: string) => {
    const items = String(content || '')
      .split('\n')
      .map((x) => x.replace(/\r/g, ''))
      .filter((x) => x.trim() !== '');
    if (items.length === 0) return;
    setLines((prev) => [
      ...prev,
      ...items.map((item) => ({ id: lineIdRef.current++, type, content: item, timestamp: new Date().toISOString() })),
    ]);
  }, []);

  React.useEffect(() => {
    if (terminalRef.current) {
      terminalRef.current.scrollTop = terminalRef.current.scrollHeight;
    }
  }, [lines]);

  React.useEffect(() => {
    if (status !== 'connected') return;
    const timer = setInterval(() => setSessionSeconds((prev) => prev + 1), 1000);
    return () => clearInterval(timer);
  }, [status]);

  const formatDuration = (total: number) => {
    const h = Math.floor(total / 3600);
    const m = Math.floor((total % 3600) / 60);
    const s = total % 60;
    return `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`;
  };

  const statusTag = React.useMemo(() => {
    if (status === 'connecting') return <Tag icon={<ClockCircleOutlined />} color="processing">连接中</Tag>;
    if (status === 'connected') return <Tag icon={<CheckCircleOutlined />} color="success">已连接</Tag>;
    if (status === 'error') return <Tag icon={<CloseCircleOutlined />} color="error">连接失败</Tag>;
    return <Tag icon={<CloseCircleOutlined />}>已断开</Tag>;
  }, [status]);

  const connect = React.useCallback(async () => {
    if (!id) return;
    setLoading(true);
    setStatus('connecting');
    try {
      const [hostResp, checkResp] = await Promise.all([Api.hosts.getHostDetail(id), Api.hosts.sshCheck(id)]);
      setHost(hostResp.data);
      const reachable = !!checkResp.data?.reachable;
      if (!reachable) {
        setStatus('error');
        pushLine('error', `SSH 连接失败: ${checkResp.data?.message || 'unknown error'}`);
        return;
      }
      setLines([]);
      setSessionSeconds(0);
      setStatus('connected');
      pushLine('system', `Connected to ${hostResp.data.name} (${hostResp.data.ip})`);
      pushLine('system', '输入命令后按 Enter 执行，执行结果来自真实主机。');
    } catch (err) {
      setStatus('error');
      pushLine('error', err instanceof Error ? err.message : '连接失败');
    } finally {
      setLoading(false);
    }
  }, [id, pushLine]);

  React.useEffect(() => {
    void connect();
  }, [connect]);

  const executeCommand = React.useCallback(async (raw: string) => {
    const cmd = raw.trim();
    if (!cmd || status !== 'connected' || !id || executing) return;

    if (cmd === 'clear') {
      setLines([]);
      setCommand('');
      return;
    }
    if (cmd === 'exit') {
      setStatus('disconnected');
      pushLine('system', '会话已断开。');
      setCommand('');
      return;
    }

    historyRef.current = [...historyRef.current, cmd];
    historyIndexRef.current = historyRef.current.length;

    pushLine('input', cmd);
    setCommand('');
    setExecuting(true);
    const start = Date.now();
    try {
      const resp = await Api.hosts.sshExec(id, cmd);
      const latency = Date.now() - start;
      setLastLatencyMs(latency);
      if (resp.data.stdout) {
        pushLine('output', resp.data.stdout);
      }
      if (resp.data.stderr) {
        pushLine(resp.data.exit_code === 0 ? 'system' : 'error', resp.data.stderr);
      }
      if (!resp.data.stdout && !resp.data.stderr) {
        pushLine('system', '(无输出)');
      }
    } catch (err) {
      pushLine('error', err instanceof Error ? err.message : '命令执行失败');
    } finally {
      setExecuting(false);
    }
  }, [executing, id, pushLine, status]);

  const onCommandKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      void executeCommand(command);
      return;
    }
    if (e.key === 'ArrowUp') {
      e.preventDefault();
      if (historyRef.current.length === 0) return;
      historyIndexRef.current = Math.max(0, historyIndexRef.current - 1);
      setCommand(historyRef.current[historyIndexRef.current] || '');
      return;
    }
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      if (historyRef.current.length === 0) return;
      historyIndexRef.current = Math.min(historyRef.current.length, historyIndexRef.current + 1);
      if (historyIndexRef.current >= historyRef.current.length) {
        setCommand('');
      } else {
        setCommand(historyRef.current[historyIndexRef.current] || '');
      }
    }
  };

  const disconnect = () => {
    Modal.confirm({
      title: '断开终端会话',
      content: '确认断开后将停止当前终端命令执行。',
      onOk: () => {
        setStatus('disconnected');
        pushLine('system', '会话已手动断开。');
      },
    });
  };

  return (
    <div className="fade-in" style={{ '--term-bg': '#070b11', '--term-panel': '#101826' } as React.CSSProperties}>
      <Breadcrumb className="mb-4">
        <Breadcrumb.Item><Link to="/hosts">主机管理</Link></Breadcrumb.Item>
        <Breadcrumb.Item><Link to={`/hosts/detail/${id}`}>{host?.name || `Host #${id}`}</Link></Breadcrumb.Item>
        <Breadcrumb.Item>真实终端</Breadcrumb.Item>
      </Breadcrumb>

      <Card
        style={{
          background: 'linear-gradient(135deg, #0f1729 0%, #131f35 100%)',
          border: '1px solid #24324a',
          borderRadius: 12,
          marginBottom: 12,
        }}
      >
        <div className="flex items-center justify-between">
          <Space size={16}>
            <Button icon={<ArrowLeftOutlined />} onClick={() => navigate(`/hosts/detail/${id}`)}>返回</Button>
            <div>
              <Typography.Title level={5} style={{ margin: 0, color: '#e6edf3' }}>{host?.name || `Host #${id}`}</Typography.Title>
              <Typography.Text type="secondary">{host?.ip || '-'}</Typography.Text>
            </div>
            {statusTag}
          </Space>
          <Space>
            <Button icon={<ReloadOutlined />} loading={loading} onClick={() => void connect()}>重连</Button>
            <Button danger icon={<DisconnectOutlined />} disabled={status !== 'connected'} onClick={disconnect}>断开</Button>
          </Space>
        </div>
      </Card>

      <Row gutter={12} style={{ marginBottom: 12 }}>
        <Col span={6}><Card size="small" style={{ borderRadius: 10 }}><Statistic title="会话时长" value={formatDuration(sessionSeconds)} /></Card></Col>
        <Col span={6}><Card size="small" style={{ borderRadius: 10 }}><Statistic title="最近延迟" value={lastLatencyMs || 0} suffix="ms" /></Card></Col>
        <Col span={6}><Card size="small" style={{ borderRadius: 10 }}><Statistic title="命令数量" value={historyRef.current.length} /></Card></Col>
        <Col span={6}><Card size="small" style={{ borderRadius: 10 }}><Statistic title="连接状态" value={status === 'connected' ? 'ONLINE' : 'OFFLINE'} valueStyle={{ color: status === 'connected' ? '#22c55e' : '#ef4444' }} /></Card></Col>
      </Row>

      <Row gutter={12}>
        <Col span={18}>
          <Card
            bodyStyle={{ padding: 0 }}
            style={{ borderRadius: 12, border: '1px solid #223149', background: '#0b1220' }}
            title={
              <div className="flex items-center justify-between">
                <span style={{ color: '#cfd6df' }}>root@{host?.name || id}</span>
                <Tag color="blue">{host?.ip || '-'}:{host?.port || 22}</Tag>
              </div>
            }
          >
            <div ref={terminalRef} style={{ height: 520, overflowY: 'auto', padding: 14, background: '#070b11', fontFamily: 'JetBrains Mono, Menlo, Monaco, Consolas, monospace', fontSize: 13 }}>
              {lines.map((line) => (
                <div key={line.id} style={{ color: lineColors[line.type], marginBottom: 6, whiteSpace: 'pre-wrap', lineHeight: 1.55 }}>
                  {line.type === 'input' ? <span style={{ color: '#22c55e', marginRight: 6 }}>$</span> : null}
                  {line.content}
                </div>
              ))}
              {lines.length === 0 ? <Typography.Text type="secondary">终端输出区</Typography.Text> : null}
            </div>
            <div style={{ borderTop: '1px solid #1f2a3c', padding: 12, background: '#0a1220' }}>
              <Input
                value={command}
                onChange={(e) => setCommand(e.target.value)}
                onKeyDown={onCommandKeyDown}
                disabled={status !== 'connected' || executing}
                placeholder={status === 'connected' ? '输入命令，Enter 执行；支持 ↑/↓ 历史命令' : '当前未连接'}
                addonBefore={<span style={{ color: '#22c55e' }}>$</span>}
                addonAfter={
                  <Button type="primary" size="small" loading={executing} onClick={() => void executeCommand(command)} icon={<ThunderboltOutlined />}>
                    执行
                  </Button>
                }
              />
            </div>
          </Card>
        </Col>

        <Col span={6}>
          <Card title="快捷命令" size="small" style={{ borderRadius: 12, marginBottom: 12 }}>
            <Space direction="vertical" style={{ width: '100%' }}>
              {quickCommands.map((cmd) => (
                <Button key={cmd} block size="small" onClick={() => void executeCommand(cmd)} disabled={status !== 'connected' || executing}>
                  {cmd}
                </Button>
              ))}
            </Space>
          </Card>

          <Card title="会话提示" size="small" style={{ borderRadius: 12 }}>
            <Alert type="info" showIcon message="这是实时 SSH 执行，不再是模拟终端。" style={{ marginBottom: 8 }} />
            <Typography.Paragraph type="secondary" style={{ marginBottom: 0, fontSize: 12 }}>
              1. `clear` 仅清空前端输出。<br />
              2. `exit` 仅断开当前 UI 会话。<br />
              3. 生产主机建议使用最小权限账号。
            </Typography.Paragraph>
          </Card>
        </Col>
      </Row>

      {status === 'error' ? (
        <Alert
          type="error"
          showIcon
          message="终端连接失败"
          description="请检查主机 SSH 凭据、网络连通性、端口与密钥配置。"
          style={{ marginTop: 12 }}
        />
      ) : null}

      {status === 'disconnected' ? (
        <Alert
          type="warning"
          showIcon
          message="终端已断开"
          description="你可以点击上方“重连”恢复会话。"
          style={{ marginTop: 12 }}
        />
      ) : null}
    </div>
  );
};

export default HostTerminalPage;

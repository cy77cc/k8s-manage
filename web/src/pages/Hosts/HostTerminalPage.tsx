import React, { useState, useEffect, useRef } from 'react';
import { 
  Card, Button, Space, Tag, Statistic, Row, Col, 
  Tooltip, Modal, message, Breadcrumb
} from 'antd';
import { 
  ArrowLeftOutlined, DesktopOutlined, ReloadOutlined,
  ExpandOutlined, CompressOutlined, DisconnectOutlined,
  WifiOutlined, CheckCircleOutlined, CloseCircleOutlined,
  SyncOutlined, SoundOutlined
} from '@ant-design/icons';
import { useNavigate, useParams, Link } from 'react-router-dom';

interface TerminalLine {
  id: number;
  type: 'input' | 'output' | 'error' | 'system';
  content: string;
  timestamp: Date;
}

type ConnectionStatus = 'connecting' | 'connected' | 'disconnected' | 'error';

const TerminalDisplay: React.FC<{ lines: TerminalLine[]; terminalRef: React.RefObject<HTMLDivElement | null> }> = ({ 
  lines, 
  terminalRef 
}) => {
  useEffect(() => {
    if (terminalRef.current) {
      terminalRef.current.scrollTop = terminalRef.current.scrollHeight;
    }
  }, [lines, terminalRef]);

  const getLineColor = (type: TerminalLine['type']) => {
    switch (type) {
      case 'error': return '#ff4757';
      case 'system': return '#ffc107';
      case 'input': return '#00d9a5';
      default: return '#e0e0e0';
    }
  };

  return (
    <div 
      ref={terminalRef}
      className="h-full overflow-auto font-mono text-sm p-4"
      style={{ background: '#0d1117', minHeight: '400px' }}
    >
      {lines.map((line) => (
        <div key={line.id} className="whitespace-pre-wrap mb-1">
          {line.type === 'input' && <span style={{ color: '#00d9a5' }}>$ </span>}
          <span style={{ color: getLineColor(line.type) }}>
            {line.content}
          </span>
        </div>
      ))}
    </div>
  );
};

const HostTerminalPage: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const [status, setStatus] = useState<ConnectionStatus>('connecting');
  const [lines, setLines] = useState<TerminalLine[]>([]);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [soundEnabled, setSoundEnabled] = useState(true);
  const [sessionDuration, setSessionDuration] = useState(0);
  const [inputCommand, setInputCommand] = useState('');
  const terminalRef = useRef<HTMLDivElement>(null);
  const lineIdRef = useRef(0);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const hostInfo = {
    id: id || '1',
    hostname: 'prod-web-01',
    privateIp: '192.168.1.10',
    status: 'online',
  };

  useEffect(() => {
    const welcomeLines: TerminalLine[] = [
      { id: lineIdRef.current++, type: 'system', content: '========================================', timestamp: new Date() },
      { id: lineIdRef.current++, type: 'system', content: '  OpsPilot 远程终端 v1.0', timestamp: new Date() },
      { id: lineIdRef.current++, type: 'system', content: '========================================', timestamp: new Date() },
      { id: lineIdRef.current++, type: 'system', content: `Connecting to ${hostInfo.hostname} (${hostInfo.privateIp})...`, timestamp: new Date() },
      { id: lineIdRef.current++, type: 'system', content: 'Connection established.', timestamp: new Date() },
      { id: lineIdRef.current++, type: 'system', content: '', timestamp: new Date() },
      { id: lineIdRef.current++, type: 'output', content: `Welcome to Ubuntu 22.04.3 LTS (GNU/Linux 5.15.0-91-generic x86_64)`, timestamp: new Date() },
      { id: lineIdRef.current++, type: 'output', content: ` * Documentation:  https://help.ubuntu.com`, timestamp: new Date() },
      { id: lineIdRef.current++, type: 'output', content: ` * Management:     https://landscape.canonical.com`, timestamp: new Date() },
      { id: lineIdRef.current++, type: 'output', content: ` * Support:        https://ubuntu.com/advantage`, timestamp: new Date() },
      { id: lineIdRef.current++, type: 'output', content: '', timestamp: new Date() },
      { id: lineIdRef.current++, type: 'output', content: `Last login: Mon Feb 23 10:25:42 2024 from 192.168.1.100`, timestamp: new Date() },
      { id: lineIdRef.current++, type: 'system', content: '', timestamp: new Date() },
    ];
    setLines(welcomeLines);
    
    const timer = setTimeout(() => {
      setStatus('connected');
    }, 1500);

    return () => clearTimeout(timer);
  }, [id]);

  useEffect(() => {
    if (status === 'connected') {
      intervalRef.current = setInterval(() => {
        setSessionDuration(prev => prev + 1);
      }, 1000);
    }
    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
      }
    };
  }, [status]);

  useEffect(() => {
    if (status !== 'connected') return;

    const messages = [
      { type: 'output' as const, content: '[system] CPU: 45%, Memory: 62%, Disk: 38%' },
      { type: 'output' as const, content: '[nginx] 192.168.1.50 - - [23/Feb/2024:10:30:15 +0800] "GET /api/health HTTP/1.1" 200 32' },
      { type: 'output' as const, content: '[sshd] Accepted publickey for root from 192.168.1.100 port 52314 ssh2' },
    ];

    const interval = setInterval(() => {
      if (Math.random() > 0.7) {
        const msg = messages[Math.floor(Math.random() * messages.length)];
        setLines(prev => [...prev, { 
          id: lineIdRef.current++, 
          type: msg.type, 
          content: msg.content, 
          timestamp: new Date() 
        }]);
      }
    }, 5000);

    return () => clearInterval(interval);
  }, [status]);

  const handleCommand = (cmd: string) => {
    if (!cmd.trim()) return;

    setLines(prev => [...prev, { 
      id: lineIdRef.current++, 
      type: 'input', 
      content: cmd, 
      timestamp: new Date() 
    }]);

    setTimeout(() => {
      let response = '';
      const command = cmd.trim().toLowerCase();
      
      if (command === 'ls') {
        response = 'bin  boot  dev  etc  home  lib  media  mnt  opt  proc  root  run  sbin  srv  sys  tmp  usr  var';
      } else if (command === 'pwd') {
        response = '/root';
      } else if (command === 'whoami') {
        response = 'root';
      } else if (command === 'hostname') {
        response = hostInfo.hostname;
      } else if (command === 'uptime') {
        response = ' 10:30:15 up 45 days,  3:22,  2 users,  load average: 0.45, 0.32, 0.28';
      } else if (command === 'top -bn1 | head -5') {
        response = 'top - 10:30:15 up 45 days,  3:22,  2 users\nTasks: 125 total,   1 running\n%Cpu(s):  4.5 us,  2.1 sy, 93.4 id\nMiB Mem :  32768.0 total,   8124.5 free';
      } else if (command === 'df -h') {
        response = 'Filesystem      Size  Used Avail Use% Mounted on\n/dev/sda1       500G  180G  320G  37% /\ntmpfs            16G     0   16G   0% /dev/shm';
      } else if (command === 'clear') {
        setLines([]);
        return;
      } else if (command === 'help') {
        response = 'Available commands: ls, pwd, whoami, hostname, uptime, top, df, clear, exit';
      } else if (command === 'exit') {
        handleDisconnect();
        return;
      } else {
        response = `bash: ${cmd}: command not found`;
      }

      setLines(prev => [...prev, { 
        id: lineIdRef.current++, 
        type: response.includes('not found') ? 'error' : 'output', 
        content: response, 
        timestamp: new Date() 
      }]);
    }, 100);

    setInputCommand('');
  };

  const handleKeyPress = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      handleCommand(inputCommand);
    }
  };

  const handleDisconnect = () => {
    Modal.confirm({
      title: '断开连接',
      content: '确定要断开当前终端连接吗？',
      okText: '确认断开',
      onOk: () => {
        setStatus('disconnected');
        setLines(prev => [...prev, { 
          id: lineIdRef.current++, 
          type: 'system', 
          content: 'Connection closed.', 
          timestamp: new Date() 
        }]);
        message.info('终端连接已断开');
      },
    });
  };

  const handleReconnect = () => {
    setStatus('connecting');
    setLines([]);
    setSessionDuration(0);
    
    setTimeout(() => {
      setStatus('connected');
      setLines(prev => [...prev, { 
        id: lineIdRef.current++, 
        type: 'system', 
        content: 'Reconnected to ' + hostInfo.hostname, 
        timestamp: new Date() 
      }]);
      message.success('重新连接成功');
    }, 1500);
  };

  const formatDuration = (seconds: number) => {
    const h = Math.floor(seconds / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    const s = seconds % 60;
    return `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`;
  };

  const getStatusConfig = () => {
    switch (status) {
      case 'connecting': return { color: 'processing', text: '连接中', icon: <SyncOutlined spin /> };
      case 'connected': return { color: 'success', text: '已连接', icon: <CheckCircleOutlined /> };
      case 'disconnected': return { color: 'default', text: '已断开', icon: <CloseCircleOutlined /> };
      case 'error': return { color: 'error', text: '连接错误', icon: <CloseCircleOutlined /> };
    }
  };

  const statusConfig = getStatusConfig();

  return (
    <div className="fade-in">
      <Breadcrumb className="mb-4">
        <Breadcrumb.Item><Link to="/hosts">主机管理</Link></Breadcrumb.Item>
        <Breadcrumb.Item><Link to={`/hosts/detail/${id}`}>{hostInfo.hostname}</Link></Breadcrumb.Item>
        <Breadcrumb.Item>终端</Breadcrumb.Item>
      </Breadcrumb>

      <Card style={{ background: '#16213e', border: '1px solid #2d3748' }} className="mb-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Button icon={<ArrowLeftOutlined />} onClick={() => navigate(`/hosts/detail/${id}`)}>
              返回
            </Button>
            <div className="flex items-center gap-2">
              <DesktopOutlined className="text-xl" />
              <div>
                <div className="font-medium">{hostInfo.hostname}</div>
                <div className="text-gray-400 text-sm">{hostInfo.privateIp}</div>
              </div>
            </div>
            <Tag color={statusConfig.color} icon={statusConfig.icon}>
              {statusConfig.text}
            </Tag>
          </div>
          <Space>
            <Tooltip title="会话时长">
              <span className="font-mono text-gray-400">
                {formatDuration(sessionDuration)}
              </span>
            </Tooltip>
            <Tooltip title={soundEnabled ? '关闭声音' : '开启声音'}>
              <Button 
                type="text" 
                icon={<SoundOutlined />} 
                onClick={() => setSoundEnabled(!soundEnabled)}
                style={{ color: soundEnabled ? '#00d9a5' : '#666' }}
              />
            </Tooltip>
            <Tooltip title={isFullscreen ? '退出全屏' : '全屏'}>
              <Button 
                type="text" 
                icon={isFullscreen ? <CompressOutlined /> : <ExpandOutlined />}
                onClick={() => setIsFullscreen(!isFullscreen)}
              />
            </Tooltip>
            <Button icon={<ReloadOutlined />} onClick={handleReconnect} disabled={status === 'connecting'}>
              重连
            </Button>
            <Button danger icon={<DisconnectOutlined />} onClick={handleDisconnect} disabled={status !== 'connected'}>
              断开
            </Button>
          </Space>
        </div>
      </Card>

      <Row gutter={16} className="mb-4">
        <Col span={6}>
          <Card style={{ background: '#16213e', border: '1px solid #2d3748' }}>
            <Statistic
              title={<span className="text-gray-400">会话时长</span>}
              value={formatDuration(sessionDuration)}
              valueStyle={{ fontFamily: 'monospace', fontSize: 18 }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card style={{ background: '#16213e', border: '1px solid #2d3748' }}>
            <Statistic
              title={<span className="text-gray-400">延迟</span>}
              value={status === 'connected' ? 12 : '-'}
              suffix="ms"
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card style={{ background: '#16213e', border: '1px solid #2d3748' }}>
            <Statistic
              title={<span className="text-gray-400">传输 (TX/RX)</span>}
              value="1.2M / 3.4M"
              suffix="B"
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card style={{ background: '#16213e', border: '1px solid #2d3748' }}>
            <div className="flex items-center gap-2">
              <WifiOutlined className={status === 'connected' ? 'text-green-500' : 'text-gray-500'} />
              <div>
                <div className="text-gray-400 text-xs">连接状态</div>
                <div className={status === 'connected' ? 'text-green-500' : 'text-gray-500'}>
                  {statusConfig.text}
                </div>
              </div>
            </div>
          </Card>
        </Col>
      </Row>

      <Card bodyStyle={{ padding: 0 }} style={{ background: '#0d1117', border: '1px solid #2d3748' }} className={isFullscreen ? 'fixed inset-0 z-50' : ''}>
        <div className="flex items-center justify-between px-4 py-2" style={{ background: '#161b22', borderBottom: '1px solid #30363d' }}>
          <div className="flex items-center gap-2">
            <div className="flex gap-1.5">
              <div className="w-3 h-3 rounded-full bg-red-500"></div>
              <div className="w-3 h-3 rounded-full bg-yellow-500"></div>
              <div className="w-3 h-3 rounded-full bg-green-500"></div>
            </div>
            <span className="ml-2 text-gray-400 text-sm">root@{hostInfo.hostname}</span>
          </div>
          <Tag>{hostInfo.privateIp}:22</Tag>
        </div>

        <div className="relative" style={{ height: isFullscreen ? 'calc(100vh - 200px)' : '500px' }}>
          <TerminalDisplay lines={lines} terminalRef={terminalRef} />
          
          {status === 'connecting' && (
            <div className="absolute inset-0 flex items-center justify-center bg-black/70">
              <div className="text-center">
                <SyncOutlined spin className="text-4xl text-blue-500 mb-4" />
                <div className="text-gray-400">正在连接...</div>
              </div>
            </div>
          )}

          {status === 'disconnected' && (
            <div className="absolute inset-0 flex items-center justify-center bg-black/70">
              <div className="text-center">
                <CloseCircleOutlined className="text-4xl text-gray-500 mb-4" />
                <div className="text-gray-400 mb-4">连接已断开</div>
                <Button type="primary" onClick={handleReconnect}>重新连接</Button>
              </div>
            </div>
          )}
        </div>

        <div className="flex items-center gap-2 px-4 py-3" style={{ background: '#161b22', borderTop: '1px solid #30363d' }}>
          <span style={{ color: '#00d9a5' }}>$</span>
          <input
            type="text"
            value={inputCommand}
            onChange={(e) => setInputCommand(e.target.value)}
            onKeyPress={handleKeyPress}
            disabled={status !== 'connected'}
            placeholder={status === 'connected' ? '输入命令后按 Enter 执行...' : '终端未连接'}
            className="flex-1 bg-transparent border-none outline-none text-white font-mono"
            style={{ caretColor: '#00d9a5' }}
            autoFocus
          />
        </div>
      </Card>

      <Card style={{ background: '#16213e', border: '1px solid #2d3748' }} className="mt-4" title="快捷命令">
        <Space wrap>
          <Button size="small" onClick={() => handleCommand('ls -la')}>ls -la</Button>
          <Button size="small" onClick={() => handleCommand('pwd')}>pwd</Button>
          <Button size="small" onClick={() => handleCommand('whoami')}>whoami</Button>
          <Button size="small" onClick={() => handleCommand('hostname')}>hostname</Button>
          <Button size="small" onClick={() => handleCommand('uptime')}>uptime</Button>
          <Button size="small" onClick={() => handleCommand('top -bn1 | head -5')}>top</Button>
          <Button size="small" onClick={() => handleCommand('df -h')}>df -h</Button>
          <Button size="small" onClick={() => handleCommand('free -m')}>free -m</Button>
        </Space>
      </Card>
    </div>
  );
};

export default HostTerminalPage;

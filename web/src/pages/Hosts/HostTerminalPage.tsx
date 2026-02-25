import React from 'react';
import { Alert, Breadcrumb, Button, Card, Col, Input, Modal, Row, Space, Spin, Tag, Typography, Upload, message } from 'antd';
import { ArrowLeftOutlined, DeleteOutlined, DownloadOutlined, EditOutlined, FileAddOutlined, FolderAddOutlined, ReloadOutlined, SaveOutlined, UploadOutlined } from '@ant-design/icons';
import { Link, useNavigate, useParams } from 'react-router-dom';
import Editor from '@monaco-editor/react';
import { Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import 'xterm/css/xterm.css';
import { Api } from '../../api';
import type { Host, HostFileItem } from '../../api/modules/hosts';

const { Text } = Typography;

type ConnStatus = 'idle' | 'connecting' | 'connected' | 'closed' | 'error';

const HostTerminalPage: React.FC = () => {
  const navigate = useNavigate();
  const { id = '' } = useParams<{ id: string }>();
  const xtermRef = React.useRef<Terminal | null>(null);
  const fitRef = React.useRef<FitAddon | null>(null);
  const resizeObserverRef = React.useRef<ResizeObserver | null>(null);
  const inputListenerRef = React.useRef<{ dispose: () => void } | null>(null);
  const wsRef = React.useRef<WebSocket | null>(null);
  const termWrapRef = React.useRef<HTMLDivElement>(null);
  const [status, setStatus] = React.useState<ConnStatus>('idle');
  const [host, setHost] = React.useState<Host | null>(null);
  const [sessionID, setSessionID] = React.useState('');
  const [cwd, setCwd] = React.useState('.');
  const [files, setFiles] = React.useState<HostFileItem[]>([]);
  const [selectedFile, setSelectedFile] = React.useState('');
  const [selectedContent, setSelectedContent] = React.useState('');
  const [filesLoading, setFilesLoading] = React.useState(false);
  const [editing, setEditing] = React.useState(false);
  const [saving, setSaving] = React.useState(false);
  const [newDirOpen, setNewDirOpen] = React.useState(false);
  const [newDirName, setNewDirName] = React.useState('');
  const [editorSize, setEditorSize] = React.useState<'sm' | 'md' | 'lg'>('md');
  const [pathInput, setPathInput] = React.useState('.');

  const pageHeight = 'calc(100vh - 112px)';
  const fileGridColumns = 'minmax(0, 1fr) 108px 88px 112px 88px';
  const rightPanelSplitMap: Record<'sm' | 'md' | 'lg', string> = {
    sm: 'minmax(0, 64fr) minmax(0, 36fr)',
    md: 'minmax(0, 58fr) minmax(0, 42fr)',
    lg: 'minmax(0, 52fr) minmax(0, 48fr)',
  };

  const setupTerminal = React.useCallback(() => {
    if (!termWrapRef.current || xtermRef.current) return;
    const term = new Terminal({
      cursorBlink: true,
      convertEol: true,
      fontFamily: 'JetBrains Mono, Menlo, Monaco, Consolas, monospace',
      fontSize: 13,
      theme: {
        background: '#0e1117',
        foreground: '#d4d4d4',
        cursor: '#8ae234',
      },
    });
    const fitAddon = new FitAddon();
    term.loadAddon(fitAddon);
    term.open(termWrapRef.current);
    fitAddon.fit();
    term.writeln('\x1b[90mConnecting to host terminal...\x1b[0m');
    xtermRef.current = term;
    fitRef.current = fitAddon;
  }, []);

  React.useEffect(() => {
    setupTerminal();
    const onResize = () => fitRef.current?.fit();
    window.addEventListener('resize', onResize);
    return () => {
      window.removeEventListener('resize', onResize);
      wsRef.current?.close();
      resizeObserverRef.current?.disconnect();
      resizeObserverRef.current = null;
      inputListenerRef.current?.dispose();
      inputListenerRef.current = null;
      xtermRef.current?.dispose();
      xtermRef.current = null;
    };
  }, [setupTerminal]);

  const wsURLFromPath = (wsPath: string) => {
    const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
    const token = localStorage.getItem('token');
    const suffix = token ? `${wsPath.includes('?') ? '&' : '?'}token=${encodeURIComponent(token)}` : '';
    return `${protocol}://${window.location.host}${wsPath}${suffix}`;
  };

  const refreshFiles = React.useCallback(async (dirPath: string) => {
    if (!id) return;
    setFilesLoading(true);
    try {
      const res = await Api.hosts.listFiles(id, dirPath);
      setFiles(res.data.list || []);
      setCwd(res.data.path || dirPath);
      setPathInput(res.data.path || dirPath);
    } catch (err) {
      message.error(err instanceof Error ? err.message : '加载文件列表失败');
    } finally {
      setFilesLoading(false);
    }
  }, [id]);

  const connect = React.useCallback(async () => {
    if (!id) return;
    setStatus('connecting');
    try {
      const [hostResp, sessResp] = await Promise.all([
        Api.hosts.getHostDetail(id),
        Api.hosts.createTerminalSession(id),
      ]);
      setHost(hostResp.data);
      setSessionID(sessResp.data.session_id);

      const ws = new WebSocket(wsURLFromPath(sessResp.data.ws_path));
      wsRef.current = ws;
      ws.onopen = () => {
        setStatus('connected');
        fitRef.current?.fit();
        const term = xtermRef.current;
        if (!term) return;
        term.focus();
        term.writeln(`\x1b[32mConnected to ${hostResp.data.name} (${hostResp.data.ip})\x1b[0m`);
        inputListenerRef.current?.dispose();
        inputListenerRef.current = term.onData((data) => {
          ws.send(JSON.stringify({ type: 'input', input: data }));
        });
        const fit = fitRef.current;
        const size = term.cols && term.rows ? { cols: term.cols, rows: term.rows } : { cols: 120, rows: 40 };
        ws.send(JSON.stringify({ type: 'resize', ...size }));
        if (fit) {
          resizeObserverRef.current?.disconnect();
          const observer = new ResizeObserver(() => {
            fit.fit();
            ws.send(JSON.stringify({ type: 'resize', cols: term.cols, rows: term.rows }));
          });
          resizeObserverRef.current = observer;
          if (termWrapRef.current) observer.observe(termWrapRef.current);
        }
      };
      ws.onmessage = (event) => {
        const term = xtermRef.current;
        if (!term) return;
        try {
          const msg = JSON.parse(String(event.data));
          if (msg.type === 'output' && msg.payload?.data) {
            term.write(String(msg.payload.data));
          }
        } catch {
          term.write(String(event.data));
        }
      };
      ws.onerror = () => {
        setStatus('error');
        xtermRef.current?.writeln('\r\n\x1b[31mTerminal websocket error\x1b[0m');
      };
      ws.onclose = () => {
        setStatus('closed');
        resizeObserverRef.current?.disconnect();
        resizeObserverRef.current = null;
        inputListenerRef.current?.dispose();
        inputListenerRef.current = null;
        xtermRef.current?.writeln('\r\n\x1b[90mSession closed\x1b[0m');
      };
      await refreshFiles('.');
    } catch (err) {
      setStatus('error');
      message.error(err instanceof Error ? err.message : '终端连接失败');
    }
  }, [id, refreshFiles]);

  React.useEffect(() => {
    void connect();
  }, [connect]);

  React.useEffect(() => {
    const raf = window.requestAnimationFrame(() => {
      fitRef.current?.fit();
    });
    return () => window.cancelAnimationFrame(raf);
  }, [selectedFile, editorSize]);

  const closeSession = React.useCallback(async () => {
    wsRef.current?.close();
    if (id && sessionID) {
      try {
        await Api.hosts.closeTerminalSession(id, sessionID);
      } catch {
        // noop
      }
    }
    setStatus('closed');
  }, [id, sessionID]);

  const openFile = async (item: HostFileItem) => {
    if (!id) return;
    if (item.is_dir) {
      await refreshFiles(item.path);
      return;
    }
    try {
      const res = await Api.hosts.readFile(id, item.path);
      setSelectedFile(item.path);
      setSelectedContent(res.data.content || '');
      setEditing(false);
    } catch (err) {
      message.error(err instanceof Error ? err.message : '读取文件失败');
    }
  };

  const saveFile = async () => {
    if (!id || !selectedFile) return;
    setSaving(true);
    try {
      await Api.hosts.writeFile(id, selectedFile, selectedContent);
      setEditing(false);
      message.success('文件已保存');
      await refreshFiles(cwd);
    } catch (err) {
      message.error(err instanceof Error ? err.message : '保存失败');
    } finally {
      setSaving(false);
    }
  };

  const removePath = (item: HostFileItem) => {
    if (!id) return;
    Modal.confirm({
      title: `删除 ${item.name}`,
      content: '此操作不可恢复，确认删除吗？',
      okButtonProps: { danger: true },
      onOk: async () => {
        await Api.hosts.deletePath(id, item.path);
        if (item.path === selectedFile) {
          setSelectedFile('');
          setSelectedContent('');
        }
        await refreshFiles(cwd);
      },
    });
  };

  const renamePath = (item: HostFileItem) => {
    if (!id) return;
    let nextName = item.name;
    Modal.confirm({
      title: '重命名',
      content: <Input defaultValue={item.name} onChange={(e) => { nextName = e.target.value; }} />,
      onOk: async () => {
        const parent = item.path.includes('/') ? item.path.slice(0, item.path.lastIndexOf('/')) : '.';
        await Api.hosts.renamePath(id, item.path, `${parent}/${nextName}`);
        await refreshFiles(cwd);
      },
    });
  };

  const downloadFile = async (item: HostFileItem) => {
    if (!id || item.is_dir) return;
    const blob = await Api.hosts.downloadFile(id, item.path);
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = item.name;
    a.click();
    URL.revokeObjectURL(url);
  };

  const toParentPath = React.useCallback((path: string) => {
    if (path === '.') return '.';
    if (!path.includes('/')) return '.';
    const parent = path.slice(0, path.lastIndexOf('/'));
    return parent || '.';
  }, []);

  const formatLsTime = React.useCallback((value?: string) => {
    if (!value) return '-';
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return '-';
    const mm = String(date.getMonth() + 1).padStart(2, '0');
    const dd = String(date.getDate()).padStart(2, '0');
    const hh = String(date.getHours()).padStart(2, '0');
    const min = String(date.getMinutes()).padStart(2, '0');
    return `${mm}-${dd} ${hh}:${min}`;
  }, []);

  return (
    <div className="fade-in host-terminal-page" style={{ height: pageHeight, overflow: 'hidden', display: 'flex', flexDirection: 'column' }}>
      <Breadcrumb className="mb-4">
        <Breadcrumb.Item><Link to="/hosts">主机管理</Link></Breadcrumb.Item>
        <Breadcrumb.Item><Link to={`/hosts/detail/${id}`}>{host?.name || `Host #${id}`}</Link></Breadcrumb.Item>
        <Breadcrumb.Item>终端与文件</Breadcrumb.Item>
      </Breadcrumb>

      <Card
        style={{ marginBottom: 8, borderRadius: 10, flex: 1, minHeight: 0, overflow: 'hidden' }}
        styles={{ body: { minHeight: 0, height: '100%' } }}
        title={
          <Space>
            <Button icon={<ArrowLeftOutlined />} onClick={() => navigate(`/hosts/detail/${id}`)}>返回</Button>
            <Text strong>{host?.name || `Host #${id}`}</Text>
            <Text type="secondary">{host?.ip || '-'}</Text>
            <Tag color={status === 'connected' ? 'success' : status === 'connecting' ? 'processing' : status === 'error' ? 'error' : 'default'}>
              {status.toUpperCase()}
            </Tag>
          </Space>
        }
        extra={
          <Space>
            <Button icon={<ReloadOutlined />} onClick={() => void connect()}>重连</Button>
            <Button danger onClick={() => void closeSession()}>关闭会话</Button>
          </Space>
        }
      >
        <Row gutter={12} style={{ height: '100%', minHeight: 0 }} align="stretch">
          <Col xs={24} xl={16} style={{ display: 'flex', minHeight: 0, minWidth: 0 }}>
            <Card
              size="small"
              styles={{ body: { padding: 0, background: '#0e1117', height: '100%', minHeight: 0 } }}
              style={{ borderRadius: 10, border: '1px solid #1f2937', width: '100%', height: '100%' }}
            >
              <div className="host-terminal-xterm" ref={termWrapRef} style={{ height: '100%', width: '100%', minHeight: 360 }} />
            </Card>
          </Col>
          <Col
            xs={24}
            xl={8}
              style={{
                display: 'grid',
                gridTemplateRows: rightPanelSplitMap[editorSize],
                gap: 8,
                minHeight: 0,
                minWidth: 0,
                overflow: 'hidden',
                height: '100%',
              }}
          >
            <Card
              size="small"
              title="文件管理"
              extra={
                <Space size={4}>
                  <Button size="small" icon={<ReloadOutlined />} onClick={() => void refreshFiles(cwd)} />
                  <Button size="small" icon={<FolderAddOutlined />} onClick={() => setNewDirOpen(true)} />
                  <Upload
                    showUploadList={false}
                    customRequest={async (opt) => {
                      const file = opt.file as File;
                      await Api.hosts.uploadFile(id, cwd, file);
                      opt.onSuccess?.({}, new XMLHttpRequest());
                      await refreshFiles(cwd);
                    }}
                  >
                    <Button size="small" icon={<UploadOutlined />} />
                  </Upload>
                </Space>
              }
              style={{ borderRadius: 10, minHeight: 0, height: '100%' }}
              styles={{ body: { display: 'flex', flexDirection: 'column', gap: 8, minHeight: 0, height: '100%', overflow: 'hidden' } }}
            >
              <Space style={{ width: '100%', justifyContent: 'space-between' }}>
                <Text type="secondary">目录: {cwd}</Text>
                <Space.Compact style={{ width: 220 }}>
                  <Input
                    size="small"
                    placeholder="输入目录并跳转"
                    value={pathInput}
                    onChange={(e) => setPathInput(e.target.value)}
                    onPressEnter={() => void refreshFiles((pathInput || '.').trim() || '.')}
                  />
                  <Button size="small" onClick={() => void refreshFiles((pathInput || '.').trim() || '.')}>跳转</Button>
                </Space.Compact>
              </Space>
              {filesLoading ? <Spin /> : null}
              <div
                style={{
                  display: 'grid',
                  gridTemplateColumns: fileGridColumns,
                  alignItems: 'center',
                  columnGap: 12,
                  fontFamily: 'JetBrains Mono, Menlo, Monaco, Consolas, monospace',
                  fontSize: 12,
                  color: '#8c8c8c',
                  padding: '2px 8px',
                }}
              >
                <span>名称</span>
                <span>修改时间</span>
                <span style={{ textAlign: 'right' }}>大小</span>
                <span>权限</span>
                <span />
              </div>
              <div style={{ width: '100%', overflowY: 'auto', overflowX: 'hidden', flex: 1, minHeight: 0 }}>
                {cwd !== '.' ? (
                  <div
                    style={{ display: 'grid', gridTemplateColumns: fileGridColumns, alignItems: 'center', columnGap: 12, borderRadius: 8, padding: '2px 8px' }}
                  >
                    <div
                      onClick={() => void refreshFiles(toParentPath(cwd))}
                      style={{ cursor: 'pointer', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}
                    >
                      <span title="..">📁 ..</span>
                    </div>
                    <span>-</span>
                    <span style={{ textAlign: 'right' }}>-</span>
                    <span>drwxr-xr-x</span>
                    <span />
                  </div>
                ) : null}
                {files.map((item) => (
                  <div
                    key={item.path}
                    style={{
                      display: 'grid',
                      gridTemplateColumns: fileGridColumns,
                      alignItems: 'center',
                      columnGap: 12,
                      borderRadius: 8,
                      padding: '2px 8px',
                      background: selectedFile === item.path ? '#e6f4ff' : 'transparent',
                    }}
                  >
                    <div
                      onClick={() => void openFile(item)}
                      style={{ cursor: 'pointer', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}
                    >
                      <span title={item.name}>
                        {item.is_dir ? '📁' : '📄'} {item.name}
                      </span>
                    </div>
                    <span>{formatLsTime(item.updated_at)}</span>
                    <span style={{ textAlign: 'right' }}>{item.is_dir ? '-' : String(item.size ?? 0)}</span>
                    <span>{item.mode || '-'}</span>
                    <Space size={0} style={{ justifyContent: 'flex-end' }}>
                      {!item.is_dir ? <Button type="text" size="small" icon={<DownloadOutlined />} onClick={() => void downloadFile(item)} /> : null}
                      <Button type="text" size="small" icon={<EditOutlined />} onClick={() => renamePath(item)} />
                      <Button type="text" size="small" danger icon={<DeleteOutlined />} onClick={() => removePath(item)} />
                    </Space>
                  </div>
                ))}
              </div>
            </Card>

            <Card
              size="small"
              title={selectedFile ? <Text ellipsis={{ tooltip: selectedFile }} style={{ maxWidth: '100%', display: 'block' }}>{`编辑: ${selectedFile}`}</Text> : '文件预览'}
              extra={selectedFile ? (
                <Space size={4}>
                  <Button size="small" type={editorSize === 'sm' ? 'primary' : 'default'} onClick={() => setEditorSize('sm')}>缩小</Button>
                  <Button size="small" type={editorSize === 'md' ? 'primary' : 'default'} onClick={() => setEditorSize('md')}>默认</Button>
                  <Button size="small" type={editorSize === 'lg' ? 'primary' : 'default'} onClick={() => setEditorSize('lg')}>放大</Button>
                  <Button size="small" icon={<SaveOutlined />} loading={saving} onClick={() => void saveFile()}>保存</Button>
                </Space>
              ) : null}
              style={{ borderRadius: 10, minHeight: 0, height: '100%' }}
              styles={{ body: { overflow: 'hidden', minHeight: 0, height: '100%', display: 'flex', flexDirection: 'column' } }}
            >
              {selectedFile ? (
                <>
                  <div style={{ flex: 1, minHeight: 0 }}>
                    <Editor
                      height="100%"
                      defaultLanguage="yaml"
                      value={selectedContent}
                      onChange={(v) => { setSelectedContent(v || ''); setEditing(true); }}
                      theme="vs-dark"
                      options={{ minimap: { enabled: false }, fontSize: 13 }}
                    />
                  </div>
                  {editing ? <Alert style={{ marginTop: 8 }} type="warning" showIcon message="内容已修改，记得保存。" /> : null}
                </>
              ) : <Text type="secondary">选择文件后在这里查看与编辑内容。</Text>}
            </Card>
          </Col>
        </Row>
      </Card>

      <div style={{ overflow: 'hidden', flexShrink: 0 }}>
        <Alert
          type="info"
          showIcon
          message="终端和文件管理都通过主机 SSH 实时执行；删除/覆盖操作请谨慎。"
        />
      </div>

      <Modal
        open={newDirOpen}
        title="新建目录"
        onOk={async () => {
          if (!newDirName.trim()) return;
          await Api.hosts.mkdir(id, `${cwd}/${newDirName.trim()}`.replace('//', '/'));
          setNewDirOpen(false);
          setNewDirName('');
          await refreshFiles(cwd);
        }}
        onCancel={() => setNewDirOpen(false)}
      >
        <Input
          prefix={<FileAddOutlined />}
          placeholder="目录名"
          value={newDirName}
          onChange={(e) => setNewDirName(e.target.value)}
        />
      </Modal>
    </div>
  );
};

export default HostTerminalPage;

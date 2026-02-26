import React, { useEffect, useState } from 'react';
import { Alert, Button, Card, Input, List, Modal, Space, Tag, Typography, message } from 'antd';
import { Api } from '../../api';
import type { AICommandHistoryItem, AICommandResult } from '../../api';

const { Paragraph, Text } = Typography;

interface CommandPanelProps {
  scene: string;
}

const CommandPanel: React.FC<CommandPanelProps> = ({ scene }) => {
  const [command, setCommand] = useState('');
  const [suggestions, setSuggestions] = useState<Array<{ command: string; hint?: string }>>([]);
  const [preview, setPreview] = useState<AICommandResult | null>(null);
  const [history, setHistory] = useState<AICommandHistoryItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [executing, setExecuting] = useState(false);
  const [detailOpen, setDetailOpen] = useState(false);
  const [detail, setDetail] = useState<{ record: AICommandHistoryItem; audit_events: any[] } | null>(null);
  const [lastPreviewParams, setLastPreviewParams] = useState<Record<string, any> | null>(null);
  const [confirmOpen, setConfirmOpen] = useState(false);

  const loadSuggestions = async () => {
    const res = await Api.ai.getCommandSuggestions();
    setSuggestions(res.data || []);
  };

  const loadHistory = async () => {
    const res = await Api.ai.getCommandHistory(20);
    setHistory(res.data?.list || []);
  };

  useEffect(() => {
    void loadSuggestions();
    void loadHistory();
  }, []);

  const pickParams = (p: AICommandResult | null): Record<string, any> => {
    if (!p) return {};
    const fromPlan = (p.plan && typeof p.plan === 'object' ? (p.plan as any).params : undefined) || {};
    const fromArtifacts = (p.artifacts && typeof p.artifacts === 'object' ? (p.artifacts as any).params : undefined) || {};
    return { ...fromPlan, ...fromArtifacts };
  };

  const buildParamDiff = (prev: Record<string, any>, next: Record<string, any>) => {
    const keys = Array.from(new Set([...Object.keys(prev), ...Object.keys(next)])).sort();
    return keys
      .map((key) => {
        const hasPrev = Object.prototype.hasOwnProperty.call(prev, key);
        const hasNext = Object.prototype.hasOwnProperty.call(next, key);
        if (hasPrev && !hasNext) {
          return { key, type: 'removed' as const, before: prev[key], after: undefined };
        }
        if (!hasPrev && hasNext) {
          return { key, type: 'added' as const, before: undefined, after: next[key] };
        }
        const beforeRaw = JSON.stringify(prev[key]);
        const afterRaw = JSON.stringify(next[key]);
        if (beforeRaw !== afterRaw) {
          return { key, type: 'changed' as const, before: prev[key], after: next[key] };
        }
        return null;
      })
      .filter(Boolean) as Array<{ key: string; type: 'added' | 'removed' | 'changed'; before: any; after: any }>;
  };

  const handlePreview = async () => {
    if (!command.trim()) return;
    setLoading(true);
    try {
      const prev = pickParams(preview);
      const res = await Api.ai.previewCommand({ command: command.trim(), scene });
      setLastPreviewParams(prev);
      setPreview(res.data);
    } finally {
      setLoading(false);
    }
  };

  const handleExecute = async () => {
    if (!preview) return;
    setConfirmOpen(true);
  };

  const doExecute = async () => {
    if (!preview) return;
    const cmdID = String(preview.artifacts?.command_id || '');
    if (!cmdID) {
      message.error('缺少 command_id，请重新预览');
      return;
    }
    setExecuting(true);
    try {
      const res = await Api.ai.executeCommand({
        command_id: cmdID,
        confirm: true,
        approval_token: String(preview.artifacts?.approval_token || ''),
      });
      setPreview(res.data);
      await loadHistory();
      message.success('命令执行完成');
      setConfirmOpen(false);
    } catch (err: any) {
      message.error(err?.response?.data?.error?.message || '命令执行失败');
    } finally {
      setExecuting(false);
    }
  };

  const openHistoryDetail = async (id: string) => {
    const res = await Api.ai.getCommandHistoryDetail(id);
    setDetail(res.data);
    setDetailOpen(true);
  };

  const currentParams = pickParams(preview);
  const paramDiff = buildParamDiff(lastPreviewParams || {}, currentParams);

  return (
    <Space className="ai-command-panel-root" direction="vertical" style={{ width: '100%' }} size={12}>
      <Card size="small" title="命令输入">
        <Space.Compact style={{ width: '100%' }}>
          <Input
            value={command}
            onChange={(e) => setCommand(e.target.value)}
            placeholder="例如: ops.aggregate.status limit=5"
            onPressEnter={() => void handlePreview()}
          />
          <Button type="primary" loading={loading} onClick={() => void handlePreview()}>预览</Button>
        </Space.Compact>
        <List
          style={{ marginTop: 12 }}
          size="small"
          dataSource={suggestions}
          renderItem={(item) => (
            <List.Item
              actions={[<Button size="small" key="use" onClick={() => setCommand(item.command)}>使用</Button>]}
            >
              <Space direction="vertical" size={0}>
                <Text code>{item.command}</Text>
                <Text type="secondary">{item.hint || ''}</Text>
              </Space>
            </List.Item>
          )}
        />
      </Card>

      <Card size="small" title="执行计划预览">
        {!preview ? <Text type="secondary">先预览命令以查看计划。</Text> : (
          <Space direction="vertical" style={{ width: '100%' }}>
            <Space>
              <Tag color={preview.status === 'blocked' ? 'orange' : preview.status === 'failed' ? 'red' : 'blue'}>{preview.status}</Tag>
              <Tag color={preview.risk === 'high' ? 'red' : preview.risk === 'low' ? 'gold' : 'blue'}>{preview.risk}</Tag>
              <Text copyable>{preview.trace_id}</Text>
            </Space>
            <Paragraph>{preview.summary}</Paragraph>
            {preview.risk === 'high' ? (
              <Alert
                type="error"
                message="高风险命令"
                description="请重点核对目标资源、关键参数和回滚路径，再执行确认。"
              />
            ) : null}
            {preview.missing && preview.missing.length > 0 ? (
              <Alert type="warning" message={`缺失参数: ${preview.missing.join(', ')}`} />
            ) : null}
            {paramDiff.length > 0 ? (
              <Card size="small" title="参数 Diff（相对上一次预览）">
                <Space direction="vertical" style={{ width: '100%' }}>
                  {paramDiff.map((row) => (
                    <div key={row.key} style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}>
                      <Tag color={row.type === 'added' ? 'green' : row.type === 'removed' ? 'default' : 'orange'}>{row.type}</Tag>
                      <Text code>{row.key}</Text>
                      <Text type="secondary">before:</Text>
                      <Text code>{row.before === undefined ? '-' : JSON.stringify(row.before)}</Text>
                      <Text type="secondary">after:</Text>
                      <Text code>{row.after === undefined ? '-' : JSON.stringify(row.after)}</Text>
                    </div>
                  ))}
                </Space>
              </Card>
            ) : null}
            <pre style={{ margin: 0, whiteSpace: 'pre-wrap' }}>{JSON.stringify(preview.plan || preview.artifacts, null, 2)}</pre>
            <Space>
              <Button type="primary" danger={preview.risk === 'high'} onClick={() => void handleExecute()} loading={executing} disabled={preview.status === 'blocked'}>确认执行</Button>
              <Button onClick={() => { setPreview(null); setLastPreviewParams(null); }}>清空</Button>
            </Space>
          </Space>
        )}
      </Card>

      <Card size="small" title="命令历史与回放">
        <List
          size="small"
          dataSource={history}
          renderItem={(item) => (
            <List.Item actions={[<Button size="small" key="detail" onClick={() => void openHistoryDetail(item.id)}>回放</Button>]}> 
              <Space direction="vertical" size={0}>
                <Space>
                  <Tag>{item.status}</Tag>
                  <Tag>{item.risk}</Tag>
                  <Text type="secondary">{item.intent}</Text>
                </Space>
                <Text>{item.command}</Text>
                <Text type="secondary">{new Date(item.created_at).toLocaleString()}</Text>
              </Space>
            </List.Item>
          )}
        />
      </Card>

      <Modal
        title="执行前确认"
        open={confirmOpen}
        onCancel={() => setConfirmOpen(false)}
        onOk={() => void doExecute()}
        confirmLoading={executing}
        okText={preview?.risk === 'high' ? '高风险，确认执行' : '确认执行'}
        okButtonProps={{ danger: preview?.risk === 'high' }}
      >
        {!preview ? null : (
          <Space direction="vertical" style={{ width: '100%' }}>
            <Space>
              <Tag color={preview.risk === 'high' ? 'red' : preview.risk === 'low' ? 'gold' : 'blue'}>{preview.risk}</Tag>
              <Tag>{preview.status}</Tag>
            </Space>
            <Text>trace_id: <Text code copyable>{preview.trace_id}</Text></Text>
            <Text>command_id: <Text code>{String(preview.artifacts?.command_id || '-')}</Text></Text>
            <Text>
              approval_token:
              <Text code>{String(preview.artifacts?.approval_token || '-')}</Text>
            </Text>
            {preview.risk === 'high' ? (
              <Alert type="error" message="高风险操作将直接影响线上状态，请确认审批与回滚方案。" />
            ) : null}
            {paramDiff.length > 0 ? (
              <Card size="small" title="参数 Diff 摘要">
                <Space direction="vertical" style={{ width: '100%' }}>
                  {paramDiff.map((row) => (
                    <div key={`confirm-${row.key}`} style={{ display: 'flex', gap: 6, flexWrap: 'wrap', alignItems: 'center' }}>
                      <Tag color={row.type === 'added' ? 'green' : row.type === 'removed' ? 'default' : 'orange'}>{row.type}</Tag>
                      <Text code>{row.key}</Text>
                      <Text type="secondary">{row.before === undefined ? '-' : JSON.stringify(row.before)}</Text>
                      <Text type="secondary">→</Text>
                      <Text>{row.after === undefined ? '-' : JSON.stringify(row.after)}</Text>
                    </div>
                  ))}
                </Space>
              </Card>
            ) : (
              <Text type="secondary">无参数变化（相对上一次预览）。</Text>
            )}
          </Space>
        )}
      </Modal>

      <Modal
        title="命令执行回放"
        open={detailOpen}
        onCancel={() => setDetailOpen(false)}
        footer={null}
        width={760}
      >
        <pre style={{ whiteSpace: 'pre-wrap', margin: 0 }}>{JSON.stringify(detail, null, 2)}</pre>
      </Modal>
    </Space>
  );
};

export default CommandPanel;

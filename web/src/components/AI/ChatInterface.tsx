import React, { useState, useRef, useEffect } from 'react';
import { Input, Button, List, Avatar, Space, Spin, Typography, Card, Divider, Tag, Collapse, Alert } from 'antd';
import { SendOutlined, MessageOutlined, ToolOutlined, BulbOutlined, WarningOutlined } from '@ant-design/icons';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { oneLight } from 'react-syntax-highlighter/dist/esm/styles/prism';
import { Api } from '../../api';
import type { AIMessage, AISession, ToolTrace } from '../../api';

const { Text, Paragraph } = Typography;

interface ChatInterfaceProps {
  sessionId?: string;
  scene?: string;
  onSessionCreate?: (session: AISession) => void;
  onSessionUpdate?: (session: AISession) => void;
  className?: string;
}

type StreamState = 'idle' | 'running' | 'timeout' | 'done' | 'error';
type LocalMessage = AIMessage & { turnId?: string };

const renderMarkdown = (content: string) => (
  <div className="ai-markdown-content">
    <ReactMarkdown
      remarkPlugins={[remarkGfm]}
      components={{
        table({ children }) {
          return (
            <div className="ai-markdown-table-wrap">
              <table className="ai-markdown-table">{children}</table>
            </div>
          );
        },
        thead({ children }) {
          return <thead className="ai-markdown-thead">{children}</thead>;
        },
        tbody({ children }) {
          return <tbody className="ai-markdown-tbody">{children}</tbody>;
        },
        tr({ children }) {
          return <tr className="ai-markdown-tr">{children}</tr>;
        },
        th({ children }) {
          return <th className="ai-markdown-th">{children}</th>;
        },
        td({ children }) {
          return <td className="ai-markdown-td">{children}</td>;
        },
        code({ className, children, ...props }) {
          const match = /language-(\w+)/.exec(className || '');
          const codeText = String(children).replace(/\n$/, '');
          if (match) {
            return (
              <SyntaxHighlighter
                style={oneLight}
                language={match[1]}
                PreTag="div"
                customStyle={{ margin: '8px 0', borderRadius: 6 }}
              >
                {codeText}
              </SyntaxHighlighter>
            );
          }
          return (
            <code className={className} style={{ background: '#f5f5f5', padding: '2px 4px', borderRadius: 4 }} {...props}>
              {children}
            </code>
          );
        },
      }}
    >
      {content}
    </ReactMarkdown>
  </div>
);

const ChatInterface: React.FC<ChatInterfaceProps> = ({
  sessionId,
  scene = 'global',
  onSessionCreate,
  onSessionUpdate,
  className,
}) => {
  const [messages, setMessages] = useState<LocalMessage[]>([]);
  const [inputValue, setInputValue] = useState('');
  const [loading, setLoading] = useState(false);
  const [streamState, setStreamState] = useState<StreamState>('idle');
  const [streamError, setStreamError] = useState('');
  const [streamNotice, setStreamNotice] = useState('');
  const [currentSession, setCurrentSession] = useState<AISession | null>(null);
  const [lastPrompt, setLastPrompt] = useState('');
  const [traceRawVisible, setTraceRawVisible] = useState<Record<string, boolean>>({});
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const scrollContainerRef = useRef<HTMLDivElement>(null);
  const shouldAutoScrollRef = useRef(true);

  const loadSession = async () => {
    try {
      if (sessionId) {
        const response = await Api.ai.getSessionDetail(sessionId);
        setMessages(response.data.messages as LocalMessage[]);
        setCurrentSession(response.data);
        return;
      }
      const res = await Api.ai.getCurrentSession(scene);
      if (res.data) {
        setMessages((res.data.messages || []) as LocalMessage[]);
        setCurrentSession(res.data);
      } else {
        setMessages([]);
        setCurrentSession(null);
      }
    } catch (error) {
      console.error('加载会话失败:', error);
    }
  };

  useEffect(() => {
    loadSession();
  }, [sessionId, scene]);

  useEffect(() => {
    if (shouldAutoScrollRef.current) {
      messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    }
  }, [messages, loading, streamState]);

  const handleScroll = () => {
    const el = scrollContainerRef.current;
    if (!el) return;
    const nearBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 80;
    shouldAutoScrollRef.current = nearBottom;
  };

  const attachTraceToAssistant = (assistantID: string, turnID: string | undefined, trace: ToolTrace) => {
    setMessages((prev) => {
      const next = [...prev];
      let idx = next.findIndex((m) => m.role === 'assistant' && turnID && m.turnId === turnID);
      if (idx < 0) {
        idx = next.findIndex((m) => m.id === assistantID);
      }
      if (idx < 0) return prev;
      const target = next[idx];
      next[idx] = {
        ...target,
        ...(turnID ? { turnId: turnID } : {}),
        traces: [...(target.traces || []), trace],
      };
      return next;
    });
  };

  const sendMessage = async (messageText: string) => {
    if (!messageText.trim() || loading) return;

    const previousSessionID = currentSession?.id;
    const userMessage: LocalMessage = {
      id: `msg-${Date.now()}`,
      role: 'user',
      content: messageText,
      timestamp: new Date().toISOString(),
    };
    const assistantMessageID = `msg-${Date.now()}-assistant`;
    const assistantPlaceholder: LocalMessage = {
      id: assistantMessageID,
      role: 'assistant',
      content: '',
      thinking: '',
      traces: [],
      timestamp: new Date().toISOString(),
    };

    setMessages((prev) => [...prev, userMessage, assistantPlaceholder]);
    setInputValue('');
    setLoading(true);
    setStreamState('running');
    setStreamError('');
    setStreamNotice('');
    setLastPrompt(messageText);

    let latestSession: AISession | undefined;
    let activeTurnID = '';

    try {
      await Api.ai.chatStream(
        {
          sessionId: currentSession?.id,
          message: messageText,
          context: { scene },
        },
        {
          onMeta: (meta) => {
            activeTurnID = meta.turn_id || activeTurnID;
            setCurrentSession((prev) => {
              if (prev?.id === meta.sessionId) {
                return prev;
              }
              return {
                id: meta.sessionId,
                title: 'AI Session',
                messages: [],
                createdAt: meta.createdAt,
                updatedAt: meta.createdAt,
              };
            });
            if (activeTurnID) {
              setMessages((prev) =>
                prev.map((item) => (item.id === assistantMessageID ? { ...item, turnId: activeTurnID } : item)),
              );
            }
          },
          onDelta: (delta) => {
            const turnID = delta.turn_id || activeTurnID;
            setMessages((prev) =>
              prev.map((item) =>
                (item.id === assistantMessageID || (turnID && item.turnId === turnID))
                  ? { ...item, turnId: turnID || item.turnId, content: `${item.content}${delta.contentChunk}` }
                  : item,
              ),
            );
          },
          onThinkingDelta: (delta) => {
            const turnID = delta.turn_id || activeTurnID;
            setMessages((prev) =>
              prev.map((item) =>
                (item.id === assistantMessageID || (turnID && item.turnId === turnID))
                  ? { ...item, turnId: turnID || item.turnId, thinking: `${item.thinking || ''}${delta.contentChunk}` }
                  : item,
              ),
            );
          },
          onToolCall: (payload) => {
            const turnID = payload.turn_id || activeTurnID;
            attachTraceToAssistant(assistantMessageID, turnID, {
              id: `trace-${Date.now()}-${Math.random().toString(36).slice(2)}`,
              type: 'tool_call',
              payload: payload as Record<string, any>,
              timestamp: new Date().toISOString(),
            });
          },
          onToolResult: (payload) => {
            const turnID = payload.turn_id || activeTurnID;
            attachTraceToAssistant(assistantMessageID, turnID, {
              id: `trace-${Date.now()}-${Math.random().toString(36).slice(2)}`,
              type: 'tool_result',
              payload: payload as Record<string, any>,
              timestamp: new Date().toISOString(),
            });
          },
          onApprovalRequired: (payload) => {
            const turnID = payload.turn_id || activeTurnID;
            attachTraceToAssistant(assistantMessageID, turnID, {
              id: `trace-${Date.now()}-${Math.random().toString(36).slice(2)}`,
              type: 'approval_required',
              payload: payload as Record<string, any>,
              timestamp: new Date().toISOString(),
            });
          },
          onDone: (done) => {
            latestSession = done.session;
            setCurrentSession(done.session);
            if (done.stream_state === 'partial') {
              setStreamState('timeout');
              setStreamError('工具结果不完整，可重试本轮对话。');
              const turnID = done.turn_id || activeTurnID;
              const missing = done.tool_summary?.missing || [];
              const missingCallIDs = done.tool_summary?.missing_call_ids || [];
              if (missing.length > 0) {
                attachTraceToAssistant(assistantMessageID, turnID, {
                  id: `trace-${Date.now()}-${Math.random().toString(36).slice(2)}`,
                  type: 'tool_missing',
                  payload: {
                    tool: 'runtime',
                    missing,
                    missing_call_ids: missingCallIDs,
                    summary: done.tool_summary,
                  },
                  timestamp: new Date().toISOString(),
                });
              }
              return;
            }
            if (done.stream_state === 'failed') {
              setStreamState('error');
              return;
            }
            setStreamState('done');
          },
          onError: (err) => {
            if (err.code === 'tool_timeout_soft') {
              setStreamNotice(err.message || '工具执行较慢，正在继续等待结果…');
              setStreamState('running');
              return;
            }
            if (err.code === 'tool_result_missing') {
              setStreamState('timeout');
              setStreamError(err.message || '工具结果不完整，可重试本轮对话。');
              return;
            }
            setStreamState(err.code === 'tool_timeout_hard' || err.message?.includes('超时') ? 'timeout' : 'error');
            setStreamError(err.message || 'AI服务暂时不可用');
          },
        },
      );

      if (latestSession !== undefined) {
        if (latestSession.id !== previousSessionID && onSessionCreate) {
          onSessionCreate(latestSession);
        } else if (onSessionUpdate) {
          onSessionUpdate(latestSession);
        }
      }
    } catch (error) {
      console.error('发送消息失败:', error);
      setStreamState('error');
      setStreamError('AI服务暂时不可用，请稍后再试。');
      setMessages((prev) =>
        prev.map((item) =>
          item.id === assistantMessageID
            ? { ...item, content: item.content || '抱歉，AI服务暂时不可用，请稍后再试。' }
            : item,
        ),
      );
    } finally {
      setLoading(false);
    }
  };

  const handleSend = async () => {
    await sendMessage(inputValue.trim());
  };

  const handleRetry = async () => {
    if (!lastPrompt) return;
    await sendMessage(lastPrompt);
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      void handleSend();
    }
  };

  return (
    <Card
      title={
        <Space>
          <MessageOutlined />
          <Text strong>AI 助手</Text>
          {streamState === 'running' ? <Tag color="processing">Streaming</Tag> : null}
        </Space>
      }
      className={`ai-chat-interface ${className || ''}`}
      style={{ height: '100%' }}
      styles={{
        body: {
          height: '100%',
          minHeight: 0,
          display: 'flex',
          flexDirection: 'column',
          padding: 0,
        },
      }}
    >
      <div ref={scrollContainerRef} onScroll={handleScroll} className="ai-chat-message-scroll">
        <List
          dataSource={messages}
          renderItem={(message) => (
            <List.Item>
              <List.Item.Meta
                avatar={
                  <Avatar
                    icon={message.role === 'user' ? <SendOutlined /> : <MessageOutlined />}
                    style={{
                      backgroundColor: message.role === 'user' ? '#1677ff' : '#52c41a',
                    }}
                  />
                }
                title={
                  <Text>
                    {message.role === 'user' ? '我' : 'AI 助手'}
                    <Text type="secondary" style={{ marginLeft: '8px' }}>
                      {new Date(message.timestamp).toLocaleString()}
                    </Text>
                  </Text>
                }
                description={
                  message.role === 'assistant' ? (
                    <div className="ai-assistant-message-bubble">
                      {message.thinking ? (
                        <div className="ai-thinking-block">
                          <Collapse
                            size="small"
                            ghost
                            items={[{
                              key: `thinking-${message.id}`,
                              label: (
                                <Space className="ai-thinking-header">
                                  <BulbOutlined />
                                  <Text className="ai-thinking-title">思考过程</Text>
                                  <Tag color="cyan">{message.thinking.length} chars</Tag>
                                </Space>
                              ),
                              children: (
                                <div className="ai-thinking-content">
                                  {renderMarkdown(message.thinking)}
                                </div>
                              ),
                            }]}
                          />
                        </div>
                      ) : null}
                      {message.traces && message.traces.length > 0 ? (
                        <div className="ai-trace-block">
                          <Collapse
                            size="small"
                            ghost
                            items={[{
                              key: `trace-${message.id}`,
                              label: (
                                <Space className="ai-trace-header">
                                  <ToolOutlined />
                                  <Text className="ai-trace-title">工具调用轨迹</Text>
                                  <Tag color="gold">{message.traces.length}</Tag>
                                </Space>
                              ),
                              children: (
                                <Space direction="vertical" style={{ width: '100%' }} size={8}>
                                  {message.traces.map((trace) => {
                                    const isRawShown = !!traceRawVisible[trace.id];
                                    const traceStatus = trace.type === 'tool_result'
                                      ? ((trace.payload?.result?.ok || trace.payload?.payload?.result?.ok) ? 'success' : 'error')
                                      : (trace.type === 'approval_required' || trace.type === 'tool_missing' ? 'warning' : 'processing');
                                    return (
                                      <Card key={trace.id} size="small" className="ai-trace-card">
                                        <Space className="ai-trace-card-head">
                                          <Space>
                                            <Tag color={traceStatus === 'success' ? 'green' : traceStatus === 'error' ? 'red' : traceStatus === 'warning' ? 'orange' : 'blue'}>
                                              {trace.type}
                                            </Tag>
                                            <Text strong>{trace.payload?.tool || trace.payload?.payload?.tool || 'unknown-tool'}</Text>
                                          </Space>
                                          <Button size="small" type="link" onClick={() => setTraceRawVisible((prev) => ({ ...prev, [trace.id]: !isRawShown }))}>
                                            {isRawShown ? '隐藏原始JSON' : '显示原始JSON'}
                                          </Button>
                                        </Space>
                                        <Text type="secondary" className="ai-trace-time">{new Date(trace.timestamp).toLocaleTimeString()}</Text>
                                        {isRawShown ? (
                                          <pre className="ai-trace-json">
                                            {JSON.stringify(trace.payload, null, 2)}
                                          </pre>
                                        ) : null}
                                      </Card>
                                    );
                                  })}
                                </Space>
                              ),
                            }]}
                          />
                        </div>
                      ) : null}
                      {renderMarkdown(message.content)}
                    </div>
                  ) : <Paragraph className="ai-chat-user-paragraph">{message.content}</Paragraph>
                }
              />
            </List.Item>
          )}
        />
        {loading && (
          <div className="ai-chat-loading-row">
            <Spin size="small" />
            <Text className="ai-chat-loading-text">AI 正在思考...</Text>
          </div>
        )}
        {streamNotice ? (
          <div style={{ padding: '4px 8px' }}>
            <Alert type="info" showIcon message={streamNotice} />
          </div>
        ) : null}
        <div ref={messagesEndRef} />
      </div>

      <Divider style={{ margin: 0 }} />

      {(streamState === 'timeout' || streamState === 'error') ? (
        <div style={{ padding: '8px 16px 0' }}>
          <Alert
            type={streamState === 'timeout' ? 'warning' : 'error'}
            showIcon
            icon={<WarningOutlined />}
            message={streamState === 'timeout' ? '工具执行可能超时' : '流式对话发生错误'}
            description={streamError || '你可以重试本轮对话。'}
            action={<Button size="small" onClick={() => void handleRetry()}>重试本轮</Button>}
          />
        </div>
      ) : null}

      <div className="ai-chat-input-wrap">
        <Input.TextArea
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="请输入您的问题..."
          rows={3}
          disabled={loading}
        />
        <div className="ai-chat-send-row">
          <Button
            type="primary"
            icon={<SendOutlined />}
            onClick={() => void handleSend()}
            loading={loading}
            disabled={!inputValue.trim() || loading}
          >
            发送
          </Button>
        </div>
      </div>
    </Card>
  );
};

export default ChatInterface;

import React, { useState, useRef, useEffect } from 'react';
import { Input, Button, List, Avatar, Space, Spin, Typography, Card, Divider, Tag, Collapse } from 'antd';
import { SendOutlined, MessageOutlined, ToolOutlined } from '@ant-design/icons';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { oneLight } from 'react-syntax-highlighter/dist/esm/styles/prism';
import { Api } from '../../api';
import type { AIMessage, AISession } from '../../api';

const { Text, Paragraph } = Typography;

interface ChatInterfaceProps {
  sessionId?: string;
  scene?: string;
  onSessionCreate?: (session: AISession) => void;
  onSessionUpdate?: (session: AISession) => void;
  className?: string;
}

const renderMarkdown = (content: string) => (
  <div style={{ color: '#1f2937', lineHeight: 1.7, fontSize: 14 }}>
    <ReactMarkdown
      remarkPlugins={[remarkGfm]}
      components={{
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
  const [messages, setMessages] = useState<AIMessage[]>([]);
  const [inputValue, setInputValue] = useState('');
  const [loading, setLoading] = useState(false);
  const [currentSession, setCurrentSession] = useState<AISession | null>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const scrollContainerRef = useRef<HTMLDivElement>(null);
  const shouldAutoScrollRef = useRef(true);

  const loadSession = async () => {
    try {
      if (sessionId) {
        const response = await Api.ai.getSessionDetail(sessionId);
        setMessages(response.data.messages);
        setCurrentSession(response.data);
        return;
      }
      const res = await Api.ai.getCurrentSession(scene);
      if (res.data) {
        setMessages(res.data.messages || []);
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
  }, [messages, loading]);

  const handleScroll = () => {
    const el = scrollContainerRef.current;
    if (!el) return;
    const nearBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 80;
    shouldAutoScrollRef.current = nearBottom;
  };

  const appendSystemTrace = (content: string) => {
    setMessages((prev) => [
      ...prev,
      {
        id: `sys-${Date.now()}-${Math.random().toString(36).slice(2)}`,
        role: 'system',
        content,
        timestamp: new Date().toISOString(),
      },
    ]);
  };

  const handleSend = async () => {
    if (!inputValue.trim() || loading) return;

    const messageText = inputValue.trim();
    const previousSessionID = currentSession?.id;

    const userMessage: AIMessage = {
      id: `msg-${Date.now()}`,
      role: 'user',
      content: messageText,
      timestamp: new Date().toISOString(),
    };
    const assistantMessageID = `msg-${Date.now()}-assistant`;
    const assistantPlaceholder: AIMessage = {
      id: assistantMessageID,
      role: 'assistant',
      content: '',
      timestamp: new Date().toISOString(),
    };

    setMessages((prev) => [...prev, userMessage, assistantPlaceholder]);
    setInputValue('');
    setLoading(true);

    let latestSession: AISession | undefined;

    try {
      await Api.ai.chatStream(
        {
          sessionId: currentSession?.id,
          message: messageText,
          context: { scene },
        },
        {
          onMeta: (meta) => {
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
          },
          onDelta: (delta) => {
            setMessages((prev) =>
              prev.map((item) =>
                item.id === assistantMessageID
                  ? { ...item, content: `${item.content}${delta.contentChunk}` }
                  : item,
              ),
            );
          },
          onThinkingDelta: (delta) => {
            setMessages((prev) =>
              prev.map((item) =>
                item.id === assistantMessageID
                  ? { ...item, thinking: `${item.thinking || ''}${delta.contentChunk}` }
                  : item,
              ),
            );
          },
          onToolCall: (payload) => {
            const toolName = payload.tool || payload.tool_calls?.[0]?.function?.name || 'unknown';
            appendSystemTrace(`**Tool Call**\n\n\`${toolName}\`\n\n\`\`\`json\n${JSON.stringify(payload, null, 2)}\n\`\`\``);
          },
          onToolResult: (payload) => {
            appendSystemTrace(`**Tool Result**\n\n\`\`\`json\n${JSON.stringify(payload, null, 2)}\n\`\`\``);
          },
          onApprovalRequired: (payload) => {
            const token = (payload as { approval_token?: string; id?: string }).approval_token || payload.id || '';
            appendSystemTrace(`**Approval Required**\n\nTool: \`${payload.tool}\`\n\nToken: \`${token}\``);
          },
          onDone: (done) => {
            latestSession = done.session;
            setCurrentSession(done.session);
          },
          onError: (err) => {
            appendSystemTrace(`**Stream Error**\n\n${err.message || 'AI服务暂时不可用'}`);
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

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  return (
    <Card
      title={
        <Space>
          <MessageOutlined />
          <Text strong>AI 助手</Text>
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
      <div
        ref={scrollContainerRef}
        onScroll={handleScroll}
        style={{
          flex: 1,
          minHeight: 0,
          overflowY: 'auto',
          padding: '12px 16px',
        }}
      >
        <List
          dataSource={messages}
          renderItem={(message) => (
            <List.Item>
              <List.Item.Meta
                avatar={
                  <Avatar
                    icon={message.role === 'system' ? <ToolOutlined /> : message.role === 'user' ? <SendOutlined /> : <MessageOutlined />}
                    style={{
                      backgroundColor:
                        message.role === 'system' ? '#faad14' : message.role === 'user' ? '#1890ff' : '#52c41a',
                    }}
                  />
                }
                title={
                  <Text>
                    {message.role === 'system' ? '系统轨迹' : message.role === 'user' ? '我' : 'AI 助手'}
                    {message.role === 'system' && <Tag color="gold" style={{ marginLeft: 8 }}>Tool</Tag>}
                    <Text type="secondary" style={{ marginLeft: '8px' }}>
                      {new Date(message.timestamp).toLocaleString()}
                    </Text>
                  </Text>
                }
                description={
                  message.role === 'assistant' || message.role === 'system' ? (
                    <div style={{ margin: 0, background: '#fff', border: '1px solid #f0f0f0', borderRadius: 8, padding: '8px 12px' }}>
                      {message.role === 'assistant' && message.thinking ? (
                        <div style={{ marginBottom: 8 }}>
                          <Collapse
                            size="small"
                            items={[{
                              key: `thinking-${message.id}`,
                              label: '展开/收起思考过程',
                              children: <div style={{ margin: 0, background: '#f8fdff', border: '1px solid #e6f7ff', borderRadius: 8, padding: '8px 12px' }}>{renderMarkdown(message.thinking)}</div>,
                            }]}
                          />
                        </div>
                      ) : null}
                      {renderMarkdown(message.content)}
                    </div>
                  ) : <Paragraph style={{ margin: 0, whiteSpace: 'pre-wrap' }}>{message.content}</Paragraph>
                }
              />
            </List.Item>
          )}
        />
        {loading && (
          <div style={{ display: 'flex', alignItems: 'center', padding: '16px' }}>
            <Spin size="small" />
            <Text style={{ marginLeft: '8px' }}>AI 正在思考...</Text>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

      <Divider style={{ margin: 0 }} />

      <div style={{ padding: '12px 16px', borderTop: '1px solid #f0f0f0' }}>
        <Input.TextArea
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="请输入您的问题..."
          rows={3}
          disabled={loading}
        />
        <Space style={{ marginTop: '8px', float: 'right' }}>
          <Button
            type="primary"
            icon={<SendOutlined />}
            onClick={handleSend}
            loading={loading}
            disabled={!inputValue.trim() || loading}
          >
            发送
          </Button>
        </Space>
      </div>
    </Card>
  );
};

export default ChatInterface;

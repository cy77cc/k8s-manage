# Design: AI 助手完全重构

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              新 AI 助手架构                                      │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  ┌───────────────────────────────────────────────────────────────────────────┐ │
│  │                        前端层 (@ant-design/x)                              │ │
│  │                                                                           │ │
│  │   ┌─────────────────────────────────────────────────────────────────────┐ │ │
│  │   │                      ChatPage (页面容器)                             │ │ │
│  │   │   ├─ ConversationSidebar    会话列表侧边栏                          │ │ │
│  │   │   │   ├─ ConversationList                                            │ │ │
│  │   │   │   └─ NewChatButton                                               │ │ │
│  │   │   │                                                                  │ │ │
│  │   │   └─ ChatMain             主聊天区域                                 │ │ │
│  │   │       ├─ MessageList                                                  │ │ │
│  │   │       │   ├─ UserBubble                                              │ │ │
│  │   │       │   ├─ AssistantBubble                                         │ │ │
│  │   │       │   │   ├─ MarkdownContent                                     │ │ │
│  │   │       │   │   ├─ ToolExecutionCard                                   │ │ │
│  │   │       │   │   └─ ConfirmationPanel                                   │ │ │
│  │   │       │   └─ SystemMessage                                           │ │ │
│  │   │       │                                                              │ │ │
│  │   │       └─ InputArea                                                    │ │ │
│  │   │           ├─ Sender (@ant-design/x)                                  │ │ │
│  │   │           └─ PromptSuggestions                                       │ │ │
│  │   └─────────────────────────────────────────────────────────────────────┘ │ │
│  │                                                                           │ │
│  │   Hooks:                                                                  │ │
│  │   ├─ useChatSession()     会话状态管理                                    │ │
│  │   ├─ useSSEConnection()   SSE 连接管理                                    │ │
│  │   └─ useConfirmation()    确认交互状态                                    │ │
│  │                                                                           │ │
│  └───────────────────────────────────────────────────────────────────────────┘ │
│                                      │                                         │
│                                      │ HTTP/SSE                                │
│                                      ▼                                         │
│  ┌───────────────────────────────────────────────────────────────────────────┐ │
│  │                        API 层 (6 个核心端点)                               │ │
│  │                                                                           │ │
│  │   POST /api/ai/chat           对话入口 (SSE 流式)                         │ │
│  │   POST /api/ai/chat/respond   响应确认请求                                │ │
│  │   GET  /api/ai/sessions       会话列表                                    │ │
│  │   GET  /api/ai/sessions/:id   会话详情                                    │ │
│  │   DELETE /api/ai/sessions/:id 删除会话                                    │ │
│  │   GET  /api/ai/tools          工具能力列表                                │ │
│  │                                                                           │ │
│  └───────────────────────────────────────────────────────────────────────────┘ │
│                                      │                                         │
│                                      ▼                                         │
│  ┌───────────────────────────────────────────────────────────────────────────┐ │
│  │                        Agent 层 (eino)                                    │ │
│  │                                                                           │ │
│  │   ┌─────────────────────────────────────────────────────────────────────┐ │ │
│  │   │                     HybridAgent                                      │ │ │
│  │   │                                                                     │ │ │
│  │   │   用户输入                                                          │ │ │
│  │   │       │                                                             │ │ │
│  │   │       ▼                                                             │ │ │
│  │   │   ┌─────────────────┐                                               │ │ │
│  │   │   │ IntentClassifier │ ──┬── simple ──▶ SimpleChatMode              │ │ │
│  │   │   │   (轻量 LLM)     │   │                                         │ │ │
│  │   │   └─────────────────┘   └── agentic ─▶ AgenticMode                  │ │ │
│  │   │                                                             │       │ │ │
│  │   │                                                             ▼       │ │ │
│  │   │                                            ┌───────────────────┐   │ │ │
│  │   │                                            │  ToolExecutor     │   │ │ │
│  │   │                                            │  ├─ execute()     │   │ │ │
│  │   │                                            │  └─ askUser()     │   │ │ │
│  │   │                                            └─────────┬─────────┘   │ │ │
│  │   │                                                      │             │ │ │
│  │   │                              高风险操作 ──────────────┘             │ │ │
│  │   │                                      │                             │ │ │
│  │   │                                      ▼                             │ │ │
│  │   │                            ┌───────────────────┐                   │ │ │
│  │   │                            │  中断 + 发送确认   │                   │ │ │
│  │   │                            │  (SSE: ask_user)  │                   │ │ │
│  │   │                            └───────────────────┘                   │ │ │
│  │   └─────────────────────────────────────────────────────────────────────┘ │ │
│  │                                                                           │ │
│  │   ┌─────────────────────────────────────────────────────────────────────┐ │ │
│  │   │                     ToolRegistry                                     │ │ │
│  │   │                                                                     │ │ │
│  │   │   K8s:              Host:              Service:         Monitor:    │ │ │
│  │   │   ├─ k8s_query      ├─ host_exec       ├─ deploy         ├─ alert  │ │ │
│  │   │   ├─ k8s_logs       └─ host_batch      └─ status         └─ metric │ │ │
│  │   │   └─ k8s_events                                                       │ │ │
│  │   │                                                                     │ │ │
│  │   │   External:                                                          │ │ │
│  │   │   └─ mcp_proxy (MCP 工具代理)                                        │ │ │
│  │   └─────────────────────────────────────────────────────────────────────┘ │ │
│  │                                                                           │ │
│  └───────────────────────────────────────────────────────────────────────────┘ │
│                                      │                                         │
│                                      ▼                                         │
│  ┌───────────────────────────────────────────────────────────────────────────┐ │
│  │                        存储层                                              │ │
│  │                                                                           │ │
│  │   ┌─────────────────────┐       ┌─────────────────────┐                  │ │
│  │   │   PostgreSQL/MySQL  │       │       Redis         │                  │ │
│  │   │                     │       │                     │                  │ │
│  │   │   ai_sessions       │       │   checkpoints:*     │                  │ │
│  │   │   ├─ id             │       │   pending_asks:*    │                  │ │
│  │   │   ├─ user_id        │       │   session_cache:*   │                  │ │
│  │   │   ├─ title          │       │                     │                  │ │
│  │   │   └─ created_at     │       │   TTL: 30 min       │                  │ │
│  │   │                     │       │                     │                  │ │
│  │   │   ai_messages       │       └─────────────────────┘                  │ │
│  │   │   ├─ id             │                                                │ │
│  │   │   ├─ session_id     │                                                │ │
│  │   │   ├─ role           │                                                │ │
│  │   │   ├─ content        │                                                │ │
│  │   │   ├─ metadata       │                                                │ │
│  │   │   └─ created_at     │                                                │ │
│  │   └─────────────────────┘                                                │ │
│  │                                                                           │ │
│  └───────────────────────────────────────────────────────────────────────────┘ │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

## Component Design

### 1. 前端组件架构

#### 1.1 ChatPage (页面容器)

```tsx
// web/src/pages/AIChat/ChatPage.tsx

import { useState } from 'react';
import { Layout } from 'antd';
import { ConversationSidebar } from './components/ConversationSidebar';
import { ChatMain } from './components/ChatMain';
import { useChatSession } from './hooks/useChatSession';

export function ChatPage() {
  const {
    sessions,
    currentSession,
    createSession,
    switchSession,
    deleteSession
  } = useChatSession();

  return (
    <Layout className="h-screen">
      <Layout.Sider width={280} theme="light">
        <ConversationSidebar
          sessions={sessions}
          currentId={currentSession?.id}
          onSelect={switchSession}
          onCreate={createSession}
          onDelete={deleteSession}
        />
      </Layout.Sider>
      <Layout.Content>
        <ChatMain session={currentSession} />
      </Layout.Content>
    </Layout>
  );
}
```

#### 1.2 ChatMain (主聊天区域)

```tsx
// web/src/pages/AIChat/components/ChatMain.tsx

import { useRef, useEffect } from 'react';
import { Sender, useSSE } from '@ant-design/x';
import { MessageList } from './MessageList';
import { useSSEConnection } from '../hooks/useSSEConnection';
import type { AISession, ChatMessage } from '../types';

interface ChatMainProps {
  session: AISession | null;
}

export function ChatMain({ session }: ChatMainProps) {
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const {
    messages,
    isLoading,
    sendMessage,
    respondToAsk
  } = useSSEConnection(session?.id);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  return (
    <div className="flex flex-col h-full">
      <div className="flex-1 overflow-auto p-4">
        <MessageList
          messages={messages}
          onRespond={respondToAsk}
        />
        <div ref={messagesEndRef} />
      </div>

      <div className="border-t p-4">
        <Sender
          placeholder="输入你的问题..."
          loading={isLoading}
          onSubmit={(text) => sendMessage(text)}
          footer={
            <div className="text-xs text-gray-400 mt-2">
              按 Enter 发送，Shift+Enter 换行
            </div>
          }
        />
      </div>
    </div>
  );
}
```

#### 1.3 MessageList (消息列表)

```tsx
// web/src/pages/AIChat/components/MessageList.tsx

import { Bubble } from '@ant-design/x';
import { MarkdownContent } from './MarkdownContent';
import { ToolExecutionCard } from './ToolExecutionCard';
import { ConfirmationPanel } from './ConfirmationPanel';
import type { ChatMessage } from '../types';

interface MessageListProps {
  messages: ChatMessage[];
  onRespond: (askId: string, response: any) => void;
}

export function MessageList({ messages, onRespond }: MessageListProps) {
  return (
    <div className="space-y-4">
      {messages.map((msg) => (
        <Bubble
          key={msg.id}
          placement={msg.role === 'user' ? 'end' : 'start'}
          avatar={{ src: msg.role === 'user' ? '/user-avatar.png' : '/ai-avatar.png' }}
          content={renderContent(msg, onRespond)}
        />
      ))}
    </div>
  );
}

function renderContent(msg: ChatMessage, onRespond: Function) {
  // 渲染确认面板
  if (msg.type === 'ask_user') {
    return (
      <ConfirmationPanel
        askId={msg.id}
        title={msg.metadata?.title}
        description={msg.metadata?.description}
        risk={msg.metadata?.risk}
        options={msg.metadata?.options}
        onRespond={onRespond}
      />
    );
  }

  // 渲染工具执行结果
  if (msg.type === 'tool_result') {
    return <ToolExecutionCard result={msg.metadata} />;
  }

  // 默认渲染 Markdown 内容
  return <MarkdownContent content={msg.content} />;
}
```

#### 1.4 ConfirmationPanel (交互式确认面板)

```tsx
// web/src/pages/AIChat/components/ConfirmationPanel.tsx

import { Card, Button, Space, Tag, Descriptions, Collapse } from 'antd';
import { WarningOutlined, CheckOutlined, CloseOutlined } from '@ant-design/icons';

interface ConfirmationPanelProps {
  askId: string;
  title: string;
  description: string;
  risk: 'low' | 'medium' | 'high';
  options?: { label: string; value: string }[];
  onRespond: (askId: string, response: any) => void;
}

export function ConfirmationPanel({
  askId,
  title,
  description,
  risk,
  options,
  onRespond
}: ConfirmationPanelProps) {
  const riskColors = {
    low: 'green',
    medium: 'orange',
    high: 'red'
  };

  return (
    <Card
      className="max-w-lg"
      title={
        <Space>
          <WarningOutlined className={risk === 'high' ? 'text-red-500' : 'text-orange-500'} />
          <span>确认操作</span>
          <Tag color={riskColors[risk]}>{risk.toUpperCase()} RISK</Tag>
        </Space>
      }
    >
      <div className="mb-4">
        <div className="font-medium text-lg mb-2">{title}</div>
        <div className="text-gray-600">{description}</div>
      </div>

      <Collapse ghost>
        <Collapse.Panel header="查看详情" key="details">
          <Descriptions column={1} size="small">
            {/* 操作详情 */}
          </Descriptions>
        </Collapse.Panel>
      </Collapse>

      <div className="flex justify-end gap-2 mt-4">
        <Button icon={<CloseOutlined />} onClick={() => onRespond(askId, { confirmed: false })}>
          取消
        </Button>
        <Button
          type="primary"
          icon={<CheckOutlined />}
          onClick={() => onRespond(askId, { confirmed: true })}
        >
          确认执行
        </Button>
      </div>
    </Card>
  );
}
```

### 2. 后端 API 设计

#### 2.1 路由定义

```go
// internal/service/ai/routes.go

package ai

import (
    "github.com/cy77cc/k8s-manage/internal/middleware"
    "github.com/cy77cc/k8s-manage/internal/svc"
    "github.com/gin-gonic/gin"
)

func RegisterAIHandlers(v1 *gin.RouterGroup, svcCtx *svc.ServiceContext) {
    h := newHandler(svcCtx)
    g := v1.Group("/ai", middleware.JWTAuth())
    {
        // 核心对话
        g.POST("/chat", h.chat)              // SSE 流式对话
        g.POST("/chat/respond", h.respond)   // 响应确认请求

        // 会话管理
        g.GET("/sessions", h.listSessions)
        g.GET("/sessions/:id", h.getSession)
        g.DELETE("/sessions/:id", h.deleteSession)

        // 工具能力
        g.GET("/tools", h.listTools)
    }
}
```

#### 2.2 Chat Handler

```go
// internal/service/ai/chat_handler.go

package ai

import (
    "net/http"
    "strings"

    "github.com/cy77cc/k8s-manage/internal/httpx"
    "github.com/cy77cc/k8s-manage/internal/xcode"
    "github.com/gin-gonic/gin"
)

type chatRequest struct {
    SessionID string         `json:"sessionId"`
    Message   string         `json:"message" binding:"required"`
    Context   map[string]any `json:"context,omitempty"`
}

// chat SSE 流式对话入口
func (h *handler) chat(c *gin.Context) {
    var req chatRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        httpx.BindErr(c, err)
        return
    }

    uid := GetUserID(c)
    msg := strings.TrimSpace(req.Message)
    if msg == "" {
        httpx.Fail(c, xcode.ParamError, "message is required")
        return
    }

    // SSE headers
    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    c.Header("Connection", "keep-alive")
    c.Header("X-Accel-Buffering", "no")
    c.Status(http.StatusOK)

    flusher, ok := c.Writer.(http.Flusher)
    if !ok {
        httpx.Fail(c, xcode.ServerError, "streaming not supported")
        return
    }

    // 获取或创建会话
    session := h.getOrCreateSession(uid, req.SessionID)

    // 调用 Agent
    h.streamAgentResponse(c.Request.Context(), session, msg, c.Writer, flusher)
}

// respond 响应确认请求
func (h *handler) respond(c *gin.Context) {
    var req struct {
        AskID     string `json:"askId" binding:"required"`
        SessionID string `json:"sessionId" binding:"required"`
        Response  any    `json:"response" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        httpx.BindErr(c, err)
        return
    }

    // 恢复 Agent 执行
    result, err := h.agent.Resume(c.Request.Context(), req.SessionID, req.AskID, req.Response)
    if err != nil {
        httpx.Fail(c, xcode.ServerError, err.Error())
        return
    }

    httpx.OK(c, result)
}
```

### 3. Agent 核心设计

#### 3.1 HybridAgent

```go
// internal/ai/agent.go

package ai

import (
    "context"
    "fmt"

    "github.com/cloudwego/eino/adk"
    "github.com/cloudwego/eino/components/model"
    "github.com/cloudwego/eino/components/tool"
)

// HybridAgent 混合模式 Agent
type HybridAgent struct {
    classifier     *IntentClassifier
    simpleChat     *SimpleChatMode
    agenticMode    *AgenticMode
    toolRegistry   *ToolRegistry
}

// AgentResult Agent 执行结果
type AgentResult struct {
    Type      string         // "text" | "tool_result" | "ask_user"
    Content   string         // 文本内容
    ToolName  string         // 工具名称（type=tool_result）
    ToolData  map[string]any // 工具结果
    Ask       *AskRequest    // 确认请求（type=ask_user）
}

// AskRequest 确认请求
type AskRequest struct {
    ID          string
    Title       string
    Description string
    Risk        string
    Details     map[string]any
}

// NewHybridAgent 创建混合 Agent
func NewHybridAgent(
    ctx context.Context,
    chatModel model.ToolCallingChatModel,
    classifierModel model.ChatModel,
    tools []tool.BaseTool,
    deps PlatformDeps,
) (*HybridAgent, error) {
    classifier := NewIntentClassifier(classifierModel)

    return &HybridAgent{
        classifier:   classifier,
        simpleChat:   NewSimpleChatMode(chatModel),
        agenticMode:  NewAgenticMode(chatModel, tools, deps),
        toolRegistry: NewToolRegistry(tools),
    }, nil
}

// Query 执行查询
func (a *HybridAgent) Query(ctx context.Context, sessionID, message string) *adk.AsyncIterator[*AgentResult] {
    iter, gen := adk.NewAsyncIteratorPair[*AgentResult]()

    go func() {
        defer gen.Close()

        // 1. 意图分类
        intent, err := a.classifier.Classify(ctx, message)
        if err != nil {
            gen.Send(&AgentResult{Type: "error", Content: err.Error()})
            return
        }

        // 2. 路由到对应模式
        switch intent {
        case IntentSimple:
            a.simpleChat.Execute(ctx, message, gen)
        case IntentAgentic:
            a.agenticMode.Execute(ctx, sessionID, message, gen)
        default:
            a.simpleChat.Execute(ctx, message, gen)
        }
    }()

    return iter
}

// Resume 恢复执行（确认后）
func (a *HybridAgent) Resume(ctx context.Context, sessionID, askID string, response any) (*AgentResult, error) {
    return a.agenticMode.Resume(ctx, sessionID, askID, response)
}
```

#### 3.2 IntentClassifier (意图分类器)

```go
// internal/ai/classifier.go

package ai

import (
    "context"
    "fmt"

    "github.com/cloudwego/eino/components/model"
)

type Intent string

const (
    IntentSimple  Intent = "simple"   // 简单问答
    IntentAgentic Intent = "agentic"  // 需要执行工具
)

// IntentClassifier 意图分类器
type IntentClassifier struct {
    model model.ChatModel
}

const classifierPrompt = `你是一个意图分类器。根据用户输入，判断是否需要调用工具来完成任务。

需要调用工具的情况：
- 查询 K8s 资源（Pod、Service、Deployment 等）
- 查看日志、事件
- 执行主机命令
- 部署服务
- 查看监控指标、告警
- 操作资源（创建、删除、更新）

不需要调用工具的情况：
- 问候、闲聊
- 知识问答（如 "什么是 Pod"）
- 总结、解释已有信息
- 简单的澄清问题

用户输入: %s

只输出一个词: "simple" 或 "agentic"`

func NewIntentClassifier(m model.ChatModel) *IntentClassifier {
    return &IntentClassifier{model: m}
}

func (c *IntentClassifier) Classify(ctx context.Context, query string) (Intent, error) {
    if c.model == nil {
        return IntentSimple, nil // 降级为简单模式
    }

    prompt := fmt.Sprintf(classifierPrompt, query)
    resp, err := c.model.Generate(ctx, []*schema.Message{
        {Role: schema.User, Content: prompt},
    })
    if err != nil {
        return IntentSimple, nil // 错误时降级
    }

    content := strings.ToLower(strings.TrimSpace(resp.Content))
    if strings.Contains(content, "agentic") {
        return IntentAgentic, nil
    }
    return IntentSimple, nil
}
```

#### 3.3 AgenticMode (Agent 执行模式)

```go
// internal/ai/agentic_mode.go

package ai

import (
    "context"
    "fmt"
    "strings"

    "github.com/cloudwego/eino/adk"
    "github.com/cloudwego/eino/components/model"
    "github.com/cloudwego/eino/components/tool"
)

// AgenticMode Agent 执行模式
type AgenticMode struct {
    agent   adk.Agent
    runner  *adk.Runner
    store   adk.CheckPointStore
}

const agenticInstruction = `你是一个专业的运维助手，可以执行以下操作：

## 核心能力

### K8s 资源管理
- k8s_query: 查询 K8s 资源 (pods/services/deployments/nodes)
- k8s_logs: 获取 Pod 日志
- k8s_events: 获取 K8s 事件

### 主机运维
- host_exec: 在主机上执行命令
- host_batch: 批量主机操作

### 服务管理
- service_deploy: 部署服务
- service_status: 查询服务状态

### 监控告警
- monitor_alert: 查询告警
- monitor_metric: 查询指标

## 执行规则

1. 直接执行工具调用，不要输出计划
2. 高风险操作会触发用户确认，等待确认后继续
3. 执行结果以清晰的方式呈现

## 注意事项

- 参数不足时询问用户
- 执行失败时给出原因和建议`

func NewAgenticMode(
    chatModel model.ToolCallingChatModel,
    tools []tool.BaseTool,
    store adk.CheckPointStore,
) *AgenticMode {
    agent := adk.NewChatModelAgent(&adk.ChatModelAgentConfig{
        Name:        "PlatformOps",
        Instruction: agenticInstruction,
        Model:       chatModel,
        ToolsConfig: adk.ToolsConfig{
            ToolsNodeConfig: compose.ToolsNodeConfig{
                Tools: tools,
            },
        },
    })

    return &AgenticMode{
        agent:  agent,
        runner: adk.NewRunner(adk.RunnerConfig{Agent: agent, CheckPointStore: store}),
        store:  store,
    }
}

// Execute 执行 Agent
func (m *AgenticMode) Execute(
    ctx context.Context,
    sessionID, message string,
    gen *adk.AsyncGenerator[*AgentResult],
) {
    iter := m.runner.Query(ctx, message, adk.WithCheckPointID(sessionID))

    for {
        event, ok := iter.Next()
        if !ok {
            break
        }

        result := m.processEvent(event)
        if result != nil {
            gen.Send(result)
        }
    }
}

// Resume 恢复执行
func (m *AgenticMode) Resume(
    ctx context.Context,
    sessionID, askID string,
    response any,
) (*AgentResult, error) {
    iter, err := m.runner.Resume(ctx, sessionID, adk.WithResumeTarget(askID, response))
    if err != nil {
        return nil, err
    }

    for {
        event, ok := iter.Next()
        if !ok {
            break
        }
        if event.Err != nil {
            return nil, event.Err
        }
        if event.Output != nil {
            return &AgentResult{
                Type:    "text",
                Content: event.Output.MessageOutput.Message.Content,
            }, nil
        }
    }

    return nil, fmt.Errorf("no output from agent")
}

func (m *AgenticMode) processEvent(event *adk.AgentEvent) *AgentResult {
    if event.Err != nil {
        return &AgentResult{Type: "error", Content: event.Err.Error()}
    }

    // 处理中断（确认请求）
    if event.Action != nil && event.Action.Interrupted != nil {
        return m.processInterrupt(event.Action.Interrupted)
    }

    // 处理消息输出
    if event.Output != nil && event.Output.MessageOutput != nil {
        msg := event.Output.MessageOutput.Message
        if msg != nil {
            return &AgentResult{
                Type:    "text",
                Content: msg.Content,
            }
        }
    }

    return nil
}

func (m *AgenticMode) processInterrupt(interrupted *adk.Interrupted) *AgentResult {
    data, ok := interrupted.Data.(*AskRequest)
    if !ok {
        return nil
    }
    return &AgentResult{
        Type: "ask_user",
        Ask:  data,
    }
}
```

### 4. 工具简化设计

#### 4.1 工具注册表

```go
// internal/ai/tools/registry.go

package tools

import (
    "context"

    "github.com/cloudwego/eino/components/tool"
)

// ToolCategory 工具分类
type ToolCategory string

const (
    CategoryK8s     ToolCategory = "k8s"
    CategoryHost    ToolCategory = "host"
    CategoryService ToolCategory = "service"
    CategoryMonitor ToolCategory = "monitor"
    CategoryMCP     ToolCategory = "mcp"
)

// ToolDefinition 工具定义
type ToolDefinition struct {
    Name        string
    Category    ToolCategory
    Description string
    Risk        ToolRisk
    Handler     ToolHandler
    Schema      map[string]any
}

// ToolRegistry 工具注册表
type ToolRegistry struct {
    tools map[string]*ToolDefinition
    mcp   *MCPClientManager
}

// NewToolRegistry 创建工具注册表
func NewToolRegistry(deps PlatformDeps, mcp *MCPClientManager) *ToolRegistry {
    r := &ToolRegistry{
        tools: make(map[string]*ToolDefinition),
        mcp:   mcp,
    }

    // 注册内置工具
    r.registerK8sTools(deps)
    r.registerHostTools(deps)
    r.registerServiceTools(deps)
    r.registerMonitorTools(deps)

    // 注册 MCP 工具
    if mcp != nil {
        r.registerMCPTools()
    }

    return r
}

// GetTools 获取所有工具
func (r *ToolRegistry) GetTools() []tool.BaseTool {
    result := make([]tool.BaseTool, 0, len(r.tools))
    for _, def := range r.tools {
        t := tool.NewTool(def.Name, def.Description, def.Handler, def.Schema)
        // 根据风险等级包装
        result = append(result, wrapTool(t, def))
    }
    return result
}
```

#### 4.2 K8s 工具简化

```go
// internal/ai/tools/k8s.go

package tools

import (
    "context"
    "strings"

    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// K8sQueryInput 统一的 K8s 查询输入
type K8sQueryInput struct {
    ClusterID int    `json:"cluster_id,omitempty" jsonschema:"description=集群 ID"`
    Namespace string `json:"namespace,omitempty" jsonschema:"description=命名空间,default=default"`
    Resource  string `json:"resource" jsonschema:"required,description=资源类型,enum=pods,enum=services,enum=deployments,enum=nodes,enum=configmaps,enum=secrets"`
    Name      string `json:"name,omitempty" jsonschema:"description=资源名称（可选，用于查询单个资源）"`
    Label     string `json:"label,omitempty" jsonschema:"description=标签选择器"`
    Limit     int    `json:"limit,omitempty" jsonschema:"description=返回数量限制,default=50"`
}

// K8sLogsInput K8s 日志输入
type K8sLogsInput struct {
    ClusterID int    `json:"cluster_id,omitempty"`
    Namespace string `json:"namespace,omitempty" jsonschema:"default=default"`
    Pod       string `json:"pod" jsonschema:"required,description=Pod 名称"`
    Container string `json:"container,omitempty" jsonschema:"description=容器名称"`
    TailLines int    `json:"tail_lines,omitempty" jsonschema:"default=200"`
    Follow    bool   `json:"follow,omitempty" jsonschema:"description=是否持续跟踪"`
}

// K8sEventsInput K8s 事件输入
type K8sEventsInput struct {
    ClusterID int    `json:"cluster_id,omitempty"`
    Namespace string `json:"namespace,omitempty"`
    Kind      string `json:"kind,omitempty" jsonschema:"description=资源类型,enum=Pod,enum=Deployment,enum=Service"`
    Name      string `json:"name,omitempty" jsonschema:"description=资源名称"`
    Limit     int    `json:"limit,omitempty" jsonschema:"default=50"`
}

func (r *ToolRegistry) registerK8sTools(deps PlatformDeps) {
    // k8s_query: 统一查询工具
    r.tools["k8s_query"] = &ToolDefinition{
        Name:        "k8s_query",
        Category:    CategoryK8s,
        Description: "查询 Kubernetes 资源",
        Risk:        ToolRiskLow,
        Handler:     makeK8sQueryHandler(deps),
    }

    // k8s_logs: 日志查询
    r.tools["k8s_logs"] = &ToolDefinition{
        Name:        "k8s_logs",
        Category:    CategoryK8s,
        Description: "获取 Pod 日志",
        Risk:        ToolRiskMedium, // 可能包含敏感信息
        Handler:     makeK8sLogsHandler(deps),
    }

    // k8s_events: 事件查询
    r.tools["k8s_events"] = &ToolDefinition{
        Name:        "k8s_events",
        Category:    CategoryK8s,
        Description: "获取 Kubernetes 事件",
        Risk:        ToolRiskLow,
        Handler:     makeK8sEventsHandler(deps),
    }
}

func makeK8sQueryHandler(deps PlatformDeps) ToolHandler {
    return func(ctx context.Context, input K8sQueryInput) (ToolResult, error) {
        cli, err := resolveK8sClient(deps, input.ClusterID)
        if err != nil {
            return ToolResult{}, err
        }

        ns := input.Namespace
        if ns == "" {
            ns = corev1.NamespaceAll
        }

        limit := input.Limit
        if limit <= 0 {
            limit = 50
        }

        var result any
        resource := strings.ToLower(input.Resource)

        switch resource {
        case "pods":
            result, err = listPods(ctx, cli, ns, input.Label, limit)
        case "services":
            result, err = listServices(ctx, cli, ns, input.Label, limit)
        case "deployments":
            result, err = listDeployments(ctx, cli, ns, input.Label, limit)
        case "nodes":
            result, err = listNodes(ctx, cli, limit)
        default:
            return ToolResult{}, fmt.Errorf("unsupported resource type: %s", resource)
        }

        if err != nil {
            return ToolResult{}, err
        }

        return ToolResult{
            OK:   true,
            Data: result,
        }, nil
    }
}
```

### 5. 存储设计

#### 5.1 数据库模型

```go
// internal/model/ai_session.go

package model

import "time"

// AIChatSession AI 会话
type AIChatSession struct {
    ID        string    `gorm:"primaryKey;type:varchar(64)"`
    UserID    uint64    `gorm:"index;not null"`
    Title     string    `gorm:"type:varchar(255)"`
    Scene     string    `gorm:"type:varchar(64);default:'default'"`
    CreatedAt time.Time `gorm:"autoCreateTime"`
    UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (AIChatSession) TableName() string {
    return "ai_sessions"
}

// AIChatMessage AI 消息
type AIChatMessage struct {
    ID        string            `gorm:"primaryKey;type:varchar(64)"`
    SessionID string            `gorm:"index;not null"`
    Role      string            `gorm:"type:varchar(16);not null"` // user/assistant/system
    Type      string            `gorm:"type:varchar(32)"`          // text/tool_result/ask_user
    Content   string            `gorm:"type:text"`
    Metadata  map[string]any    `gorm:"type:jsonb"`                // 工具结果、确认请求等
    CreatedAt time.Time         `gorm:"autoCreateTime"`
}

func (AIChatMessage) TableName() string {
    return "ai_messages"
}
```

#### 5.2 Session Store

```go
// internal/ai/session_store.go

package ai

import (
    "context"
    "fmt"
    "time"

    "github.com/cy77cc/k8s-manage/internal/model"
    "github.com/redis/go-redis/v9"
    "gorm.io/gorm"
)

// SessionStore 会话存储
type SessionStore struct {
    db    *gorm.DB
    redis redis.UniversalClient
}

// NewSessionStore 创建会话存储
func NewSessionStore(db *gorm.DB, redis redis.UniversalClient) *SessionStore {
    return &SessionStore{db: db, redis: redis}
}

// GetSession 获取会话
func (s *SessionStore) GetSession(ctx context.Context, sessionID string) (*model.AIChatSession, error) {
    // 先查 Redis 缓存
    if s.redis != nil {
        cached, err := s.redis.Get(ctx, "session:"+sessionID).Bytes()
        if err == nil {
            var session model.AIChatSession
            if json.Unmarshal(cached, &session) == nil {
                return &session, nil
            }
        }
    }

    // 查数据库
    var session model.AIChatSession
    if err := s.db.WithContext(ctx).First(&session, "id = ?", sessionID).Error; err != nil {
        return nil, err
    }

    // 写入缓存
    if s.redis != nil {
        data, _ := json.Marshal(session)
        s.redis.Set(ctx, "session:"+sessionID, data, 30*time.Minute)
    }

    return &session, nil
}

// AppendMessage 追加消息
func (s *SessionStore) AppendMessage(ctx context.Context, sessionID string, msg *model.AIChatMessage) error {
    // 存入数据库
    if err := s.db.WithContext(ctx).Create(msg).Error; err != nil {
        return err
    }

    // 更新会话时间
    s.db.WithContext(ctx).Model(&model.AIChatSession{}).
        Where("id = ?", sessionID).
        Update("updated_at", time.Now())

    // 清除缓存
    if s.redis != nil {
        s.redis.Del(ctx, "session:"+sessionID)
    }

    return nil
}

// GetMessages 获取会话消息
func (s *SessionStore) GetMessages(ctx context.Context, sessionID string, limit int) ([]model.AIChatMessage, error) {
    var messages []model.AIChatMessage
    query := s.db.WithContext(ctx).
        Where("session_id = ?", sessionID).
        Order("created_at ASC")

    if limit > 0 {
        query = query.Limit(limit)
    }

    if err := query.Find(&messages).Error; err != nil {
        return nil, err
    }

    return messages, nil
}
```

## SSE Event Format

### 统一事件格式

```typescript
// SSE 事件类型
interface SSEEvent {
  type: 'text' | 'text_delta' | 'tool_start' | 'tool_result' | 'ask_user' | 'done' | 'error';
  payload: any;
}

// 文本事件
interface TextEvent {
  type: 'text' | 'text_delta';
  payload: {
    content: string;
    done: boolean;
  };
}

// 工具开始事件
interface ToolStartEvent {
  type: 'tool_start';
  payload: {
    tool: string;
    arguments: Record<string, any>;
  };
}

// 工具结果事件
interface ToolResultEvent {
  type: 'tool_result';
  payload: {
    tool: string;
    ok: boolean;
    data?: any;
    error?: string;
    latency_ms: number;
  };
}

// 用户确认事件
interface AskUserEvent {
  type: 'ask_user';
  payload: {
    ask_id: string;
    session_id: string;
    title: string;
    description: string;
    risk: 'low' | 'medium' | 'high';
    details: Record<string, any>;
  };
}

// 完成事件
interface DoneEvent {
  type: 'done';
  payload: {
    session_id: string;
    message_id: string;
  };
}

// 错误事件
interface ErrorEvent {
  type: 'error';
  payload: {
    code: string;
    message: string;
    recoverable: boolean;
  };
}
```

## Migration Strategy

### 渐进式迁移

```
Phase 1: 后端重构 (不改动前端)
├─ 新 Agent 实现与旧实现并存
├─ 通过 feature flag 切换
└─ 保持现有 API 兼容

Phase 2: 前端重写
├─ 新前端组件独立开发
├─ 通过路由切换新旧版本
└─ 逐步迁移用户

Phase 3: 清理旧代码
├─ 移除旧 Agent 实现
├─ 移除旧前端组件
└─ 清理未使用的路由
```

### Feature Flag

```yaml
# config.yaml
ai:
  agent:
    version: "v2"  # v1 | v2
  features:
    intent_classifier: true
    interactive_confirmation: true
    tool_simplification: true
```

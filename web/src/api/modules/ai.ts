import apiService from '../api';
import type { ApiResponse } from '../api';

// AI对话消息数据结构
export interface AIMessage {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  thinking?: string;
  timestamp: string;
}

// AI对话会话数据结构
export interface AISession {
  id: string;
  title: string;
  messages: AIMessage[];
  createdAt: string;
  updatedAt: string;
}

// AI推荐数据结构
export interface AIRecommendation {
  id: string;
  type: string;
  title: string;
  content: string;
  relevance: number;
  action?: string;
  params?: Record<string, any>;
}

// AI分析结果数据结构
export interface AIAnalysis {
  id: string;
  type: string;
  title: string;
  summary: string;
  details: any;
  createdAt: string;
}

// AI对话请求参数
export interface AIChatParams {
  sessionId?: string;
  message: string;
  context?: any;
}

// AI分析请求参数
export interface AIAnalysisParams {
  type: string;
  data: any;
  context?: any;
}

// AI推荐请求参数
export interface AIRecommendationParams {
  type: string;
  context: any;
  limit?: number;
}

export interface AIActionPreviewParams {
  action: string;
  params: Record<string, any>;
}

export interface AIActionExecuteParams {
  action?: string;
  approval_token: string;
}

interface SSEMetaEvent {
  sessionId: string;
  createdAt: string;
}

interface SSEDeltaEvent {
  contentChunk: string;
}

interface SSEDoneEvent {
  session: AISession;
}

interface SSEErrorEvent {
  message: string;
}
interface SSEThinkingEvent {
  contentChunk: string;
}

export interface AIChatStreamHandlers {
  onMeta?: (payload: SSEMetaEvent) => void;
  onDelta?: (payload: SSEDeltaEvent) => void;
  onDone?: (payload: SSEDoneEvent) => void;
  onError?: (payload: SSEErrorEvent) => void;
  onThinkingDelta?: (payload: SSEThinkingEvent) => void;
  onToolCall?: (payload: { tool?: string; params?: Record<string, any>; tool_calls?: Array<{ function?: { name?: string; arguments?: string } }> }) => void;
  onToolResult?: (payload: { tool?: string; result?: { ok: boolean; data?: any; error?: string; source?: string; latency_ms?: number } }) => void;
  onApprovalRequired?: (payload: ApprovalTicket & { approval_required?: boolean; previewDiff?: string }) => void;
}

export type RiskLevel = 'low' | 'medium' | 'high';

export interface AICapability {
  name: string;
  description: string;
  mode: 'readonly' | 'mutating';
  risk: RiskLevel;
  provider: 'local' | 'mcp';
  schema?: Record<string, any>;
  permission?: string;
}

export interface ToolCallTrace {
  tool: string;
  params?: Record<string, any>;
  at?: string;
}

export interface ApprovalTicket {
  id: string;
  tool: string;
  params: Record<string, any>;
  risk: RiskLevel;
  mode: 'readonly' | 'mutating';
  status: 'pending' | 'approved' | 'rejected' | 'expired';
  createdAt: string;
  expiresAt: string;
}

export interface AIToolExecution {
  id: string;
  tool: string;
  params: Record<string, any>;
  mode: 'readonly' | 'mutating';
  status: 'running' | 'succeeded' | 'failed';
  approvalId?: string;
  createdAt: string;
  finishedAt?: string;
  error?: string;
  result?: {
    ok: boolean;
    data?: any;
    error?: string;
    source: string;
    latency_ms: number;
  };
}

// AI功能API
export const aiApi = {
  // AI对话（SSE流式）
  async chatStream(params: AIChatParams, handlers: AIChatStreamHandlers): Promise<void> {
    const base = import.meta.env.VITE_API_BASE || '/api/v1';
    const token = localStorage.getItem('token');
    const projectId = localStorage.getItem('projectId');

    const response = await fetch(`${base}/ai/chat`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...(projectId ? { 'X-Project-ID': projectId } : {}),
      },
      body: JSON.stringify(params),
    });

    if (!response.ok || !response.body) {
      throw new Error(`请求失败: ${response.status}`);
    }

    const reader = response.body.getReader();
    const decoder = new TextDecoder('utf-8');
    let buffer = '';

    const dispatchEvent = (chunk: string) => {
      const lines = chunk.split('\n');
      let eventType = 'message';
      const dataLines: string[] = [];

      lines.forEach((line) => {
        if (line.startsWith('event:')) {
          eventType = line.slice(6).trim();
          return;
        }
        if (line.startsWith('data:')) {
          dataLines.push(line.slice(5).trim());
        }
      });

      if (dataLines.length === 0) {
        return;
      }

      const rawData = dataLines.join('\n');
      let payload: unknown = rawData;
      try {
        payload = JSON.parse(rawData);
      } catch {
        payload = { message: rawData };
      }

      if (eventType === 'meta') {
        handlers.onMeta?.(payload as SSEMetaEvent);
      } else if (eventType === 'delta') {
        handlers.onDelta?.(payload as SSEDeltaEvent);
      } else if (eventType === 'done') {
        handlers.onDone?.(payload as SSEDoneEvent);
      } else if (eventType === 'error') {
        handlers.onError?.(payload as SSEErrorEvent);
      } else if (eventType === 'thinking_delta') {
        handlers.onThinkingDelta?.(payload as SSEThinkingEvent);
      } else if (eventType === 'tool_call') {
        handlers.onToolCall?.(payload as { tool: string; params?: Record<string, any> });
      } else if (eventType === 'tool_result') {
        handlers.onToolResult?.(payload as { tool?: string; result?: { ok: boolean; data?: any; error?: string; source?: string; latency_ms?: number } });
      } else if (eventType === 'approval_required') {
        handlers.onApprovalRequired?.(payload as ApprovalTicket & { approval_required?: boolean; previewDiff?: string });
      }
    };

    while (true) {
      const { done, value } = await reader.read();
      if (done) {
        break;
      }

      buffer += decoder.decode(value, { stream: true }).replace(/\r/g, '');
      const segments = buffer.split('\n\n');
      buffer = segments.pop() || '';
      segments.forEach(dispatchEvent);
    }

    if (buffer.trim()) {
      dispatchEvent(buffer);
    }
  },

  // 获取对话会话列表
  async getSessions(scene?: string): Promise<ApiResponse<AISession[]>> {
    return apiService.get('/ai/sessions', scene ? { params: { scene } } : undefined);
  },

  async getCurrentSession(scene?: string): Promise<ApiResponse<AISession | null>> {
    return apiService.get('/ai/sessions/current', scene ? { params: { scene } } : undefined);
  },

  // 获取对话会话详情
  async getSessionDetail(id: string): Promise<ApiResponse<AISession>> {
    return apiService.get(`/ai/sessions/${id}`);
  },

  // 删除对话会话
  async deleteSession(id: string): Promise<ApiResponse<void>> {
    return apiService.delete(`/ai/sessions/${id}`);
  },

  // AI分析
  async analyze(params: AIAnalysisParams): Promise<ApiResponse<AIAnalysis>> {
    return apiService.post('/ai/analyze', params);
  },

  // AI推荐
  async getRecommendations(params: AIRecommendationParams): Promise<ApiResponse<AIRecommendation[]>> {
    return apiService.post('/ai/recommendations', params);
  },

  // AI辅助功能
  async getAssistance(params: {
    type: string;
    context: any;
  }): Promise<ApiResponse<{
    assistance: string;
    actions?: Array<{
      label: string;
      value: string;
    }>;
  }>> {
    return apiService.post('/ai/assist', params);
  },

  async previewAction(params: AIActionPreviewParams): Promise<ApiResponse<{
    approval_token: string;
    intent: string;
    risk: string;
    params: Record<string, any>;
    previewDiff: string;
  }>> {
    return apiService.post('/ai/actions/preview', params);
  },

  async executeAction(params: AIActionExecuteParams): Promise<ApiResponse<Record<string, any>>> {
    return apiService.post('/ai/actions/execute', params);
  },

  async k8sAnalyze(params: { cluster_id: number; namespace?: string; question?: string; context?: Record<string, any> }): Promise<ApiResponse<{
    insights: string[];
    risks: string[];
    recommended_actions: Array<{ action: string; params?: Record<string, any>; reason?: string }>;
  }>> {
    return apiService.post('/ai/k8s/analyze', params);
  },

  async previewK8sAction(params: AIActionPreviewParams): Promise<ApiResponse<any>> {
    return apiService.post('/ai/k8s/actions/preview', params);
  },

  async executeK8sAction(params: AIActionExecuteParams): Promise<ApiResponse<Record<string, any>>> {
    return apiService.post('/ai/k8s/actions/execute', params);
  },

  async getCapabilities(): Promise<ApiResponse<AICapability[]>> {
    return apiService.get('/ai/capabilities');
  },

  async previewTool(params: { tool: string; params?: Record<string, any> }): Promise<ApiResponse<Record<string, any>>> {
    return apiService.post('/ai/tools/preview', params);
  },

  async executeTool(params: { tool: string; params?: Record<string, any>; approval_token?: string }): Promise<ApiResponse<AIToolExecution>> {
    return apiService.post('/ai/tools/execute', params);
  },

  async getExecution(id: string): Promise<ApiResponse<AIToolExecution>> {
    return apiService.get(`/ai/executions/${id}`);
  },

  async createApproval(params: { tool: string; params?: Record<string, any> }): Promise<ApiResponse<ApprovalTicket>> {
    return apiService.post('/ai/approvals', params);
  },

  async confirmApproval(id: string, approve: boolean): Promise<ApiResponse<ApprovalTicket>> {
    return apiService.post(`/ai/approvals/${id}/confirm`, { approve });
  },
};

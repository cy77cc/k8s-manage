export interface AIChatToolTrace {
  id: string;
  type: 'tool_call' | 'tool_result';
  tool: string;
  callId?: string;
  timestamp?: string;
  payload?: Record<string, unknown>;
  retry?: boolean;
}

export interface AIChatAskRequest {
  id: string;
  kind?: 'approval' | 'confirmation' | 'review' | 'interrupt';
  title: string;
  description?: string;
  risk?: string;
  status?: 'pending' | 'approved' | 'rejected' | 'confirmed' | 'cancelled' | 'expired';
  details?: Record<string, unknown>;
}

export interface AIChatMessage {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  thinking?: string;
  createdAt?: string;
  traces?: AIChatToolTrace[];
  ask?: AIChatAskRequest | null;
}

export interface AIChatSession {
  id: string;
  title: string;
  scene?: string;
  createdAt?: string;
  updatedAt?: string;
  messages: AIChatMessage[];
}

export interface AIChatRecommendation {
  id: string;
  type: string;
  title: string;
  content: string;
  followupPrompt?: string;
  reasoning?: string;
  relevance?: number;
}

export interface AIChatStreamState {
  status?: 'completed' | 'interrupted' | 'error';
  interrupted?: boolean;
  error?: string;
}

export interface AIChatDonePayload {
  sessionId?: string;
  streamState?: AIChatStreamState;
  turnRecommendations?: AIChatRecommendation[];
  toolSummary?: {
    totalCalls?: number;
    completedCalls?: number;
    failedCalls?: number;
  };
}

export interface AIChatPendingAsk {
  messageId: string;
  ask: AIChatAskRequest;
}

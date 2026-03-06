import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { aiApi } from '../../../api/modules/ai';
import type { AIChatAskRequest, AIChatDonePayload, AIChatMessage, AIChatRecommendation, AIChatSession, AIChatToolTrace } from '../types';

interface UseSSEConnectionOptions {
  scene?: string;
  sessionId?: string;
  initialMessages?: AIChatMessage[];
  runtimeContext?: Record<string, unknown>;
  onSessionResolved?: (session: AIChatSession) => void;
}

function toolTraceFromPayload(
  type: AIChatToolTrace['type'],
  payload: Record<string, unknown> | undefined,
  fallbackTool = '',
): AIChatToolTrace {
  const nestedPayload = payload?.payload && typeof payload.payload === 'object' ? (payload.payload as Record<string, unknown>) : undefined;
  return {
    id: `${type}-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
    type,
    tool: String(payload?.tool || nestedPayload?.tool || fallbackTool || ''),
    callId: String(payload?.call_id || ''),
    timestamp: String(payload?.ts || new Date().toISOString()),
    payload,
    retry: Boolean(payload?.retry),
  };
}

function mapSessionFromDone(session: any): AIChatSession {
  return {
    id: String(session?.id || ''),
    title: String(session?.title || 'AI Session'),
    scene: String(session?.scene || ''),
    createdAt: String(session?.createdAt || ''),
    updatedAt: String(session?.updatedAt || ''),
    messages: Array.isArray(session?.messages)
      ? session.messages.map((item: any) => ({
          id: String(item?.id || ''),
          role: item?.role === 'user' || item?.role === 'assistant' || item?.role === 'system' ? item.role : 'assistant',
          content: String(item?.content || ''),
          thinking: item?.thinking ? String(item.thinking) : undefined,
          createdAt: item?.timestamp ? String(item.timestamp) : undefined,
        }))
      : [],
  };
}

export function useSSEConnection(options: UseSSEConnectionOptions = {}) {
  const scene = options.scene || 'global';
  const [messages, setMessages] = useState<AIChatMessage[]>(options.initialMessages || []);
  const [isLoading, setIsLoading] = useState(false);
  const [streamState, setStreamState] = useState<'idle' | 'running' | 'done' | 'error'>('idle');
  const [streamError, setStreamError] = useState('');
  const [recommendations, setRecommendations] = useState<AIChatRecommendation[]>([]);
  const lastPromptRef = useRef('');
  const activeSessionIdRef = useRef(options.sessionId);
  const activeAssistantIdRef = useRef('');
  const activeTurnIdRef = useRef('');

  useEffect(() => {
    setMessages(options.initialMessages || []);
    activeSessionIdRef.current = options.sessionId;
  }, [options.initialMessages, options.sessionId]);

  const patchAssistant = useCallback((patch: (item: AIChatMessage) => AIChatMessage) => {
    setMessages((prev) =>
      prev.map((item) => {
        if (item.id !== activeAssistantIdRef.current) {
          return item;
        }
        return patch(item);
      }),
    );
  }, []);

  const sendMessage = useCallback(
    async (message: string, extraContext?: Record<string, unknown>) => {
      const trimmed = message.trim();
      if (!trimmed || isLoading) {
        return;
      }

      const userMessage: AIChatMessage = {
        id: `user-${Date.now()}`,
        role: 'user',
        content: trimmed,
        createdAt: new Date().toISOString(),
      };
      const assistantId = `assistant-${Date.now()}`;
      const assistantPlaceholder: AIChatMessage = {
        id: assistantId,
        role: 'assistant',
        content: '',
        createdAt: new Date().toISOString(),
        traces: [],
      };

      activeAssistantIdRef.current = assistantId;
      activeTurnIdRef.current = '';
      lastPromptRef.current = trimmed;
      setStreamError('');
      setStreamState('running');
      setIsLoading(true);
      setMessages((prev) => [...prev, userMessage, assistantPlaceholder]);

      try {
        await aiApi.chatStream(
          {
            sessionId: activeSessionIdRef.current,
            message: trimmed,
            context: { scene, ...(options.runtimeContext || {}), ...(extraContext || {}) },
          },
          {
            onMeta: (payload) => {
              activeTurnIdRef.current = payload.turn_id || '';
              activeSessionIdRef.current = payload.sessionId;
            },
            onDelta: (payload) => {
              patchAssistant((item) => ({
                ...item,
                content: `${item.content || ''}${payload.contentChunk || ''}`,
              }));
            },
            onThinkingDelta: (payload) => {
              patchAssistant((item) => ({
                ...item,
                thinking: `${item.thinking || ''}${payload.contentChunk || ''}`,
              }));
            },
            onToolCall: (payload) => {
              patchAssistant((item) => ({
                ...item,
                traces: [...(item.traces || []), toolTraceFromPayload('tool_call', payload as Record<string, unknown>)],
              }));
            },
            onToolResult: (payload) => {
              patchAssistant((item) => ({
                ...item,
                traces: [...(item.traces || []), toolTraceFromPayload('tool_result', payload as Record<string, unknown>)],
              }));
            },
            onApprovalRequired: (payload) => {
              const extendedPayload = payload as unknown as Record<string, unknown>;
              const ask: AIChatAskRequest = {
                id: String(payload.id || payload.turn_id || Date.now()),
                kind: 'approval',
                title: payload.tool || '待审批操作',
                description: '高风险操作需要审批后继续执行',
                risk: payload.risk,
                status: payload.status,
                details: {
                  checkpointId: extendedPayload.checkpoint_id,
                  sessionId: extendedPayload.sessionId,
                  interruptTargets: extendedPayload.interrupt_targets,
                  approvalToken: payload.id,
                  params: payload.params,
                  preview: extendedPayload.preview,
                },
              };
              patchAssistant((item) => ({ ...item, ask }));
            },
            onConfirmationRequired: (payload) => {
              const ask: AIChatAskRequest = {
                id: String(payload.confirmation_token || payload.turn_id || Date.now()),
                kind: 'confirmation',
                title: payload.tool || '待确认操作',
                description: payload.message || '需要确认后继续执行',
                risk: 'medium',
                status: 'pending',
                details: {
                  confirmationToken: payload.confirmation_token,
                  preview: payload.preview,
                },
              };
              patchAssistant((item) => ({ ...item, ask }));
            },
            onDone: (payload) => {
              const donePayload: AIChatDonePayload = {
                sessionId: payload.session?.id,
                streamState: {
                  status:
                    payload.stream_state === 'failed'
                      ? 'error'
                      : payload.stream_state === 'partial'
                        ? 'interrupted'
                        : 'completed',
                },
                turnRecommendations: payload.turn_recommendations || [],
                toolSummary: payload.tool_summary
                  ? {
                      totalCalls: payload.tool_summary.calls,
                      completedCalls: payload.tool_summary.results,
                      failedCalls: payload.tool_summary.missing?.length || 0,
                    }
                  : undefined,
              };
              if (Array.isArray(donePayload.turnRecommendations)) {
                setRecommendations(donePayload.turnRecommendations);
              }
              if (payload.session) {
                const session = mapSessionFromDone(payload.session);
                activeSessionIdRef.current = session.id;
                options.onSessionResolved?.(session);
                setMessages((prev) => {
                  const latest = prev.filter((item) => !payload.session.messages?.some((msg: any) => String(msg?.id || '') === item.id));
                  return [...latest.filter((item) => item.role === 'user'), ...session.messages];
                });
              }
              setStreamState(donePayload.streamState?.status === 'error' ? 'error' : 'done');
              setIsLoading(false);
            },
            onError: (payload) => {
              setStreamError(payload.message || '流式对话失败');
              setStreamState('error');
              setIsLoading(false);
              patchAssistant((item) => ({
                ...item,
                content: item.content || payload.message || '本轮执行失败',
              }));
            },
          },
        );
      } catch (error) {
        const messageText = error instanceof Error ? error.message : '流式请求失败';
        setStreamError(messageText);
        setStreamState('error');
        patchAssistant((item) => ({ ...item, content: item.content || messageText }));
      } finally {
        setIsLoading(false);
      }
    },
    [isLoading, options, patchAssistant, scene],
  );

  return useMemo(
    () => ({
      messages,
      setMessages,
      isLoading,
      streamState,
      streamError,
      recommendations,
      lastPrompt: lastPromptRef.current,
      sendMessage,
    }),
    [messages, isLoading, streamState, streamError, recommendations, sendMessage],
  );
}

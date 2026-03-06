import { useCallback, useEffect, useMemo, useState } from 'react';
import { aiApi } from '../../../api/modules/ai';
import type { AIMessage, AISession } from '../../../api/modules/ai';
import type { AIChatMessage, AIChatSession } from '../types';

function mapMessage(message: AIMessage): AIChatMessage {
  return {
    id: message.id,
    role: message.role,
    content: message.content,
    thinking: message.thinking,
    createdAt: message.timestamp,
    traces: (message.traces || []).map((trace) => ({
      id: trace.id,
      type: trace.type === 'tool_call' || trace.type === 'tool_result' ? trace.type : 'tool_result',
      tool: String(trace.payload?.tool || trace.payload?.payload?.tool || ''),
      callId: String(trace.payload?.call_id || ''),
      timestamp: trace.timestamp,
      payload: trace.payload,
    })),
  };
}

function mapSession(session: AISession): AIChatSession {
  return {
    id: session.id,
    title: session.title || 'AI Session',
    createdAt: session.createdAt,
    updatedAt: session.updatedAt,
    messages: (session.messages || []).map(mapMessage),
  };
}

interface UseChatSessionOptions {
  scene?: string;
}

export function useChatSession(options: UseChatSessionOptions = {}) {
  const scene = options.scene || 'global';
  const [sessions, setSessions] = useState<AIChatSession[]>([]);
  const [currentSessionId, setCurrentSessionId] = useState<string>();
  const [currentSession, setCurrentSession] = useState<AIChatSession | null>(null);
  const [loading, setLoading] = useState(false);

  const refreshSessions = useCallback(async () => {
    setLoading(true);
    try {
      const res = await aiApi.getSessions(scene);
      const next = (res.data || []).map(mapSession).sort((a, b) => {
        return new Date(b.updatedAt || 0).getTime() - new Date(a.updatedAt || 0).getTime();
      });
      setSessions(next);
      setCurrentSession((prev) => {
        if (!prev) {
          return prev;
        }
        const matched = next.find((item) => item.id === prev.id);
        return matched ? { ...prev, title: matched.title, updatedAt: matched.updatedAt } : prev;
      });
    } finally {
      setLoading(false);
    }
  }, [scene]);

  const loadCurrent = useCallback(async () => {
    setLoading(true);
    try {
      if (currentSessionId) {
        const res = await aiApi.getSessionDetail(currentSessionId);
        setCurrentSession(mapSession(res.data));
        return;
      }
      const res = await aiApi.getCurrentSession(scene);
      if (res.data) {
        const next = mapSession(res.data);
        setCurrentSession(next);
        setCurrentSessionId(next.id);
      } else {
        setCurrentSession(null);
      }
    } finally {
      setLoading(false);
    }
  }, [currentSessionId, scene]);

  useEffect(() => {
    void refreshSessions();
  }, [refreshSessions]);

  useEffect(() => {
    void loadCurrent();
  }, [loadCurrent]);

  const createSession = useCallback(() => {
    setCurrentSessionId(undefined);
    setCurrentSession(null);
  }, []);

  const switchSession = useCallback((id: string) => {
    setCurrentSessionId(id);
  }, []);

  const deleteSession = useCallback(
    async (id: string) => {
      await aiApi.deleteSession(id);
      setSessions((prev) => prev.filter((item) => item.id !== id));
      if (currentSessionId === id) {
        setCurrentSessionId(undefined);
        setCurrentSession(null);
      }
      await refreshSessions();
    },
    [currentSessionId, refreshSessions],
  );

  const updateSession = useCallback((session: AIChatSession) => {
    setCurrentSession(session);
    setCurrentSessionId(session.id);
    setSessions((prev) => {
      const existing = prev.find((item) => item.id === session.id);
      const next = existing
        ? prev.map((item) => (item.id === session.id ? { ...item, ...session } : item))
        : [{ ...session, messages: [] }, ...prev];
      return next.sort((a, b) => {
        return new Date(b.updatedAt || 0).getTime() - new Date(a.updatedAt || 0).getTime();
      });
    });
  }, []);

  return useMemo(
    () => ({
      sessions,
      currentSession,
      currentSessionId,
      loading,
      createSession,
      switchSession,
      deleteSession,
      refreshSessions,
      updateSession,
      setCurrentSession,
    }),
    [
      sessions,
      currentSession,
      currentSessionId,
      loading,
      createSession,
      switchSession,
      deleteSession,
      refreshSessions,
      updateSession,
    ],
  );
}

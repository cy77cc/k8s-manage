import { useCallback, useMemo } from 'react';
import { aiApi } from '../../../api/modules/ai';
import type { AIChatMessage, AIChatPendingAsk } from '../types';

interface UseConfirmationOptions {
  messages: AIChatMessage[];
  setMessages: React.Dispatch<React.SetStateAction<AIChatMessage[]>>;
  lastPrompt?: string;
  sendMessage: (message: string, extraContext?: Record<string, unknown>) => Promise<void>;
}

export function useConfirmation(options: UseConfirmationOptions) {
  const { messages, setMessages, lastPrompt, sendMessage } = options;

  const pendingAsks = useMemo<AIChatPendingAsk[]>(
    () =>
      messages
        .filter((item) => item.ask && item.ask.status !== 'approved' && item.ask.status !== 'confirmed')
        .map((item) => ({ messageId: item.id, ask: item.ask! })),
    [messages],
  );

  const updateAsk = useCallback(
    (messageId: string, patch: (message: AIChatMessage) => AIChatMessage) => {
      setMessages((prev) => prev.map((item) => (item.id === messageId ? patch(item) : item)));
    },
    [setMessages],
  );

  const respondToAsk = useCallback(
    async (messageId: string, approved: boolean) => {
      const target = messages.find((item) => item.id === messageId);
      const ask = target?.ask;
      if (!ask) {
        return;
      }

      if (ask.kind === 'approval') {
        const details = ask.details || {};
        const checkpointId = String(details.checkpointId || details.sessionId || '').trim();
        const interruptTargets = Array.isArray(details.interruptTargets)
          ? details.interruptTargets.map((item) => String(item || '').trim()).filter(Boolean)
          : [];
        const approvalToken = String(details.approvalToken || '').trim();

        if (checkpointId && interruptTargets.length > 0) {
          const resp = await aiApi.respondApproval({
            checkpoint_id: checkpointId,
            target: interruptTargets[0],
            approved,
          });
          updateAsk(messageId, (item) => ({
            ...item,
            content: resp.data.content ? `${item.content}\n\n${resp.data.content}`.trim() : item.content,
            ask: item.ask
              ? {
                  ...item.ask,
                  status: resp.data.interrupted ? 'pending' : approved ? 'approved' : 'rejected',
                  details: {
                    ...(item.ask.details || {}),
                    interruptTargets: resp.data.interrupt_targets || interruptTargets,
                  },
                }
              : item.ask,
          }));
          return;
        }

        if (approvalToken) {
          await aiApi.confirmApproval(approvalToken, approved);
          updateAsk(messageId, (item) => ({
            ...item,
            ask: item.ask ? { ...item.ask, status: approved ? 'approved' : 'rejected' } : item.ask,
          }));
        }
        return;
      }

      if (ask.kind === 'confirmation') {
        const token = String(ask.details?.confirmationToken || '').trim();
        if (!token) {
          return;
        }
        await aiApi.confirmConfirmation(token, approved);
        updateAsk(messageId, (item) => ({
          ...item,
          ask: item.ask ? { ...item.ask, status: approved ? 'confirmed' : 'cancelled' } : item.ask,
        }));
        if (approved && lastPrompt) {
          await sendMessage(lastPrompt, { confirmation_token: token });
        }
      }
    },
    [lastPrompt, messages, sendMessage, updateAsk],
  );

  return useMemo(
    () => ({
      pendingAsks,
      respondToAsk,
    }),
    [pendingAsks, respondToAsk],
  );
}

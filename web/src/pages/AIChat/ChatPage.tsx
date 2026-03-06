import { useMemo, useState } from 'react';
import { ConversationSidebar, type ConversationSummary } from './components/ConversationSidebar';
import { ChatMain } from './components/ChatMain';
import { useAIChatShortcuts, useChatSession, useConfirmation, useSSEConnection } from './hooks';
import './ai-chat.css';

export default function ChatPage() {
  const [draft, setDraft] = useState('');
  const { sessions, currentSession, currentSessionId, createSession, switchSession, deleteSession, updateSession } =
    useChatSession({ scene: 'global' });
  const conversationItems: ConversationSummary[] = useMemo(
    () =>
      sessions.map((item) => ({
        id: item.id,
        title: item.title,
        scene: item.scene,
        updatedAt: item.updatedAt ? new Date(item.updatedAt).toLocaleString() : '刚刚',
        preview: [...item.messages].reverse().find((msg) => msg.role === 'assistant')?.content || '点击继续对话',
      })),
    [sessions],
  );
  const {
    messages,
    setMessages,
    isLoading,
    streamState,
    streamError,
    recommendations,
    lastPrompt,
    sendMessage,
  } = useSSEConnection({
    scene: 'global',
    sessionId: currentSessionId,
    initialMessages: currentSession?.messages || [],
    onSessionResolved: updateSession,
  });
  const { pendingAsks, respondToAsk } = useConfirmation({
    messages,
    setMessages,
    lastPrompt,
    sendMessage,
  });
  useAIChatShortcuts({
    draft,
    disabled: isLoading,
    onCreateConversation: createSession,
    onDraftChange: setDraft,
    onSend: (value) => {
      void sendMessage(value);
    },
  });

  return (
    <div className="ai-chat-page">
      <div className="ai-chat-page__shell">
        <aside className="ai-chat-page__sidebar ai-chat-page__panel-enter" data-stagger="1">
          <ConversationSidebar
            conversations={conversationItems}
            currentId={currentSessionId}
            onSelect={switchSession}
            onCreate={createSession}
            onDelete={deleteSession}
            extraHeader={
              <div className="ai-chat-surface rounded-[32px] p-4">
                <div className="mb-3 flex items-center justify-between gap-3">
                  <span className="inline-flex items-center rounded-full bg-[#fde68a] px-3 py-1 text-xs font-semibold text-[#7c4d02]">
                    AI Chat V2
                  </span>
                  <button
                    type="button"
                    className="inline-flex h-9 w-9 items-center justify-center rounded-full text-lg text-slate-600 transition hover:bg-black/5"
                    onClick={() => createSession()}
                  >
                    +
                  </button>
                </div>
                <h2 className="mb-2 text-xl font-semibold text-[#112018]">对话工作台</h2>
                <p className="mb-4 text-sm leading-6 text-slate-500">
                  会话、工具轨迹、审批流和建议动作统一收敛到同一张操作台。
                </p>
                <div className="ai-chat-mobile-actions flex flex-wrap items-center gap-2">
                  <span className="ai-chat-kbd">/</span>
                  <span className="ai-chat-kbd">
                    Cmd
                  </span>
                  <span className="ai-chat-kbd">↵</span>
                  <span className="text-xs text-slate-500">聚焦输入 / 发送 / 新建会话</span>
                </div>
              </div>
            }
          />
        </aside>
        <main className="ai-chat-page__content">
          <div className="ai-chat-page__panel-enter" data-stagger="2">
            <ChatMain
              title={currentSession?.title || '新会话'}
              scene={currentSession?.scene || 'global'}
              messages={messages}
              draft={draft}
              isLoading={isLoading}
              streamState={streamState}
              streamError={streamError}
              recommendations={recommendations}
              pendingAskCount={pendingAsks.length}
              onDraftChange={setDraft}
              onRespondToAsk={respondToAsk}
              onCreateConversation={createSession}
              onSubmit={(value) => void sendMessage(value)}
            />
          </div>
        </main>
      </div>
    </div>
  );
}

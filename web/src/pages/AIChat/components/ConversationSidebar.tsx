import type { ReactNode } from 'react';
import { useDeferredValue, useMemo, useState } from 'react';

export interface ConversationSummary {
  id: string;
  title: string;
  preview: string;
  updatedAt: string;
  scene?: string;
  unread?: number;
  pinned?: boolean;
}

interface ConversationSidebarProps {
  conversations: ConversationSummary[];
  currentId?: string;
  onSelect?: (id: string) => void;
  onCreate?: () => void;
  onDelete?: (id: string) => void | Promise<void>;
  extraHeader?: ReactNode;
}

const sceneColorMap: Record<string, string> = {
  monitoring: 'red',
  k8s: 'cyan',
  host: 'geekblue',
};

export function ConversationSidebar({
  conversations,
  currentId,
  onSelect,
  onCreate,
  onDelete,
  extraHeader,
}: ConversationSidebarProps) {
  const [keyword, setKeyword] = useState('');
  const deferredKeyword = useDeferredValue(keyword);

  const filtered = useMemo(() => {
    const q = deferredKeyword.trim().toLowerCase();
    if (!q) {
      return conversations;
    }
    return conversations.filter((item) =>
      `${item.title} ${item.preview} ${item.scene ?? ''}`.toLowerCase().includes(q),
    );
  }, [conversations, deferredKeyword]);

  return (
    <div className="flex h-full flex-col gap-4 p-5">
      {extraHeader}
      <label className="relative block">
        <span className="pointer-events-none absolute left-4 top-1/2 -translate-y-1/2 text-slate-400">⌕</span>
        <input
          type="text"
          placeholder="搜索会话、场景或问题"
          value={keyword}
          onChange={(event) => setKeyword(event.target.value)}
          className="w-full rounded-2xl border border-black/8 bg-white/85 py-3 pl-11 pr-4 text-sm text-slate-900 outline-none transition placeholder:text-slate-400 focus:border-[#173022] focus:ring-2 focus:ring-[#173022]/10"
        />
      </label>
      <div className="flex items-center justify-between px-1">
        <span className="text-sm font-semibold text-slate-900">最近会话</span>
        <button
          type="button"
          className="rounded-full px-3 py-1 text-sm text-slate-500 transition hover:bg-black/5 hover:text-slate-900"
          onClick={() => onCreate?.()}
        >
          新建
        </button>
      </div>
      <div className="min-h-0 flex-1 overflow-y-auto pr-1">
        {filtered.length === 0 ? (
          <div className="ai-chat-surface rounded-3xl border border-dashed border-black/10 bg-white/60 py-12">
            <div className="flex flex-col items-center justify-center gap-3 px-6 text-center">
              <div className="text-3xl text-slate-300">∅</div>
              <div className="text-sm text-slate-500">没有匹配的会话</div>
            </div>
          </div>
        ) : (
          <div className="flex flex-col gap-3">
            {filtered.map((item, index) => {
              const active = item.id === currentId;
              const sceneColor = sceneColorMap[item.scene ?? ''] ?? 'default';
              return (
                <button
                  key={item.id}
                  type="button"
                  onClick={() => onSelect?.(item.id)}
                  className={`ai-chat-sidebar-card ai-chat-page__panel-enter w-full rounded-[28px] border px-4 py-4 text-left transition ${
                    active
                      ? 'border-[#173022] bg-[linear-gradient(145deg,_#163321_0%,_#102618_100%)] text-white shadow-[0_16px_38px_rgba(15,23,42,0.18)]'
                      : 'border-black/5 bg-white/82 text-slate-900 shadow-[0_8px_24px_rgba(15,23,42,0.05)]'
                  }`}
                  data-stagger={index > 2 ? '3' : String(index + 1)}
                >
                  <div className="mb-3 flex items-start justify-between gap-3">
                    <div className="flex min-w-0 items-center gap-3">
                      <span
                        className={`flex h-10 w-10 items-center justify-center rounded-full text-sm font-semibold ${
                          active ? 'bg-white/15 text-white' : 'bg-[#efe6d2] text-[#7c5a1b]'
                        }`}
                      >
                        {String(index + 1).padStart(2, '0')}
                      </span>
                      <div className="min-w-0">
                        <div className="truncate font-semibold">{item.title}</div>
                        <div
                          className={`mt-1 flex items-center gap-2 text-xs ${active ? 'text-white/70' : 'text-slate-500'}`}
                        >
                          <span>◷</span>
                          <span>{item.updatedAt}</span>
                        </div>
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      {item.pinned ? (
                        <span className={active ? 'text-white/70' : 'text-slate-400'}>⌖</span>
                      ) : null}
                      {onDelete ? (
                        <button
                          type="button"
                          className={`rounded-full px-2 py-1 text-xs ${active ? 'text-white/70 hover:bg-white/10' : 'text-slate-400 hover:bg-slate-100'}`}
                          onClick={(event) => {
                            event.stopPropagation();
                            void onDelete(item.id);
                          }}
                        >
                          删除
                        </button>
                      ) : null}
                    </div>
                  </div>
                  <p className={`mb-3 line-clamp-2 text-sm ${active ? 'text-white/78' : 'text-slate-600'}`}>
                    {item.preview}
                  </p>
                  <div className="flex items-center justify-between gap-3">
                    <span
                      className={`inline-flex rounded-full px-3 py-1 text-xs font-medium ${
                        sceneColor === 'red'
                          ? 'bg-rose-100 text-rose-700'
                          : sceneColor === 'cyan'
                            ? 'bg-cyan-100 text-cyan-700'
                            : sceneColor === 'geekblue'
                              ? 'bg-blue-100 text-blue-700'
                              : 'bg-slate-100 text-slate-600'
                      }`}
                    >
                      {item.scene || 'general'}
                    </span>
                    {item.unread ? (
                      <span
                        className={`inline-flex min-w-6 items-center justify-center rounded-full px-2 py-1 text-xs font-semibold ${
                          active ? 'bg-amber-300 text-slate-900' : 'bg-slate-900 text-white'
                        }`}
                      >
                        {item.unread}
                      </span>
                    ) : (
                      <span className={`text-xs ${active ? 'text-white/50' : 'text-slate-400'}`}>
                        已读
                      </span>
                    )}
                  </div>
                </button>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}

import { useMemo } from 'react';
import type { AIChatMessage } from '../types';

interface ChatMainProps {
  title: string;
  scene?: string;
  messages: AIChatMessage[];
  draft: string;
  isLoading?: boolean;
  streamState?: string;
  streamError?: string;
  recommendations?: Array<{ id: string; content: string; title: string }>;
  pendingAskCount?: number;
  onDraftChange: (value: string) => void;
  onRespondToAsk: (messageId: string, approved: boolean) => void | Promise<void>;
  onCreateConversation?: () => void;
  onSubmit: (value: string) => void;
}

export function ChatMain({
  title,
  scene,
  messages,
  draft,
  isLoading,
  streamState,
  streamError,
  recommendations,
  pendingAskCount,
  onDraftChange,
  onRespondToAsk,
  onCreateConversation,
  onSubmit,
}: ChatMainProps) {
  const toolTimeline = useMemo(
    () =>
      messages.flatMap((item) =>
        (item.traces ?? []).map((trace) => ({
          key: trace.id,
          tool: trace.tool,
          description: trace.type === 'tool_result' ? '工具已返回结果' : '工具调用中',
          type: trace.type,
        })),
      ),
    [messages],
  );

  return (
    <div className="flex h-full flex-col gap-5">
      <section
        className="ai-chat-surface ai-chat-page__panel-enter overflow-hidden rounded-[32px] bg-[linear-gradient(135deg,_rgba(16,24,40,0.98)_0%,_rgba(27,55,44,0.96)_100%)] text-white shadow-[0_28px_60px_rgba(15,23,42,0.18)]"
        data-stagger="1"
      >
        <div className="flex flex-wrap items-center justify-between gap-4 p-6">
          <div>
            <div className="flex flex-wrap items-center gap-2">
              <span className="inline-flex rounded-full bg-[#fde68a] px-3 py-1 text-xs font-semibold text-[#7c4d02]">
                {scene || 'general'}
              </span>
              <span className="inline-flex rounded-full bg-[#d9f99d] px-3 py-1 text-xs font-semibold text-[#365314]">
                流式对话架构
              </span>
            </div>
            <h1 className="mt-3 text-3xl font-semibold">{title}</h1>
            <p className="mt-2 max-w-3xl text-sm leading-6 text-white/72">
              新页面统一承载消息气泡、工具轨迹、审批确认和下一步建议，当前实现已经接上真实会话与 SSE 流。
            </p>
          </div>
          <div className="flex flex-wrap gap-3">
            <div className="rounded-3xl border border-white/10 bg-white/10 px-4 py-3">
              <div className="text-xs uppercase tracking-[0.2em] text-white/45">Messages</div>
              <div className="mt-1 text-2xl font-semibold">{messages.length}</div>
            </div>
            <div className="rounded-3xl border border-white/10 bg-white/10 px-4 py-3">
              <div className="text-xs uppercase tracking-[0.2em] text-white/45">Pending</div>
              <div className="mt-1 text-2xl font-semibold">{pendingAskCount || 0}</div>
            </div>
            <div className="rounded-3xl border border-white/10 bg-white/10 px-4 py-3">
              <div className="text-xs uppercase tracking-[0.2em] text-white/45">Shortcuts</div>
              <div className="mt-1 flex items-center gap-2 text-sm font-medium">
                <span className="ai-chat-kbd">/</span>
                <span className="ai-chat-kbd">Cmd</span>
                <span className="ai-chat-kbd">↵</span>
              </div>
            </div>
          </div>
        </div>
      </section>

      <div className="grid min-h-0 flex-1 gap-5 xl:grid-cols-[minmax(0,1fr)_320px]">
        <section
          className="ai-chat-surface ai-chat-page__panel-enter flex min-h-[560px] flex-col rounded-[32px] bg-white/86 shadow-[0_18px_45px_rgba(15,23,42,0.07)]"
          data-stagger="2"
        >
          <div className="mb-4 flex flex-wrap items-center justify-between gap-3 p-6 pb-0">
            <div>
              <div className="text-sm text-slate-500">主聊天区域</div>
              <h2 className="mt-1 text-2xl font-semibold text-slate-900">ChatMain</h2>
            </div>
            <div className="flex flex-wrap gap-2">
              <button
                type="button"
                className="rounded-full border border-black/8 bg-white px-4 py-2 text-sm text-slate-700 transition hover:bg-slate-50"
                onClick={() => onCreateConversation?.()}
              >
                新建会话
              </button>
              <button
                type="button"
                className="rounded-full border border-black/8 bg-white px-4 py-2 text-sm text-slate-700 transition hover:bg-slate-50"
                onClick={() => onDraftChange('')}
              >
                清空草稿
              </button>
              <span className="inline-flex items-center rounded-full bg-[#173022] px-4 py-2 text-sm font-medium text-white">
                {streamState === 'running' ? '流式处理中' : scene || 'global'}
              </span>
            </div>
          </div>

          {streamError ? (
            <div className="mx-6 mb-4 rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">
              {streamError}
            </div>
          ) : null}

          <div className="min-h-0 flex-1 overflow-y-auto px-6 pr-4">
            <div className="flex flex-col gap-4">
              {messages.map((message) => (
                <div
                  key={message.id}
                  className={`ai-chat-message-enter flex flex-col gap-3 ${
                    message.role === 'user' ? 'items-end' : 'items-start'
                  }`}
                >
                  <div
                    className={`max-w-[min(780px,100%)] rounded-[28px] px-5 py-4 shadow-[0_12px_30px_rgba(15,23,42,0.06)] ${
                      message.role === 'user'
                        ? 'bg-[#111827] text-white'
                        : 'border border-black/6 bg-white text-slate-900'
                    }`}
                  >
                    <div className="mb-3 flex items-center gap-3">
                      <span
                        className={`inline-flex h-10 w-10 items-center justify-center rounded-full text-sm font-semibold ${
                          message.role === 'user' ? 'bg-white/10 text-white' : 'bg-[#ede9fe] text-[#7c3aed]'
                        }`}
                      >
                        {message.role === 'user' ? 'U' : 'AI'}
                      </span>
                      <div className="text-sm font-semibold">{message.role === 'user' ? '你' : 'AI 助手'}</div>
                    </div>
                    <div className="space-y-3">
                      <p className="whitespace-pre-wrap text-sm leading-7">{message.content}</p>
                      {message.ask ? (
                        <div className="rounded-2xl border border-amber-200 bg-amber-50 p-4 text-slate-900">
                          <div className="mb-2 flex items-center gap-2">
                            <span className="text-amber-600">⚠</span>
                            <span className="font-semibold">{message.ask.title}</span>
                            <span className="rounded-full bg-rose-100 px-3 py-1 text-xs font-semibold text-rose-700">
                              {message.ask.risk || 'high'}
                            </span>
                          </div>
                          <p className="mb-3 text-sm leading-6 text-slate-600">{message.ask.description}</p>
                          <div className="mb-3 flex gap-2">
                            <button
                              type="button"
                              className="rounded-full bg-[#173022] px-4 py-2 text-sm font-medium text-white transition hover:bg-[#102618]"
                              onClick={() => void onRespondToAsk(message.id, true)}
                            >
                              确认
                            </button>
                            <button
                              type="button"
                              className="rounded-full border border-black/10 bg-white px-4 py-2 text-sm font-medium text-slate-700 transition hover:bg-slate-50"
                              onClick={() => void onRespondToAsk(message.id, false)}
                            >
                              取消
                            </button>
                          </div>
                          <pre className="m-0 overflow-x-auto rounded-xl bg-black/85 p-3 text-xs text-emerald-200">
                            {JSON.stringify(message.ask.details ?? {}, null, 2)}
                          </pre>
                        </div>
                      ) : null}
                    </div>
                  </div>
                  {message.createdAt ? (
                    <span className="px-2 text-xs text-slate-400">{message.createdAt}</span>
                  ) : null}
                </div>
              ))}
              {isLoading ? (
                <div className="ai-chat-message-enter flex items-center gap-3 pl-2">
                  <span className="inline-flex h-10 w-10 items-center justify-center rounded-full bg-[#20543d] text-white">AI</span>
                  <div className="rounded-[22px] border border-[#d7e4dc] bg-[#f3f8f4] px-4 py-3">
                    <div className="mb-2 text-sm font-medium text-[#20543d]">Agent 正在整理上下文</div>
                    <div className="ai-chat-loading-dots" aria-label="loading">
                      <span />
                      <span />
                      <span />
                    </div>
                  </div>
                </div>
              ) : null}
            </div>
          </div>

          <div className="mx-6 my-5 h-px bg-black/6" />

          <div className="p-6 pt-0">
            <div className="rounded-[28px] border border-black/5 bg-[#f7f4ee] p-4" data-ai-chat-composer="true">
              <textarea
                value={draft}
                onChange={(event) => onDraftChange(event.target.value)}
                placeholder="输入你的运维问题，后续任务会接入真实会话与 SSE 流..."
                className="min-h-[140px] w-full resize-none rounded-[24px] border border-black/5 bg-white/70 p-4 text-sm leading-7 text-slate-900 outline-none transition placeholder:text-slate-400 focus:border-[#173022] focus:ring-2 focus:ring-[#173022]/10"
              />
              <div className="mt-3 flex flex-wrap items-center justify-between gap-3 text-xs text-slate-500">
                <span className="flex items-center gap-2">
                  <span className="ai-chat-kbd">/</span>
                  聚焦输入
                  <span className="ai-chat-kbd">Cmd</span>
                  <span className="ai-chat-kbd">↵</span>
                  发送
                </span>
                <div className="flex items-center gap-2">
                  <span>Esc 清空并失焦，N 新建会话</span>
                  <button
                    type="button"
                    disabled={isLoading}
                    className="rounded-full bg-[#173022] px-4 py-2 text-sm font-medium text-white transition hover:bg-[#102618] disabled:cursor-not-allowed disabled:opacity-60"
                    onClick={() => onSubmit(draft)}
                  >
                    发送
                  </button>
                </div>
              </div>
            </div>
          </div>
        </section>

        <div className="flex min-h-[560px] flex-col gap-5">
          <section
            className="ai-chat-surface ai-chat-page__panel-enter rounded-[32px] bg-[rgba(247,244,238,0.96)] p-6 shadow-[0_18px_45px_rgba(15,23,42,0.06)]"
            data-stagger="2"
          >
            <div className="text-sm text-slate-500">工具执行轨迹</div>
            <h3 className="mt-1 text-2xl font-semibold text-slate-900">Tool Timeline</h3>
            {toolTimeline.length === 0 ? (
              <p className="mt-4 text-sm leading-6 text-slate-500">当前会话还没有工具事件，后续会在这里展示调用链与结果。</p>
            ) : (
              <div className="mt-4 flex flex-col gap-3">
                {toolTimeline.map((item) => (
                  <div key={item.key} className="rounded-2xl border border-black/5 bg-white/70 px-4 py-3">
                    <div className="flex items-center gap-3">
                      <span
                        className={`inline-flex h-3 w-3 rounded-full ${
                          item.type === 'tool_result' ? 'bg-emerald-500' : 'bg-sky-500'
                        }`}
                      />
                      <div>
                        <div className="text-sm font-semibold text-slate-900">{item.tool}</div>
                        <div className="text-xs text-slate-500">{item.description}</div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </section>

          <section
            className="ai-chat-surface ai-chat-page__panel-enter rounded-[32px] bg-[rgba(255,255,255,0.86)] p-6 shadow-[0_18px_45px_rgba(15,23,42,0.06)]"
            data-stagger="3"
          >
            <div className="text-sm text-slate-500">下一步建议</div>
            <h3 className="mt-1 text-2xl font-semibold text-slate-900">Prompt Suggestions</h3>
            <div className="mt-4 flex flex-col gap-3">
              {(recommendations?.length
                ? recommendations.map((item) => item.title || item.content)
                : ['继续查看相关日志', '生成一份执行前预检查', '汇总当前结论并给出回滚建议']
              ).map((item) => (
                <button
                  key={item}
                  type="button"
                  className="ai-chat-recommendation rounded-2xl border border-black/5 bg-[#f6f8fb] px-4 py-3 text-left text-sm transition hover:bg-[#eef3fb]"
                  onClick={() => onDraftChange(item)}
                >
                  {item}
                </button>
              ))}
            </div>
          </section>
        </div>
      </div>
    </div>
  );
}

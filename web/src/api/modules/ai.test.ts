import { afterEach, describe, expect, it, vi } from 'vitest';
import { aiApi } from './ai';

function buildStream(chunks: string[]) {
  const encoder = new TextEncoder();
  return new ReadableStream<Uint8Array>({
    start(controller) {
      chunks.forEach((chunk) => controller.enqueue(encoder.encode(chunk)));
      controller.close();
    },
  });
}

describe('aiApi.chatStream', () => {
  afterEach(() => {
    vi.restoreAllMocks();
    localStorage.clear();
  });

  it('maps legacy message events to onDelta content', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue({
      ok: true,
      body: buildStream([
        'event: meta\ndata: {"sessionId":"sess-1"}\n\n',
        'event: message\ndata: {"content":"hello from backend"}\n\n',
        'event: done\ndata: {"stream_state":"ok"}\n\n',
      ]),
    } as Response);

    const onMeta = vi.fn();
    const onDelta = vi.fn();
    const onDone = vi.fn();

    await aiApi.chatStream(
      { message: 'hi', context: { scene: 'global' } },
      { onMeta, onDelta, onDone }
    );

    expect(fetchMock).toHaveBeenCalled();
    expect(onMeta).toHaveBeenCalledWith(expect.objectContaining({ sessionId: 'sess-1' }));
    expect(onDelta).toHaveBeenCalledWith(expect.objectContaining({ contentChunk: 'hello from backend' }));
    expect(onDone).toHaveBeenCalled();
  });

  it('dispatches high-level orchestration events', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue({
      ok: true,
      body: buildStream([
        'event: rewrite_result\ndata: {"user_visible_summary":"rewrite ok"}\n\n',
        'event: plan_created\ndata: {"user_visible_summary":"plan ok"}\n\n',
        'event: step_update\ndata: {"step_id":"step-1","status":"running","user_visible_summary":"executing"}\n\n',
        'event: summary\ndata: {"output":{"summary":"done","conclusion":"finished"}}\n\n',
      ]),
    } as Response);

    const onRewriteResult = vi.fn();
    const onPlanCreated = vi.fn();
    const onStepUpdate = vi.fn();
    const onSummary = vi.fn();

    await aiApi.chatStream(
      { message: 'hi', context: { scene: 'global' } },
      { onRewriteResult, onPlanCreated, onStepUpdate, onSummary }
    );

    expect(onRewriteResult).toHaveBeenCalledWith(expect.objectContaining({ user_visible_summary: 'rewrite ok' }));
    expect(onPlanCreated).toHaveBeenCalledWith(expect.objectContaining({ user_visible_summary: 'plan ok' }));
    expect(onStepUpdate).toHaveBeenCalledWith(expect.objectContaining({ step_id: 'step-1', status: 'running' }));
    expect(onSummary).toHaveBeenCalledWith(expect.objectContaining({ output: expect.objectContaining({ summary: 'done' }) }));
  });
});

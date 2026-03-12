import type { AIReplayBlock, AIReplayTurn } from '../../api/modules/ai';
import type { ChatTurn, ChatTurnStatus, EmbeddedRecommendation, TurnBlock } from './types';

export type DisplayMode = 'normal' | 'debug';

export function createAssistantTurn(turnId: string, patch?: Partial<ChatTurn>): ChatTurn {
  const now = new Date().toISOString();
  return {
    id: turnId,
    role: 'assistant',
    status: 'streaming',
    phase: 'rewrite',
    blocks: [],
    createdAt: now,
    updatedAt: now,
    ...patch,
  };
}

export function turnFromReplay(turn: AIReplayTurn): ChatTurn {
  return {
    id: turn.id,
    role: turn.role,
    status: normalizeTurnStatus(turn.status),
    phase: turn.phase,
    traceId: turn.traceId,
    parentTurnId: turn.parentTurnId,
    blocks: (turn.blocks || []).map(blockFromReplay),
    createdAt: turn.createdAt,
    updatedAt: turn.updatedAt,
    completedAt: turn.completedAt,
  };
}

export function applyTurnStarted(
  current: ChatTurn | undefined,
  payload: { turn_id: string; phase?: string; status?: string; role?: string },
  traceId?: string,
): ChatTurn {
  const base = current && current.id === payload.turn_id
    ? current
    : createAssistantTurn(payload.turn_id);
  return {
    ...base,
    role: payload.role === 'user' ? 'user' : 'assistant',
    phase: payload.phase || base.phase,
    status: normalizeTurnStatus(payload.status, base.status),
    traceId: traceId || base.traceId,
    updatedAt: new Date().toISOString(),
  };
}

export function applyBlockOpen(
  current: ChatTurn | undefined,
  payload: {
    turn_id: string;
    block_id: string;
    block_type: string;
    position?: number;
    status?: string;
    title?: string;
    payload?: Record<string, unknown>;
  },
): ChatTurn {
  const turn = current && current.id === payload.turn_id ? current : createAssistantTurn(payload.turn_id);
  const block = upsertBlock(turn.blocks, {
    id: payload.block_id,
    type: normalizeBlockType(payload.block_type),
    position: payload.position ?? nextBlockPosition(turn.blocks),
    status: payload.status,
    title: payload.title,
    data: payload.payload,
    content: resolveBlockContent(undefined, payload.payload),
    streaming: true,
  });
  return {
    ...turn,
    blocks: block,
    updatedAt: new Date().toISOString(),
  };
}

export function applyBlockDelta(
  current: ChatTurn | undefined,
  payload: {
    turn_id: string;
    block_id: string;
    block_type?: string;
    patch?: Record<string, unknown>;
  },
): ChatTurn {
  const turn = current && current.id === payload.turn_id ? current : createAssistantTurn(payload.turn_id);
  const existing = findBlock(turn.blocks, payload.block_id);
  const patch = payload.patch || {};
  const nextContent = appendContent(
    existing?.content,
    resolveBlockContent(patch, patch.payload as Record<string, unknown> | undefined),
  );
  return {
    ...turn,
    blocks: upsertBlock(turn.blocks, {
      id: payload.block_id,
      type: normalizeBlockType(payload.block_type || existing?.type || 'status'),
      position: existing?.position ?? nextBlockPosition(turn.blocks),
      title: asString(patch.title) || existing?.title,
      status: asString(patch.status) || existing?.status,
      content: nextContent || existing?.content,
      data: mergePayload(existing?.data, patch.payload as Record<string, unknown> | undefined, patch),
      streaming: patch.streaming === false ? false : true,
    }),
    updatedAt: new Date().toISOString(),
  };
}

export function applyBlockReplace(
  current: ChatTurn | undefined,
  payload: {
    turn_id: string;
    block_id: string;
    block_type?: string;
    payload?: Record<string, unknown>;
  },
): ChatTurn {
  const turn = current && current.id === payload.turn_id ? current : createAssistantTurn(payload.turn_id);
  const existing = findBlock(turn.blocks, payload.block_id);
  return {
    ...turn,
    blocks: upsertBlock(turn.blocks, {
      id: payload.block_id,
      type: normalizeBlockType(payload.block_type || existing?.type || 'status'),
      position: existing?.position ?? nextBlockPosition(turn.blocks),
      title: asString(payload.payload?.title) || existing?.title,
      status: asString(payload.payload?.status) || existing?.status,
      content: resolveBlockContent(payload.payload, payload.payload) || existing?.content,
      data: payload.payload || existing?.data,
      streaming: payload.payload?.streaming === true,
    }),
    updatedAt: new Date().toISOString(),
  };
}

export function applyBlockClose(
  current: ChatTurn | undefined,
  payload: { turn_id: string; block_id: string; status?: string },
): ChatTurn {
  const turn = current && current.id === payload.turn_id ? current : createAssistantTurn(payload.turn_id);
  const existing = findBlock(turn.blocks, payload.block_id);
  if (!existing) {
    return turn;
  }
  return {
    ...turn,
    blocks: upsertBlock(turn.blocks, {
      ...existing,
      status: payload.status || existing.status,
      streaming: false,
    }),
    updatedAt: new Date().toISOString(),
  };
}

export function applyTurnState(
  current: ChatTurn | undefined,
  payload: { turn_id: string; status?: string; phase?: string },
): ChatTurn {
  const turn = current && current.id === payload.turn_id ? current : createAssistantTurn(payload.turn_id);
  return {
    ...turn,
    status: normalizeTurnStatus(payload.status, turn.status),
    phase: payload.phase || turn.phase,
    updatedAt: new Date().toISOString(),
  };
}

export function applyTurnDone(
  current: ChatTurn | undefined,
  payload: { turn_id: string; status?: string; phase?: string },
): ChatTurn {
  const turn = current && current.id === payload.turn_id ? current : createAssistantTurn(payload.turn_id);
  const doneAt = new Date().toISOString();
  return {
    ...turn,
    status: normalizeTurnStatus(payload.status, 'completed'),
    phase: payload.phase || 'done',
    updatedAt: doneAt,
    completedAt: doneAt,
    blocks: turn.blocks.map((block) => ({ ...block, streaming: false })),
  };
}

export function projectTurnSummary(turn: ChatTurn | undefined): {
  content: string;
  thinking?: string;
  rawEvidence?: string[];
  recommendations?: EmbeddedRecommendation[];
} {
  if (!turn) {
    return { content: '' };
  }

  const orderedBlocks = [...turn.blocks].sort((a, b) => a.position - b.position);
  const textContent = orderedBlocks
    .filter((block) => block.type === 'text')
    .map((block) => block.content || '')
    .join('');
  const thinking = orderedBlocks
    .filter((block) => block.type === 'thinking')
    .map((block) => block.content || '')
    .join('');
  const rawEvidence = orderedBlocks
    .filter((block) => block.type === 'evidence')
    .flatMap((block) => {
      const items = block.data?.items;
      if (Array.isArray(items)) {
        return items.map((item) => String(item));
      }
      return block.content ? [block.content] : [];
    });
  const recommendations = orderedBlocks
    .filter((block) => block.type === 'recommendations')
    .flatMap((block) => {
      const items = block.data?.recommendations || block.data?.items;
      return Array.isArray(items) ? (items as EmbeddedRecommendation[]) : [];
    });

  return {
    content: textContent,
    thinking: thinking || undefined,
    rawEvidence: rawEvidence.length > 0 ? rawEvidence : undefined,
    recommendations: recommendations.length > 0 ? recommendations : undefined,
  };
}

export function getTurnBlocksForDisplay(
  turn: ChatTurn | undefined,
  displayMode: DisplayMode,
  reducedMotion: boolean,
): TurnBlock[] {
  if (!turn) {
    return [];
  }
  return turn.blocks
    .filter((block) => displayMode === 'debug' || block.type !== 'thinking')
    .map((block) => {
      if (displayMode === 'normal' && block.type === 'tool') {
        const payload = block.data || {};
        return {
          ...block,
          streaming: reducedMotion ? false : block.streaming,
          data: {
            tool_name: payload.tool_name,
            tool: payload.tool,
            error: payload.error,
            result: payload.result
              ? {
                  ok: (payload.result as Record<string, unknown>).ok,
                  error: (payload.result as Record<string, unknown>).error,
                  latency_ms: (payload.result as Record<string, unknown>).latency_ms,
                }
              : undefined,
          },
        };
      }

      if (displayMode === 'normal' && block.type === 'evidence') {
        return {
          ...block,
          streaming: false,
          data: {
            items: Array.isArray(block.data?.items) ? (block.data?.items as unknown[]).slice(0, 4) : undefined,
          },
        };
      }

      return {
        ...block,
        streaming: reducedMotion ? false : block.streaming,
      };
    });
}

function blockFromReplay(block: AIReplayBlock): TurnBlock {
  return {
    id: block.id,
    type: normalizeBlockType(block.blockType),
    position: block.position,
    status: block.status,
    title: block.title,
    content: block.contentText,
    data: block.contentJson,
    streaming: block.streaming,
  };
}

function normalizeTurnStatus(status: string | undefined, fallback: ChatTurnStatus = 'streaming'): ChatTurnStatus {
  switch (status) {
    case 'waiting_user':
    case 'waiting_approval':
      return 'waiting_user';
    case 'completed':
    case 'success':
    case 'approved':
    case 'rejected':
    case 'cancelled':
      return 'completed';
    case 'failed':
    case 'error':
      return 'error';
    case 'streaming':
    case 'running':
    default:
      return fallback;
  }
}

function normalizeBlockType(type: string | undefined): TurnBlock['type'] {
  switch (type) {
    case 'text':
    case 'status':
    case 'plan':
    case 'tool':
    case 'approval':
    case 'evidence':
    case 'thinking':
    case 'error':
    case 'recommendations':
      return type;
    default:
      return 'status';
  }
}

function nextBlockPosition(blocks: TurnBlock[]): number {
  return blocks.reduce((max, block) => Math.max(max, block.position), 0) + 1;
}

function findBlock(blocks: TurnBlock[], blockID: string): TurnBlock | undefined {
  return blocks.find((block) => block.id === blockID);
}

function upsertBlock(blocks: TurnBlock[], patch: TurnBlock): TurnBlock[] {
  const index = blocks.findIndex((block) => block.id === patch.id);
  if (index === -1) {
    return [...blocks, patch].sort((a, b) => a.position - b.position);
  }

  return blocks.map((block, blockIndex) => (blockIndex === index ? { ...block, ...patch } : block));
}

function appendContent(current: string | undefined, delta: string | undefined): string | undefined {
  if (!delta) {
    return current;
  }
  return `${current || ''}${delta}`;
}

function resolveBlockContent(
  patch?: Record<string, unknown>,
  payload?: Record<string, unknown>,
): string | undefined {
  const content = patch?.content_chunk ?? patch?.content ?? patch?.text ?? payload?.content_chunk ?? payload?.content ?? payload?.text ?? payload?.summary;
  return typeof content === 'string' ? content : undefined;
}

function mergePayload(
  existing: Record<string, unknown> | undefined,
  payload: Record<string, unknown> | undefined,
  patch: Record<string, unknown>,
): Record<string, unknown> | undefined {
  const merged = {
    ...(existing || {}),
    ...(payload || {}),
  };

  if (Object.prototype.hasOwnProperty.call(patch, 'status')) {
    merged.status = patch.status;
  }
  if (Object.prototype.hasOwnProperty.call(patch, 'title')) {
    merged.title = patch.title;
  }
  if (Object.keys(merged).length === 0) {
    return undefined;
  }
  return merged;
}

function asString(value: unknown): string | undefined {
  return typeof value === 'string' && value.trim() ? value : undefined;
}

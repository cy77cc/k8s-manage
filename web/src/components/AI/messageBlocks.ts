import type { EmbeddedRecommendation, TurnBlock } from './types';

export interface AssistantMessageInput {
  content?: string;
  thinking?: string;
  showThinking?: boolean;
  rawEvidence?: string[];
  recommendations?: EmbeddedRecommendation[];
  isStreaming?: boolean;
}

interface BaseBlock {
  id: string;
}

export interface ThinkingBlock extends BaseBlock {
  type: 'thinking';
  content: string;
  isStreaming?: boolean;
}

export interface MarkdownBlock extends BaseBlock {
  type: 'markdown';
  content: string;
  streaming?: boolean;
}

export interface StatusBlock extends BaseBlock {
  type: 'status';
  title?: string;
  content: string;
  status?: string;
}

export interface PlanBlock extends BaseBlock {
  type: 'plan';
  title?: string;
  content?: string;
  payload?: Record<string, unknown>;
}

export interface ToolExecutionBlock extends BaseBlock {
  type: 'tool';
  title?: string;
  status?: string;
  payload?: Record<string, unknown>;
}

export interface ApprovalBlock extends BaseBlock {
  type: 'approval';
  title?: string;
  payload?: Record<string, unknown>;
}

export interface EvidenceBlock extends BaseBlock {
  type: 'evidence';
  title?: string;
  items: string[];
  payload?: Record<string, unknown>;
}

export interface ErrorBlock extends BaseBlock {
  type: 'error';
  title?: string;
  content: string;
}

export interface RecommendationsBlock extends BaseBlock {
  type: 'recommendations';
  recommendations: EmbeddedRecommendation[];
}

export interface RawEvidenceBlock extends BaseBlock {
  type: 'raw_evidence';
  items: string[];
}

export interface FallbackBlock extends BaseBlock {
  type: 'fallback';
  content: string;
}

export type AssistantMessageBlock =
  | ThinkingBlock
  | MarkdownBlock
  | StatusBlock
  | PlanBlock
  | ToolExecutionBlock
  | ApprovalBlock
  | EvidenceBlock
  | ErrorBlock
  | RecommendationsBlock
  | RawEvidenceBlock
  | FallbackBlock;

export function normalizeAssistantMessage(input: AssistantMessageInput): AssistantMessageBlock[] {
  const blocks: AssistantMessageBlock[] = [];

  if (input.showThinking || input.thinking) {
    blocks.push({
      id: 'thinking',
      type: 'thinking',
      content: input.thinking || '',
      isStreaming: input.isStreaming && !input.content,
    });
  }

  if (input.content) {
    blocks.push({
      id: 'markdown',
      type: 'markdown',
      content: input.content,
    });
  }

  if (input.recommendations && input.recommendations.length > 0) {
    blocks.push({
      id: 'recommendations',
      type: 'recommendations',
      recommendations: input.recommendations,
    });
  }

  if (input.rawEvidence && input.rawEvidence.length > 0) {
    blocks.push({
      id: 'raw_evidence',
      type: 'raw_evidence',
      items: input.rawEvidence,
    });
  }

  return blocks;
}

export function normalizeTurnBlocks(turnBlocks: TurnBlock[] | undefined): AssistantMessageBlock[] {
  if (!turnBlocks || turnBlocks.length === 0) {
    return [];
  }

  return [...turnBlocks]
    .sort((a, b) => a.position - b.position)
    .map((block) => {
      switch (block.type) {
        case 'text':
          return {
            id: block.id,
            type: 'markdown',
            content: block.content || '',
            streaming: block.streaming,
          } satisfies MarkdownBlock;
        case 'thinking':
          return {
            id: block.id,
            type: 'thinking',
            content: block.content || '',
            isStreaming: block.streaming,
          } satisfies ThinkingBlock;
        case 'status':
          return {
            id: block.id,
            type: 'status',
            title: block.title,
            content: block.content || stringifyBlockPayload(block.data),
            status: block.status,
          } satisfies StatusBlock;
        case 'plan':
          return {
            id: block.id,
            type: 'plan',
            title: block.title,
            content: block.content,
            payload: block.data,
          } satisfies PlanBlock;
        case 'tool':
          return {
            id: block.id,
            type: 'tool',
            title: block.title,
            status: block.status,
            payload: block.data,
          } satisfies ToolExecutionBlock;
        case 'approval':
          return {
            id: block.id,
            type: 'approval',
            title: block.title,
            payload: block.data,
          } satisfies ApprovalBlock;
        case 'evidence':
          return {
            id: block.id,
            type: 'evidence',
            title: block.title,
            items: extractEvidenceItems(block),
            payload: block.data,
          } satisfies EvidenceBlock;
        case 'error':
          return {
            id: block.id,
            type: 'error',
            title: block.title,
            content: block.content || stringifyBlockPayload(block.data),
          } satisfies ErrorBlock;
        case 'recommendations':
          return {
            id: block.id,
            type: 'recommendations',
            recommendations: ((block.data?.recommendations || block.data?.items || []) as EmbeddedRecommendation[]),
          } satisfies RecommendationsBlock;
        default:
          return {
            id: block.id,
            type: 'fallback',
            content: block.content || stringifyBlockPayload(block.data),
          } satisfies FallbackBlock;
      }
    });
}

function stringifyBlockPayload(payload: Record<string, unknown> | undefined): string {
  if (!payload) {
    return '';
  }
  if (typeof payload.content_chunk === 'string') {
    return payload.content_chunk;
  }
  if (typeof payload.summary === 'string') {
    return payload.summary;
  }
  try {
    return JSON.stringify(payload, null, 2);
  } catch {
    return String(payload);
  }
}

function extractEvidenceItems(block: TurnBlock): string[] {
  const items = block.data?.items;
  if (Array.isArray(items)) {
    return items.map((item) => String(item));
  }
  if (block.content) {
    return [block.content];
  }
  const payload = stringifyBlockPayload(block.data);
  return payload ? [payload] : [];
}

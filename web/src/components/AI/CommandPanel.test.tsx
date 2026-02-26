import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import CommandPanel from './CommandPanel';

const mockApi = vi.hoisted(() => ({
  ai: {
    getCommandSuggestions: vi.fn(),
    getCommandHistory: vi.fn(),
    previewCommand: vi.fn(),
    executeCommand: vi.fn(),
    getCommandHistoryDetail: vi.fn(),
  },
}));

vi.mock('../../api', () => ({ Api: mockApi }));

describe('CommandPanel', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockApi.ai.getCommandSuggestions.mockResolvedValue({ data: [{ command: 'ops.aggregate.status limit=5', hint: '聚合' }] });
    mockApi.ai.getCommandHistory.mockResolvedValue({ data: { list: [], total: 0 } });
    mockApi.ai.previewCommand.mockResolvedValue({
      data: {
        status: 'previewed',
        summary: 'ok',
        artifacts: { command_id: 'cmd-1' },
        trace_id: 'trace-1',
        next_actions: [],
        risk: 'readonly',
        plan: { steps: [] },
      },
    });
    mockApi.ai.executeCommand.mockResolvedValue({
      data: {
        status: 'succeeded',
        summary: 'done',
        artifacts: { command_id: 'cmd-1' },
        trace_id: 'trace-1',
        next_actions: [],
        risk: 'readonly',
      },
    });
  });

  it('previews then executes command', async () => {
    render(<CommandPanel scene="scene:test" />);
    await screen.findByText('ops.aggregate.status limit=5');

    fireEvent.change(screen.getByPlaceholderText('例如: ops.aggregate.status limit=5'), { target: { value: 'ops.aggregate.status limit=5' } });
    fireEvent.click(screen.getByRole('button', { name: /预\s*览/ }));

    await waitFor(() => expect(mockApi.ai.previewCommand).toHaveBeenCalled());
    fireEvent.click(screen.getByRole('button', { name: '确认执行' }));
    await waitFor(() => expect(mockApi.ai.executeCommand).toHaveBeenCalled());
    await waitFor(() => expect(mockApi.ai.getCommandHistory).toHaveBeenCalledTimes(2));
  });
});

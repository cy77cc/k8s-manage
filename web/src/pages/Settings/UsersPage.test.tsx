import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import UsersPage from './UsersPage';

const mockApi = vi.hoisted(() => ({
  rbac: {
    getUserList: vi.fn(),
    createUser: vi.fn(),
    deleteUser: vi.fn(),
  },
}));

vi.mock('../../api', () => ({ Api: mockApi }));

describe('UsersPage accessibility interactions', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockApi.rbac.getUserList.mockResolvedValue({
      data: {
        list: [
          {
            id: '1',
            username: 'alice',
            name: 'Alice',
            email: 'alice@example.com',
            roles: ['admin'],
            status: 'active',
            createdAt: '2026-01-01T00:00:00Z',
            updatedAt: '2026-01-01T00:00:00Z',
          },
        ],
      },
    });
  });

  it('opens detail drawer with keyboard Enter on interactive row', async () => {
    render(<UsersPage />);

    await waitFor(() => expect(screen.getByText('alice')).toBeInTheDocument());

    const rowTrigger = screen.getByLabelText('查看用户 alice 详情');
    fireEvent.keyDown(rowTrigger, { key: 'Enter' });

    expect(await screen.findByText('用户详情')).toBeInTheDocument();
    const dialogs = screen.getAllByRole('dialog');
    const detailDialog = dialogs[dialogs.length - 1];
    expect(within(detailDialog).getByText('alice@example.com')).toBeInTheDocument();
  });
});

import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import Authorized from './Authorized';

const mockUsePermission = vi.hoisted(() => vi.fn());

vi.mock('./PermissionContext', () => ({
  usePermission: mockUsePermission,
}));

describe('Authorized', () => {
  it('renders fallback when permission denied', () => {
    mockUsePermission.mockReturnValue({ loading: false, hasPermission: () => false });

    render(
      <MemoryRouter>
        <Authorized resource="rbac" action="read" fallback={<div>denied</div>}>
          <div>allowed</div>
        </Authorized>
      </MemoryRouter>,
    );

    expect(screen.getByText('denied')).toBeInTheDocument();
    expect(screen.queryByText('allowed')).not.toBeInTheDocument();
  });
});

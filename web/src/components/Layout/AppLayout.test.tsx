import { cleanup, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import AppLayout from './AppLayout';

const mockUseAuth = vi.hoisted(() => vi.fn());
const mockUsePermission = vi.hoisted(() => vi.fn());
const mockUseI18n = vi.hoisted(() => vi.fn());

vi.mock('../Auth/AuthContext', () => ({
  useAuth: mockUseAuth,
}));

vi.mock('../RBAC', () => ({
  usePermission: mockUsePermission,
}));

vi.mock('../../i18n', () => ({
  useI18n: mockUseI18n,
}));

vi.mock('../Project/ProjectSwitcher', () => ({
  default: () => <div data-testid="project-switcher" />,
}));

vi.mock('../AI/GlobalAIAssistant', () => ({
  default: () => <div data-testid="ai-assistant" />,
}));

describe('AppLayout governance menu', () => {
  beforeEach(() => {
    cleanup();
    vi.clearAllMocks();
  });

  const renderWithRouter = () => render(
    <MemoryRouter>
      <AppLayout>
        <div>content</div>
      </AppLayout>
    </MemoryRouter>,
  );

  it('shows governance menu for users with rbac read permission', () => {
    mockUseAuth.mockReturnValue({ logout: vi.fn() });
    mockUsePermission.mockReturnValue({ hasPermission: vi.fn(() => true) });
    mockUseI18n.mockReturnValue({
      t: (key: string) => key,
      lang: 'zh-CN',
      setLang: vi.fn(),
    });

    renderWithRouter();

    expect(screen.getByText('访问治理')).toBeInTheDocument();
  });

  it('hides governance menu for users without rbac read permission', () => {
    mockUseAuth.mockReturnValue({ logout: vi.fn() });
    mockUsePermission.mockReturnValue({ hasPermission: vi.fn(() => false) });
    mockUseI18n.mockReturnValue({
      t: (key: string) => key,
      lang: 'zh-CN',
      setLang: vi.fn(),
    });

    renderWithRouter();

    expect(screen.queryByText('访问治理')).not.toBeInTheDocument();
  });
});

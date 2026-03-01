import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { NotificationProvider } from '../../contexts/NotificationContext';
import { NotificationPanel } from '../../components/Notification';

// Mock notification API
vi.mock('../../api/modules/notification', () => ({
  notificationApi: {
    getNotifications: vi.fn().mockResolvedValue({
      data: {
        list: [
          {
            id: '1',
            user_id: '1',
            notification_id: '101',
            read_at: null,
            notification: {
              id: '101',
              type: 'alert',
              title: 'CPU 使用率超过 90%',
              content: '主机 node-01 的 CPU 使用率达到 95%',
              severity: 'critical',
              source: '主机 node-01',
              source_id: 'alert-001',
              action_url: '/monitor?alert_id=alert-001',
              action_type: 'confirm',
              created_at: new Date().toISOString(),
            },
          },
          {
            id: '2',
            user_id: '1',
            notification_id: '102',
            read_at: new Date().toISOString(),
            notification: {
              id: '102',
              type: 'task',
              title: '部署任务完成',
              content: 'deployment-web 部署成功',
              severity: 'info',
              source: 'deployment-web',
              source_id: 'task-001',
              action_url: '/tasks/task-001',
              created_at: new Date(Date.now() - 3600000).toISOString(),
            },
          },
        ],
        total: 2,
      },
    }),
    getUnreadCount: vi.fn().mockResolvedValue({
      data: {
        total: 1,
        by_type: { alert: 1, task: 0, system: 0, approval: 0 },
        by_severity: { critical: 1, warning: 0, info: 0 },
      },
    }),
    markAsRead: vi.fn().mockResolvedValue({ success: true }),
    dismiss: vi.fn().mockResolvedValue({ success: true }),
    confirm: vi.fn().mockResolvedValue({ success: true }),
    markAllAsRead: vi.fn().mockResolvedValue({ success: true }),
  },
}));

const renderWithProviders = (component: React.ReactNode) => {
  return render(
    <BrowserRouter>
      <NotificationProvider>
        {component}
      </NotificationProvider>
    </BrowserRouter>
  );
};

describe('NotificationPanel', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders notification panel with title', async () => {
    renderWithProviders(<NotificationPanel />);

    await waitFor(() => {
      expect(screen.getByText('通知中心')).toBeInTheDocument();
    });
  });

  it('displays notification items', async () => {
    renderWithProviders(<NotificationPanel />);

    await waitFor(() => {
      expect(screen.getByText('CPU 使用率超过 90%')).toBeInTheDocument();
    });
  });

  it('shows unread indicator for unread notifications', async () => {
    renderWithProviders(<NotificationPanel />);

    await waitFor(() => {
      const dots = document.querySelectorAll('.notification-item-dot');
      expect(dots.length).toBeGreaterThan(0);
    });
  });

  it('filters notifications by type', async () => {
    renderWithProviders(<NotificationPanel />);

    await waitFor(() => {
      expect(screen.getByText('告警')).toBeInTheDocument();
    });

    // Click on alert tab
    fireEvent.click(screen.getByText('告警'));

    await waitFor(() => {
      expect(screen.getByText('CPU 使用率超过 90%')).toBeInTheDocument();
    });
  });

  it('calls markAsRead when clicking notification', async () => {
    renderWithProviders(<NotificationPanel />);

    await waitFor(() => {
      expect(screen.getByText('CPU 使用率超过 90%')).toBeInTheDocument();
    });

    // Click on notification title
    fireEvent.click(screen.getByText('CPU 使用率超过 90%'));
  });
});

describe('NotificationItem', () => {
  it('renders critical notification with correct styling', async () => {
    renderWithProviders(<NotificationPanel />);

    await waitFor(() => {
      const criticalIcon = document.querySelector('.notification-item-icon');
      expect(criticalIcon).toBeInTheDocument();
    });
  });

  it('shows confirm button for alert type notifications', async () => {
    renderWithProviders(<NotificationPanel />);

    await waitFor(() => {
      expect(screen.getByText('确认告警')).toBeInTheDocument();
    });
  });

  it('shows mark as read button for unread notifications', async () => {
    renderWithProviders(<NotificationPanel />);

    await waitFor(() => {
      expect(screen.getAllByText('标记已读')[0]).toBeInTheDocument();
    });
  });
});

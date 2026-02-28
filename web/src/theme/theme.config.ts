export const theme = {
  colors: {
    background: {
      primary: 'var(--color-bg-app)',
      secondary: 'var(--color-bg-surface)',
    },
    text: {
      primary: 'var(--color-text-primary)',
      secondary: 'var(--color-text-secondary)',
      tertiary: '#64748B',
      quaternary: '#94A3B8',
    },
    brand: {
      primary: 'var(--color-brand-500)',
    },
    status: {
      running: 'var(--color-success)',
      warning: 'var(--color-warning)',
      error: 'var(--color-error)',
      offline: '#8C8C8C',
    },
  },
  statusMap: {
    RUNNING: 'running',
    WARNING: 'warning',
    ERROR: 'error',
    OFFLINE: 'offline',
  },
  animations: {
    pulse: {
      duration: 2000,
      fadeIn: 500,
      fadeOut: 1500,
    },
  },
};

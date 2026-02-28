/** @type {import('tailwindcss').Config} */
export default {
  content: [
    './index.html',
    './src/**/*.{js,ts,jsx,tsx}',
  ],
  theme: {
    extend: {
      colors: {
        'bg-primary': 'var(--color-bg-app)',
        'bg-secondary': 'var(--color-bg-surface)',
        'text-primary': 'var(--color-text-primary)',
        'text-secondary': 'var(--color-text-secondary)',
        'text-tertiary': '#64748B',
        'text-quaternary': '#94A3B8',
        'brand-primary': 'var(--color-brand-500)',
        'status-running': 'var(--color-success)',
        'status-warning': 'var(--color-warning)',
        'status-error': 'var(--color-error)',
        'status-offline': '#8C8C8C',
      },
      animation: {
        'status-pulse': 'status-pulse 2s ease-in-out infinite',
      },
      keyframes: {
        'status-pulse': {
          '0%, 100%': {
            opacity: '0.4',
          },
          '50%': {
            opacity: '1',
          },
        },
      },
    },
  },
  plugins: [],
};

/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        'bg-primary': '#0B0E14',
        'bg-secondary': '#161B22',
        'text-primary': '#FFFFFF',
        'text-secondary': '#E5E7EB',
        'text-tertiary': '#D1D5DB',
        'text-quaternary': '#9CA3AF',
        'brand-primary': '#1677FF',
        'status-running': '#52C41A',
        'status-warning': '#FAAD14',
        'status-error': '#FF4D4F',
        'status-offline': '#8C8C8C'
      },
      animation: {
        'status-pulse': 'status-pulse 2s ease-in-out infinite'
      },
      keyframes: {
        'status-pulse': {
          '0%, 100%': {
            opacity: '0.4'
          },
          '50%': {
            opacity: '1'
          }
        }
      }
    },
  },
  plugins: [],
}
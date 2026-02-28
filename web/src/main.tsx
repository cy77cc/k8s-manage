import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { ConfigProvider } from 'antd';
import './index.css';
import App from './App.tsx';
import { I18nProvider } from './i18n';

document.body.dataset.uiTheme = import.meta.env.VITE_UI_THEME_LEGACY === 'true' ? 'legacy' : 'nebula';

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <ConfigProvider
      theme={{
        token: {
          colorPrimary: 'var(--color-brand-500)',
          colorSuccess: 'var(--color-success)',
          colorWarning: 'var(--color-warning)',
          colorError: 'var(--color-error)',
          colorInfo: 'var(--color-info)',
          borderRadius: 12,
          fontFamily: 'var(--font-body)',
        },
      }}
    >
      <I18nProvider>
        <App />
      </I18nProvider>
    </ConfigProvider>
  </StrictMode>,
);

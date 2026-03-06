import React, { Suspense, lazy } from 'react';
import { BrowserRouter, Navigate, Route, Routes, useLocation } from 'react-router-dom';
import { AuthProvider, useAuth } from './components/Auth/AuthContext';
import { NotificationProvider } from './contexts/NotificationContext';
import { PermissionProvider } from './components/RBAC/PermissionContext';
import { usePermission } from './components/RBAC/PermissionContext';

const LoginPage = lazy(() => import('./pages/Auth/LoginPage'));
const RegisterPage = lazy(() => import('./pages/Auth/RegisterPage'));
const AIChatPage = lazy(() => import('./pages/AIChat/ChatPage'));
const ProtectedApp = lazy(() => import('./ProtectedApp'));

const RouteFallback: React.FC = () => (
  <div className="min-h-screen flex items-center justify-center">加载中...</div>
);

const ProtectedRoute: React.FC<{ children: React.ReactElement }> = ({ children }) => {
  const { isAuthenticated, loading } = useAuth();
  const location = useLocation();

  if (loading) {
    return <RouteFallback />;
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />;
  }

  return children;
};

const AIProtectedRoute: React.FC = () => {
  const { user } = useAuth();
  const { hasPermission } = usePermission();

  if (!hasPermission('ai', 'read')) {
    return (
      <div className="min-h-screen bg-[radial-gradient(circle_at_top,_#f5f0e2,_#e4efe6_48%,_#f3f7f4_100%)] px-6 py-10 text-slate-900">
        <div className="mx-auto max-w-3xl rounded-[32px] border border-black/5 bg-white/88 p-10 shadow-[0_24px_60px_rgba(15,23,42,0.08)]">
          <div className="mb-4 text-xs uppercase tracking-[0.24em] text-slate-400">AI Workspace</div>
          <h1 className="text-3xl font-semibold text-slate-900">没有访问 AI 工作台的权限</h1>
          <p className="mt-4 text-base leading-7 text-slate-600">
            当前账号 {user?.username ? `(${user.username})` : ''} 无法访问该页面。请联系管理员授予 `ai:read` 权限。
          </p>
        </div>
      </div>
    );
  }

  return (
    <NotificationProvider userId={user?.id}>
      <AIChatPage />
    </NotificationProvider>
  );
};

const App: React.FC = () => {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Suspense fallback={<RouteFallback />}>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/register" element={<RegisterPage />} />
            <Route
              path="/ai"
              element={
                <ProtectedRoute>
                  <PermissionProvider>
                    <AIProtectedRoute />
                  </PermissionProvider>
                </ProtectedRoute>
              }
            />
            <Route
              path="/*"
              element={
                <ProtectedRoute>
                  <ProtectedApp />
                </ProtectedRoute>
              }
            />
          </Routes>
        </Suspense>
      </BrowserRouter>
    </AuthProvider>
  );
};

export default App;

import React, { lazy, Suspense } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import LoadingSkeleton from '../components/LoadingSkeleton';

// 立即加载的核心页面
import AppLayout from '../components/Layout/AppLayout';
import Dashboard from '../pages/Dashboard/Dashboard';

// 懒加载的页面组件
const ServiceListPage = lazy(() => import('../pages/Services/ServiceListPage'));
const ServiceDetailPage = lazy(() => import('../pages/Services/ServiceDetailPage'));
const ServiceProvisionPage = lazy(() => import('../pages/Services/ServiceProvisionPage'));

const DeploymentListPage = lazy(() => import('../pages/Deployment/DeploymentListPage'));
const DeploymentCreatePage = lazy(() => import('../pages/Deployment/DeploymentCreatePage'));
const DeploymentDetailPage = lazy(() => import('../pages/Deployment/DeploymentDetailPage'));

const HostListPage = lazy(() => import('../pages/Hosts/HostListPage'));
const HostDetailPage = lazy(() => import('../pages/Hosts/HostDetailPage'));

const ConfigPage = lazy(() => import('../pages/Config/ConfigPage'));
const TasksPage = lazy(() => import('../pages/Tasks/TasksPage'));
const SettingsPage = lazy(() => import('../pages/Settings/SettingsPage'));

// 懒加载包装器
const LazyPage: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <Suspense fallback={<LoadingSkeleton type="detail" count={1} />}>{children}</Suspense>
);

/**
 * 应用路由配置
 *
 * 特性:
 * - 路由级代码分割
 * - 懒加载非核心页面
 * - 加载时显示骨架屏
 */
const AppRoutes: React.FC = () => {
  return (
    <Routes>
      <Route path="/" element={<AppLayout><Dashboard /></AppLayout>} />

      {/* 服务管理 */}
      <Route
        path="/services"
        element={
          <AppLayout>
            <LazyPage>
              <ServiceListPage />
            </LazyPage>
          </AppLayout>
        }
      />
      <Route
        path="/services/:id"
        element={
          <AppLayout>
            <LazyPage>
              <ServiceDetailPage />
            </LazyPage>
          </AppLayout>
        }
      />
      <Route
        path="/services/provision"
        element={
          <AppLayout>
            <LazyPage>
              <ServiceProvisionPage />
            </LazyPage>
          </AppLayout>
        }
      />

      {/* 部署管理 */}
      <Route
        path="/deployment"
        element={
          <AppLayout>
            <LazyPage>
              <DeploymentListPage />
            </LazyPage>
          </AppLayout>
        }
      />
      <Route
        path="/deployment/create"
        element={
          <AppLayout>
            <LazyPage>
              <DeploymentCreatePage />
            </LazyPage>
          </AppLayout>
        }
      />
      <Route
        path="/deployment/:id"
        element={
          <AppLayout>
            <LazyPage>
              <DeploymentDetailPage />
            </LazyPage>
          </AppLayout>
        }
      />

      {/* 主机管理 */}
      <Route
        path="/hosts"
        element={
          <AppLayout>
            <LazyPage>
              <HostListPage />
            </LazyPage>
          </AppLayout>
        }
      />
      <Route
        path="/hosts/detail/:id"
        element={
          <AppLayout>
            <LazyPage>
              <HostDetailPage />
            </LazyPage>
          </AppLayout>
        }
      />

      {/* 其他页面 */}
      <Route
        path="/config"
        element={
          <AppLayout>
            <LazyPage>
              <ConfigPage />
            </LazyPage>
          </AppLayout>
        }
      />
      <Route
        path="/tasks"
        element={
          <AppLayout>
            <LazyPage>
              <TasksPage />
            </LazyPage>
          </AppLayout>
        }
      />
      <Route
        path="/settings"
        element={
          <AppLayout>
            <LazyPage>
              <SettingsPage />
            </LazyPage>
          </AppLayout>
        }
      />

      {/* 默认重定向 */}
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
};

export default AppRoutes;

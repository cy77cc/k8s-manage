import React from 'react';
import { BrowserRouter, Routes, Route, Navigate, useLocation } from 'react-router-dom';
import AppLayout from './components/Layout/AppLayout';
import { PermissionProvider, Authorized } from './components/RBAC';
import { AuthProvider, useAuth } from './components/Auth/AuthContext';
import Dashboard from './pages/Dashboard/Dashboard';
import HostListPage from './pages/Hosts/HostListPage';
import HostDetailPage from './pages/Hosts/HostDetailPage';
import HostOnboardingPage from './pages/Hosts/HostOnboardingPage';
import HostTerminalPage from './pages/Hosts/HostTerminalPage';
import HostKeysPage from './pages/Hosts/HostKeysPage';
import HostCloudImportPage from './pages/Hosts/HostCloudImportPage';
import HostVirtualizationPage from './pages/Hosts/HostVirtualizationPage';
import ConfigPage from './pages/Config/ConfigPage';
import TasksPage from './pages/Tasks/TasksPage';
import K8sPage from './pages/K8s/K8sPage';
import MonitorPage from './pages/Monitor/MonitorPage';
import ToolsPage from './pages/Tools/ToolsPage';
import SettingsPage from './pages/Settings/SettingsPage';
import UsersPage from './pages/Settings/UsersPage';
import RolesPage from './pages/Settings/RolesPage';
import PermissionsPage from './pages/Settings/PermissionsPage';
import ServiceListPage from './pages/Services/ServiceListPage';
import ServiceDetailPage from './pages/Services/ServiceDetailPage';
import ServiceProvisionPage from './pages/Services/ServiceProvisionPage';
import CMDBPage from './pages/CMDB/CMDBPage';
import AutomationPage from './pages/Automation/AutomationPage';
import CICDPage from './pages/CICD/CICDPage';
import JobListPage from './pages/Jobs/JobListPage';
import JobCreationPage from './pages/Jobs/JobCreationPage';
import ExecutionHistoryPage from './pages/Jobs/ExecutionHistoryPage';
import JobCalendarPage from './pages/Jobs/JobCalendarPage';
import ConfigAppsPage from './pages/ConfigCenter/ConfigAppsPage';
import ConfigListPage from './pages/ConfigCenter/ConfigListPage';
import ConfigDiffPage from './pages/ConfigCenter/ConfigDiffPage';
import ConfigMultiEnvPage from './pages/ConfigCenter/ConfigMultiEnvPage';
import AuditLogsPage from './pages/ConfigCenter/AuditLogsPage';
import LoginPage from './pages/Auth/LoginPage';
import RegisterPage from './pages/Auth/RegisterPage';

const ProtectedRoute: React.FC<{ children: React.ReactElement }> = ({ children }) => {
  const { isAuthenticated, loading } = useAuth();
  const location = useLocation();

  if (loading) {
    return <div className="min-h-screen flex items-center justify-center">加载中...</div>;
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />;
  }

  return children;
};

const ProtectedApp: React.FC = () => {
  const withAuth = (resource: string, action: string, element: React.ReactElement) => (
    <Authorized resource={resource} action={action}>
      {element}
    </Authorized>
  );

  return (
    <PermissionProvider>
      <AppLayout>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/hosts" element={withAuth('host', 'read', <HostListPage />)} />
          <Route path="/hosts/onboarding" element={withAuth('host', 'write', <HostOnboardingPage />)} />
          <Route path="/hosts/keys" element={withAuth('host', 'write', <HostKeysPage />)} />
          <Route path="/hosts/cloud-import" element={withAuth('host', 'write', <HostCloudImportPage />)} />
          <Route path="/hosts/virtualization" element={withAuth('host', 'write', <HostVirtualizationPage />)} />
          <Route path="/hosts/detail/:id" element={withAuth('host', 'read', <HostDetailPage />)} />
          <Route path="/hosts/terminal/:id" element={withAuth('host', 'write', <HostTerminalPage />)} />
          <Route path="/config" element={withAuth('config', 'read', <ConfigPage />)} />
          <Route path="/config/:id" element={withAuth('config', 'write', <ConfigPage />)} />
          <Route path="/configcenter/apps" element={withAuth('config', 'read', <ConfigAppsPage />)} />
          <Route path="/configcenter/list" element={withAuth('config', 'read', <ConfigListPage />)} />
          <Route path="/configcenter/diff" element={withAuth('config', 'read', <ConfigDiffPage />)} />
          <Route path="/configcenter/multienv" element={withAuth('config', 'write', <ConfigMultiEnvPage />)} />
          <Route path="/configcenter/logs" element={withAuth('config', 'read', <AuditLogsPage />)} />
          <Route path="/tasks" element={withAuth('task', 'read', <TasksPage />)} />
          <Route path="/tasks/create" element={withAuth('task', 'write', <TasksPage />)} />
          <Route path="/tasks/:id" element={withAuth('task', 'read', <TasksPage />)} />
          <Route path="/jobs" element={withAuth('task', 'read', <JobListPage />)} />
          <Route path="/jobs/create" element={withAuth('task', 'write', <JobCreationPage />)} />
          <Route path="/jobs/:id/edit" element={withAuth('task', 'write', <JobCreationPage />)} />
          <Route path="/jobs/:jobId/history" element={withAuth('task', 'read', <ExecutionHistoryPage />)} />
          <Route path="/jobs/calendar" element={withAuth('task', 'read', <JobCalendarPage />)} />
          <Route path="/k8s" element={withAuth('kubernetes', 'read', <K8sPage />)} />
          <Route path="/k8s/:cluster" element={withAuth('kubernetes', 'read', <K8sPage />)} />
          <Route path="/monitor" element={withAuth('monitoring', 'read', <MonitorPage />)} />
          <Route path="/monitor/dashboard" element={withAuth('monitoring', 'read', <MonitorPage />)} />
          <Route path="/monitor/alerts" element={withAuth('monitoring', 'read', <MonitorPage />)} />
          <Route path="/monitor/rules" element={withAuth('monitoring', 'read', <MonitorPage />)} />
          <Route path="/tools" element={<ToolsPage />} />
          <Route path="/tools/nightingale" element={<ToolsPage />} />
          <Route path="/tools/jenkins" element={<ToolsPage />} />
          <Route path="/tools/jumpserver" element={<ToolsPage />} />
          <Route path="/tools/kuboard" element={<ToolsPage />} />
          <Route path="/tools/cmdb" element={<ToolsPage />} />
          <Route path="/tools/archery" element={<ToolsPage />} />
          <Route path="/settings" element={<SettingsPage />} />
          <Route path="/settings/users" element={withAuth('rbac', 'read', <UsersPage />)} />
          <Route path="/settings/roles" element={withAuth('rbac', 'read', <RolesPage />)} />
          <Route path="/settings/permissions" element={withAuth('rbac', 'read', <PermissionsPage />)} />
          <Route path="/services" element={<ServiceListPage />} />
          <Route path="/services/provision" element={<ServiceProvisionPage />} />
          <Route path="/services/:id" element={<ServiceDetailPage />} />
          <Route path="/cmdb/assets" element={withAuth('cmdb', 'read', <CMDBPage />)} />
          <Route path="/automation" element={withAuth('automation', 'read', <AutomationPage />)} />
          <Route path="/cicd" element={withAuth('cicd', 'read', <CICDPage />)} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </AppLayout>
    </PermissionProvider>
  );
};

const App: React.FC = () => {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />
          <Route
            path="/*"
            element={
              <ProtectedRoute>
                <ProtectedApp />
              </ProtectedRoute>
            }
          />
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  );
};

export default App;

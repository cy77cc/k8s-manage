import React, { useState } from 'react';
import { Layout, Menu, Breadcrumb, Avatar, Dropdown, Badge, Input, Tooltip, Button } from 'antd';
import type { MenuProps } from 'antd';
import {
  DashboardOutlined,
  DesktopOutlined,
  SettingOutlined,
  AlertOutlined,
  CloudOutlined,
  ClockCircleOutlined,
  ToolOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  BellOutlined,
  SearchOutlined,
  UserOutlined,
  LogoutOutlined,
  QuestionCircleOutlined,
  CloudServerOutlined,
  AppstoreOutlined,
  DeploymentUnitOutlined,
  SafetyCertificateOutlined,
} from '@ant-design/icons';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../Auth/AuthContext';
import ProjectSwitcher from '../Project/ProjectSwitcher';
import GlobalAIAssistant from '../AI/GlobalAIAssistant';
import { useI18n } from '../../i18n';
import { usePermission } from '../RBAC';
import { brand, getAppTitle } from '../../brand/brand';
import BrandLogo from '../../brand/BrandLogo';

const { Header, Sider, Content } = Layout;
type MenuItem = Required<MenuProps>['items'][number];

interface AppLayoutProps {
  children: React.ReactNode;
}

const PARENT_MENU_MAP: Record<string, string> = {
  '/': '/workspace',
  '/hosts': '/workspace',
  '/services': '/workspace',
  '/cmdb/assets': '/workspace',
  '/automation': '/workspace',
  '/config': '/delivery',
  '/tasks': '/delivery',
  '/deployment': '/delivery',
  '/cicd': '/delivery',
  '/ai': '/delivery',
  '/monitor': '/observe',
  '/tools': '/observe',
  '/governance/users': '/governance',
  '/governance/roles': '/governance',
  '/governance/permissions': '/governance',
};

const AppLayout: React.FC<AppLayoutProps> = ({ children }) => {
  const [collapsed, setCollapsed] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const { logout } = useAuth();
  const { hasPermission } = usePermission();
  const { t, lang, setLang } = useI18n();
  const governanceMenuEnabled = import.meta.env.VITE_FEATURE_GOVERNANCE_MENU !== 'false';
  const canReadGovernance = hasPermission('rbac', 'read');

  const workspaceMenuItems: MenuItem[] = [
    { key: '/', icon: <DashboardOutlined />, label: t('menu.dashboard') },
    { key: '/hosts', icon: <DesktopOutlined />, label: t('menu.hosts') },
    { key: '/services', icon: <CloudServerOutlined />, label: t('menu.services') },
    { key: '/cmdb/assets', icon: <CloudServerOutlined />, label: t('menu.cmdb') },
    { key: '/automation', icon: <ToolOutlined />, label: t('menu.automation') },
  ];

  const deliveryMenuItems: MenuItem[] = [
    { key: '/config', icon: <SettingOutlined />, label: t('menu.config') },
    { key: '/tasks', icon: <ClockCircleOutlined />, label: t('menu.tasks') },
    { key: '/deployment', icon: <CloudOutlined />, label: '部署管理' },
    { key: '/cicd', icon: <DeploymentUnitOutlined />, label: 'CI/CD' },
    { key: '/ai', icon: <ToolOutlined />, label: 'AI 命令中心' },
  ];

  const observabilityMenuItems: MenuItem[] = [
    { key: '/monitor', icon: <AlertOutlined />, label: t('menu.monitor') },
    { key: '/tools', icon: <ToolOutlined />, label: t('menu.tools') },
  ];

  const governanceChildren: MenuItem[] =
    governanceMenuEnabled && canReadGovernance
      ? [
          { key: '/governance/users', label: '用户管理' },
          { key: '/governance/roles', label: '角色管理' },
          { key: '/governance/permissions', label: '权限列表' },
        ]
      : [];

  const menuItems: MenuItem[] = [
    {
      key: '/workspace',
      icon: <AppstoreOutlined />,
      label: '资源与资产',
      children: workspaceMenuItems,
    },
    {
      key: '/delivery',
      icon: <DeploymentUnitOutlined />,
      label: '交付与变更',
      children: deliveryMenuItems,
    },
    {
      key: '/observe',
      icon: <AlertOutlined />,
      label: '可观测与工具',
      children: observabilityMenuItems,
    },
    ...(governanceChildren.length > 0
      ? [
          {
            key: '/governance',
            icon: <SafetyCertificateOutlined />,
            label: '访问治理',
            children: governanceChildren,
          },
        ]
      : []),
    ...(governanceMenuEnabled ? [] : [{ key: '/settings', icon: <SettingOutlined />, label: '系统设置' }]),
  ];

  const activeMenuKey = React.useMemo(() => {
    if (location.pathname.startsWith('/jobs')) return '/tasks';
    if (location.pathname.startsWith('/configcenter')) return '/config';
    if (location.pathname.startsWith('/k8s')) return '/deployment';
    if (location.pathname.startsWith('/governance/users')) return '/governance/users';
    if (location.pathname.startsWith('/governance/roles')) return '/governance/roles';
    if (location.pathname.startsWith('/governance/permissions')) return '/governance/permissions';
    return location.pathname;
  }, [location.pathname]);

  const openMenuKeys = React.useMemo(() => {
    const parent = PARENT_MENU_MAP[activeMenuKey];
    return parent ? [parent] : [];
  }, [activeMenuKey]);

  const userMenuItems = [
    { key: 'profile', icon: <UserOutlined />, label: '个人中心' },
    { key: 'settings', icon: <SettingOutlined />, label: '系统设置' },
    { type: 'divider' as const },
    { key: 'logout', icon: <LogoutOutlined />, label: '退出登录' },
  ];

  const getLabelByKey = React.useCallback((key: string): string | null => {
    for (const item of menuItems) {
      if (!item) continue;
      if ((item as any).key === key && (item as any).label) {
        return String((item as any).label);
      }
      const children = (item as any).children;
      if (Array.isArray(children)) {
        const child = children.find((entry: any) => entry && entry.key === key);
        if (child?.label) return String(child.label);
      }
    }
    return null;
  }, [menuItems]);

  const breadcrumbItems = React.useMemo(() => {
    const paths = location.pathname.split('/').filter(Boolean);
    const items = [{ title: '首页', path: '/' }];
    let currentPath = '';

    paths.forEach((path) => {
      currentPath += `/${path}`;
      const label = getLabelByKey(currentPath);
      if (label) {
        items.push({ title: label, path: currentPath });
      }
    });

    return items;
  }, [getLabelByKey, location.pathname]);

  const healthScore = React.useMemo(() => {
    if (location.pathname.startsWith('/monitor')) return 91;
    if (location.pathname.startsWith('/deployment')) return 94;
    return 97;
  }, [location.pathname]);

  React.useEffect(() => {
    const pageTitle = breadcrumbItems[breadcrumbItems.length - 1]?.title;
    document.title = getAppTitle(pageTitle === '首页' ? '控制台' : pageTitle);
  }, [breadcrumbItems]);

  return (
    <Layout className="min-h-screen">
      <Sider
        trigger={null}
        collapsible
        collapsed={collapsed}
        width={272}
        className="fixed left-0 top-0 bottom-0 z-50"
        style={{
          background: 'linear-gradient(180deg, var(--color-sider-start) 0%, var(--color-sider-end) 100%)',
          borderRight: '1px solid rgba(148, 163, 184, 0.18)',
          boxShadow: 'var(--shadow-soft)',
        }}
      >
        <div className="h-16 flex items-center px-4 border-b border-slate-600/40">
          <div className="flex items-center gap-3 w-full">
            <BrandLogo variant="simplified" width={34} height={34} />
            {!collapsed && (
              <div className="leading-tight">
                <div className="text-slate-50 font-semibold tracking-wide">{brand.canonicalName}</div>
                <div className="text-[11px] text-slate-300">{brand.tagline}</div>
              </div>
            )}
          </div>
        </div>

        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[activeMenuKey]}
          defaultOpenKeys={openMenuKeys}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
          style={{ marginTop: 8, border: 'none' }}
        />

        <div className="absolute bottom-4 left-0 right-0 px-4">
          <Button
            type="text"
            icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={() => setCollapsed(!collapsed)}
            className="w-full text-gray-400 hover:text-white"
            style={{ color: '#d4deef' }}
          />
        </div>
      </Sider>

      <Layout style={{ marginLeft: collapsed ? 80 : 272, transition: 'margin-left 0.2s' }}>
        <Header
          className="h-[78px] px-5 py-2 flex flex-col justify-center gap-2"
          style={{
            background: 'var(--color-bg-surface)',
            borderBottom: '1px solid var(--color-border)',
            position: 'sticky',
            top: 0,
            zIndex: 40,
          }}
        >
          <div className="flex items-center justify-between text-xs text-slate-300">
            <div className="flex items-center gap-3">
              <span className="inline-flex items-center gap-2 rounded-full border border-slate-700/60 bg-slate-900/70 px-3 py-1">
                <span className={`h-2 w-2 rounded-full ${healthScore >= 95 ? 'bg-emerald-400' : healthScore >= 90 ? 'bg-amber-400' : 'bg-red-400'}`} />
                系统健康度 {healthScore}%
              </span>
              <span className="text-slate-400">自动刷新: 30s</span>
            </div>
            <span className="text-slate-500 hidden md:block">面向 SRE 的监控优先控制台</span>
          </div>

          <div className="flex items-center justify-between gap-4">
            <Breadcrumb
              items={breadcrumbItems.map((item, index) => ({
                title:
                  index === breadcrumbItems.length - 1 ? item.title : <a onClick={() => navigate(item.path!)}>{item.title}</a>,
              }))}
              separator="/"
              style={{ color: 'var(--color-text-secondary)' }}
            />
            <div className="flex items-center gap-3">
              <ProjectSwitcher />
              <select
                value={lang}
                onChange={(e) => setLang(e.target.value as 'zh-CN' | 'en-US')}
                style={{ border: '1px solid var(--color-border)', borderRadius: 6, height: 32, padding: '0 8px' }}
              >
                <option value="zh-CN">{t('lang.zh')}</option>
                <option value="en-US">{t('lang.en')}</option>
              </select>
              <Input
                placeholder="搜索主机、配置、任务..."
                prefix={<SearchOutlined style={{ color: '#8d98a8' }} />}
                className="header-search-input hidden lg:inline-flex"
                style={{ width: 280, background: 'var(--color-bg-surface)', border: '1px solid var(--color-border)', color: 'var(--color-text-primary)' }}
              />

              <Tooltip title="帮助文档">
                <Button type="text" icon={<QuestionCircleOutlined />} style={{ color: 'var(--color-text-secondary)' }} />
              </Tooltip>
              <GlobalAIAssistant inlineTrigger />

              <Badge count={3} size="small">
                <Button type="text" icon={<BellOutlined />} style={{ color: 'var(--color-text-secondary)' }} />
              </Badge>

              <Dropdown
                menu={{
                  items: userMenuItems,
                  onClick: ({ key }) => {
                    if (key === 'logout') {
                      logout();
                      navigate('/login', { replace: true });
                    }
                    if (key === 'settings') {
                      navigate('/settings');
                    }
                  },
                }}
                placement="bottomRight"
              >
                <Avatar style={{ backgroundColor: 'var(--color-brand-500)', cursor: 'pointer' }} icon={<UserOutlined />} />
              </Dropdown>
            </div>
          </div>
        </Header>

        <Content className="p-4 md:p-6 min-h-[calc(100vh-78px)]">{children}</Content>
      </Layout>
    </Layout>
  );
};

export default AppLayout;

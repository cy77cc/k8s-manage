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
} from '@ant-design/icons';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../Auth/AuthContext';
import ProjectSwitcher from '../Project/ProjectSwitcher';
import GlobalAIAssistant from '../AI/GlobalAIAssistant';
import { useI18n } from '../../i18n';
import { usePermission } from '../RBAC';

const { Header, Sider, Content } = Layout;
type MenuItem = Required<MenuProps>['items'][number];

interface AppLayoutProps {
  children: React.ReactNode;
}

const AppLayout: React.FC<AppLayoutProps> = ({ children }) => {
  const [collapsed, setCollapsed] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const { logout } = useAuth();
  const { hasPermission } = usePermission();
  const { t, lang, setLang } = useI18n();
  const governanceMenuEnabled = import.meta.env.VITE_FEATURE_GOVERNANCE_MENU !== 'false';
  const canReadGovernance = hasPermission('rbac', 'read');

  const baseMenuItems: MenuItem[] = [
    { key: '/', icon: <DashboardOutlined />, label: t('menu.dashboard') },
    { key: '/hosts', icon: <DesktopOutlined />, label: t('menu.hosts') },
    { key: '/services', icon: <CloudServerOutlined />, label: t('menu.services') },
    { key: '/cmdb/assets', icon: <CloudServerOutlined />, label: t('menu.cmdb') },
    { key: '/automation', icon: <ToolOutlined />, label: t('menu.automation') },
    { key: '/cicd', icon: <ToolOutlined />, label: 'CI/CD' },
    { key: '/ai', icon: <ToolOutlined />, label: 'AI命令中心' },
    { key: '/config', icon: <SettingOutlined />, label: t('menu.config') },
    { key: '/tasks', icon: <ClockCircleOutlined />, label: t('menu.tasks') },
    { key: '/deployment', icon: <CloudOutlined />, label: '部署管理' },
    { key: '/monitor', icon: <AlertOutlined />, label: t('menu.monitor') },
    { key: '/tools', icon: <ToolOutlined />, label: t('menu.tools') },
    ...(governanceMenuEnabled ? [] : [{ key: '/settings', icon: <SettingOutlined />, label: '系统设置' }]),
  ];
  const governanceMenuItems: MenuItem[] =
    governanceMenuEnabled && canReadGovernance
      ? [
          {
            key: '/governance',
            icon: <UserOutlined />,
            label: '访问治理',
            children: [
              { key: '/governance/users', label: '用户管理' },
              { key: '/governance/roles', label: '角色管理' },
              { key: '/governance/permissions', label: '权限列表' },
            ],
          },
        ]
      : [];

  const menuItems = [...baseMenuItems, ...governanceMenuItems];

  const activeMenuKey = React.useMemo(() => {
    if (location.pathname.startsWith('/jobs')) return '/tasks';
    if (location.pathname.startsWith('/configcenter')) return '/config';
    if (location.pathname.startsWith('/k8s')) return '/deployment';
    if (location.pathname.startsWith('/governance/users')) return '/governance/users';
    if (location.pathname.startsWith('/governance/roles')) return '/governance/roles';
    if (location.pathname.startsWith('/governance/permissions')) return '/governance/permissions';
    return location.pathname;
  }, [location.pathname]);

  const userMenuItems = [
    { key: 'profile', icon: <UserOutlined />, label: '个人中心' },
    { key: 'settings', icon: <SettingOutlined />, label: '系统设置' },
    { type: 'divider' as const },
    { key: 'logout', icon: <LogoutOutlined />, label: '退出登录' },
  ];

  const getBreadcrumbItems = () => {
    const paths = location.pathname.split('/').filter(Boolean);
    const items = [{ title: '首页', path: '/' }];
    let currentPath = '';
    
    paths.forEach((path) => {
      currentPath += `/${path}`;
      const menuItem = menuItems.find((item) => item && (item as any).key === currentPath) as any;
      if (menuItem?.label) {
        items.push({ title: String(menuItem.label), path: currentPath });
      }
      const governanceRoot = menuItems.find((item) => item && (item as any).key === '/governance') as any;
      if (governanceRoot?.children && Array.isArray(governanceRoot.children)) {
        const child = governanceRoot.children.find((sub: any) => sub && sub.key === currentPath);
        if (child?.label) {
          items.push({ title: String(child.label), path: currentPath });
        }
      }
    });
    
    return items;
  };

  return (
    <Layout className="min-h-screen">
      <Sider 
        trigger={null} 
        collapsible 
        collapsed={collapsed}
        width={260}
        className="fixed left-0 top-0 bottom-0 z-50"
        style={{ background: 'linear-gradient(180deg, var(--color-sider-start) 0%, var(--color-sider-end) 100%)', borderRight: '1px solid rgba(255,255,255,0.08)' }}
      >
        <div className="h-16 flex items-center justify-center border-b border-gray-700">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-500 to-cyan-400 flex items-center justify-center">
              <CloudOutlined className="text-white text-lg" />
            </div>
            {!collapsed && (
              <span className="text-white font-bold text-lg tracking-wider">OpsPilot</span>
            )}
          </div>
        </div>
        
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[activeMenuKey]}
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
      
      <Layout style={{ marginLeft: collapsed ? 80 : 260, transition: 'margin-left 0.2s' }}>
        <Header 
          className="h-16 px-6 flex items-center justify-between"
          style={{ 
            background: '#ffffff', 
            borderBottom: '1px solid var(--color-border)',
            position: 'sticky',
            top: 0,
            zIndex: 40,
          }}
        >
          <div className="flex items-center gap-4">
            <Breadcrumb 
              items={getBreadcrumbItems().map((item, index) => ({
                title: index === getBreadcrumbItems().length - 1 ? item.title : <a onClick={() => navigate(item.path!)}>{item.title}</a>,
              }))}
              separator="/"
              style={{ color: 'var(--color-text-secondary)' }}
            />
          </div>
          
          <div className="flex items-center gap-4">
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
              className="header-search-input"
              style={{ width: 280, background: '#ffffff', border: '1px solid var(--color-border)', color: '#1f2329' }}
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
              <Avatar 
                style={{ backgroundColor: 'var(--color-theme)', cursor: 'pointer' }}
                icon={<UserOutlined />}
              />
            </Dropdown>
          </div>
        </Header>
        
        <Content className="p-6 min-h-[calc(100vh-64px)]">
          {children}
        </Content>
      </Layout>
    </Layout>
  );
};

export default AppLayout;

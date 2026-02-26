import React from 'react';
import { Button, Drawer, Segmented, Space, Tabs, Typography } from 'antd';
import { MessageOutlined } from '@ant-design/icons';
import { useLocation } from 'react-router-dom';
import ChatInterface from './ChatInterface';
import CommandPanel from './CommandPanel';
import './ai-assistant.css';

const { Text } = Typography;
const MemoChatInterface = React.memo(ChatInterface);
const MemoCommandPanel = React.memo(CommandPanel);
const DRAWER_WIDTH_STORAGE_KEY = 'ai:assistant:drawer:width';
const MOBILE_BREAKPOINT = 768;
const DRAWER_MIN_WIDTH = 620;
const DRAWER_MAX_GUTTER = 24;

const clampDrawerWidth = (width: number, viewportWidth: number): number => {
  const max = Math.max(DRAWER_MIN_WIDTH, viewportWidth - DRAWER_MAX_GUTTER);
  return Math.min(max, Math.max(DRAWER_MIN_WIDTH, width));
};

interface GlobalAIAssistantProps {
  inlineTrigger?: boolean;
}

const sceneFromPath = (pathname: string): string => {
  const seg = pathname.split('/').filter(Boolean)[0];
  return seg ? `scene:${seg}` : 'scene:home';
};

const GlobalAIAssistant: React.FC<GlobalAIAssistantProps> = ({ inlineTrigger = false }) => {
  const [open, setOpen] = React.useState(false);
  const [scope, setScope] = React.useState<'scene' | 'global'>('scene');
  const [tabKey, setTabKey] = React.useState<'chat' | 'command'>('chat');
  const [viewportWidth, setViewportWidth] = React.useState(() => window.innerWidth);
  const [drawerWidth, setDrawerWidth] = React.useState(() => {
    const maxWidth = window.innerWidth - DRAWER_MAX_GUTTER;
    const fallback = Math.min(920, maxWidth);
    const saved = Number(localStorage.getItem(DRAWER_WIDTH_STORAGE_KEY));
    if (Number.isFinite(saved) && saved > 0) {
      return clampDrawerWidth(saved, window.innerWidth);
    }
    return clampDrawerWidth(fallback, window.innerWidth);
  });
  const location = useLocation();
  const isMobile = viewportWidth < MOBILE_BREAKPOINT;
  const pageScene = React.useMemo(() => sceneFromPath(location.pathname), [location.pathname]);
  const currentScene = scope === 'global' ? 'global' : pageScene;
  const resizingRef = React.useRef<{ startX: number; startWidth: number } | null>(null);
  const pendingWidthRef = React.useRef(drawerWidth);
  const rafRef = React.useRef<number | null>(null);

  React.useEffect(() => {
    const onResize = () => {
      const nextViewportWidth = window.innerWidth;
      setViewportWidth(nextViewportWidth);
      setDrawerWidth((prev) => clampDrawerWidth(prev, nextViewportWidth));
    };
    window.addEventListener('resize', onResize);
    return () => window.removeEventListener('resize', onResize);
  }, []);

  React.useEffect(() => {
    if (!isMobile) {
      localStorage.setItem(DRAWER_WIDTH_STORAGE_KEY, String(drawerWidth));
    }
  }, [drawerWidth, isMobile]);

  const handleResizeStart = (event: React.MouseEvent<HTMLDivElement>) => {
    if (isMobile) return;
    event.preventDefault();
    resizingRef.current = { startX: event.clientX, startWidth: drawerWidth };

    const onMouseMove = (moveEvent: MouseEvent) => {
      const state = resizingRef.current;
      if (!state) return;
      const delta = state.startX - moveEvent.clientX;
      pendingWidthRef.current = clampDrawerWidth(state.startWidth + delta, window.innerWidth);
      if (rafRef.current !== null) {
        return;
      }
      rafRef.current = window.requestAnimationFrame(() => {
        rafRef.current = null;
        setDrawerWidth(pendingWidthRef.current);
      });
    };

    const onMouseUp = () => {
      if (rafRef.current !== null) {
        window.cancelAnimationFrame(rafRef.current);
        rafRef.current = null;
      }
      setDrawerWidth(pendingWidthRef.current);
      localStorage.setItem(DRAWER_WIDTH_STORAGE_KEY, String(pendingWidthRef.current));
      resizingRef.current = null;
      document.body.style.userSelect = '';
      window.removeEventListener('mousemove', onMouseMove);
      window.removeEventListener('mouseup', onMouseUp);
    };

    document.body.style.userSelect = 'none';
    window.addEventListener('mousemove', onMouseMove);
    window.addEventListener('mouseup', onMouseUp);
  };

  React.useEffect(() => {
    return () => {
      if (rafRef.current !== null) {
        window.cancelAnimationFrame(rafRef.current);
      }
    };
  }, []);

  const tabItems = React.useMemo(() => ([
    {
      key: 'chat',
      label: '对话',
      children: (
        <div className="ai-assistant-tabpane-wrap">
          <MemoChatInterface className="ai-chat-interface" scene={currentScene} />
        </div>
      ),
    },
    {
      key: 'command',
      label: '命令中心',
      children: (
        <div className="ai-assistant-tabpane-wrap">
          <MemoCommandPanel scene={currentScene} />
        </div>
      ),
    },
  ]), [currentScene]);

  return (
    <>
      <Button
        type={inlineTrigger ? 'default' : 'primary'}
        shape={inlineTrigger ? 'default' : 'circle'}
        size={inlineTrigger ? 'middle' : 'large'}
        icon={<MessageOutlined />}
        style={inlineTrigger ? undefined : { position: 'fixed', right: 28, bottom: 28, zIndex: 1000, boxShadow: '0 10px 30px rgba(0,0,0,0.2)' }}
        onClick={() => setOpen(true)}
      >
        {inlineTrigger ? 'AI助手' : null}
      </Button>
      <Drawer
        rootClassName="ai-assistant-drawer"
        title={<Space><MessageOutlined /><Text>AI 助手</Text></Space>}
        open={open}
        onClose={() => setOpen(false)}
        width={isMobile ? '100vw' : drawerWidth}
        styles={{ body: { position: 'relative', padding: 12, height: 'calc(100vh - 56px)', overflow: 'hidden', display: 'flex', flexDirection: 'column' } }}
        extra={(
          <Segmented
            size="small"
            value={scope}
            options={[
              { label: '当前场景', value: 'scene' },
              { label: '全局', value: 'global' },
            ]}
            onChange={(v) => setScope(v as 'scene' | 'global')}
          />
        )}
      >
        {!isMobile ? (
          <div
            className="ai-assistant-resize-handle"
            onMouseDown={handleResizeStart}
            role="separator"
            aria-orientation="vertical"
            aria-label="调整 AI 助手宽度"
          />
        ) : null}
        <div className="ai-assistant-layout">
          <div className="ai-assistant-hero">
            <Text className="ai-assistant-hero-title">智能运维助手</Text>
            <Text type="secondary" className="ai-assistant-hero-subtitle">
              对话诊断、命令预览、确认执行与历史回放
            </Text>
          </div>
          <div className="ai-assistant-chat-wrap">
            <Tabs
              className="ai-assistant-tabs"
              activeKey={tabKey}
              onChange={(v) => setTabKey(v as 'chat' | 'command')}
              items={tabItems}
            />
          </div>
        </div>
      </Drawer>
    </>
  );
};

export default GlobalAIAssistant;

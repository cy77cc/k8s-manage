import React from 'react';
import { Button, Drawer, Segmented, Space, Typography } from 'antd';
import { MessageOutlined } from '@ant-design/icons';
import { useLocation } from 'react-router-dom';
import ChatInterface from './ChatInterface';
import RecommendationPanel from './RecommendationPanel';

const { Text } = Typography;

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
  const [recRefreshSignal, setRecRefreshSignal] = React.useState(0);
  const location = useLocation();
  const drawerWidth = React.useMemo(() => (window.innerWidth < 768 ? '100vw' : 760), []);
  const pageScene = React.useMemo(() => sceneFromPath(location.pathname), [location.pathname]);
  const currentScene = scope === 'global' ? 'global' : pageScene;

  const context = React.useMemo(() => ({
    page: location.pathname,
    scene: currentScene,
    projectId: localStorage.getItem('projectId'),
  }), [location.pathname, currentScene]);

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
        title={<Space><MessageOutlined /><Text>AI 助手</Text></Space>}
        open={open}
        onClose={() => setOpen(false)}
        width={drawerWidth}
        styles={{ body: { padding: 12, height: 'calc(100vh - 56px)', overflow: 'hidden' } }}
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
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12, height: '100%' }}>
          <div style={{ maxHeight: 220, overflow: 'auto' }}>
            <RecommendationPanel type="suggestion" context={context} refreshSignal={recRefreshSignal} />
          </div>
          <div style={{ flex: 1, minHeight: 0, overflow: 'hidden' }}>
            <ChatInterface
              scene={currentScene}
              onSessionCreate={() => setRecRefreshSignal((v) => v + 1)}
              onSessionUpdate={() => setRecRefreshSignal((v) => v + 1)}
            />
          </div>
        </div>
      </Drawer>
    </>
  );
};

export default GlobalAIAssistant;

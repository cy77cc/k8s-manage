import React from 'react';
import { Button, Drawer, Segmented, Space, Typography } from 'antd';
import { MessageOutlined } from '@ant-design/icons';
import { useLocation } from 'react-router-dom';
import ChatInterface from './ChatInterface';
import RecommendationPanel from './RecommendationPanel';
import './ai-assistant.css';

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
  const [recommendationLoading, setRecommendationLoading] = React.useState(false);
  const [isRefreshingRecommendations, setIsRefreshingRecommendations] = React.useState(false);
  const recLoadedRef = React.useRef(false);
  const location = useLocation();
  const drawerWidth = React.useMemo(() => {
    if (window.innerWidth < 768) return '100vw';
    return Math.min(920, window.innerWidth - 32);
  }, []);
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
        rootClassName="ai-assistant-drawer"
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
        <div className="ai-assistant-layout">
          <div className="ai-assistant-recommendation-wrap">
            {recommendationLoading ? (
              <div className="ai-recommendation-loading-banner">
                <div className="ai-recommendation-loading-strip" />
                <Text type="secondary" className="ai-recommendation-loading-text">
                  {isRefreshingRecommendations ? '建议更新中...' : '建议加载中...'}
                </Text>
              </div>
            ) : null}
            <RecommendationPanel
              type="suggestion"
              context={context}
              refreshSignal={recRefreshSignal}
              className="ai-recommendation-panel"
              onLoadingChange={(loading) => {
                setRecommendationLoading(loading);
                if (loading) {
                  setIsRefreshingRecommendations(recLoadedRef.current);
                } else {
                  recLoadedRef.current = true;
                  setIsRefreshingRecommendations(false);
                }
              }}
            />
          </div>
          <div className="ai-assistant-chat-wrap">
            <ChatInterface
              className="ai-chat-interface"
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

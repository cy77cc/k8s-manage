import React from 'react';
import { Button, Drawer, Segmented, Space, Tabs, Typography } from 'antd';
import { MessageOutlined } from '@ant-design/icons';
import { useLocation } from 'react-router-dom';
import ChatInterface from './ChatInterface';
import CommandPanel from './CommandPanel';
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
  const [tabKey, setTabKey] = React.useState<'chat' | 'command'>('chat');
  const location = useLocation();
  const drawerWidth = React.useMemo(() => {
    if (window.innerWidth < 768) return '100vw';
    return Math.min(920, window.innerWidth - 32);
  }, []);
  const pageScene = React.useMemo(() => sceneFromPath(location.pathname), [location.pathname]);
  const currentScene = scope === 'global' ? 'global' : pageScene;

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
        styles={{ body: { padding: 12, height: 'calc(100vh - 56px)', overflow: 'hidden', display: 'flex', flexDirection: 'column' } }}
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
              items={[
                {
                  key: 'chat',
                  label: '对话',
                  children: (
                    <div className="ai-assistant-tabpane-wrap">
                      <ChatInterface className="ai-chat-interface" scene={currentScene} />
                    </div>
                  ),
                },
                {
                  key: 'command',
                  label: '命令中心',
                  children: (
                    <div className="ai-assistant-tabpane-wrap">
                      <CommandPanel scene={currentScene} />
                    </div>
                  ),
                },
              ]}
            />
          </div>
        </div>
      </Drawer>
    </>
  );
};

export default GlobalAIAssistant;

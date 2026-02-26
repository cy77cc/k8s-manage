import React from 'react';
import { Card, Space, Typography } from 'antd';
import CommandPanel from '../../components/AI/CommandPanel';

const { Title, Text } = Typography;

const AICommandCenterPage: React.FC = () => {
  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Card>
        <Title level={4} style={{ marginBottom: 0 }}>AI 命令中心</Title>
        <Text type="secondary">通过命令预览、确认执行、历史回放，完成跨域运维操作。</Text>
      </Card>
      <CommandPanel scene="scene:ai" />
    </Space>
  );
};

export default AICommandCenterPage;

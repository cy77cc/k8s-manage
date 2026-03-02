import React from 'react';
import { Card, Steps, Tag, Space } from 'antd';
import { CheckCircleOutlined, LoadingOutlined, CloseCircleOutlined, ClockCircleOutlined } from '@ant-design/icons';

const { Step } = Steps;

interface BootstrapPhase {
  name: string;
  status: 'pending' | 'running' | 'success' | 'failed';
  startTime?: string;
  endTime?: string;
}

interface BootstrapPhaseTrackerProps {
  phases: BootstrapPhase[];
  currentPhase?: string;
}

const BootstrapPhaseTracker: React.FC<BootstrapPhaseTrackerProps> = ({ phases, currentPhase }) => {
  const getPhaseIcon = (status: string) => {
    switch (status) {
      case 'success':
        return <CheckCircleOutlined style={{ color: '#52c41a' }} />;
      case 'running':
        return <LoadingOutlined style={{ color: '#1890ff' }} />;
      case 'failed':
        return <CloseCircleOutlined style={{ color: '#ff4d4f' }} />;
      case 'pending':
        return <ClockCircleOutlined style={{ color: '#d9d9d9' }} />;
      default:
        return null;
    }
  };

  const getPhaseStatus = (status: string): 'wait' | 'process' | 'finish' | 'error' => {
    switch (status) {
      case 'success':
        return 'finish';
      case 'running':
        return 'process';
      case 'failed':
        return 'error';
      default:
        return 'wait';
    }
  };

  const getCurrentStep = () => {
    const runningIndex = phases.findIndex((p) => p.status === 'running');
    if (runningIndex !== -1) return runningIndex;
    const lastSuccessIndex = phases.findLastIndex((p) => p.status === 'success');
    if (lastSuccessIndex !== -1) return lastSuccessIndex + 1;
    return 0;
  };

  return (
    <Card title="初始化阶段">
      <Steps current={getCurrentStep()} direction="vertical">
        {phases.map((phase, index) => (
          <Step
            key={index}
            title={
              <Space>
                <span>{phase.name}</span>
                {phase.status !== 'pending' && (
                  <Tag color={phase.status === 'success' ? 'success' : phase.status === 'failed' ? 'error' : 'processing'}>
                    {phase.status}
                  </Tag>
                )}
              </Space>
            }
            status={getPhaseStatus(phase.status)}
            icon={getPhaseIcon(phase.status)}
            description={
              phase.status !== 'pending' && (
                <div className="text-xs text-gray-500">
                  {phase.startTime && `开始: ${new Date(phase.startTime).toLocaleString()}`}
                  {phase.endTime && ` | 结束: ${new Date(phase.endTime).toLocaleString()}`}
                </div>
              )
            }
          />
        ))}
      </Steps>
    </Card>
  );
};

export default BootstrapPhaseTracker;

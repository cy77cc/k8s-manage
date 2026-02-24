import React, { useState } from 'react';
import { Card, Typography, Collapse, Space, Button, Input, Tag } from 'antd';
import { QuestionCircleOutlined, SearchOutlined, BookOutlined, MessageOutlined } from '@ant-design/icons';

const { Text, Paragraph } = Typography;
const { Panel } = Collapse;
const { Search } = Input;

// 帮助文档数据结构
export interface HelpDocument {
  id: string;
  title: string;
  content: string;
  category: string;
  tags: string[];
  difficulty: 'beginner' | 'intermediate' | 'advanced';
}

// 帮助中心属性
interface HelpCenterProps {
  onAskAI?: () => void;
  className?: string;
}

// 帮助中心组件
const HelpCenter: React.FC<HelpCenterProps> = ({ 
  onAskAI, 
  className 
}) => {
  const [searchValue, setSearchValue] = useState('');
  const [activeCategory, setActiveCategory] = useState<string | null>(null);

  // 帮助文档数据
  const helpDocuments: HelpDocument[] = [
    {
      id: 'doc-1',
      title: '如何添加新主机？',
      content: '1. 进入主机管理页面\n2. 点击「添加主机」按钮\n3. 填写主机信息，包括名称、IP地址、状态等\n4. 点击「确认」完成添加',
      category: '主机管理',
      tags: ['主机', '添加', '新手'],
      difficulty: 'beginner'
    },
    {
      id: 'doc-2',
      title: '如何创建定时任务？',
      content: '1. 进入任务管理页面\n2. 点击「创建任务」按钮\n3. 填写任务信息，包括名称、类型、调度时间等\n4. 点击「确认」完成创建',
      category: '任务管理',
      tags: ['任务', '定时', '新手'],
      difficulty: 'beginner'
    },
    {
      id: 'doc-3',
      title: '如何配置Kubernetes集群？',
      content: '1. 进入Kubernetes管理页面\n2. 点击「添加集群」按钮\n3. 填写集群信息，包括名称、版本、连接信息等\n4. 点击「确认」完成配置',
      category: 'Kubernetes',
      tags: ['Kubernetes', '集群', '中级'],
      difficulty: 'intermediate'
    },
    {
      id: 'doc-4',
      title: '如何设置告警规则？',
      content: '1. 进入监控告警页面\n2. 点击「添加告警规则」按钮\n3. 填写规则信息，包括名称、条件、严重程度等\n4. 点击「确认」完成设置',
      category: '监控告警',
      tags: ['监控', '告警', '中级'],
      difficulty: 'intermediate'
    },
    {
      id: 'doc-5',
      title: '如何使用AI助手？',
      content: '1. 在页面右侧找到AI助手图标\n2. 点击图标打开对话窗口\n3. 输入您的问题，点击发送\n4. 等待AI助手的回复',
      category: 'AI功能',
      tags: ['AI', '助手', '新手'],
      difficulty: 'beginner'
    },
    {
      id: 'doc-6',
      title: '如何管理用户权限？',
      content: '1. 进入权限管理页面\n2. 选择「用户管理」或「角色管理」\n3. 对用户或角色进行编辑，设置相应的权限\n4. 点击「保存」完成设置',
      category: '权限管理',
      tags: ['权限', '用户', '高级'],
      difficulty: 'advanced'
    }
  ];

  // 获取所有分类
  const categories = [...new Set(helpDocuments.map(doc => doc.category))];

  // 过滤文档
  const filteredDocuments = helpDocuments.filter(doc => {
    const matchesSearch = doc.title.toLowerCase().includes(searchValue.toLowerCase()) || 
                        doc.content.toLowerCase().includes(searchValue.toLowerCase()) ||
                        doc.tags.some(tag => tag.toLowerCase().includes(searchValue.toLowerCase()));
    const matchesCategory = !activeCategory || doc.category === activeCategory;
    return matchesSearch && matchesCategory;
  });

  // 获取难度对应的颜色
  const getDifficultyColor = (difficulty: HelpDocument['difficulty']) => {
    const colors = {
      beginner: 'green',
      intermediate: 'blue',
      advanced: 'orange'
    };
    return colors[difficulty];
  };

  // 获取难度对应的文本
  const getDifficultyText = (difficulty: HelpDocument['difficulty']) => {
    const texts = {
      beginner: '新手',
      intermediate: '中级',
      advanced: '高级'
    };
    return texts[difficulty];
  };

  return (
    <Card 
      title={
        <Space>
          <QuestionCircleOutlined />
          <Text strong>帮助中心</Text>
        </Space>
      }
      className={`help-center ${className || ''}`}
    >
      <Space direction="vertical" style={{ width: '100%', marginBottom: '16px' }}>
        {/* 搜索框 */}
        <Search
          placeholder="搜索帮助文档"
          allowClear
          enterButton={<SearchOutlined />}
          value={searchValue}
          onChange={(e) => setSearchValue(e.target.value)}
          style={{ width: '100%', maxWidth: 600, marginBottom: '16px' }}
        />

        {/* 分类标签 */}
        <Space wrap style={{ marginBottom: '16px' }}>
          <Tag 
            color={!activeCategory ? 'blue' : 'default'}
            onClick={() => setActiveCategory(null)}
          >
            全部
          </Tag>
          {categories.map(category => (
            <Tag
              key={category}
              color={activeCategory === category ? 'blue' : 'default'}
              onClick={() => setActiveCategory(category)}
            >
              {category}
            </Tag>
          ))}
        </Space>

        {/* 文档列表 */}
        <Collapse
          defaultActiveKey={[]}
          style={{ width: '100%' }}
        >
          {filteredDocuments.map(doc => (
            <Panel
              key={doc.id}
              header={
                <Space style={{ width: '100%', justifyContent: 'space-between' }}>
                  <Text strong>{doc.title}</Text>
                  <Space>
                    <Tag color={getDifficultyColor(doc.difficulty)}>
                      {getDifficultyText(doc.difficulty)}
                    </Tag>
                    <Tag>{doc.category}</Tag>
                  </Space>
                </Space>
              }
            >
              <Paragraph>
                {doc.content.split('\n').map((line, index) => (
                  <div key={index}>{line}</div>
                ))}
              </Paragraph>
              <Space>
                {doc.tags.map(tag => (
                  <Tag key={tag}>{tag}</Tag>
                ))}
              </Space>
            </Panel>
          ))}
        </Collapse>

        {/* 空状态 */}
        {filteredDocuments.length === 0 && (
          <div style={{ padding: '32px', textAlign: 'center' }}>
            <Text type="secondary">未找到相关帮助文档</Text>
          </div>
        )}

        {/* 底部操作 */}
        <Space style={{ marginTop: '16px', justifyContent: 'center' }}>
          <Button 
            type="primary" 
            icon={<MessageOutlined />} 
            onClick={onAskAI}
          >
            咨询AI助手
          </Button>
          <Button 
            icon={<BookOutlined />}
          >
            查看完整文档
          </Button>
        </Space>
      </Space>
    </Card>
  );
};

export default HelpCenter;
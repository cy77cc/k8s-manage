import React, { useEffect, useMemo, useState } from 'react';
import { Button, Card, Collapse, Descriptions, List, Modal, Skeleton, Space, Spin, Tag, Typography, message } from 'antd';
import { AlertOutlined, BookOutlined, BulbOutlined, RocketOutlined } from '@ant-design/icons';
import { Api } from '../../api';
import type { AIRecommendation } from '../../api';

const { Text, Paragraph } = Typography;

interface RecommendationPanelProps {
  type: string;
  context: any;
  limit?: number;
  className?: string;
  refreshSignal?: number;
  onLoadingChange?: (loading: boolean) => void;
}

const recommendationIcons: Record<string, React.ReactNode> = {
  suggestion: <BulbOutlined />,
  action: <RocketOutlined />,
  knowledge: <BookOutlined />,
  warning: <AlertOutlined />,
};

const RecommendationPanel: React.FC<RecommendationPanelProps> = ({ type, context, limit = 5, className, refreshSignal = 0, onLoadingChange }) => {
  const [recommendations, setRecommendations] = useState<AIRecommendation[]>([]);
  const [loading, setLoading] = useState(false);
  const [previewLoading, setPreviewLoading] = useState(false);
  const [execLoading, setExecLoading] = useState(false);
  const [preview, setPreview] = useState<any>(null);
  const [previewFor, setPreviewFor] = useState<AIRecommendation | null>(null);

  const loadRecommendations = async () => {
    try {
      setLoading(true);
      onLoadingChange?.(true);
      const response = await Api.ai.getRecommendations({ type, context, limit });
      const list = response.data || [];
      if (list.length === 0) {
        setRecommendations([{
          id: `fallback-${Date.now()}`,
          type: 'suggestion',
          title: '通用建议',
          content: '建议先查询资源健康状态，再执行变更操作。',
          relevance: 0.7,
        }]);
        return;
      }
      setRecommendations(list);
    } catch (error) {
      console.error(error);
      setRecommendations([{
        id: `fallback-error-${Date.now()}`,
        type: 'suggestion',
        title: '推荐服务暂不可用',
        content: '当前未获取到智能推荐，可先在聊天窗口直接描述你的目标。',
        relevance: 0.5,
      }]);
    } finally {
      setLoading(false);
      onLoadingChange?.(false);
    }
  };

  useEffect(() => {
    loadRecommendations();
  }, [type, context, limit, refreshSignal]);

  const openPreview = async (rec: AIRecommendation) => {
    if (!rec.action) return;
    try {
      setPreviewLoading(true);
      setPreviewFor(rec);
      const params = rec.params || { page: context?.page, project_id: context?.projectId };
      const res = await Api.ai.previewAction({ action: rec.action, params });
      setPreview(res.data);
    } catch (error: any) {
      message.error(error?.response?.data?.message || '预览失败');
      setPreviewFor(null);
      setPreview(null);
    } finally {
      setPreviewLoading(false);
    }
  };

  const execute = async () => {
    if (!preview?.approval_token) return;
    try {
      setExecLoading(true);
      await Api.ai.executeAction({ approval_token: preview.approval_token });
      message.success('执行成功');
      setPreview(null);
      setPreviewFor(null);
      await loadRecommendations();
    } catch (error: any) {
      message.error(error?.response?.data?.message || '执行失败');
    } finally {
      setExecLoading(false);
    }
  };

  const visible = useMemo(() => !!previewFor, [previewFor]);

  return (
    <>
      <Card
        title={
          <Space>
            {recommendationIcons[type] || <BulbOutlined />}
            <Text strong>智能推荐</Text>
          </Space>
        }
        className={`ai-recommendation-panel ${className || ''}`}
      >
        {loading ? <div className="ai-recommendation-header-progress" /> : null}
        {loading ? (
          <div className="ai-recommendation-skeleton-wrap">
            <Skeleton active paragraph={{ rows: 2 }} title={{ width: '45%' }} />
            <Skeleton active paragraph={{ rows: 2 }} title={{ width: '50%' }} />
            <Skeleton active paragraph={{ rows: 1 }} title={{ width: '40%' }} />
          </div>
        ) : (
          <List
            dataSource={recommendations}
            locale={{ emptyText: '暂无推荐内容' }}
            renderItem={(item) => (
              <List.Item
                actions={
                  item.action
                    ? [<Button size="small" type="link" onClick={() => openPreview(item)}>预览</Button>]
                    : []
                }
              >
                <List.Item.Meta
                  avatar={<Tag icon={recommendationIcons[item.type]}>{item.type}</Tag>}
                  title={
                    <Space>
                      <Text strong>{item.title}</Text>
                      <Tag color="blue">{Math.round((item.relevance || 0) * 100)}%</Tag>
                    </Space>
                  }
                  description={(
                    <Space direction="vertical" size={6} style={{ width: '100%' }}>
                      <Text type="secondary">{item.content}</Text>
                      {item.reasoning ? (
                        <Collapse
                          className="ai-recommendation-reasoning"
                          size="small"
                          ghost
                          items={[{
                            key: `reasoning-${item.id}`,
                            label: <Text type="secondary">建议思考摘要</Text>,
                            children: <Text type="secondary">{item.reasoning}</Text>,
                          }]}
                        />
                      ) : null}
                    </Space>
                  )}
                />
              </List.Item>
            )}
          />
        )}
      </Card>

      <Modal
        title="AI 动作预览"
        open={visible}
        confirmLoading={execLoading}
        okText="确认执行"
        cancelText="取消"
        onCancel={() => {
          if (!execLoading) {
            setPreview(null);
            setPreviewFor(null);
          }
        }}
        onOk={execute}
      >
        {previewLoading ? (
          <div style={{ textAlign: 'center', padding: 20 }}><Spin /></div>
        ) : preview ? (
          <Space direction="vertical" style={{ width: '100%' }} size="middle">
            <Descriptions column={1} size="small" bordered>
              <Descriptions.Item label="Intent">{preview.intent || previewFor?.action}</Descriptions.Item>
              <Descriptions.Item label="Risk">{preview.risk || 'medium'}</Descriptions.Item>
              <Descriptions.Item label="Params">
                <pre style={{ margin: 0, whiteSpace: 'pre-wrap' }}>{JSON.stringify(preview.params || {}, null, 2)}</pre>
              </Descriptions.Item>
            </Descriptions>
            <div>
              <Text strong>Preview Diff</Text>
              <Paragraph style={{ whiteSpace: 'pre-wrap', marginTop: 8 }}>
                {preview.previewDiff || '-'}
              </Paragraph>
            </div>
          </Space>
        ) : (
          <Text type="secondary">暂无预览数据</Text>
        )}
      </Modal>
    </>
  );
};

export default RecommendationPanel;

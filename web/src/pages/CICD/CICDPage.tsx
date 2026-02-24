import React, { useState } from 'react';
import { Button, Card, Form, Input, InputNumber, Modal, Space, Table, message } from 'antd';
import { Api } from '../../api';

const CICDPage: React.FC = () => {
  const [open, setOpen] = useState(false);
  const [pipeline, setPipeline] = useState<any>(null);
  const [logs, setLogs] = useState<any[]>([]);
  const [form] = Form.useForm();

  const createPipeline = async () => {
    const values = await form.validateFields();
    const res = await Api.cicd.createPipeline(values);
    setPipeline(res.data);
    setOpen(false);
    message.success('流水线创建成功');
  };

  const runPipeline = async () => {
    if (!pipeline?.id) return;
    const runRes = await Api.cicd.runPipeline(String(pipeline.id));
    const runId = String(runRes.data.id);
    if (runId) {
      const logRes = await Api.cicd.getRunLogs(runId);
      setLogs(logRes.data || []);
    }
    message.success(`运行状态: ${runRes.data.status}`);
  };

  return (
    <Card
      title="CI/CD"
      extra={<Space><Button type="primary" onClick={() => setOpen(true)}>创建流水线</Button><Button disabled={!pipeline} onClick={runPipeline}>运行</Button></Space>}
    >
      {pipeline ? <pre>{JSON.stringify(pipeline, null, 2)}</pre> : <div>暂无流水线</div>}
      <Table rowKey="id" dataSource={logs} columns={[{ title: '步骤', dataIndex: 'step' }, { title: '日志', dataIndex: 'message' }, { title: '时间', dataIndex: 'created_at' }]} />
      <Modal title="创建流水线（Git -> Build -> Deploy）" open={open} onCancel={() => setOpen(false)} onOk={createPipeline}>
        <Form form={form} layout="vertical" initialValues={{ branch: 'main', build_cmd: 'npm run build', namespace: 'default', cluster_id: 1 }}>
          <Form.Item name="service_name" label="服务名" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="repo_url" label="Git仓库地址" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="branch" label="分支"><Input /></Form.Item>
          <Form.Item name="build_cmd" label="构建命令"><Input /></Form.Item>
          <Form.Item name="dockerfile_path" label="Dockerfile路径"><Input placeholder="Dockerfile" /></Form.Item>
          <Form.Item name="image_repo" label="镜像仓库"><Input /></Form.Item>
          <Form.Item name="cluster_id" label="Cluster ID"><InputNumber min={1} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="namespace" label="Namespace"><Input /></Form.Item>
        </Form>
      </Modal>
    </Card>
  );
};

export default CICDPage;

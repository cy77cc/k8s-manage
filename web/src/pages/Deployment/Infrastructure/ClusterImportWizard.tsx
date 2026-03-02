import React, { useState } from 'react';
import { Steps, Form, Input, Button, Card, Space, message, Upload, Alert, Result } from 'antd';
import { ArrowLeftOutlined, UploadOutlined, CheckCircleOutlined, LoadingOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { Api } from '../../../api';
import type { Cluster } from '../../../api/modules/cluster';

const { TextArea } = Input;

const ClusterImportWizard: React.FC = () => {
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const [currentStep, setCurrentStep] = useState(0);
  const [loading, setLoading] = useState(false);
  const [validating, setValidating] = useState(false);
  const [validationResult, setValidationResult] = useState<{
    valid: boolean;
    message: string;
    endpoint?: string;
    version?: string;
  } | null>(null);
  const [importedCluster, setImportedCluster] = useState<Cluster | null>(null);

  const handleValidate = async () => {
    try {
      const kubeconfig = form.getFieldValue('kubeconfig');
      if (!kubeconfig) {
        message.error('请输入 kubeconfig');
        return;
      }

      setValidating(true);
      const res = await Api.cluster.validateImport({ kubeconfig });
      setValidationResult(res.data);

      if (res.data.valid) {
        message.success('kubeconfig 验证成功');
      } else {
        message.error(res.data.message);
      }
    } catch (err) {
      message.error(err instanceof Error ? err.message : '验证失败');
      setValidationResult({ valid: false, message: err instanceof Error ? err.message : '验证失败' });
    } finally {
      setValidating(false);
    }
  };

  const handleImport = async () => {
    try {
      setLoading(true);
      const values = await form.validateFields();

      const res = await Api.cluster.importCluster({
        name: values.name,
        description: values.description,
        kubeconfig: values.kubeconfig,
      });

      setImportedCluster(res.data);
      setCurrentStep(3);
      message.success('集群导入成功');
    } catch (err) {
      message.error(err instanceof Error ? err.message : '导入失败');
    } finally {
      setLoading(false);
    }
  };

  const handleFileUpload = (file: File) => {
    const reader = new FileReader();
    reader.onload = (e) => {
      const content = e.target?.result as string;
      form.setFieldsValue({ kubeconfig: content });
      message.success('文件已加载');
    };
    reader.onerror = () => {
      message.error('读取文件失败');
    };
    reader.readAsText(file);
    return false; // Prevent default upload behavior
  };

  const renderStep0 = () => (
    <Card title="基本信息">
      <Form form={form} layout="vertical">
        <Form.Item
          name="name"
          label="集群名称"
          rules={[{ required: true, message: '请输入集群名称' }]}
        >
          <Input placeholder="例如: production-k8s" />
        </Form.Item>
        <Form.Item name="description" label="描述">
          <TextArea rows={2} placeholder="集群描述（可选）" />
        </Form.Item>
      </Form>
    </Card>
  );

  const renderStep1 = () => (
    <Card title="上传 kubeconfig">
      <Form form={form} layout="vertical">
        <Form.Item
          name="kubeconfig"
          label="Kubeconfig 内容"
          rules={[{ required: true, message: '请输入或上传 kubeconfig' }]}
          extra={
            <div className="mt-2">
              <Upload
                beforeUpload={handleFileUpload}
                accept=".yaml,.yml,.conf"
                showUploadList={false}
              >
                <Button icon={<UploadOutlined />}>上传 kubeconfig 文件</Button>
              </Upload>
            </div>
          }
        >
          <TextArea
            rows={12}
            placeholder="粘贴 kubeconfig 内容，或点击上方按钮上传文件"
            style={{ fontFamily: 'monospace' }}
          />
        </Form.Item>

        <Space className="mt-4">
          <Button type="primary" onClick={handleValidate} loading={validating}>
            验证连接
          </Button>
        </Space>

        {validationResult && (
          <Alert
            className="mt-4"
            type={validationResult.valid ? 'success' : 'error'}
            message={validationResult.valid ? '验证成功' : '验证失败'}
            description={
              validationResult.valid ? (
                <div>
                  <p>Endpoint: {validationResult.endpoint}</p>
                  <p>Kubernetes 版本: {validationResult.version}</p>
                </div>
              ) : (
                validationResult.message
              )
            }
            showIcon
          />
        )}
      </Form>
    </Card>
  );

  const renderStep2 = () => (
    <Card title="确认导入">
      <Form form={form} layout="vertical">
        <Form.Item label="集群名称">
          <Input value={form.getFieldValue('name')} disabled />
        </Form.Item>
        <Form.Item label="验证状态">
          {validationResult?.valid ? (
            <div className="flex items-center gap-2 text-green-600">
              <CheckCircleOutlined />
              <span>验证通过</span>
            </div>
          ) : (
            <Alert type="warning" message="请先验证 kubeconfig" />
          )}
        </Form.Item>
        {validationResult?.valid && (
          <>
            <Form.Item label="集群地址">
              <Input value={validationResult.endpoint} disabled />
            </Form.Item>
            <Form.Item label="Kubernetes 版本">
              <Input value={validationResult.version} disabled />
            </Form.Item>
          </>
        )}
      </Form>
    </Card>
  );

  const renderStep3 = () => (
    <Result
      status="success"
      title="集群导入成功"
      subTitle={`集群 "${importedCluster?.name}" 已成功导入`}
      extra={[
        <Button type="primary" key="detail" onClick={() => navigate(`/deployment/infrastructure/clusters/${importedCluster?.id}`)}>
          查看集群
        </Button>,
        <Button key="list" onClick={() => navigate('/deployment/infrastructure/clusters')}>
          返回列表
        </Button>,
      ]}
    />
  );

  const steps = [
    { title: '基本信息', content: renderStep0() },
    { title: '上传配置', content: renderStep1() },
    { title: '确认导入', content: renderStep2() },
    { title: '完成', content: renderStep3() },
  ];

  const canProceed = () => {
    switch (currentStep) {
      case 0:
        return !!form.getFieldValue('name');
      case 1:
        return !!form.getFieldValue('kubeconfig') && validationResult?.valid;
      default:
        return true;
    }
  };

  const handleNext = async () => {
    if (currentStep === 0) {
      try {
        await form.validateFields(['name']);
        setCurrentStep(currentStep + 1);
      } catch {
        // Validation failed
      }
    } else if (currentStep === 1) {
      if (validationResult?.valid) {
        setCurrentStep(currentStep + 1);
      } else {
        message.warning('请先验证 kubeconfig');
      }
    } else if (currentStep === 2) {
      handleImport();
    }
  };

  const handlePrev = () => {
    setCurrentStep(currentStep - 1);
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/deployment/infrastructure/clusters')}>
          返回
        </Button>
        <div>
          <h1 className="text-2xl font-semibold text-gray-900">导入集群</h1>
          <p className="text-sm text-gray-500 mt-1">导入已存在的 Kubernetes 集群</p>
        </div>
      </div>

      {currentStep < 3 && (
        <Steps current={currentStep} items={steps.map(s => ({ title: s.title }))} />
      )}

      <div className="min-h-[400px]">
        {steps[currentStep].content}
      </div>

      {currentStep < 3 && (
        <div className="flex justify-between">
          <Button onClick={() => navigate('/deployment/infrastructure/clusters')}>
            取消
          </Button>
          <Space>
            {currentStep > 0 && (
              <Button onClick={handlePrev}>
                上一步
              </Button>
            )}
            {currentStep < 2 && (
              <Button type="primary" onClick={handleNext} disabled={!canProceed()}>
                下一步
              </Button>
            )}
            {currentStep === 2 && (
              <Button
                type="primary"
                onClick={handleNext}
                loading={loading}
                disabled={!validationResult?.valid}
              >
                确认导入
              </Button>
            )}
          </Space>
        </div>
      )}
    </div>
  );
};

export default ClusterImportWizard;

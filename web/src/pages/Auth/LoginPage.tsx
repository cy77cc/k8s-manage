import React from 'react';
import { Alert, Button, Card, Form, Input, Typography } from 'antd';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../../components/Auth/AuthContext';

const { Title, Text } = Typography;

const LoginPage: React.FC = () => {
  const [loading, setLoading] = React.useState(false);
  const [error, setError] = React.useState<string | null>(null);
  const { login } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  const onFinish = async (values: { username: string; password: string }) => {
    try {
      setLoading(true);
      setError(null);
      await login(values);
      const redirect = (location.state as { from?: string } | null)?.from || '/';
      navigate(redirect, { replace: true });
    } catch (err) {
      setError(err instanceof Error ? err.message : '登录失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div
      className="min-h-screen flex items-center justify-center px-4"
      style={{
        background:
          'radial-gradient(circle at 8% 4%, rgba(59,130,246,0.12), transparent 30%), radial-gradient(circle at 94% 0%, rgba(14,165,233,0.1), transparent 22%), var(--color-bg-app)',
      }}
    >
      <Card className="w-full max-w-md shadow-lg">
        <Title level={3}>登录 OpsPilot</Title>
        <Text type="secondary">使用账号密码访问控制台</Text>
        <Form layout="vertical" className="mt-6" onFinish={onFinish}>
          {error && <Alert type="error" message={error} className="mb-4" />}
          <Form.Item name="username" label="用户名" rules={[{ required: true, message: '请输入用户名' }]}>
            <Input placeholder="admin" />
          </Form.Item>
          <Form.Item name="password" label="密码" rules={[{ required: true, message: '请输入密码' }]}>
            <Input.Password placeholder="******" />
          </Form.Item>
          <Button type="primary" htmlType="submit" loading={loading} block>
            登录
          </Button>
          <div className="mt-4 text-center">
            <Text type="secondary">
              还没有账号？<Link to="/register">立即注册</Link>
            </Text>
          </div>
        </Form>
      </Card>
    </div>
  );
};

export default LoginPage;

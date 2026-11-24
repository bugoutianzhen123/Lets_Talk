import React, { useState } from 'react';
import { Card, Form, Input, Button, Typography, message, Space, Alert } from 'antd';
import { UserOutlined, LockOutlined, MessageTwoTone } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useChatStore } from '../store';

const { Title, Text } = Typography;

export const Login: React.FC = () => {
  const [isLogin, setIsLogin] = useState(true);
  const [loading, setLoading] = useState(false);
  const [successMsg, setSuccessMsg] = useState('');
  const [form] = Form.useForm();
  const navigate = useNavigate();
  const { setCurrentUser } = useChatStore();

  const onFinish = async (values: any) => {
    setLoading(true);
    setSuccessMsg('');
    try {
      const endpoint = isLogin ? '/user/login' : '/user/register';
      // 使用相对路径：本地由 Vite 代理、线上由 Nginx 反代到后端
      const response = await fetch(endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(values),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || '操作失败');
      }

      if (isLogin) {
        message.success('登录成功！');
        setCurrentUser(data.user);
        navigate('/chat');
      } else {
        // 注册成功后切回登录页，并给出明显的常驻提示，同时回填用户名
        setSuccessMsg(`账号「${values.username}」注册成功，请直接输入密码登录。`);
        setIsLogin(true);
        form.setFieldsValue({ username: values.username, password: '' });
      }
    } catch (error: any) {
      message.error(error.message);
    } finally {
      setLoading(false);
    }
  };

  const switchMode = () => {
    setIsLogin(!isLogin);
    setSuccessMsg('');
    form.resetFields();
  };

  return (
    <div style={{
      height: '100vh',
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      background: 'linear-gradient(135deg, #f0f2f5 0%, #e6f7ff 100%)'
    }}>
      <Card
        style={{ width: 400, boxShadow: '0 4px 12px rgba(0,0,0,0.1)', borderRadius: 12 }}
        styles={{ body: { padding: '32px' } }}
      >
        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          <MessageTwoTone style={{ fontSize: 48, marginBottom: 16 }} />
          <Title level={2} style={{ margin: 0 }}>LetsTalk</Title>
          <Text type="secondary">企业级实时通讯系统</Text>
        </div>

        {successMsg && (
          <Alert
            type="success"
            showIcon
            message="注册成功"
            description={successMsg}
            style={{ marginBottom: 24 }}
          />
        )}

        <Form
          form={form}
          name="auth"
          onFinish={onFinish}
          layout="vertical"
          size="large"
        >
          <Form.Item
            name="username"
            rules={[{ required: true, message: '请输入用户名！' }]}
          >
            <Input prefix={<UserOutlined />} placeholder="用户名" />
          </Form.Item>

          <Form.Item
            name="password"
            rules={[{ required: true, message: '请输入密码！' }]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="密码" />
          </Form.Item>

          <Form.Item>
            <Button type="primary" htmlType="submit" block loading={loading}>
              {isLogin ? '登 录' : '注 册'}
            </Button>
          </Form.Item>
        </Form>

        <div style={{ textAlign: 'center' }}>
          <Space>
            <Text type="secondary">
              {isLogin ? '还没有账号？' : '已有账号？'}
            </Text>
            <Button type="link" style={{ padding: 0 }} onClick={switchMode}>
              {isLogin ? '立即注册' : '返回登录'}
            </Button>
          </Space>
        </div>
      </Card>
    </div>
  );
};
import React, { useState } from 'react';
import { useAuth } from '../context/AuthContext';

const LoginPage = () => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const { login, register, isLoading } = useAuth();
  const [isRegisterMode, setIsRegisterMode] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');

    if (isRegisterMode) {
      // 注册
      const result = await register({
        username,
        password
      });

      if (!result.success) {
        setError(result.error);
      }
    } else {
      // 登录
      const result = await login({
        username,
        password
      });

      if (!result.success) {
        setError(result.error);
      }
    }
  };

  return (
    <div className="login-container">
      <div className="login-form">
        <h2>{isRegisterMode ? '注册账号' : '用户登录'}</h2>
        {error && <div className="error-message">{error}</div>}

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="username">用户名</label>
            <input
              type="text"
              id="username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
              placeholder="请输入用户名"
            />
          </div>

          <div className="form-group">
            <label htmlFor="password">密码</label>
            <input
              type="password"
              id="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              placeholder="请输入密码"
            />
          </div>

          <button
            type="submit"
            className="submit-button"
            disabled={isLoading}
          >
            {isLoading ? '处理中...' : (isRegisterMode ? '注册' : '登录')}
          </button>
        </form>

        <div className="switch-mode">
          {isRegisterMode ? (
            <>
              已有账号？
              <button
                type="button"
                className="switch-button"
                onClick={() => setIsRegisterMode(false)}
              >
                去登录
              </button>
            </>
          ) : (
            <>
              还没有账号？
              <button
                type="button"
                className="switch-button"
                onClick={() => setIsRegisterMode(true)}
              >
                去注册
              </button>
            </>
          )}
        </div>
      </div>
    </div>
  );
};

export default LoginPage;
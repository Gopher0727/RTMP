import { createContext, useState, useEffect, useContext } from 'react';
import { authApi, userApi } from '../services/api';
import wsService from '../services/websocket';

// 创建认证上下文
const AuthContext = createContext(null);

// 认证上下文提供器组件
export const AuthProvider = ({ children }) => {
  const [currentUser, setCurrentUser] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isAuthenticated, setIsAuthenticated] = useState(false);

  // 初始化 - 检查本地存储中的token
  useEffect(() => {
    const initAuth = async () => {
      const token = localStorage.getItem('token');
      if (token) {
        try {
          // 获取用户信息
          const userData = await userApi.getCurrentUser();
          setCurrentUser(userData);
          setIsAuthenticated(true);
        } catch (error) {
          console.error('验证用户信息失败:', error);
          // 清除无效token
          localStorage.removeItem('token');
        }
      }
      setIsLoading(false);
    };

    initAuth();

    return;
  }, []);

  // 登录
  const login = async (credentials) => {
    try {
      setIsLoading(true);
      const responseData = await authApi.login(credentials);

      // 确保token已被保存（authApi.login内部已经处理）
      // 获取用户信息
      const userData = await userApi.getCurrentUser();
      setCurrentUser(userData);
      setIsAuthenticated(true);

      // 移除：不再在登录时连接WebSocket
      // wsService.connect();
      // wsService.startHeartbeat();

      // 登录成功后重定向到聊天页面
      window.location.href = '/chat';

      return { success: true };
    } catch (error) {
      console.error('登录失败:', error);
      return {
        success: false,
        error: error.response?.data?.message || '登录失败，请检查用户名和密码'
      };
    } finally {
      setIsLoading(false);
    }
  };

  // 注册
  const register = async ({ username, password, email }) => {
    try {
      setIsLoading(true);
      const responseData = await authApi.register({ username, password, email });

      // 检查是否注册成功且有token（authApi.register内部已处理token保存）
      if (responseData.data && responseData.data.token) {
        // 获取用户信息
        const userInfo = await userApi.getCurrentUser();
        setCurrentUser(userInfo);
        setIsAuthenticated(true);

        // 移除：不再在注册时连接WebSocket
        // wsService.connect();
        // wsService.startHeartbeat();
        return { success: true };
      }
      return { success: false, error: '注册失败，未返回有效token' };
    } catch (error) {
      console.error('注册错误:', error);
      return { success: false, error: error.response?.data?.message || '注册失败' };
    } finally {
      setIsLoading(false);
    }
  };

  // 登出
  const logout = async () => {
    try {
      // 修改：确保在登出时断开WebSocket连接
      wsService.stopHeartbeat();
      wsService.disconnect();

      // 清除本地存储的token
      localStorage.removeItem('token');

      // 更新状态
      setCurrentUser(null);
      setIsAuthenticated(false);

      // 重定向到登录页面
      window.location.href = '/login';
    } catch (error) {
      console.error('登出失败:', error);
    }
  };

  // 暴露给子组件的值
  const value = {
    currentUser,
    user: currentUser, // 为了兼容ChatRoom组件中使用的user变量名
    isAuthenticated,
    isLoading,
    login,
    register,
    logout
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};

// 自定义Hook，方便组件使用认证上下文
export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth必须在AuthProvider内部使用');
  }
  return context;
}
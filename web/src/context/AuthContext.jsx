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

                    // 连接WebSocket
                    wsService.connect();
                    wsService.startHeartbeat();
                } catch (error) {
                    console.error('验证用户信息失败:', error);
                    // 清除无效token
                    localStorage.removeItem('token');
                }
            }
            setIsLoading(false);
        };

        initAuth();

        // 清理函数
        return () => {
            wsService.stopHeartbeat();
        };
    }, []);

    // 登录
    const login = async (credentials) => {
        try {
            setIsLoading(true);
            const data = await authApi.login(credentials);

            // 获取用户信息
            const userData = await userApi.getCurrentUser();
            setCurrentUser(userData);
            setIsAuthenticated(true);

            // 连接WebSocket
            wsService.connect();
            wsService.startHeartbeat();

            return { success: true };
        } catch (error) {
            console.error('登录失败:', error);
            return {
                success: false,
                error: error.response?.data?.error || '登录失败，请检查用户名和密码'
            };
        } finally {
            setIsLoading(false);
        }
    };

    // 注册
    const register = async (userData) => {
        try {
            setIsLoading(true);
            const data = await authApi.register(userData);
            return { success: true, data };
        } catch (error) {
            console.error('注册失败:', error);
            return {
                success: false,
                error: error.response?.data?.error || '注册失败，请稍后重试'
            };
        } finally {
            setIsLoading(false);
        }
    };

    // 登出
    const logout = async () => {
        try {
            setIsLoading(true);
            // 断开WebSocket连接
            wsService.disconnect();
            wsService.stopHeartbeat();

            // 调用后端登出API
            await authApi.logout();

            // 清除本地状态
            setCurrentUser(null);
            setIsAuthenticated(false);
        } catch (error) {
            console.error('登出失败:', error);
        } finally {
            setIsLoading(false);
        }
    };

    // 提供的上下文值
    const contextValue = {
        currentUser,
        isLoading,
        isAuthenticated,
        login,
        register,
        logout
    };

    return (
        <AuthContext.Provider value={contextValue}>
            {children}
        </AuthContext.Provider>
    );
};

// 自定义Hook，方便使用认证上下文
export const useAuth = () => {
    const context = useContext(AuthContext);
    if (!context) {
        throw new Error('useAuth必须在AuthProvider内部使用');
    }
    return context;
};
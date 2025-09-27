import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import LoginPage from '../pages/LoginPage';
import ChatPage from '../pages/ChatPage';

// 受保护的路由组件
const ProtectedRoute = ({ children }) => {
  const { isAuthenticated, isLoading } = useAuth();

  if (isLoading) {
    return <div className="loading-screen">加载中...</div>;
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return children;
};

// 路由配置组件
const AppRoutes = () => {
  return (
    <Router>
      <Routes>
        {/* 登录页面 - 公开访问 */}
        <Route
          path="/login"
          element={<LoginPage />}
        />

        {/* 聊天页面 - 受保护 */}
        <Route
          path="/chat"
          element={
            <ProtectedRoute>
              <ChatPage />
            </ProtectedRoute>
          }
        />

        {/* 根路径重定向 */}
        <Route
          path="/"
          element={<Navigate to="/login" replace />}
        />

        {/* 404页面 */}
        <Route
          path="*"
          element={<Navigate to="/login" replace />}
        />
      </Routes>
    </Router>
  );
};

export default AppRoutes;
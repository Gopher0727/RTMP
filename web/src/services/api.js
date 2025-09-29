import axios from 'axios';

// 创建axios实例
const api = axios.create({
  baseURL: 'http://localhost:8080/api/v1',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
});

// 请求拦截器 - 添加JWT token
api.interceptors.request.use(
  config => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  error => {
    return Promise.reject(error);
  }
);

// 认证相关API
export const authApi = {
  // 用户登录
  login: async (credentials) => {
    const response = await api.post('/auth/login', credentials);
    // 检查嵌套的data对象中的token
    if (response.data && response.data.data && response.data.data.token) {
      localStorage.setItem('token', response.data.data.token);
    }
    return response.data;
  },

  // 用户注册
  register: async (userData) => {
    const response = await api.post('/auth/register', userData);
    // 检查嵌套的data对象中的token并保存
    if (response.data && response.data.data && response.data.data.token) {
      localStorage.setItem('token', response.data.data.token);
    }
    return response.data;
  },

  // 用户登出
  logout: async () => {
    const response = await api.post('/auth/logout');
    localStorage.removeItem('token');
    return response.data;
  }
};

// 消息相关API
export const messageApi = {
  // 发送消息
  sendMessage: async (messageData) => {
    const response = await api.post('/messages', messageData);
    return response.data;
  },

  // 获取用户消息
  getUserMessages: async (userId) => {
    const response = await api.get(`/messages/user/${userId}`);
    return response.data;
  },

  // 获取房间消息
  getRoomMessages: async (roomId) => {
    const response = await api.get(`/messages/room/${roomId}`);
    return response.data;
  }
};

// 房间相关API
export const roomApi = {
  // 获取房间列表
  getRooms: async () => {
    const response = await api.get('/rooms');
    return response.data;
  },

  // 获取房间详情
  getRoom: async (roomId) => {
    const response = await api.get(`/rooms/${roomId}`);
    return response.data;
  },

  // 获取房间成员
  getMembers: async (roomId) => {
    const response = await api.get(`/rooms/${roomId}/members`);
    return response.data;
  }
};

// 用户相关API
export const userApi = {
  // 获取当前用户信息
  getCurrentUser: async () => {
    const response = await api.get('/users/me');
    return response.data;
  },

  // 获取在线用户
  getOnlineUsers: async () => {
    const response = await api.get('/online');
    return response.data;
  },

  // 获取用户列表
  getUsers: async () => {
    const response = await api.get('/users');
    return response.data;
  }
};

export default api;
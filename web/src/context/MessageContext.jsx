import React, { createContext, useState, useEffect, useContext, useCallback } from 'react';
import axios from 'axios';
import wsService from '../services/websocket';
import { useAuth } from './AuthContext';

// 创建消息上下文
const MessageContext = createContext();

// 消息提供者组件
export const MessageProvider = ({ children }) => {
  // 使用认证上下文获取用户信息
  const { user } = useAuth();
  
  // 状态管理
  const [rooms, setRooms] = useState([]);
  const [selectedRoom, setSelectedRoom] = useState(null);
  const [messages, setMessages] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [unreadCount, setUnreadCount] = useState({});
  const [wsConnected, setWsConnected] = useState(false);

  // 加载房间列表
  const loadRooms = useCallback(async () => {
    if (!user) return;
    
    try {
      setLoading(true);
      setError(null);
      
      const response = await axios.get('/api/v1/rooms', {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`
        }
      });
      
      setRooms(response.data);
      
      // 如果有房间但未选择，选择第一个
      if (response.data.length > 0 && !selectedRoom) {
        handleSelectRoom(response.data[0].id);
      }
    } catch (err) {
      console.error('加载房间失败:', err);
      setError('加载房间失败，请重试');
    } finally {
      setLoading(false);
    }
  }, [user, selectedRoom]);

  // 选择房间
  const handleSelectRoom = useCallback(async (roomId) => {
    try {
      setLoading(true);
      setError(null);
      
      // 获取房间信息
      const roomResponse = await axios.get(`/api/v1/rooms/${roomId}`, {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`
        }
      });
      
      // 获取房间消息
      const messagesResponse = await axios.get(`/api/v1/rooms/${roomId}/messages`, {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`
        }
      });
      
      setSelectedRoom(roomResponse.data);
      setMessages(messagesResponse.data);
      
      // 重置该房间的未读消息数
      setUnreadCount(prev => ({
        ...prev,
        [roomId]: 0
      }));
      
      // 连接WebSocket（如果尚未连接）
      if (!wsService.isConnected()) {
        wsService.connect();
        wsService.startHeartbeat();
      }
      
      // 加入房间
      wsService.send({
        type: 'join_room',
        room_id: roomId
      });
    } catch (err) {
      console.error('选择房间失败:', err);
      setError('选择房间失败，请重试');
    } finally {
      setLoading(false);
    }
  }, []);

  // 创建房间
  const createRoom = useCallback(async (roomData) => {
    if (!user) return;
    
    try {
      setLoading(true);
      setError(null);
      
      const response = await axios.post('/api/v1/rooms', roomData, {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`
        }
      });
      
      // 更新房间列表
      setRooms(prev => [...prev, response.data]);
      
      return response.data;
    } catch (err) {
      console.error('创建房间失败:', err);
      setError('创建房间失败，请重试');
      throw err;
    } finally {
      setLoading(false);
    }
  }, [user]);

  // 发送消息到房间（带HTTP备选）
  const sendMessageToRoom = useCallback(async (roomId, messageText) => {
    if (!user || !messageText.trim()) return;
    
    const messageData = {
      room_id: roomId,
      content: messageText.trim(),
      sender_id: user.id,
      sender_name: user.username,
      created_at: new Date().toISOString()
    };
    
    try {
      // 1. 尝试通过WebSocket发送
      const wsSent = wsService.send({
        type: 'send_message',
        ...messageData
      });
      
      // 如果WebSocket发送成功，直接更新本地消息列表
      if (wsSent) {
        setMessages(prev => [...prev, messageData]);
        return;
      }
      
      // 2. 如果WebSocket发送失败，使用HTTP作为备选
      console.log('WebSocket发送失败，使用HTTP作为备选');
      const response = await axios.post(`/api/v1/rooms/${roomId}/messages`, 
        { content: messageText.trim() },
        {
          headers: {
            Authorization: `Bearer ${localStorage.getItem('token')}`
          }
        }
      );
      
      // 更新本地消息列表
      setMessages(prev => [...prev, response.data]);
    } catch (err) {
      console.error('发送消息失败:', err);
      setError('发送消息失败，请重试');
      throw err;
    }
  }, [user]);

  // 处理WebSocket消息
  const handleWebSocketMessage = useCallback((data) => {
    if (data.type === 'new_message') {
      // 如果消息来自当前选中的房间，添加到消息列表
      if (selectedRoom && data.room_id === selectedRoom.id) {
        setMessages(prev => [...prev, data]);
      } else {
        // 否则增加未读消息计数
        setUnreadCount(prev => ({
          ...prev,
          [data.room_id]: (prev[data.room_id] || 0) + 1
        }));
      }
    } else if (data.type === 'room_update') {
      // 更新房间信息
      setRooms(prev => prev.map(room => 
        room.id === data.room_id ? { ...room, ...data } : room
      ));
    }
  }, [selectedRoom]);

  // 处理WebSocket错误
  const handleWebSocketError = useCallback((error) => {
    console.error('WebSocket错误:', error);
  }, []);

  // 处理连接状态变化
  const handleConnectionChange = useCallback((connected) => {
    setWsConnected(connected);
  }, []);

  // 初始化时加载房间列表
  useEffect(() => {
    if (user) {
      loadRooms();
    }
  }, [user, loadRooms]);

  // 初始化WebSocket连接事件监听
  useEffect(() => {
    // 监听WebSocket消息
    wsService.on('message', handleWebSocketMessage);
    wsService.on('error', handleWebSocketError);
    wsService.on('connection_change', handleConnectionChange);

    // 清理函数
    return () => {
      wsService.off('message', handleWebSocketMessage);
      wsService.off('error', handleWebSocketError);
      wsService.off('connection_change', handleConnectionChange);
    };
  }, [handleWebSocketMessage, handleWebSocketError, handleConnectionChange]);

  // 当组件卸载时，如果WebSocket已连接，断开连接
  useEffect(() => {
    return () => {
      if (wsService.isConnected()) {
        wsService.disconnect();
      }
    };
  }, []);

  // 提供上下文值
  const contextValue = {
    rooms,
    selectedRoom,
    messages,
    loading,
    error,
    unreadCount,
    wsConnected,
    selectRoom: handleSelectRoom,
    sendMessageToRoom,
    createRoom,
    loadRooms
  };

  return (
    <MessageContext.Provider value={contextValue}>
      {children}
    </MessageContext.Provider>
  );
};

// 自定义Hook用于访问消息上下文
export const useMessages = () => {
  const context = useContext(MessageContext);
  if (!context) {
    throw new Error('useMessages必须在MessageProvider内部使用');
  }
  return context;
};
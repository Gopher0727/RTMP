import { createContext, useState, useEffect, useContext } from 'react';
import { messageApi, roomApi } from '../services/api';
import wsService from '../services/websocket';
import { useAuth } from './AuthContext';

// 创建消息上下文
const MessageContext = createContext(null);

// 消息上下文提供器组件
export const MessageProvider = ({ children }) => {
  const { currentUser, isAuthenticated } = useAuth();
  const [messages, setMessages] = useState([]);
  const [rooms, setRooms] = useState([]);
  const [selectedRoom, setSelectedRoom] = useState(null);
  const [isLoading, setIsLoading] = useState(false);
  const [unreadCounts, setUnreadCounts] = useState({});

  // 加载房间列表
  const loadRooms = async () => {
    if (!isAuthenticated) return;

    try {
      setIsLoading(true);
      const data = await roomApi.getRooms();
      setRooms(data);

      // 自动选择第一个房间
      if (data.length > 0 && !selectedRoom) {
        setSelectedRoom(data[0]);
        loadRoomMessages(data[0].id);
      }
    } catch (error) {
      console.error('加载房间列表失败:', error);
    } finally {
      setIsLoading(false);
    }
  };

  // 加载房间消息
  const loadRoomMessages = async (roomId) => {
    if (!isAuthenticated) return;

    try {
      setIsLoading(true);
      const data = await messageApi.getRoomMessages(roomId);
      setMessages(data);

      // 清除未读计数
      setUnreadCounts(prev => ({
        ...prev,
        [roomId]: 0
      }));
    } catch (error) {
      console.error('加载房间消息失败:', error);
    } finally {
      setIsLoading(false);
    }
  };

  // 发送消息到房间
  const sendMessageToRoom = async (roomId, content) => {
    if (!isAuthenticated) return;

    try {
      const messageData = {
        type: 'room',
        target_id: roomId,
        content: content,
        sender_id: currentUser.id
      };

      // 发送到服务器
      await messageApi.sendMessage(messageData);

      // 也可以通过WebSocket发送
      wsService.send({
        type: 'message',
        ...messageData
      });
    } catch (error) {
      console.error('发送消息失败:', error);
      throw error;
    }
  };

  // 发送私聊消息
  const sendPrivateMessage = async (userId, content) => {
    if (!isAuthenticated) return;

    try {
      const messageData = {
        type: 'user',
        target_id: userId,
        content: content,
        sender_id: currentUser.id
      };

      // 发送到服务器
      await messageApi.sendMessage(messageData);

      // 通过WebSocket发送
      wsService.send({
        type: 'message',
        ...messageData
      });
    } catch (error) {
      console.error('发送私聊消息失败:', error);
      throw error;
    }
  };

  // 选择房间
  const selectRoom = (room) => {
    setSelectedRoom(room);
    loadRoomMessages(room.id);
  };

  // WebSocket消息处理
  useEffect(() => {
    if (!isAuthenticated) return;

    // 监听新消息
    const handleNewMessage = (message) => {
      if (message.type === 'message') {
        setMessages(prev => [...prev, message]);

        // 如果消息不是来自当前选中的房间，增加未读计数
        if (message.target_id !== selectedRoom?.id) {
          setUnreadCounts(prev => ({
            ...prev,
            [message.target_id]: (prev[message.target_id] || 0) + 1
          }));
        }
      }
    };

    // 监听用户状态变化
    const handleUserStatus = (data) => {
      // 处理用户上线/下线状态更新
      console.log('用户状态更新:', data);
    };

    // 注册监听器
    wsService.on('message', handleNewMessage);
    wsService.on('message:user_status', handleUserStatus);

    // 清理函数
    return () => {
      wsService.off('message', handleNewMessage);
      wsService.off('message:user_status', handleUserStatus);
    };
  }, [isAuthenticated, selectedRoom]);

  // 认证状态变化时加载房间
  useEffect(() => {
    if (isAuthenticated) {
      loadRooms();
    } else {
      // 清除数据
      setMessages([]);
      setRooms([]);
      setSelectedRoom(null);
      setUnreadCounts({});
    }
  }, [isAuthenticated]);

  // 提供的上下文值
  const contextValue = {
    messages,
    rooms,
    selectedRoom,
    isLoading,
    unreadCounts,
    sendMessageToRoom,
    sendPrivateMessage,
    selectRoom,
    loadRooms,
    loadRoomMessages
  };

  return (
    <MessageContext.Provider value={contextValue}>
      {children}
    </MessageContext.Provider>
  );
};

// 自定义Hook，方便使用消息上下文
export const useMessages = () => {
  const context = useContext(MessageContext);
  if (!context) {
    throw new Error('useMessages必须在MessageProvider内部使用');
  }
  return context;
};
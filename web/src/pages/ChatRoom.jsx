import React, { useState, useEffect, useRef } from 'react';
import { useAuth } from '../context/AuthContext';
import { useMessages } from '../context/MessageContext';
import './ChatRoom.css';

const ChatRoom = () => {
  const { user, logout } = useAuth();
  const { 
    rooms, 
    selectedRoom, 
    messages, 
    unreadCount, 
    loading, 
    error, 
    selectRoom, 
    sendMessageToRoom, 
    createRoom,
    loadRooms
  } = useMessages();
  
  const [messageInput, setMessageInput] = useState('');
  const [showSettings, setShowSettings] = useState(false);
  const [showCreateRoom, setShowCreateRoom] = useState(false);
  const [newRoomData, setNewRoomData] = useState({
    name: '',
    description: ''
  });
  const [settingsData, setSettingsData] = useState({
    username: user?.username || '',
    avatarUrl: user?.avatar_url || ''
  });
  
  const messageEndRef = useRef(null);
  const messagesRef = useRef(null);

  // 自动滚动到最新消息
  useEffect(() => {
    if (selectedRoom && messages[selectedRoom] && messages[selectedRoom].length > 0) {
      messageEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    }
  }, [selectedRoom, messages]);

  // 处理发送消息
  const handleSendMessage = async () => {
    if (!messageInput.trim() || !selectedRoom) return;
    
    try {
      await sendMessageToRoom(selectedRoom, messageInput);
      setMessageInput('');
    } catch (error) {
      console.error('发送消息失败:', error);
    }
  };

  // 处理键盘事件
  const handleKeyPress = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSendMessage();
    }
  };

  // 处理房间选择
  const handleRoomSelect = async (roomId) => {
    await selectRoom(roomId);
  };

  // 处理创建房间
  const handleCreateRoom = async (e) => {
    e.preventDefault();
    
    if (!newRoomData.name.trim()) return;
    
    try {
      const roomData = {
        name: newRoomData.name.trim(),
        description: newRoomData.description.trim(),
        is_private: false
      };
      
      const newRoom = await createRoom(roomData);
      
      if (newRoom) {
        setShowCreateRoom(false);
        setNewRoomData({ name: '', description: '' });
        await selectRoom(newRoom.id);
        await loadRooms(); // 重新加载房间列表
      }
    } catch (error) {
      console.error('创建房间失败:', error);
    }
  };

  // 处理更新设置
  const handleUpdateSettings = async (e) => {
    e.preventDefault();
    // 这里可以添加更新用户设置的API调用
    setShowSettings(false);
  };

  // 获取用户头像
  const getUserAvatar = (userObj) => {
    if (!userObj) return null;
    
    if (userObj.avatar_url) {
      return <img src={userObj.avatar_url} alt={userObj.username} className="avatar-image" />;
    }
    
    // 使用用户名首字母作为默认头像
    const initial = userObj.username ? userObj.username.charAt(0).toUpperCase() : '?';
    return <div className="avatar-placeholder">{initial}</div>;
  };

  // 渲染房间信息头部
  const renderRoomHeader = () => {
    if (!selectedRoom) return null;
    
    const room = rooms.find(r => r.id === selectedRoom);
    if (!room) return null;
    
    return (
      <div className="chat-header">
        <div className="room-info-header">
          <h2>{room.name}</h2>
          {room.is_private && <span className="private-badge">私密</span>}
        </div>
        <div className="room-stats">
          <span>创建于: {new Date(room.created_at).toLocaleDateString()}</span>
        </div>
      </div>
    );
  };

  // 渲染消息列表
  const renderMessages = () => {
    if (!selectedRoom) {
      return (
        <div className="no-room-selected">
          <div className="no-room-icon">💬</div>
          <h3>欢迎使用聊天系统</h3>
          <p>请选择一个房间开始聊天，或者创建一个新房间</p>
        </div>
      );
    }
    
    const roomMessages = messages[selectedRoom] || [];
    
    if (roomMessages.length === 0) {
      return <div className="no-messages">暂无消息，开始发送第一条消息吧！</div>;
    }
    
    return (
      <div className="message-list" ref={messagesRef}>
        {roomMessages.map((msg) => {
          const isOwnMessage = msg.user_id === user?.id;
          return (
            <div key={msg.id} className={`message-item ${isOwnMessage ? 'own' : ''}`}>
              <div className="message-avatar">
                {getUserAvatar(msg.user)}
              </div>
              <div className="message-content">
                <div className="message-header">
                  <span className="message-sender">{isOwnMessage ? '我' : msg.user?.username || 'Unknown'}</span>
                  <span className="message-time">
                    {new Date(msg.created_at || Date.now()).toLocaleTimeString()}
                  </span>
                </div>
                <div className="message-text">{msg.content}</div>
              </div>
            </div>
          );
        })}
        <div ref={messageEndRef} />
      </div>
    );
  };

  // 渲染个人设置模态框
  const renderSettingsModal = () => {
    if (!showSettings) return null;
    
    return (
      <div className="modal-overlay" onClick={() => setShowSettings(false)}>
        <div className="modal-content" onClick={(e) => e.stopPropagation()}>
          <div className="modal-header">
            <h2>个人设置</h2>
            <button className="close-button" onClick={() => setShowSettings(false)}>&times;</button>
          </div>
          <div className="modal-body">
            <form onSubmit={handleUpdateSettings}>
              <div className="form-group">
                <label>用户名</label>
                <input
                  type="text"
                  value={settingsData.username}
                  disabled
                />
              </div>
              <div className="form-group">
                <label>头像URL</label>
                <input
                  type="text"
                  value={settingsData.avatarUrl}
                  onChange={(e) => setSettingsData({ ...settingsData, avatarUrl: e.target.value })}
                  placeholder="输入头像URL"
                />
              </div>
              <div className="form-actions">
                <button type="button" className="cancel-button" onClick={() => setShowSettings(false)}>取消</button>
                <button type="submit" className="save-button">保存</button>
              </div>
            </form>
          </div>
        </div>
      </div>
    );
  };

  // 渲染创建房间模态框
  const renderCreateRoomModal = () => {
    if (!showCreateRoom) return null;
    
    return (
      <div className="modal-overlay" onClick={() => setShowCreateRoom(false)}>
        <div className="modal-content" onClick={(e) => e.stopPropagation()}>
          <div className="modal-header">
            <h2>创建新房间</h2>
            <button className="close-button" onClick={() => setShowCreateRoom(false)}>&times;</button>
          </div>
          <div className="modal-body">
            <form onSubmit={handleCreateRoom}>
              <div className="form-group">
                <label>房间名称</label>
                <input
                  type="text"
                  value={newRoomData.name}
                  onChange={(e) => setNewRoomData({ ...newRoomData, name: e.target.value })}
                  placeholder="输入房间名称"
                  required
                />
              </div>
              <div className="form-group">
                <label>房间描述</label>
                <textarea
                  value={newRoomData.description}
                  onChange={(e) => setNewRoomData({ ...newRoomData, description: e.target.value })}
                  placeholder="输入房间描述"
                  rows="3"
                />
              </div>
              <div className="form-actions">
                <button type="button" className="cancel-button" onClick={() => setShowCreateRoom(false)}>取消</button>
                <button type="submit" className="create-button">创建</button>
              </div>
            </form>
          </div>
        </div>
      </div>
    );
  };

  if (loading && rooms.length === 0) {
    return <div className="loading">加载中...</div>;
  }

  return (
    <div className="chat-container">
      {/* 左侧房间侧边栏 */}
      <div className="room-sidebar">
        <div className="user-info">
          <div className="user-avatar" onClick={() => setShowSettings(true)}>
            {getUserAvatar(user)}
          </div>
          <div className="user-details">
            <h3>{user?.username || '用户'}</h3>
            <div className="user-status">在线</div>
          </div>
          <button className="settings-button" onClick={() => setShowSettings(true)}>
            ⚙️
          </button>
        </div>
        
        <div className="rooms-header">
          <h2>房间列表</h2>
          <span className="rooms-count">({rooms.length})</span>
          <button className="create-room-button" onClick={() => setShowCreateRoom(true)}>
            +
          </button>
        </div>
        
        <div className="rooms-list">
          {rooms.map((room) => (
            <div 
              key={room.id} 
              className={`room-item ${selectedRoom === room.id ? 'active' : ''}`}
              onClick={() => handleRoomSelect(room.id)}
            >
              <div className="room-info">
                <div className="room-name">{room.name}</div>
                <div className="room-description">{room.description || '暂无描述'}</div>
              </div>
              {unreadCount[room.id] > 0 && (
                <div className="unread-badge">{unreadCount[room.id]}</div>
              )}
            </div>
          ))}
        </div>
      </div>
      
      {/* 右侧聊天区域 */}
      <div className="chat-main">
        {renderRoomHeader()}
        {renderMessages()}
        
        {selectedRoom && (
          <div className="message-input-container">
            <textarea
              className="message-input"
              placeholder="输入消息..."
              value={messageInput}
              onChange={(e) => setMessageInput(e.target.value)}
              onKeyPress={handleKeyPress}
              rows="3"
            />
            <button 
              className="send-button" 
              onClick={handleSendMessage}
              disabled={!messageInput.trim()}
            >
              发送
            </button>
          </div>
        )}
      </div>
      
      {renderSettingsModal()}
      {renderCreateRoomModal()}
      
      {error && (
        <div className="error-toast">{error}</div>
      )}
    </div>
  );
};

export default ChatRoom;
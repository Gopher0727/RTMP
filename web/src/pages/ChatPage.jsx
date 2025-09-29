import React, { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import { useMessages } from '../context/MessageContext';

const ChatPage = () => {
  const { currentUser, logout } = useAuth();
  const {
    messages,
    rooms,
    selectedRoom,
    isLoading,
    unreadCounts,
    sendMessageToRoom,
    selectRoom
  } = useMessages();

  const [messageInput, setMessageInput] = useState('');
  const [isSending, setIsSending] = useState(false);

  // 发送消息
  const handleSendMessage = async () => {
    if (!messageInput.trim() || !selectedRoom || isSending) return;

    try {
      setIsSending(true);
      await sendMessageToRoom(selectedRoom.id, messageInput.trim());
      setMessageInput('');
    } catch (error) {
      console.error('发送消息失败:', error);
      alert('发送消息失败，请稍后重试');
    } finally {
      setIsSending(false);
    }
  };

  // 处理键盘事件
  const handleKeyPress = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSendMessage();
    }
  };

  // 格式化时间
  const formatTime = (timestamp) => {
    const date = new Date(timestamp);
    return date.toLocaleTimeString('zh-CN', {
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  return (
    <div className="chat-container">
      {/* 顶部导航栏 */}
      <div className="chat-header">
        <div className="header-info">
          <h1>实时通信系统</h1>
          {currentUser && (
            <span className="user-info">欢迎，{currentUser.username}</span>
          )}
        </div>
        <button className="logout-button" onClick={logout}>
          退出登录
        </button>
      </div>

      <div className="chat-content">
        {/* 左侧房间列表 */}
        <div className="room-list">
          <h2>聊天房间</h2>
          {isLoading && rooms.length === 0 ? (
            <div className="loading">加载中...</div>
          ) : (
            <ul>
              {rooms.map(room => (
                <li
                  key={room.id}
                  className={`room-item ${selectedRoom?.id === room.id ? 'active' : ''}`}
                  onClick={() => selectRoom(room)}
                >
                  <div className="room-name">{room.name}</div>
                  {unreadCounts[room.id] > 0 && (
                    <span className="unread-count">{unreadCounts[room.id]}</span>
                  )}
                </li>
              ))}
            </ul>
          )}
        </div>

        {/* 右侧聊天区域 */}
        <div className="chat-area">
          {!selectedRoom ? (
            <div className="no-room-selected">
              <p>请选择一个房间开始聊天</p>
            </div>
          ) : (
            <>
              <div className="room-header">
                <h2>{selectedRoom.name}</h2>
                {selectedRoom.description && (
                  <p className="room-description">{selectedRoom.description}</p>
                )}
              </div>

              <div className="messages-container">
                {messages.length === 0 ? (
                  <div className="no-messages">暂无消息</div>
                ) : (
                  messages.map((message, index) => (
                    <div
                      key={index}
                      className={`message-item ${message.sender_id === currentUser?.id ? 'own' : 'other'}`}
                    >
                      <div className="message-sender">
                        {message.sender_id === currentUser?.id ? '我' : message.sender_name}
                      </div>
                      <div className="message-content">{message.content}</div>
                      <div className="message-time">{formatTime(message.created_at)}</div>
                    </div>
                  ))
                )}
                {isLoading && <div className="loading">加载中...</div>}
              </div>

              <div className="message-input-container">
                <textarea
                  className="message-input"
                  value={messageInput}
                  onChange={(e) => setMessageInput(e.target.value)}
                  onKeyPress={handleKeyPress}
                  placeholder="输入消息..."
                  rows={3}
                />
                <button
                  className="send-button"
                  onClick={handleSendMessage}
                  disabled={isSending || !messageInput.trim()}
                >
                  {isSending ? '发送中...' : '发送'}
                </button>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
};

export default ChatPage;
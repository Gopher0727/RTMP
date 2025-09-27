// WebSocket服务管理类
class WebSocketService {
  constructor() {
    this.ws = null;
    this.url = 'ws://localhost:8080/api/v1/ws';
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.reconnectDelay = 1000;
    this.listeners = new Map();
    this.isConnecting = false;
  }

  // 连接WebSocket
  connect() {
    if (this.ws && (this.ws.readyState === WebSocket.OPEN || this.ws.readyState === WebSocket.CONNECTING)) {
      return;
    }

    if (this.isConnecting) {
      return;
    }

    this.isConnecting = true;
    const token = localStorage.getItem('token');
    const url = `${this.url}?token=${token}`;

    try {
      this.ws = new WebSocket(url);

      this.ws.onopen = () => {
        console.log('WebSocket连接已建立');
        this.reconnectAttempts = 0;
        this.isConnecting = false;
        this.emit('open');
      };

      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          this.emit('message', data);

          // 根据消息类型分发
          if (data.type) {
            this.emit(`message:${data.type}`, data);
          }
        } catch (error) {
          console.error('解析WebSocket消息失败:', error);
        }
      };

      this.ws.onerror = (error) => {
        console.error('WebSocket错误:', error);
        this.isConnecting = false;
        this.emit('error', error);
      };

      this.ws.onclose = () => {
        console.log('WebSocket连接已关闭');
        this.isConnecting = false;
        this.emit('close');

        // 尝试重连
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
          setTimeout(() => {
            this.reconnectAttempts++;
            console.log(`尝试第 ${this.reconnectAttempts} 次重连...`);
            this.connect();
          }, this.reconnectDelay * Math.pow(2, this.reconnectAttempts));
        }
      };
    } catch (error) {
      console.error('创建WebSocket连接失败:', error);
      this.isConnecting = false;
    }
  }

  // 发送消息
  send(data) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data));
      return true;
    }
    console.error('WebSocket连接未建立或已关闭');
    return false;
  }

  // 关闭连接
  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
      this.reconnectAttempts = 0;
    }
  }

  // 添加事件监听器
  on(event, callback) {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, []);
    }
    this.listeners.get(event).push(callback);
  }

  // 移除事件监听器
  off(event, callback) {
    if (this.listeners.has(event)) {
      const callbacks = this.listeners.get(event);
      const index = callbacks.indexOf(callback);
      if (index !== -1) {
        callbacks.splice(index, 1);
      }
    }
  }

  // 触发事件
  emit(event, data) {
    if (this.listeners.has(event)) {
      this.listeners.get(event).forEach(callback => {
        try {
          callback(data);
        } catch (error) {
          console.error(`处理事件 ${event} 时出错:`, error);
        }
      });
    }
  }

  // 获取连接状态
  getReadyState() {
    if (!this.ws) return 0;
    return this.ws.readyState;
  }

  // 发送心跳保持连接
  startHeartbeat(interval = 30000) {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
    }

    this.heartbeatInterval = setInterval(() => {
      if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        this.send({ type: 'ping' });
      }
    }, interval);
  }

  // 停止心跳
  stopHeartbeat() {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }
  }
}

// 导出单例实例
const wsService = new WebSocketService();
export default wsService;
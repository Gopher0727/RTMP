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
    this.heartbeatInterval = null;
    this.heartbeatIntervalTime = 30000; // 30秒
    this.autoReconnect = true;
    this.manualDisconnect = false;
  }

  // 检查连接状态
  isConnected() {
    return this.ws && this.ws.readyState === WebSocket.OPEN;
  }

  // 连接WebSocket
  connect() {
    // 检查是否已连接
    if (this.isConnected()) {
      console.log('WebSocket already connected');
      return;
    }

    if (this.isConnecting) {
      return;
    }

    this.isConnecting = true;
    this.manualDisconnect = false;

    // 获取token
    const token = localStorage.getItem('token');
    if (!token || token === 'null') {
      console.error('Cannot connect to WebSocket: No authentication token');
      this.emit('error', new Error('No authentication token'));
      this.isConnecting = false;
      return;
    }

    try {
      // 使用配置的url或根据当前页面动态构建
      const wsUrl = `${this.url}?token=${token}`;

      this.ws = new WebSocket(wsUrl);

      // 设置事件监听器
      this.ws.onopen = () => {
        console.log('WebSocket connected');
        this.reconnectAttempts = 0;
        this.isConnecting = false;
        this.emit('open');
        this.emit('connection_change', true);
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
          console.error('Failed to parse WebSocket message:', error);
          this.emit('error', error);
        }
      };

      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        this.isConnecting = false;
        this.emit('error', error);
      };

      this.ws.onclose = () => {
        console.log('WebSocket disconnected');
        this.isConnecting = false;
        this.emit('close');
        this.emit('connection_change', false);
        this.stopHeartbeat();

        // 尝试重连
        if (this.autoReconnect && !this.manualDisconnect && this.reconnectAttempts < this.maxReconnectAttempts) {
          setTimeout(() => {
            this.reconnectAttempts++;
            console.log(`尝试第 ${this.reconnectAttempts} 次重连...`);
            this.connect();
          }, this.reconnectDelay * Math.pow(2, this.reconnectAttempts));
        }
      };
    } catch (error) {
      console.error('Failed to create WebSocket connection:', error);
      this.isConnecting = false;
      this.emit('error', error);
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
    this.manualDisconnect = true;
    if (this.ws) {
      this.ws.close();
      this.ws = null;
      this.reconnectAttempts = 0;
    }
    this.stopHeartbeat();
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
    if (!this.listeners.has(event)) return;
    const callbacks = this.listeners.get(event);
    const index = callbacks.indexOf(callback);
    if (index > -1) {
      callbacks.splice(index, 1);
    }
  }

  // 触发事件
  emit(event, data) {
    if (!this.listeners.has(event)) return;
    this.listeners.get(event).forEach(callback => {
      try {
        callback(data);
      } catch (error) {
        console.error(`事件监听器错误 [${event}]:`, error);
      }
    });
  }

  // 开始心跳
  startHeartbeat() {
    this.stopHeartbeat(); // 确保没有重复的心跳
    this.heartbeatInterval = setInterval(() => {
      this.send({
        type: 'heartbeat',
        timestamp: Date.now()
      });
    }, this.heartbeatIntervalTime);
  }

  // 停止心跳
  stopHeartbeat() {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }
  }
}

// 导出单例
export default new WebSocketService();
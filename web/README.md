# 实时聊天客户端

这是一个基于React和WebSocket的实时聊天客户端，与Go后端服务配合使用，提供即时消息功能。

## 功能特性

- 🔐 用户认证：登录、注册和JWT token管理
- 💬 实时通信：基于WebSocket的双向消息传输
- 📱 响应式设计：适配不同屏幕尺寸
- 📦 消息管理：支持发送和接收文本消息
- 🏠 房间系统：支持多房间聊天功能

## 技术栈

- **前端框架**：React 18
- **构建工具**：Vite
- **路由管理**：React Router
- **HTTP请求**：Axios
- **状态管理**：React Context
- **样式**：纯CSS

## 开始使用

### 前置条件

- Node.js 16.x 或更高版本
- 确保Go后端服务正在运行（默认端口8080）

### 安装与运行

1. 安装依赖：

```bash
npm install
```

2. 启动开发服务器：

```bash
npm run dev
```

3. 构建生产版本：

```bash
npm run build
```

## 项目结构

```
web/
├── src/
│   ├── components/     # 可复用组件
│   ├── context/        # React上下文（认证和消息状态）
│   ├── hooks/          # 自定义Hooks
│   ├── pages/          # 页面组件（登录页和聊天页）
│   ├── services/       # API和WebSocket服务
│   ├── App.jsx         # 应用入口组件
│   └── main.jsx        # 渲染入口
└── vite.config.js      # Vite配置（包含API代理）
```

## 代理配置

开发服务器配置了以下代理：

- `/api` 请求转发到 `http://localhost:8080`
- `/ws` WebSocket连接转发到 `ws://localhost:8080`

## 注意事项

- 首次使用需要注册账号
- 登录后才能访问聊天功能
- 确保后端服务正在运行

# 实时消息推送系统

## 消息流

1. 客户端通过 /api/v1/ws 建立 WebSocket（带 JWT），注册到 Hub。
2. 外部系统调用 /api/v1/push，服务端将消息封装并发送到 Kafka topic。
3. Kafka consumer 接收消息，根据路由规则：
   - 若目标用户在线（Hub 中存在连接），直接推送。
   - 若离线且 persist=true，写入 Redis 离线队列与数据库备份。
4. 当用户重连，Hub 会拉取 Redis 离线消息并发送，清理离线队列。
5. 支持长轮询回退（当 WebSocket 不可用时），客户端可使用 /api/v1/longpoll。

## 扩展性

1. 水平扩展
   - 多实例部署，使用 Kafka 解耦，Redis Cluster 做状态缓存，Hub 只管理本实例连接。
   - 通过 Nginx 将 WebSocket 请求分发到后端实例，建议使用 ip_hash 或 sticky session（若需要）或用共享连接路由层（如消息转发服务）
2. 高可用
   - Redis 使用 Cluster 与持久化（AOF/RDB）
   - Kafka 三副本、合理分区
   - 健康检查 + 自动扩容
3. 性能与可靠性
   - 写入离线消息到 Redis 时使用 LPUSH + LTRIM 控制队列长度
   - 限流、熔断：push API、连接数控制
   - 心跳与超时：client 超时后自动注销并通知其他服务
4. 安全
   - JWT 签发与过期策略、服务间 token（rbac）
   - 请求签名、IP 白名单（管理接口）
   - Nginx 作为边界代理，TLS 终止
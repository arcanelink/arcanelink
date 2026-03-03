# 架构设计

## 整体架构

### 系统组成

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  客户端 A    │     │  客户端 B    │     │  客户端 C    │
│ @alice:srv-a │     │ @bob:srv-b   │     │ @carol:srv-c │
└──────┬───────┘     └──────┬───────┘     └──────┬───────┘
       │                    │                    │
       │ HTTP长轮询          │                    │
       │                    │                    │
┌──────▼───────┐     ┌──────▼───────┐     ┌──────▼───────┐
│ Homeserver A │◄───►│ Homeserver B │◄───►│ Homeserver C │
│  srv-a.com   │     │  srv-b.com   │     │  srv-c.com   │
└──────────────┘     └──────────────┘     └──────────────┘
      联邦协议              联邦协议
```

### Homeserver架构

```
┌─────────────────────────────────────────┐
│           Homeserver                    │
├─────────────────────────────────────────┤
│  客户端API层                             │
│  - 长轮询管理                            │
│  - 消息发送接口                          │
│  - Room管理接口                          │
├─────────────────────────────────────────┤
│  业务逻辑层                              │
│  - 用户管理                              │
│  - 私聊消息路由                          │
│  - Room事件处理                          │
│  - 在线状态管理                          │
├─────────────────────────────────────────┤
│  联邦协议层                              │
│  - 服务器发现                            │
│  - 消息转发                              │
│  - Room同步                              │
├─────────────────────────────────────────┤
│  存储层                                  │
│  - 用户数据                              │
│  - 私聊消息                              │
│  - Room数据                              │
│  - 消息队列                              │
└─────────────────────────────────────────┘
```

## 双通道消息模型

### 私聊通道（Direct Channel）

特点：
- 不创建Room对象
- 消息直接路由
- 轻量级处理
- 双方服务器各自存储

处理流程：
```
1. Alice发送消息给Bob
2. Alice的客户端 POST到 Server-A
3. Server-A验证并存储消息
4. Server-A通过联邦协议转发到Server-B
5. Server-B存储到Bob的消息队列
6. Bob的长轮询请求收到响应
```

### 群聊通道（Room Channel）

特点：
- 使用Room对象
- 维护成员列表
- 保存完整历史
- Room所在服务器为主节点

处理流程：
```
1. Alice在Room中发送消息
2. Alice的客户端 POST到 Room所在的Server
3. Server验证权限并存储事件
4. Server并行转发到所有成员的Homeserver
5. 各成员的Homeserver存储到消息队列
6. 各成员的长轮询请求收到响应
```

## HTTP长轮询机制

### 工作原理

```
客户端                          服务器
  │                              │
  ├─ GET /sync?since=token ─────►│
  │                              │ 检查新消息
  │                              │ 无消息，hold住请求
  │                              │ 等待...
  │                              │ 新消息到达！
  │◄──── 返回消息 ───────────────┤
  │                              │
  ├─ GET /sync?since=new_token ─►│
  │                              │ 立即发起下一次请求
```

### 参数说明

- `since`：上次同步的token，标记已读位置
- `timeout`：超时时间（毫秒），建议30000（30秒）

### 服务器端实现

```
伪代码：
function handleSync(user_id, since_token, timeout):
    messages = getNewMessages(user_id, since_token)

    if messages.length > 0:
        return {
            next_token: generateToken(),
            messages: messages
        }

    # 无新消息，等待
    wait_result = waitForMessages(user_id, timeout)

    if wait_result.has_messages:
        return {
            next_token: generateToken(),
            messages: wait_result.messages
        }
    else:
        # 超时
        return {
            next_token: since_token,
            messages: []
        }
```

## 服务器联邦

### 服务器发现

方式1：DNS SRV记录
```
_matrix-fed._tcp.example.com. 3600 IN SRV 10 0 8448 matrix.example.com.
```

方式2：.well-known
```
GET https://example.com/.well-known/matrix/server
{
  "m.server": "matrix.example.com:8448"
}
```

### 联邦连接

- 协议：HTTPS
- 端口：默认8448
- 路径前缀：`/_fed/v1/`

### 消息转发流程

私聊消息转发：
```
Server-A                    Server-B
    │                          │
    ├─ POST /_fed/v1/send_direct ─►│
    │  {sender, recipient, ...}    │
    │                          │ 验证
    │                          │ 存储
    │◄──── 200 OK ─────────────┤
```

群聊消息转发：
```
Room Server                Member Server
    │                          │
    ├─ POST /_fed/v1/send_room ──►│
    │  {room_id, event, ...}      │
    │                          │ 验证
    │                          │ 存储
    │◄──── 200 OK ─────────────┤
```

## 数据存储模型

### 用户数据

```sql
CREATE TABLE users (
    user_id VARCHAR(255) PRIMARY KEY,
    username VARCHAR(100),
    password_hash VARCHAR(255),
    created_at TIMESTAMP,
    profile_data JSON
);
```

### 私聊消息

```sql
CREATE TABLE direct_messages (
    msg_id VARCHAR(100) PRIMARY KEY,
    sender VARCHAR(255),
    recipient VARCHAR(255),
    content JSON,
    timestamp BIGINT,
    INDEX idx_recipient_time (recipient, timestamp)
);
```

### Room数据

```sql
CREATE TABLE rooms (
    room_id VARCHAR(255) PRIMARY KEY,
    creator VARCHAR(255),
    name VARCHAR(200),
    topic TEXT,
    created_at TIMESTAMP
);

CREATE TABLE room_members (
    room_id VARCHAR(255),
    user_id VARCHAR(255),
    joined_at TIMESTAMP,
    PRIMARY KEY (room_id, user_id)
);

CREATE TABLE room_events (
    event_id VARCHAR(100) PRIMARY KEY,
    room_id VARCHAR(255),
    sender VARCHAR(255),
    event_type VARCHAR(50),
    content JSON,
    timestamp BIGINT,
    INDEX idx_room_time (room_id, timestamp)
);
```

### 消息队列

```sql
CREATE TABLE message_queue (
    queue_id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(255),
    message_type ENUM('direct', 'room'),
    message_id VARCHAR(100),
    created_at TIMESTAMP,
    INDEX idx_user_created (user_id, created_at)
);
```

## 在线状态管理

### 状态判定

- 在线：最近60秒内有sync请求
- 离线：超过60秒无sync请求

### 状态存储

```sql
CREATE TABLE presence (
    user_id VARCHAR(255) PRIMARY KEY,
    status ENUM('online', 'offline'),
    last_active TIMESTAMP
);
```

### 状态更新

每次sync请求时更新：
```sql
UPDATE presence
SET status = 'online', last_active = NOW()
WHERE user_id = ?;
```

定期清理（每分钟执行）：
```sql
UPDATE presence
SET status = 'offline'
WHERE last_active < NOW() - INTERVAL 60 SECOND;
```

## 性能优化

### 消息队列优化

使用内存队列 + 持久化：
- 新消息先入内存队列
- 异步写入数据库
- 用户上线时从内存队列读取
- 历史消息从数据库读取

### 长轮询优化

- 使用事件通知机制（epoll/kqueue）
- 避免轮询数据库
- 单个服务器支持数万并发连接

### 联邦缓存

- 缓存服务器发现结果
- 缓存用户所属服务器
- 减少DNS查询

## 扩展性设计

### 水平扩展

- 多个Homeserver实例
- 负载均衡器分发请求
- 共享数据库或分布式存储

### 消息分片

- 按用户ID哈希分片
- 每个分片独立处理
- 提升并发能力

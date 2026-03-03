# 联邦API规范

## 基础信息

- 协议：HTTPS
- 端口：8448（默认联邦端口）
- 路径前缀：`/_fed/v1/`
- 认证方式：服务器签名（可选实现）

## 服务器发现

### 方式1：DNS SRV记录

```
查询：_matrix-fed._tcp.example.com

记录格式：
_matrix-fed._tcp.example.com. 3600 IN SRV 10 0 8448 matrix.example.com.

参数说明：
- 10: 优先级
- 0: 权重
- 8448: 端口
- matrix.example.com: 目标主机
```

### 方式2：.well-known

```
GET https://example.com/.well-known/matrix/server

响应：
{
  "m.server": "matrix.example.com:8448"
}
```

### 服务器连接流程

```
1. 解析目标用户ID：@bob:example.com
2. 提取域名：example.com
3. 查询DNS SRV记录或.well-known
4. 获取实际服务器地址和端口
5. 建立HTTPS连接
```

## 私聊消息转发

### 发送私聊消息

```
POST /_fed/v1/send_direct

请求体：
{
  "sender": "@alice:server-a.com",
  "recipient": "@bob:server-b.com",
  "msg_id": "msg_abc123",
  "content": {
    "msgtype": "m.text",
    "body": "Hello"
  },
  "timestamp": 1234567890000
}

响应：
{
  "success": true,
  "received_at": 1234567890100
}

错误响应：
{
  "error": "USER_NOT_FOUND",
  "message": "Recipient does not exist"
}
```

### 验证流程

接收服务器应验证：
1. recipient确实属于本服务器
2. sender格式正确
3. msg_id唯一（防止重复）
4. timestamp合理（不能太旧或太新）

## Room消息转发

### 发送Room事件

```
POST /_fed/v1/send_room

请求体：
{
  "room_id": "!abc123:server-a.com",
  "event_id": "evt_xyz789",
  "sender": "@alice:server-a.com",
  "event_type": "m.room.message",
  "content": {
    "msgtype": "m.text",
    "body": "Hello everyone"
  },
  "timestamp": 1234567890000
}

响应：
{
  "success": true,
  "received_at": 1234567890100
}

错误响应：
{
  "error": "NOT_IN_ROOM",
  "message": "No local users in this room"
}
```

### 验证流程

接收服务器应验证：
1. 本服务器有用户在该Room中
2. sender有权限发送消息
3. event_id唯一
4. room_id格式正确

## 用户查询

### 查询用户是否存在

```
GET /_fed/v1/query_user/@user:domain.com

响应：
{
  "user_id": "@user:domain.com",
  "exists": true,
  "presence": "online",
  "last_active": 1234567890000
}

用户不存在：
{
  "user_id": "@user:domain.com",
  "exists": false
}
```

## Room操作

### 邀请用户加入Room

```
POST /_fed/v1/room/invite

请求体：
{
  "room_id": "!abc123:server-a.com",
  "inviter": "@alice:server-a.com",
  "invitee": "@bob:server-b.com",
  "room_name": "My Group",
  "room_topic": "Discussion"
}

响应：
{
  "success": true
}
```

### 加入Room

```
POST /_fed/v1/room/join

请求体：
{
  "room_id": "!abc123:server-a.com",
  "user_id": "@bob:server-b.com"
}

响应：
{
  "success": true,
  "room_state": {
    "name": "My Group",
    "topic": "Discussion",
    "members": [
      "@alice:server-a.com",
      "@carol:server-c.com"
    ]
  }
}

错误响应：
{
  "error": "ROOM_NOT_FOUND",
  "message": "Room does not exist"
}
```

### 离开Room

```
POST /_fed/v1/room/leave

请求体：
{
  "room_id": "!abc123:server-a.com",
  "user_id": "@bob:server-b.com"
}

响应：
{
  "success": true
}
```

### 获取Room状态

```
GET /_fed/v1/room/{room_id}/state

响应：
{
  "room_id": "!abc123:server-a.com",
  "name": "My Group",
  "topic": "Discussion",
  "creator": "@alice:server-a.com",
  "created_at": 1234567890000,
  "member_count": 5
}
```

### 获取Room成员列表

```
GET /_fed/v1/room/{room_id}/members

响应：
{
  "members": [
    {
      "user_id": "@alice:server-a.com",
      "joined_at": 1234567890000
    },
    {
      "user_id": "@bob:server-b.com",
      "joined_at": 1234567891000
    }
  ]
}
```

## 服务器健康检查

### 健康检查接口

```
GET /_fed/v1/health

响应：
{
  "status": "ok",
  "version": "1.0",
  "server_name": "example.com",
  "timestamp": 1234567890000
}
```

## 错误处理

### 错误响应格式

```json
{
  "error": "ERROR_CODE",
  "message": "Human readable error message"
}
```

### 常见错误码

- `USER_NOT_FOUND`: 用户不存在
- `ROOM_NOT_FOUND`: Room不存在
- `NOT_IN_ROOM`: 用户不在Room中
- `FORBIDDEN`: 无权限执行操作
- `INVALID_REQUEST`: 请求格式错误
- `SERVER_ERROR`: 服务器内部错误
- `TIMEOUT`: 请求超时

### HTTP状态码

- 200: 成功
- 400: 请求错误
- 403: 禁止访问
- 404: 资源不存在
- 500: 服务器错误
- 503: 服务不可用

## 重试机制

### 消息转发重试

当消息转发失败时：
1. 立即重试1次
2. 如果仍失败，等待5秒后重试
3. 如果仍失败，等待30秒后重试
4. 如果仍失败，等待5分钟后重试
5. 最多重试5次，之后放弃

### 指数退避

```
重试间隔 = min(base_delay * 2^retry_count, max_delay)

base_delay = 5秒
max_delay = 5分钟
```

## 安全考虑

### 防止消息伪造

建议实现：
- 服务器签名验证
- TLS证书验证
- 消息ID去重

### 防止滥用

- 速率限制：每服务器每秒最多100个请求
- 连接限制：每服务器最多10个并发连接
- 消息大小限制：单条消息最大1MB

### 防止循环

- 消息转发时检查sender的域名
- 不转发来自目标服务器的消息
- 维护消息ID历史，防止重复处理

## 性能优化

### 连接复用

- 使用HTTP/2或HTTP/1.1 Keep-Alive
- 维护服务器间的长连接池
- 批量发送消息

### 并行转发

Room消息转发时：
- 并行向所有成员服务器发送
- 不等待所有响应
- 异步处理失败重试

### 缓存

- 缓存服务器发现结果（TTL: 1小时）
- 缓存Room状态（TTL: 5分钟）
- 缓存用户所属服务器（TTL: 1天）

## 实现建议

### 最小实现

必须实现的接口：
- `POST /_fed/v1/send_direct`
- `POST /_fed/v1/send_room`
- `GET /_fed/v1/query_user/{user_id}`
- `GET /_fed/v1/health`

### 完整实现

建议实现所有接口以支持完整功能。

### 测试

提供测试工具验证：
- 服务器发现
- 消息转发
- Room操作
- 错误处理
- 重试机制
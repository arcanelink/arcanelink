# 客户端API规范

## 基础信息

- 协议：HTTPS
- 端口：443（标准HTTPS端口）
- 路径前缀：`/_api/v1/`
- 认证方式：Bearer Token

## 认证

### 登录

```
POST /_api/v1/login

请求体：
{
  "user_id": "@alice:example.com",
  "password": "password123"
}

响应：
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user_id": "@alice:example.com",
  "expires_in": 86400
}
```

### 注销

```
POST /_api/v1/logout

请求头：
Authorization: Bearer <access_token>

响应：
{
  "success": true
}
```

## 消息同步（长轮询）

### 同步接口

```
GET /_api/v1/sync?since={token}&timeout={ms}

请求头：
Authorization: Bearer <access_token>

参数：
- since: 上次同步的token（首次可省略）
- timeout: 超时时间（毫秒），建议30000

响应：
{
  "next_token": "t_12345_67890",
  "direct_messages": [
    {
      "msg_id": "msg_001",
      "sender": "@bob:example.com",
      "content": {
        "msgtype": "m.text",
        "body": "Hello"
      },
      "timestamp": 1234567890000
    }
  ],
  "room_events": [
    {
      "event_id": "evt_001",
      "room_id": "!abc:example.com",
      "sender": "@carol:example.com",
      "event_type": "m.room.message",
      "content": {
        "msgtype": "m.text",
        "body": "Hi everyone"
      },
      "timestamp": 1234567891000
    }
  ]
}
```

## 私聊消息

### 发送私聊消息

```
POST /_api/v1/send_direct

请求头：
Authorization: Bearer <access_token>

请求体：
{
  "recipient": "@bob:example.com",
  "content": {
    "msgtype": "m.text",
    "body": "Hello Bob"
  }
}

响应：
{
  "msg_id": "msg_12345",
  "timestamp": 1234567890000
}
```

### 获取私聊历史

```
GET /_api/v1/direct_history?peer={user_id}&limit={n}&before={token}

请求头：
Authorization: Bearer <access_token>

参数：
- peer: 对方用户ID
- limit: 返回消息数量（默认50，最大100）
- before: 获取此token之前的消息（可选）

响应：
{
  "messages": [
    {
      "msg_id": "msg_100",
      "sender": "@alice:example.com",
      "recipient": "@bob:example.com",
      "content": {
        "msgtype": "m.text",
        "body": "Message text"
      },
      "timestamp": 1234567890000
    }
  ],
  "prev_token": "t_prev_token",
  "has_more": true
}
```

## Room管理

### 创建Room

```
POST /_api/v1/rooms/create

请求头：
Authorization: Bearer <access_token>

请求体：
{
  "name": "My Group Chat",
  "topic": "Discussion about project",
  "invite": ["@bob:example.com", "@carol:example.com"]
}

响应：
{
  "room_id": "!abc123:example.com",
  "created_at": 1234567890000
}
```

### 加入Room

```
POST /_api/v1/rooms/join

请求头：
Authorization: Bearer <access_token>

请求体：
{
  "room_id": "!abc123:example.com"
}

响应：
{
  "success": true,
  "joined_at": 1234567890000
}
```

### 离开Room

```
POST /_api/v1/rooms/leave

请求头：
Authorization: Bearer <access_token>

请求体：
{
  "room_id": "!abc123:example.com"
}

响应：
{
  "success": true
}
```

**注意**：Room创建者不能离开Room，只能删除Room。

### 邀请用户

```
POST /_api/v1/rooms/invite

请求头：
Authorization: Bearer <access_token>

请求体：
{
  "room_id": "!abc123:example.com",
  "user_id": "@dave:example.com"
}

响应：
{
  "success": true
}
```

**注意**：系统会验证被邀请用户是否存在，以及是否已经是Room成员。

### 删除Room

```
POST /_api/v1/rooms/delete

请求头：
Authorization: Bearer <access_token>

请求体：
{
  "room_id": "!abc123:example.com"
}

响应：
{
  "success": true
}
```

**权限要求**：只有Room创建者可以删除Room。

**行为**：删除Room时会自动清理所有成员。

### 获取Room状态

```
GET /_api/v1/rooms/state?room_id=!abc123:example.com

请求头：
Authorization: Bearer <access_token>

响应：
{
  "room_id": "!abc123:example.com",
  "name": "My Group Chat",
  "topic": "Discussion about project",
  "creator": "@alice:example.com",
  "created_at": 1234567890000
}
```

### 获取Room成员

```
GET /_api/v1/rooms/members?room_id=!abc123:example.com

请求头：
Authorization: Bearer <access_token>

响应：
{
  "members": [
    {
      "user_id": "@alice:example.com",
      "joined_at": 1234567890000
    },
    {
      "user_id": "@bob:example.com",
      "joined_at": 1234567891000
    }
  ]
}
```

### 获取Room列表

```
GET /_api/v1/rooms

请求头：
Authorization: Bearer <access_token>

响应：
{
  "rooms": [
    {
      "room_id": "!abc:example.com",
      "name": "My Group",
      "topic": "Discussion",
      "member_count": 5,
      "last_activity": 1234567890000
    }
  ]
}
```

## Room消息

### 发送Room消息

```
POST /_api/v1/send_room

请求头：
Authorization: Bearer <access_token>

请求体：
{
  "room_id": "!abc123:example.com",
  "content": {
    "msgtype": "m.text",
    "body": "Hello everyone"
  }
}

响应：
{
  "event_id": "evt_12345",
  "timestamp": 1234567890000
}
```

### 获取Room历史

```
GET /_api/v1/rooms/history?room_id=!abc123:example.com&limit={n}&before={token}

请求头：
Authorization: Bearer <access_token>

参数：
- room_id: Room ID
- limit: 返回事件数量（默认50，最大100）
- before: 获取此token之前的事件（可选）

响应：
{
  "events": [
    {
      "event_id": "evt_100",
      "sender": "@alice:example.com",
      "event_type": "m.room.message",
      "content": {
        "msgtype": "m.text",
        "body": "Message text"
      },
      "timestamp": 1234567890000
    }
  ],
  "prev_token": "t_prev_token",
  "has_more": true
}
```

## 链接预览

### 获取链接预览

```
GET /_api/v1/link_preview?url={encoded_url}

请求头：
Authorization: Bearer <access_token>

参数：
- url: URL编码的网页地址

响应：
{
  "url": "https://example.com",
  "title": "Example Website",
  "description": "This is an example website",
  "image": "https://example.com/image.jpg",
  "site_name": "Example"
}
```

**功能说明**：
- 自动抓取网页的Open Graph和Twitter Card元数据
- 如果无法获取元数据，返回基本信息（URL作为标题）
- 用于在消息中显示富链接预览卡片

## 用户信息

### 获取用户资料

```
GET /_api/v1/user/{user_id}/profile

响应：
{
  "user_id": "@alice:example.com",
  "display_name": "Alice",
  "avatar_url": "https://example.com/avatar.jpg"
}
```

### 更新用户资料

```
PUT /_api/v1/user/profile

请求头：
Authorization: Bearer <access_token>

请求体：
{
  "display_name": "Alice Smith",
  "avatar_url": "https://example.com/new_avatar.jpg"
}

响应：
{
  "success": true
}
```

## 在线状态

### 获取在线状态

```
GET /_api/v1/presence/{user_id}

响应：
{
  "user_id": "@alice:example.com",
  "presence": "online",
  "last_active": 1234567890000
}
```

## 错误响应

所有API在出错时返回统一格式：

```json
{
  "error": "ERROR_CODE",
  "message": "Human readable error message"
}
```

常见错误码：
- `UNAUTHORIZED`: 未授权或token无效
- `FORBIDDEN`: 无权限执行操作
- `NOT_FOUND`: 资源不存在
- `BAD_REQUEST`: 请求参数错误
- `RATE_LIMITED`: 请求过于频繁
- `SERVER_ERROR`: 服务器内部错误

## 速率限制

为防止滥用，API实施速率限制：

- 登录：每IP每分钟5次
- 发送消息：每用户每秒10条
- 创建Room：每用户每小时10个
- 其他API：每用户每秒100次

超过限制返回429状态码：
```json
{
  "error": "RATE_LIMITED",
  "message": "Too many requests",
  "retry_after": 30
}
```
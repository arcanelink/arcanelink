# 消息格式规范

## 消息类型概述

协议定义两大类消息：
1. 私聊消息（Direct Message）
2. Room事件（Room Event）

## 私聊消息格式

### 基本结构

```json
{
  "msg_id": "msg_abc123",
  "sender": "@alice:example.com",
  "recipient": "@bob:example.com",
  "content": {
    "msgtype": "m.text",
    ...
  },
  "timestamp": 1234567890000
}
```

### 字段说明

- `msg_id`: 消息唯一标识符，格式：`msg_` + 随机字符串
- `sender`: 发送者用户ID
- `recipient`: 接收者用户ID
- `content`: 消息内容对象
- `timestamp`: Unix时间戳（毫秒）

### 文本消息

```json
{
  "msgtype": "m.text",
  "body": "Hello, how are you?"
}
```

**链接检测**：客户端应自动检测消息中的URL并将其渲染为可点击链接。

**链接预览**：客户端可以调用`/link_preview` API获取链接的元数据（标题、描述、图片）并显示富预览卡片。

示例带链接的消息：
```json
{
  "msgtype": "m.text",
  "body": "Check out this website: https://example.com"
}
```

**表情符号**：客户端应支持Unicode表情符号的输入和显示。

### 图片消息

```json
{
  "msgtype": "m.image",
  "body": "image.jpg",
  "url": "https://example.com/files/image.jpg",
  "info": {
    "mimetype": "image/jpeg",
    "size": 102400,
    "w": 1920,
    "h": 1080
  }
}
```

### 文件消息

```json
{
  "msgtype": "m.file",
  "body": "document.pdf",
  "url": "https://example.com/files/document.pdf",
  "info": {
    "mimetype": "application/pdf",
    "size": 524288
  }
}
```

### 音频消息

```json
{
  "msgtype": "m.audio",
  "body": "voice.mp3",
  "url": "https://example.com/files/voice.mp3",
  "info": {
    "mimetype": "audio/mpeg",
    "size": 204800,
    "duration": 30000
  }
}
```

### 视频消息

```json
{
  "msgtype": "m.video",
  "body": "video.mp4",
  "url": "https://example.com/files/video.mp4",
  "info": {
    "mimetype": "video/mp4",
    "size": 10485760,
    "duration": 60000,
    "w": 1920,
    "h": 1080
  }
}
```

### 位置消息

```json
{
  "msgtype": "m.location",
  "body": "My Location",
  "geo_uri": "geo:37.7749,-122.4194",
  "info": {
    "latitude": 37.7749,
    "longitude": -122.4194
  }
}
```

## Room事件格式

### 基本结构

```json
{
  "event_id": "evt_xyz789",
  "room_id": "!abc123:example.com",
  "sender": "@alice:example.com",
  "event_type": "m.room.message",
  "content": {
    ...
  },
  "timestamp": 1234567890000
}
```

### 字段说明

- `event_id`: 事件唯一标识符，格式：`evt_` + 随机字符串
- `room_id`: Room ID
- `sender`: 发送者用户ID
- `event_type`: 事件类型
- `content`: 事件内容对象
- `timestamp`: Unix时间戳（毫秒）

## Room事件类型

### m.room.message - Room消息

与私聊消息的content格式相同：

```json
{
  "event_type": "m.room.message",
  "content": {
    "msgtype": "m.text",
    "body": "Hello everyone"
  }
}
```

### m.room.create - 创建Room

```json
{
  "event_type": "m.room.create",
  "content": {
    "creator": "@alice:example.com",
    "room_version": "1"
  }
}
```

### m.room.name - 设置Room名称

```json
{
  "event_type": "m.room.name",
  "content": {
    "name": "My Group Chat"
  }
}
```

### m.room.topic - 设置Room主题

```json
{
  "event_type": "m.room.topic",
  "content": {
    "topic": "Discussion about the project"
  }
}
```

### m.room.member - 成员变更

加入：
```json
{
  "event_type": "m.room.member",
  "content": {
    "membership": "join",
    "user_id": "@bob:example.com"
  }
}
```

离开：
```json
{
  "event_type": "m.room.member",
  "content": {
    "membership": "leave",
    "user_id": "@bob:example.com"
  }
}
```

邀请：
```json
{
  "event_type": "m.room.member",
  "content": {
    "membership": "invite",
    "user_id": "@carol:example.com",
    "inviter": "@alice:example.com"
  }
}
```

踢出：
```json
{
  "event_type": "m.room.member",
  "content": {
    "membership": "kick",
    "user_id": "@dave:example.com",
    "reason": "Violation of rules"
  }
}
```

## 扩展字段

### 回复消息

```json
{
  "msgtype": "m.text",
  "body": "That's a great idea!",
  "m.relates_to": {
    "rel_type": "m.reply",
    "event_id": "evt_original_message"
  }
}
```

### 编辑消息

```json
{
  "msgtype": "m.text",
  "body": "Corrected message",
  "m.relates_to": {
    "rel_type": "m.replace",
    "event_id": "evt_original_message"
  }
}
```

### 删除消息

```json
{
  "msgtype": "m.text",
  "body": "Message deleted",
  "m.relates_to": {
    "rel_type": "m.delete",
    "event_id": "evt_message_to_delete"
  }
}
```

### 表情回应

```json
{
  "msgtype": "m.reaction",
  "m.relates_to": {
    "rel_type": "m.annotation",
    "event_id": "evt_target_message",
    "key": "👍"
  }
}
```

## 用户资料格式

```json
{
  "user_id": "@alice:example.com",
  "display_name": "Alice Smith",
  "avatar_url": "https://example.com/avatars/alice.jpg",
  "status_msg": "Working on the project"
}
```

## Room状态格式

```json
{
  "room_id": "!abc123:example.com",
  "name": "My Group Chat",
  "topic": "Discussion about the project",
  "creator": "@alice:example.com",
  "created_at": 1234567890000,
  "member_count": 5,
  "avatar_url": "https://example.com/room_avatars/abc123.jpg"
}
```

## 在线状态格式

```json
{
  "user_id": "@alice:example.com",
  "presence": "online",
  "last_active": 1234567890000,
  "status_msg": "Available"
}
```

presence可选值：
- `online`: 在线
- `offline`: 离线
- `away`: 离开
- `busy`: 忙碌

## 同步响应格式

```json
{
  "next_token": "t_12345_67890",
  "direct_messages": [
    {
      "msg_id": "msg_001",
      "sender": "@bob:example.com",
      "content": {...},
      "timestamp": 1234567890000
    }
  ],
  "room_events": [
    {
      "event_id": "evt_001",
      "room_id": "!abc:example.com",
      "sender": "@carol:example.com",
      "event_type": "m.room.message",
      "content": {...},
      "timestamp": 1234567891000
    }
  ],
  "presence_updates": [
    {
      "user_id": "@dave:example.com",
      "presence": "online",
      "last_active": 1234567892000
    }
  ]
}
```

## 数据类型

### 用户ID

格式：`@username:domain.com`

规则：
- 以`@`开头
- username只能包含小写字母、数字、`.`、`_`、`-`
- domain为有效域名

示例：
- `@alice:example.com` ✓
- `@bob_123:chat.example.org` ✓
- `@Alice:example.com` ✗（大写字母）

### Room ID

格式：`!roomid:domain.com`

规则：
- 以`!`开头
- roomid为随机字符串
- domain为创建Room的服务器域名

示例：
- `!abc123xyz:example.com` ✓

### 消息ID

格式：`msg_` + 随机字符串

示例：
- `msg_a1b2c3d4e5f6` ✓

### 事件ID

格式：`evt_` + 随机字符串

示例：
- `evt_x1y2z3a4b5c6` ✓

### 时间戳

Unix时间戳（毫秒）

示例：
- `1234567890000`

### URL

必须是HTTPS URL（生产环境）

示例：
- `https://example.com/files/image.jpg` ✓
- `http://example.com/files/image.jpg` ✗（HTTP不安全）

## 大小限制

- 文本消息body：最大64KB
- 文件URL：最大2KB
- 用户display_name：最大100字符
- Room名称：最大200字符
- Room主题：最大1000字符
- 单个事件总大小：最大1MB

## 字符编码

所有文本必须使用UTF-8编码。

## 扩展性

### 自定义消息类型

可以定义自定义msgtype，建议使用命名空间：

```json
{
  "msgtype": "com.example.custom",
  "body": "Fallback text",
  "com.example.data": {
    "custom_field": "value"
  }
}
```

### 自定义事件类型

可以定义自定义event_type，建议使用命名空间：

```json
{
  "event_type": "com.example.custom_event",
  "content": {
    "custom_data": "value"
  }
}
```

## 向后兼容

- 客户端应忽略未知的msgtype和event_type
- 客户端应忽略未知的字段
- 新增字段应设置合理的默认值
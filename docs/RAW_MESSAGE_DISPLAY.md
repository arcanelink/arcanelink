# 消息原始报文显示功能

## 功能说明

为每条消息添加了查看原始 JSON 报文的功能，方便开发和调试。

## 使用方法

### 1. 查看原始报文

在每条消息的右下角，点击 🔍 图标即可展开原始报文。

```
┌─────────────────────────────┐
│ Hello, world!               │
│                             │
│ 10:30 AM              🔍   │  ← 点击这里
└─────────────────────────────┘
```

### 2. 展开后的显示

```
┌─────────────────────────────┐
│ Hello, world!               │
│                             │
│ 10:30 AM              📋   │  ← 图标变化
├─────────────────────────────┤
│ Raw Message    📋 Copy     │  ← 标题和复制按钮
│ ┌─────────────────────────┐ │
│ │ {                       │ │
│ │   "msg_id": "...",      │ │
│ │   "sender": "...",      │ │
│ │   "content": {          │ │
│ │     "msgtype": "m.text",│ │
│ │     "body": "Hello..."  │ │
│ │   },                    │ │
│ │   "timestamp": 1234567  │ │
│ │ }                       │ │
│ └─────────────────────────┘ │
└─────────────────────────────┘
```

### 3. 复制原始报文

点击 "📋 Copy" 按钮，将原始 JSON 复制到剪贴板。

## 功能特性

### 显示内容

原始报文包含完整的消息数据：

```json
{
  "msg_id": "msg_123456",
  "sender": "@alice:localhost",
  "recipient": "@bob:localhost",
  "content": {
    "msgtype": "m.text",
    "body": "Hello, world!"
  },
  "timestamp": 1709712000000
}
```

### 文件消息

对于文件消息，会显示完整的文件信息：

```json
{
  "msg_id": "msg_789012",
  "sender": "@alice:localhost",
  "room_id": "!room123:localhost",
  "content": {
    "msgtype": "m.image",
    "body": "今天的晚餐",
    "url": "/_api/v1/files/550e8400-e29b-41d4-a716-446655440000",
    "info": {
      "size": 2621440,
      "mimetype": "image/jpeg",
      "filename": "photo.jpg"
    }
  },
  "timestamp": 1709712000000
}
```

### 群聊消息

群聊消息包含 `room_id` 字段：

```json
{
  "msg_id": "event_456789",
  "sender": "@alice:localhost",
  "room_id": "!room123:localhost",
  "content": {
    "msgtype": "m.text",
    "body": "Hello everyone!"
  },
  "timestamp": 1709712000000
}
```

## UI 设计

### 切换按钮

- **未展开**: 🔍 (放大镜图标)
- **已展开**: 📋 (剪贴板图标)
- **位置**: 消息时间戳右侧
- **样式**: 半透明，悬停时高亮

### 原始报文容器

- **背景**: 半透明黑色/白色（根据消息方向）
- **字体**: 等宽字体 (Courier New)
- **格式**: JSON 格式化，缩进 2 空格
- **滚动**: 最大高度 300px，超出可滚动
- **边框**: 顶部分隔线

### 复制按钮

- **位置**: 原始报文标题右侧
- **文字**: "📋 Copy"
- **反馈**: 点击后弹出提示

## 技术实现

### 状态管理

```typescript
const [showRawMessage, setShowRawMessage] = useState(false)

const toggleRawMessage = () => {
  setShowRawMessage(!showRawMessage)
}
```

### 复制功能

```typescript
const copyRawMessage = () => {
  const rawJson = JSON.stringify(message, null, 2)
  navigator.clipboard.writeText(rawJson)
    .then(() => alert('Raw message copied to clipboard!'))
    .catch(() => alert('Failed to copy'))
}
```

### 渲染逻辑

```typescript
<div className="message-footer">
  <div className="message-time">{formatTime(message.timestamp)}</div>
  <button
    className="raw-message-toggle"
    onClick={toggleRawMessage}
    title="Show raw message"
  >
    {showRawMessage ? '📋' : '🔍'}
  </button>
</div>

{showRawMessage && (
  <div className="raw-message-container">
    <div className="raw-message-header">
      <span>Raw Message</span>
      <button className="copy-raw-btn" onClick={copyRawMessage}>
        📋 Copy
      </button>
    </div>
    <pre className="raw-message-content">
      {JSON.stringify(message, null, 2)}
    </pre>
  </div>
)}
```

## 样式细节

### 自适应颜色

```css
/* 自己发送的消息（蓝色背景） */
.message-item.own .raw-message-content {
  background-color: rgba(0, 0, 0, 0.2);
  color: rgba(255, 255, 255, 0.9);
}

/* 别人发送的消息（灰色背景） */
.message-item.other .raw-message-content {
  background-color: rgba(0, 0, 0, 0.05);
  color: var(--text-primary);
}
```

### 滚动条样式

```css
.raw-message-content {
  max-height: 300px;
  overflow-y: auto;
  overflow-x: auto;
}
```

### 响应式设计

- 小屏幕：原始报文容器自动换行
- 大屏幕：保持固定宽度
- 触摸设备：增大按钮点击区域

## 使用场景

### 1. 开发调试

查看消息的完整结构，验证字段是否正确：

```json
{
  "msg_id": "msg_123",
  "content": {
    "msgtype": "m.text",
    "body": "Test message"
  }
}
```

### 2. 问题排查

当消息显示异常时，查看原始数据：

```json
{
  "content": {
    "msgtype": "m.image",
    "url": null,  // ← 发现问题：URL 为空
    "body": "photo.jpg"
  }
}
```

### 3. API 学习

了解消息的数据结构，学习 API 格式：

```json
{
  "msg_id": "msg_456",
  "sender": "@alice:localhost",
  "recipient": "@bob:localhost",
  "content": {
    "msgtype": "m.file",
    "body": "document.pdf",
    "url": "/_api/v1/files/xxx",
    "info": {
      "size": 1024000,
      "mimetype": "application/pdf",
      "filename": "document.pdf"
    }
  },
  "timestamp": 1709712000000
}
```

### 4. 数据导出

复制原始报文用于：
- 保存到文件
- 分享给其他开发者
- 提交 bug 报告
- 编写测试用例

## 键盘快捷键（未来）

可以考虑添加键盘快捷键：

- `Ctrl/Cmd + I`: 切换原始报文显示
- `Ctrl/Cmd + C`: 复制原始报文（当展开时）
- `Esc`: 关闭原始报文

## 性能考虑

### JSON 格式化

```typescript
// 使用 JSON.stringify 的第三个参数格式化
JSON.stringify(message, null, 2)
```

### 按需渲染

只有在展开时才渲染原始报文内容：

```typescript
{showRawMessage && (
  <div className="raw-message-container">
    {/* 内容 */}
  </div>
)}
```

### 内存优化

- 不缓存格式化后的 JSON
- 每次展开时重新生成
- 关闭时立即释放

## 安全考虑

### 敏感信息

原始报文可能包含敏感信息：
- 用户 ID
- 文件 URL
- 时间戳

**建议**: 在生产环境中可以添加开关，只在开发模式下启用此功能。

### 实现方式

```typescript
// 只在开发环境显示
const isDevelopment = process.env.NODE_ENV === 'development'

{isDevelopment && (
  <button className="raw-message-toggle" onClick={toggleRawMessage}>
    {showRawMessage ? '📋' : '🔍'}
  </button>
)}
```

## 未来改进

- [ ] 添加语法高亮
- [ ] 支持展开/折叠 JSON 节点
- [ ] 添加搜索功能
- [ ] 支持导出为文件
- [ ] 添加键盘快捷键
- [ ] 支持比较两条消息
- [ ] 添加字段说明提示
- [ ] 支持编辑和重发（开发模式）

## 总结

原始报文显示功能为开发者提供了强大的调试工具：

- ✅ 查看完整消息结构
- ✅ 一键复制到剪贴板
- ✅ 优雅的 UI 设计
- ✅ 自适应颜色主题
- ✅ 支持所有消息类型
- ✅ 性能优化
- ✅ 易于使用

这个功能对于开发、调试和学习 API 都非常有帮助！

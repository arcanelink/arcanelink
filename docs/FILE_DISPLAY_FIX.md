# 文件消息显示优化

## 问题

之前的实现中，文件消息显示时会从 URL 中提取文件名，导致显示的是文件 ID 而不是原始文件名。

### 之前的问题
```
URL: /_api/v1/files/550e8400-e29b-41d4-a716-446655440000
显示: 550e8400-e29b-41d4-a716-446655440000 ❌
期望: document.pdf ✅
```

## 解决方案

### 1. 存储原始文件名

在发送消息时，将原始文件名存储在 `info.filename` 中：

```typescript
info = {
  size: attachedFile.fileInfo.file_size,
  mimetype: attachedFile.fileInfo.content_type,
  filename: attachedFile.fileInfo.filename, // 存储原始文件名
}
```

### 2. 智能识别 body 内容

在显示消息时，智能判断 `body` 是文件名还是用户说明：

```typescript
// 如果 body 有文件扩展名，认为是文件名
const hasExtension = body && /\.[a-zA-Z0-9]+$/.test(body)
const filename = hasExtension ? body : (info?.filename || body)
const caption = hasExtension ? null : body
```

### 3. 显示逻辑

```typescript
// 场景 1: 只发送文件（无说明）
body = "document.pdf"
info.filename = "document.pdf"
→ 显示: document.pdf

// 场景 2: 文件 + 说明
body = "这是项目文档"
info.filename = "document.pdf"
→ 显示: document.pdf + "这是项目文档"

// 场景 3: 旧消息（无 info.filename）
body = "document.pdf"
info.filename = undefined
→ 显示: document.pdf (降级处理)
```

## 消息结构

### 完整的文件消息结构

```json
{
  "msgtype": "m.file",
  "body": "这是项目文档",
  "url": "/_api/v1/files/550e8400-e29b-41d4-a716-446655440000",
  "info": {
    "size": 1024000,
    "mimetype": "application/pdf",
    "filename": "document.pdf"
  }
}
```

### 字段说明

- **msgtype**: 消息类型（m.text, m.image, m.audio, m.video, m.file）
- **body**: 用户输入的说明文字，或文件名（如果没有说明）
- **url**: 文件下载 URL
- **info.size**: 文件大小（字节）
- **info.mimetype**: MIME 类型
- **info.filename**: 原始文件名（新增）

## 显示效果

### 图片消息

#### 只发送图片
```
┌─────────────────────┐
│   [图片预览]         │
│                     │
│ photo.jpg (2.5 MB)  │
└─────────────────────┘
```

#### 图片 + 说明
```
┌─────────────────────┐
│   [图片预览]         │
│ ─────────────────── │
│ 今天的晚餐           │
│ photo.jpg (2.5 MB)  │
└─────────────────────┘
```

### 文件消息

#### 只发送文件
```
┌─────────────────────┐
│ 📎 document.pdf  ⬇️ │
│ document.pdf (1 MB) │
└─────────────────────┘
```

#### 文件 + 说明
```
┌─────────────────────┐
│ 📎 document.pdf  ⬇️ │
│ ─────────────────── │
│ 请查看这个报告       │
└─────────────────────┘
```

## 代码实现

### 发送消息时

```typescript
// ChatWindow.tsx
if (attachedFile) {
  body = messageText.trim() || attachedFile.fileInfo.filename
  url = attachedFile.fileInfo.url
  info = {
    size: attachedFile.fileInfo.file_size,
    mimetype: attachedFile.fileInfo.content_type,
    filename: attachedFile.fileInfo.filename, // 关键：存储原始文件名
  }
}
```

### 显示消息时

```typescript
// MessageItem.tsx
const renderFileMessage = () => {
  const { msgtype, body, url, info } = message.content

  // 智能判断 body 是文件名还是说明
  const hasExtension = body && /\.[a-zA-Z0-9]+$/.test(body)
  const filename = hasExtension ? body : (info?.filename || body)
  const caption = hasExtension ? null : body

  return (
    <div className="file-message">
      {/* 文件预览/下载 */}
      <FileComponent url={url} filename={filename} />

      {/* 用户说明（如果有） */}
      {caption && <div className="file-caption">{caption}</div>}

      {/* 文件信息 */}
      <div className="file-info">
        <span className="file-name">{filename}</span>
        <span className="file-size">{fileSize}</span>
      </div>
    </div>
  )
}
```

## 兼容性

### 向后兼容

对于旧消息（没有 `info.filename`），使用降级策略：

```typescript
const filename = hasExtension ? body : (info?.filename || body)
```

这样即使旧消息没有 `info.filename`，也能正常显示。

### 不同场景

| 场景 | body | info.filename | 显示文件名 | 显示说明 |
|------|------|---------------|-----------|---------|
| 只发送文件 | "doc.pdf" | "doc.pdf" | doc.pdf | 无 |
| 文件+说明 | "重要文档" | "doc.pdf" | doc.pdf | 重要文档 |
| 旧消息 | "doc.pdf" | undefined | doc.pdf | 无 |
| 旧消息+说明 | "重要文档" | undefined | 重要文档 | 无 |

## 测试用例

### 测试 1: 发送图片（无说明）
```typescript
// 发送
body: "photo.jpg"
info: { filename: "photo.jpg", size: 2621440 }

// 显示
文件名: photo.jpg
说明: 无
```

### 测试 2: 发送图片（有说明）
```typescript
// 发送
body: "今天的晚餐"
info: { filename: "photo.jpg", size: 2621440 }

// 显示
文件名: photo.jpg
说明: 今天的晚餐
```

### 测试 3: 发送文档（无说明）
```typescript
// 发送
body: "report.pdf"
info: { filename: "report.pdf", size: 1024000 }

// 显示
文件名: report.pdf
说明: 无
```

### 测试 4: 发送文档（有说明）
```typescript
// 发送
body: "请查看这个报告"
info: { filename: "report.pdf", size: 1024000 }

// 显示
文件名: report.pdf
说明: 请查看这个报告
```

## 优势

1. **清晰的文件名**: 显示原始文件名而不是 UUID
2. **灵活的说明**: 支持为文件添加说明文字
3. **智能识别**: 自动区分文件名和说明
4. **向后兼容**: 支持旧消息格式
5. **用户友好**: 符合用户期望的显示方式

## 注意事项

1. **文件名唯一性**: 原始文件名可能重复，但 URL 是唯一的
2. **扩展名检测**: 使用正则表达式检测文件扩展名
3. **降级处理**: 如果 `info.filename` 不存在，使用 `body` 作为文件名
4. **特殊字符**: 文件名可能包含特殊字符，需要正确处理

## 总结

通过在 `info` 中存储原始文件名，并在显示时智能判断 `body` 的内容，我们实现了：
- ✅ 显示正确的文件名
- ✅ 支持文件说明
- ✅ 向后兼容
- ✅ 用户体验优化

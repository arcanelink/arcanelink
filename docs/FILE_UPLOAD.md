# 文件发送功能使用说明

## 功能概述

系统现已支持在私聊和群聊中发送文件。文件上传后会存储在服务器，并返回文件 URL 供消息引用。

## API 端点

### 1. 上传文件

**端点**: `POST /_api/v1/files`

**请求头**:
- `Authorization: Bearer <token>` (必需)
- `Content-Type: multipart/form-data`

**请求体**:
- `file`: 文件数据 (最大 50MB)

**响应示例**:
```json
{
  "file_id": "550e8400-e29b-41d4-a716-446655440000",
  "filename": "document.pdf",
  "content_type": "application/pdf",
  "file_size": 1024000,
  "url": "/_api/v1/files/550e8400-e29b-41d4-a716-446655440000"
}
```

### 2. 下载文件

**端点**: `GET /_api/v1/files/{file_id}`

**请求头**:
- `Authorization: Bearer <token>` (必需)

**响应**: 文件二进制数据流

### 3. 获取文件信息

**端点**: `GET /_api/v1/files/{file_id}/info`

**请求头**:
- `Authorization: Bearer <token>` (必需)

**响应示例**:
```json
{
  "file_id": "550e8400-e29b-41d4-a716-446655440000",
  "filename": "document.pdf",
  "content_type": "application/pdf",
  "file_size": 1024000,
  "url": "/_api/v1/files/550e8400-e29b-41d4-a716-446655440000"
}
```

## 发送文件消息

### 私聊发送文件

**端点**: `POST /_api/v1/messages`

**请求体**:
```json
{
  "recipient": "@bob:localhost",
  "content": {
    "msgtype": "m.file",
    "body": "document.pdf",
    "url": "/_api/v1/files/550e8400-e29b-41d4-a716-446655440000",
    "info": {
      "size": 1024000,
      "mimetype": "application/pdf"
    }
  }
}
```

### 群聊发送文件

**端点**: `POST /_api/v1/rooms/{room_id}/messages`

**请求体**:
```json
{
  "content": {
    "msgtype": "m.file",
    "body": "document.pdf",
    "url": "/_api/v1/files/550e8400-e29b-41d4-a716-446655440000",
    "info": {
      "size": 1024000,
      "mimetype": "application/pdf"
    }
  }
}
```

## 完整流程示例

### 1. 上传文件

```bash
curl -X POST http://localhost:8080/_api/v1/files \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "file=@/path/to/document.pdf"
```

### 2. 发送文件消息（私聊）

```bash
curl -X POST http://localhost:8080/_api/v1/messages \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "recipient": "@bob:localhost",
    "content": {
      "msgtype": "m.file",
      "body": "document.pdf",
      "url": "/_api/v1/files/550e8400-e29b-41d4-a716-446655440000",
      "info": {
        "size": 1024000,
        "mimetype": "application/pdf"
      }
    }
  }'
```

### 3. 接收方下载文件

```bash
curl -X GET http://localhost:8080/_api/v1/files/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -o downloaded_file.pdf
```

## 支持的消息类型

- `m.file`: 通用文件
- `m.image`: 图片文件
- `m.audio`: 音频文件
- `m.video`: 视频文件

所有类型都使用相同的上传和下载流程，只是 `msgtype` 字段不同。

## 配置

### 环境变量

- `FILE_STORAGE_PATH`: 文件存储路径（默认: `./data/files`）

### 数据库迁移

系统会自动创建 `file_storage` 表来存储文件元数据。确保运行最新的数据库迁移：

```bash
# 迁移文件位置
migrations/008_create_file_storage.up.sql
```

## 注意事项

1. **文件大小限制**: 当前限制为 50MB
2. **存储结构**: 文件按日期组织存储 (YYYY/MM/DD)
3. **去重**: 使用 SHA256 哈希避免重复存储相同文件
4. **权限**: 只有文件上传者可以删除文件
5. **清理**: 需要定期清理未被引用的文件（待实现）

## 安全考虑

- 所有文件操作都需要认证
- 文件存储路径不暴露给客户端
- 支持的文件类型可以通过 Content-Type 验证
- 建议在生产环境中添加病毒扫描

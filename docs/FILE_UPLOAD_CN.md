# 文件发送功能

## 概述

ArcaneLink 现已支持在私聊和群聊中发送文件。用户可以上传文件到服务器，然后在消息中引用文件 URL。

## 主要特性

- **文件上传**: 支持最大 50MB 的文件上传
- **文件下载**: 通过文件 ID 下载文件
- **文件去重**: 使用 SHA256 哈希避免重复存储
- **日期组织**: 文件按日期（年/月/日）组织存储
- **多种类型**: 支持 m.file、m.image、m.audio、m.video 等消息类型
- **私聊和群聊**: 可在私聊和群聊中发送文件

## 使用流程

### 1. 上传文件

```bash
curl -X POST http://localhost:8080/_api/v1/files \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "file=@/path/to/file.pdf"
```

响应：
```json
{
  "file_id": "550e8400-e29b-41d4-a716-446655440000",
  "filename": "file.pdf",
  "content_type": "application/pdf",
  "file_size": 1024000,
  "url": "/_api/v1/files/550e8400-e29b-41d4-a716-446655440000"
}
```

### 2. 发送文件消息

#### 私聊

```bash
curl -X POST http://localhost:8080/_api/v1/messages \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "recipient": "@bob:localhost",
    "content": {
      "msgtype": "m.file",
      "body": "file.pdf",
      "url": "/_api/v1/files/550e8400-e29b-41d4-a716-446655440000",
      "info": {
        "size": 1024000,
        "mimetype": "application/pdf"
      }
    }
  }'
```

#### 群聊

```bash
curl -X POST http://localhost:8080/_api/v1/rooms/!roomid:localhost/messages \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "content": {
      "msgtype": "m.file",
      "body": "file.pdf",
      "url": "/_api/v1/files/550e8400-e29b-41d4-a716-446655440000",
      "info": {
        "size": 1024000,
        "mimetype": "application/pdf"
      }
    }
  }'
```

### 3. 下载文件

```bash
curl -X GET http://localhost:8080/_api/v1/files/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -o downloaded_file.pdf
```

## 数据库结构

文件元数据存储在 `file_storage` 表中：

```sql
CREATE TABLE file_storage (
    file_id VARCHAR(255) PRIMARY KEY,
    uploader VARCHAR(255) NOT NULL,
    filename VARCHAR(512) NOT NULL,
    content_type VARCHAR(128) NOT NULL,
    file_size BIGINT NOT NULL,
    storage_path TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## 文件存储

- **存储路径**: 默认为 `./data/files`，可通过环境变量 `FILE_STORAGE_PATH` 配置
- **目录结构**: `YYYY/MM/DD/[SHA256_HASH][.ext]`
- **示例**: `2026/03/06/abc123def456...xyz.pdf`

## 配置

### 环境变量

在 `docker-compose.yml` 或环境中设置：

```yaml
environment:
  - FILE_STORAGE_PATH=/data/files
```

### Docker 卷

```yaml
volumes:
  - file_storage:/data/files
```

## 测试

运行测试脚本：

```bash
chmod +x scripts/test_file_upload.sh
./scripts/test_file_upload.sh
```

测试脚本会：
1. 注册测试用户
2. 上传文件
3. 获取文件信息
4. 下载文件并验证
5. 发送文件消息（私聊）
6. 创建房间并发送文件消息（群聊）

## Web 客户端集成

参考 `docs/file_upload_example.js` 中的示例代码：

```javascript
// 上传文件
const fileInfo = await uploadFile(file, token);

// 发送文件消息
await sendFileMessage(recipient, fileInfo, token);
```

## 安全注意事项

1. **文件大小限制**: 当前限制为 50MB，可在代码中调整
2. **文件类型验证**: 建议添加文件类型白名单
3. **病毒扫描**: 生产环境建议集成病毒扫描
4. **访问控制**: 所有文件操作都需要认证
5. **存储清理**: 建议定期清理未被引用的文件

## 未来改进

- [ ] 文件缩略图生成（图片、视频）
- [ ] 文件预览功能
- [ ] 文件分享链接（带过期时间）
- [ ] 文件上传进度显示
- [ ] 断点续传支持
- [ ] 文件压缩和优化
- [ ] 垃圾文件自动清理
- [ ] 文件访问权限控制

## 相关文档

- [English Documentation](FILE_UPLOAD.md)
- [Web Client Example](file_upload_example.js)
- [API Reference](../README.md#file-operations)

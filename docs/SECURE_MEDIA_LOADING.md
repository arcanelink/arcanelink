# 文件访问认证问题修复

## 问题描述

当发送图片、音频或视频消息时，浏览器尝试加载媒体文件会收到 401 未授权错误。

### 原因

```html
<!-- 浏览器直接加载图片 -->
<img src="http://localhost:3000/_api/v1/files/FILE_ID" />

<!-- 问题：浏览器不会自动携带 Authorization header -->
```

HTML 的 `<img>`, `<audio>`, `<video>` 标签在加载资源时不会自动携带自定义的 HTTP 头（如 Authorization），导致后端认证失败。

## 解决方案

使用 Blob URL 方案：先通过 fetch API 下载文件（携带认证 token），然后创建 Blob URL 供媒体标签使用。

### 流程

```
1. 用户查看消息
   ↓
2. 组件挂载，触发 useEffect
   ↓
3. 使用 fetch + Authorization header 下载文件
   ↓
4. 将响应转换为 Blob
   ↓
5. 创建 Blob URL (blob:http://...)
   ↓
6. 设置到 <img>/<audio>/<video> 的 src
   ↓
7. 组件卸载时清理 Blob URL
```

## 实现

### 1. SecureImage 组件

用于安全加载图片：

```typescript
// SecureImage.tsx
export function SecureImage({ src, alt, className }: SecureImageProps) {
  const [blobUrl, setBlobUrl] = useState<string>('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)

  useEffect(() => {
    let objectUrl: string | null = null

    const loadImage = async () => {
      const token = apiClient.getToken()

      // Fetch with authentication
      const response = await fetch(src, {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })

      const blob = await response.blob()
      objectUrl = URL.createObjectURL(blob)
      setBlobUrl(objectUrl)
    }

    loadImage()

    // Cleanup
    return () => {
      if (objectUrl) {
        URL.revokeObjectURL(objectUrl)
      }
    }
  }, [src])

  return <img src={blobUrl} alt={alt} className={className} />
}
```

### 2. SecureMedia 组件

用于安全加载音频和视频：

```typescript
// SecureMedia.tsx
export function SecureMedia({ src, type, className }: SecureMediaProps) {
  const [blobUrl, setBlobUrl] = useState<string>('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)

  useEffect(() => {
    // 同样的逻辑
    const loadMedia = async () => {
      const token = apiClient.getToken()
      const response = await fetch(src, {
        headers: { 'Authorization': `Bearer ${token}` }
      })
      const blob = await response.blob()
      const objectUrl = URL.createObjectURL(blob)
      setBlobUrl(objectUrl)
    }

    loadMedia()
    return () => URL.revokeObjectURL(objectUrl)
  }, [src])

  if (type === 'audio') {
    return <audio controls src={blobUrl} className={className} />
  }
  return <video controls src={blobUrl} className={className} />
}
```

### 3. 使用组件

```typescript
// MessageItem.tsx
if (msgtype === 'm.image') {
  return (
    <div className="file-message image-message">
      <SecureImage src={url!} alt={filename} className="message-image" />
      {/* ... */}
    </div>
  )
}

if (msgtype === 'm.audio') {
  return (
    <div className="file-message audio-message">
      <SecureMedia src={url!} type="audio" className="message-audio" />
      {/* ... */}
    </div>
  )
}

if (msgtype === 'm.video') {
  return (
    <div className="file-message video-message">
      <SecureMedia src={url!} type="video" className="message-video" />
      {/* ... */}
    </div>
  )
}
```

## 技术细节

### Blob URL

```typescript
// 创建 Blob URL
const blob = await response.blob()
const blobUrl = URL.createObjectURL(blob)
// 结果: "blob:http://localhost:3000/550e8400-e29b-41d4-a716-446655440000"

// 使用
<img src={blobUrl} />

// 清理（重要！）
URL.revokeObjectURL(blobUrl)
```

### 内存管理

```typescript
useEffect(() => {
  let objectUrl: string | null = null

  const loadImage = async () => {
    // ...
    objectUrl = URL.createObjectURL(blob)
    setBlobUrl(objectUrl)
  }

  loadImage()

  // 组件卸载时清理
  return () => {
    if (objectUrl) {
      URL.revokeObjectURL(objectUrl)
    }
  }
}, [src])
```

### 加载状态

```typescript
const [loading, setLoading] = useState(true)
const [error, setError] = useState(false)

if (loading) {
  return <div className="loading-spinner">Loading...</div>
}

if (error) {
  return <div className="error-message">Failed to load image</div>
}

return <img src={blobUrl} alt={alt} className={className} />
```

## 优势

1. **安全性**: 所有文件访问都需要认证
2. **兼容性**: 适用于所有媒体类型
3. **用户体验**: 显示加载状态和错误提示
4. **内存管理**: 自动清理 Blob URL，防止内存泄漏

## 性能考虑

### 缓存

可以添加缓存机制避免重复下载：

```typescript
const mediaCache = new Map<string, string>()

const loadMedia = async () => {
  // 检查缓存
  if (mediaCache.has(src)) {
    setBlobUrl(mediaCache.get(src)!)
    return
  }

  // 下载并缓存
  const response = await fetch(src, { headers })
  const blob = await response.blob()
  const objectUrl = URL.createObjectURL(blob)

  mediaCache.set(src, objectUrl)
  setBlobUrl(objectUrl)
}
```

### 预加载

对于图片密集的聊天，可以预加载可见区域的图片：

```typescript
// 使用 Intersection Observer
const observer = new IntersectionObserver((entries) => {
  entries.forEach(entry => {
    if (entry.isIntersecting) {
      loadImage()
    }
  })
})
```

## 替代方案

### 方案 A: URL 参数传递 Token（不推荐）

```typescript
// 后端支持
GET /files/{id}?token=xxx

// 前端使用
<img src={`${url}?token=${token}`} />
```

**缺点**:
- Token 暴露在 URL 中
- 可能被日志记录
- 安全性较低

### 方案 B: Cookie 认证

```typescript
// 后端设置 HttpOnly Cookie
// 前端自动携带

<img src={url} />
```

**缺点**:
- 需要修改后端认证策略
- CSRF 风险
- 跨域问题

### 方案 C: 代理服务器

```typescript
// Nginx 配置
location /files/ {
  proxy_pass http://backend;
  proxy_set_header Authorization $http_authorization;
}
```

**缺点**:
- 需要额外配置
- 增加复杂度

## 测试

### 测试用例

1. **正常加载**
   - 发送图片 → 图片正常显示
   - 发送音频 → 音频可播放
   - 发送视频 → 视频可播放

2. **认证失败**
   - 清除 token → 显示错误提示
   - Token 过期 → 显示错误提示

3. **网络错误**
   - 断网 → 显示错误提示
   - 超时 → 显示错误提示

4. **内存泄漏**
   - 快速滚动消息列表 → 内存不增长
   - 切换聊天 → Blob URL 被清理

### 测试代码

```typescript
// 测试 Blob URL 清理
it('should cleanup blob URL on unmount', () => {
  const { unmount } = render(<SecureImage src="..." alt="test" />)

  // 验证 Blob URL 被创建
  expect(URL.createObjectURL).toHaveBeenCalled()

  // 卸载组件
  unmount()

  // 验证 Blob URL 被清理
  expect(URL.revokeObjectURL).toHaveBeenCalled()
})
```

## 总结

通过使用 Blob URL 方案，我们成功解决了媒体文件的认证问题：

- ✅ 所有文件访问都需要认证
- ✅ 支持图片、音频、视频
- ✅ 良好的加载状态提示
- ✅ 自动内存管理
- ✅ 不需要修改后端

这是一个安全、可靠、用户友好的解决方案。

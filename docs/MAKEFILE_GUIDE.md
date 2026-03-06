# Makefile 使用指南

## 基本命令

### 服务管理

```bash
# 查看所有可用命令
make help

# 构建所有服务
make build

# 启动所有服务（后台运行）
make run

# 停止所有服务
make stop

# 重启所有服务
make restart

# 开发模式启动（前台运行，显示日志）
make dev

# 查看运行状态
make ps

# 完全清理（删除容器和数据卷）
make clean
```

### 日志查看

```bash
# 查看所有服务日志
make logs

# 查看特定服务日志
make logs-api          # API Gateway
make logs-auth         # Auth Service
make logs-message      # Message Service
make logs-room         # Room Service
make logs-federation   # Federation Service
make logs-presence     # Presence Service
```

## 数据库管理

### 迁移命令

```bash
# 运行所有迁移（包括新的文件存储表）
make migrate-up

# 回滚所有迁移
make migrate-down

# 查看迁移状态（显示所有表）
make migrate-status

# 重置数据库（先 down 再 up）
make db-reset

# 打开数据库 shell
make db-shell
```

### 迁移说明

`migrate-up` 会按顺序执行以下迁移：
1. `001_create_users.up.sql` - 用户表
2. `002_create_direct_messages.up.sql` - 私聊消息表
3. `003_create_rooms.up.sql` - 房间表
4. `004_create_room_members.up.sql` - 房间成员表
5. `005_create_room_events.up.sql` - 房间事件表
6. `006_create_message_queue.up.sql` - 消息队列表
7. `007_create_presence.up.sql` - 在线状态表
8. `008_create_file_storage.up.sql` - **文件存储表（新增）**

## 文件存储管理

### 文件上传测试

```bash
# 运行完整的文件上传测试
make test-file-upload
```

这个命令会：
- 注册测试用户
- 上传测试文件
- 下载并验证文件
- 发送文件消息（私聊和群聊）

### 文件存储信息

```bash
# 查看文件存储信息
make file-storage-info
```

显示：
- 存储路径配置
- 已上传的文件列表
- 数据库中的文件记录（最近 10 条）

### 清理文件存储

```bash
# 清理所有上传的文件（需要确认）
make file-storage-clean
```

⚠️ **警告**: 这会删除所有上传的文件和数据库记录，操作不可逆！

## Web 客户端

### 开发

```bash
# 启动 Web 客户端开发服务器
make web-client
```

这会：
1. 安装依赖（如果需要）
2. 启动开发服务器（http://localhost:3000）
3. 启用热重载

### 构建

```bash
# 构建生产版本
make web-client-build
```

### 测试

```bash
# 运行 Web 客户端测试
make web-client-test
```

## 开发工具

### 代码质量

```bash
# 格式化 Go 代码
make fmt

# 运行 linter
make lint

# 下载/更新依赖
make deps
```

### 测试

```bash
# 运行所有测试
make test

# 运行测试并生成覆盖率报告
make test-coverage
```

### 本地构建

```bash
# 本地构建二进制文件（需要安装 Go）
make build-local
```

生成的二进制文件位于 `./bin/` 目录。

### Protocol Buffers

```bash
# 重新生成 gRPC 代码
make proto
```

需要安装：
- `protoc`
- `protoc-gen-go`
- `protoc-gen-go-grpc`

## 完整工作流示例

### 首次启动

```bash
# 1. 构建服务
make build

# 2. 启动服务
make run

# 3. 查看日志确认启动成功
make logs

# 4. 运行迁移（如果需要）
make migrate-up

# 5. 启动 Web 客户端
make web-client
```

### 开发流程

```bash
# 1. 修改代码后重新构建
make build

# 2. 重启服务
make restart

# 3. 查看日志
make logs-api

# 4. 运行测试
make test
```

### 测试文件上传功能

```bash
# 1. 确保服务运行
make ps

# 2. 运行文件上传测试
make test-file-upload

# 3. 查看文件存储信息
make file-storage-info

# 4. 在 Web 客户端测试
make web-client
# 然后在浏览器中测试文件上传
```

### 数据库维护

```bash
# 1. 查看当前表结构
make migrate-status

# 2. 打开数据库 shell 查询数据
make db-shell
# 然后执行 SQL 查询

# 3. 如果需要重置数据库
make db-reset

# 4. 清理文件存储
make file-storage-clean
```

### 完全清理和重启

```bash
# 1. 停止并删除所有数据
make clean

# 2. 重新构建
make build

# 3. 启动服务
make run

# 4. 运行迁移
make migrate-up
```

## 常见问题

### 迁移失败

如果 `make migrate-up` 失败：

```bash
# 1. 检查数据库是否运行
make ps

# 2. 查看数据库日志
docker-compose logs postgres

# 3. 手动连接数据库检查
make db-shell

# 4. 如果需要，重置数据库
make db-reset
```

### 文件上传测试失败

```bash
# 1. 检查 API Gateway 是否运行
make logs-api

# 2. 检查文件存储目录权限
ls -la data/files

# 3. 查看文件存储信息
make file-storage-info

# 4. 检查数据库中的文件记录
make db-shell
# SELECT * FROM file_storage;
```

### 端口冲突

如果端口被占用：

```bash
# 1. 停止服务
make stop

# 2. 检查端口占用
lsof -i :8080  # API Gateway
lsof -i :5432  # PostgreSQL
lsof -i :6379  # Redis

# 3. 修改 docker-compose.yml 中的端口映射

# 4. 重新启动
make run
```

## 环境变量

可以通过环境变量覆盖默认配置：

```bash
# 文件存储路径
FILE_STORAGE_PATH=/custom/path make run

# 数据库配置
DB_HOST=localhost DB_PORT=5432 make migrate-up

# JWT 密钥
JWT_SECRET=your-secret-key make run
```

## 提示

1. **开发时使用 `make dev`**: 可以实时查看日志
2. **定期运行 `make test`**: 确保代码质量
3. **使用 `make help`**: 查看所有可用命令
4. **备份数据**: 在运行 `make clean` 前备份重要数据
5. **查看日志**: 遇到问题先查看相关服务的日志

## 快捷命令组合

```bash
# 完整重启
make stop && make build && make run

# 重置并测试
make db-reset && make test-file-upload

# 开发环境启动
make run && make web-client

# 查看所有状态
make ps && make migrate-status && make file-storage-info
```

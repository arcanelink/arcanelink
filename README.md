# 分布式IM通信协议 / Distributed IM Protocol

基于Matrix协议改进的分布式即时通讯协议，简化设计、提升性能。

An improved distributed instant messaging protocol based on Matrix, with simplified design and enhanced performance.

## 特性 / Features

- **私聊不使用Room** / **Private chat without Room**: 消息直接点对点路由，无需创建Room对象 / Direct point-to-point message routing
- **群聊使用Room** / **Group chat with Room**: 保留Room概念用于群组通信 / Retains Room concept for group communication
- **HTTP长轮询** / **HTTP long polling**: 使用标准HTTP协议，更简单、更兼容 / Uses standard HTTP, simpler and more compatible
- **可选加密** / **Optional encryption**: 不强制端到端加密，降低实现复杂度 / E2EE not mandatory, reducing complexity
- **服务器联邦** / **Server federation**: 去中心化架构，支持跨服务器通信 / Decentralized architecture with cross-server communication

## 与Matrix的主要区别 / Main Differences from Matrix

| 特性 / Feature | Matrix | 本协议 / This Protocol |
|----------------|--------|------------------------|
| 私聊机制 / Private Chat | 使用Room / Uses Room | 直接点对点 / Direct P2P |
| 群聊机制 / Group Chat | 使用Room / Uses Room | 使用Room / Uses Room |
| 通信方式 / Communication | WebSocket/HTTP | HTTP长轮询 / HTTP Long Polling |
| 加密要求 / Encryption | 支持E2EE / Supports E2EE | 可选 / Optional |
| 实现复杂度 / Complexity | 较高 / Higher | 简化 / Simplified |

## 文档 / Documentation

完整的协议规范文档位于 `spec/` 目录，提供中英文双语版本。

Complete protocol specification is available in the `spec/` directory in both Chinese and English.

### 中文文档 / Chinese Documentation

- [协议概述](./spec/zh-CN/01-overview.md) - 设计目标、核心特性
- [架构设计](./spec/zh-CN/02-architecture.md) - 系统架构、双通道模型、长轮询机制
- [客户端API](./spec/zh-CN/03-client-api.md) - 客户端接口规范
- [联邦API](./spec/zh-CN/04-federation-api.md) - 服务器间通信接口
- [消息格式](./spec/zh-CN/05-message-format.md) - 消息和事件数据结构

### English Documentation

- [Protocol Overview](./spec/en/01-overview.md) - Design goals, core features
- [Architecture Design](./spec/en/02-architecture.md) - System architecture, dual-channel model, long polling
- [Client API](./spec/en/03-client-api.md) - Client interface specification
- [Federation API](./spec/en/04-federation-api.md) - Inter-server communication interface
- [Message Format](./spec/en/05-message-format.md) - Message and event data structures

## 快速开始 / Quick Start

### 基本概念 / Basic Concepts

**用户ID / User ID**: `@username:domain.com`
- 示例 / Example: `@alice:example.com`

**Room ID**: `!roomid:domain.com`
- 示例 / Example: `!abc123:example.com`

### 消息流程 / Message Flow

**私聊 / Private Chat**:
```
发送方客户端 → 发送方服务器 → 接收方服务器 → 接收方客户端
Sender Client → Sender Server → Recipient Server → Recipient Client
```

**群聊 / Group Chat**:
```
发送方客户端 → Room服务器 → 成员服务器 → 成员客户端
Sender Client → Room Server → Member Servers → Member Clients
```

## 协议层次 / Protocol Layers

```
客户端应用层 / Client Application Layer
    ↓
客户端API层 / Client API Layer (HTTP长轮询 / HTTP Long Polling + REST API)
    ↓
Homeserver (用户管理、消息路由、存储 / User Management, Message Routing, Storage)
    ↓
联邦协议层 / Federation Protocol Layer (服务器间通信 / Inter-server Communication)
    ↓
传输层 / Transport Layer (HTTP/1.1 或 / or HTTP/2)
```

## API示例 / API Examples

### 客户端同步 / Client Sync (Long Polling)

```http
GET /_api/v1/sync?since=token&timeout=30000
Authorization: Bearer <access_token>
```

### 发送私聊消息 / Send Direct Message

```http
POST /_api/v1/send_direct
Authorization: Bearer <access_token>

{
  "recipient": "@bob:example.com",
  "content": {
    "msgtype": "m.text",
    "body": "Hello"
  }
}
```

### 发送群聊消息 / Send Room Message

```http
POST /_api/v1/send_room
Authorization: Bearer <access_token>

{
  "room_id": "!abc123:example.com",
  "content": {
    "msgtype": "m.text",
    "body": "Hello everyone"
  }
}
```

## 实现建议 / Implementation Recommendations

### 最小实现 / Minimal Implementation

必须实现的核心功能 / Core features that must be implemented:

1. 用户认证 / User authentication
2. HTTP长轮询同步 / HTTP long polling sync
3. 私聊消息发送和接收 / Direct message send/receive
4. 基本的联邦消息转发 / Basic federation message forwarding

### 完整实现 / Complete Implementation

建议实现的完整功能 / Recommended full features:

1. Room创建和管理 / Room creation and management
2. 成员邀请和权限 / Member invitation and permissions
3. 在线状态管理 / Presence management
4. 消息历史查询 / Message history query
5. 多媒体消息支持 / Multimedia message support

## 技术栈建议 / Recommended Tech Stack

### 服务器端 / Server-side

- **语言 / Language**: Go, Rust, Node.js, Python
- **数据库 / Database**: PostgreSQL, MySQL, MongoDB
- **缓存 / Cache**: Redis (用于消息队列和在线状态 / for message queue and presence)
- **Web框架 / Web Framework**: 支持长轮询的HTTP框架 / HTTP framework with long polling support

### 客户端 / Client-side

- **Web**: JavaScript/TypeScript + React/Vue
- **移动端 / Mobile**: Swift (iOS), Kotlin (Android), Flutter
- **桌面端 / Desktop**: Electron, Qt

## 性能指标 / Performance Metrics

- **长轮询超时 / Long polling timeout**: 30秒 / 30 seconds
- **单服务器并发连接 / Concurrent connections per server**: 10,000+
- **消息延迟 / Message latency**: < 100ms (同服务器 / same server), < 500ms (跨服务器 / cross-server)
- **消息大小限制 / Message size limit**: 1MB

## 安全考虑 / Security Considerations

- **传输加密 / Transport encryption**: 强制HTTPS (生产环境 / production)
- **认证 / Authentication**: Bearer Token (JWT推荐 / JWT recommended)
- **速率限制 / Rate limiting**: 防止滥用 / Prevent abuse
- **消息验证 / Message validation**: 防止注入攻击 / Prevent injection attacks

## 开发路线图 / Roadmap

- [x] 协议规范设计 / Protocol specification design
- [ ] 参考实现 / Reference implementation
  - [ ] 服务器端 / Server-side
  - [ ] 客户端SDK / Client SDK
- [ ] 测试工具 / Testing tools
- [ ] 性能基准测试 / Performance benchmarks
- [ ] 生产部署指南 / Production deployment guide

## 贡献 / Contributing

欢迎贡献代码、文档改进和问题反馈。

Contributions for code, documentation improvements, and issue reports are welcome.

## 许可证 / License

待定 / To be determined

## 联系方式 / Contact

项目讨论和问题反馈请使用GitHub Issues。

For project discussions and issue reports, please use GitHub Issues.

---

**版本 / Version**: 1.0
**状态 / Status**: 草案 / Draft
**最后更新 / Last Updated**: 2026-03-02

# Distributed IM Protocol - Overview

## Version Information

- Protocol Version: 1.0
- Document Date: 2026-03-02
- Status: Draft

## Protocol Introduction

This protocol is a distributed instant messaging protocol improved from the Matrix protocol, aiming to simplify Matrix's complexity while maintaining its distributed architecture advantages.

### Design Goals

1. **Distributed Architecture**: Maintain server federation model, anyone can run their own server
2. **Simplified Design**: Remove unnecessary complexity, improve performance and usability
3. **Dual-Channel Model**: Private chat and group chat use different processing mechanisms
4. **Easy to Implement**: Reduce implementation difficulty, use standard HTTP protocol

### Core Features

- **Private Chat Without Room**: Private messages are routed point-to-point directly, no need to create Room objects
- **Group Chat Uses Room**: Group chat retains the Room concept, maintaining member lists and history
- **HTTP Long Polling**: Use HTTP long polling instead of WebSocket, simpler and more compatible
- **Optional Encryption**: End-to-end encryption not mandatory, reducing implementation complexity
- **Server Federation**: Support cross-server communication, users can communicate with users on any server

## Main Differences from Matrix

| Feature | Matrix | This Protocol |
|---------|--------|---------------|
| Private Chat | Uses Room | Direct point-to-point routing |
| Group Chat | Uses Room | Uses Room |
| Communication | WebSocket/HTTP | HTTP Long Polling |
| Encryption | Supports E2EE | Optional, not mandatory |
| Message Model | Unified event model | Dual-channel model |

## Basic Concepts

### User Identifier

User ID format: `@username:domain.com`

Example: `@alice:example.com`

### Server Identifier

Server ID is the domain name: `domain.com`

### Room Identifier

Room ID format: `!roomid:domain.com`

Example: `!abc123:example.com`

## Protocol Layers

```
Client Application Layer
    ↓
Client API Layer (HTTP Long Polling + REST API)
    ↓
Homeserver (User Management, Message Routing, Storage)
    ↓
Federation Protocol Layer (Inter-server Communication)
    ↓
Transport Layer (HTTP/1.1 or HTTP/2)
```

## Message Flow

### Private Message Flow
```
Sender Client → Sender HS → Recipient HS → Recipient Client
```

### Group Message Flow
```
Sender Client → Room HS → Member HSs → Member Clients
```

## Document Structure

1. **01-overview.md**: Protocol overview
2. **02-architecture.md**: Architecture design details
3. **03-client-api.md**: Client API specification
4. **04-federation-api.md**: Federation API specification
5. **05-message-format.md**: Message format specification

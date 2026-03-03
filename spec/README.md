# Distributed IM Protocol Specification

This directory contains the complete specification for a distributed instant messaging protocol improved from Matrix.

## Languages

- [English](./en/) - English version
- [中文](./zh-CN/) - Chinese version

## Documents

1. **01-overview.md** - Protocol overview, design goals, and basic concepts
2. **02-architecture.md** - Architecture design, dual-channel model, long polling mechanism
3. **03-client-api.md** - Client API specification
4. **04-federation-api.md** - Federation API specification for inter-server communication
5. **05-message-format.md** - Message and event format specification

## Key Features

- **Private chat without Room**: Direct point-to-point routing for private messages
- **Group chat with Room**: Maintains member lists and history for group conversations
- **HTTP long polling**: Simpler and more compatible than WebSocket
- **Optional encryption**: Not mandatory, reducing implementation complexity
- **Server federation**: Decentralized architecture supporting cross-server communication

## Quick Start

Start by reading the overview document in your preferred language:
- English: [en/01-overview.md](./en/01-overview.md)
- 中文: [zh-CN/01-overview.md](./zh-CN/01-overview.md)

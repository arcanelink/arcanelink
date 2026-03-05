# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ArcaneLink is a distributed instant messaging protocol implementation inspired by Matrix but with simplified design. It uses a microservices architecture with Go backend services communicating via gRPC, and a React/TypeScript web client.

**Key Design Principles:**
- Private chat uses direct messages (no Room concept)
- Group chat uses Room concept
- HTTP long polling for real-time updates (not WebSocket)
- Optional E2EE (not mandatory)
- Microservices architecture with API Gateway pattern

## Development Commands

### Backend Services

```bash
# Start all services (Docker Compose)
make run
# or
docker-compose up -d

# Stop all services
make stop

# View logs
make logs                  # All services
make logs-api             # API Gateway only
make logs-auth            # Auth service only
make logs-message         # Message service only
make logs-room            # Room service only

# Rebuild services
make build

# Clean up (removes volumes)
make clean
```

### Testing

```bash
# Run all tests
make test

# Run tests with coverage report
make test-coverage

# Run specific service tests
go test -v ./internal/auth/service/...
go test -v ./internal/message/service/...
go test -v ./internal/room/service/...
```

### Local Development

```bash
# Build binaries locally (requires Go 1.24+)
make build-local

# Format code
make fmt

# Run linter (requires golangci-lint)
make lint

# Download/update dependencies
make deps
```

### Protocol Buffers

```bash
# Regenerate gRPC code from .proto files
make proto
```

### Web Client

```bash
cd web-client
npm install
npm run dev      # Development server at http://localhost:3000
npm run build    # Production build
npm run lint     # ESLint
```

## Architecture

### Microservices Structure

The system consists of 6 microservices:

1. **API Gateway** (port 8080) - HTTP/REST interface for clients
   - Location: `cmd/api-gateway/`, `internal/api-gateway/`
   - Routes client requests to appropriate services via gRPC
   - Handles authentication middleware and rate limiting
   - CORS enabled for web client

2. **Auth Service** (gRPC port 50051)
   - Location: `cmd/auth/`, `internal/auth/`
   - User registration, login, JWT generation/validation
   - Password hashing with bcrypt

3. **Message Service** (gRPC port 50052)
   - Location: `cmd/message/`, `internal/message/`
   - Direct message send/receive and history
   - Long polling manager for real-time updates
   - Redis pub/sub for message notifications

4. **Room Service** (gRPC port 50053)
   - Location: `cmd/room/`, `internal/room/`
   - Room creation, management, member management
   - Room message handling and event distribution

5. **Presence Service** (gRPC port 50054)
   - Location: `cmd/presence/`, `internal/presence/`
   - Online status tracking
   - Automatic cleanup of stale presence

6. **Federation Service** (gRPC port 50055, HTTP port 8448)
   - Location: `cmd/federation/`, `internal/federation/`
   - Cross-server communication (basic structure)
   - Server discovery and message forwarding

### Key Components

**Long Polling Manager** (`internal/message/longpoll/manager.go`):
- Manages client subscriptions with 30-second timeout
- Notifies clients of new direct messages and room events
- Automatic cleanup of expired subscriptions

**API Routes** (`internal/api-gateway/router/router.go`):
- Public: `/auth/login`, `/auth/register`
- Protected: `/sync`, `/send_direct`, `/send_room`, `/rooms/*`, `/direct_history`
- All protected routes require Bearer token authentication

**Database Models** (`pkg/models/`):
- User, DirectMessage, Room, RoomMember, RoomEvent, Presence
- PostgreSQL for persistent storage

**Configuration** (`pkg/config/config.go`):
- Environment variable based configuration
- Service addresses, database, Redis, JWT settings

### Communication Flow

```
Client (HTTP) → API Gateway (HTTP/REST)
                    ↓ (gRPC)
              Microservices (Auth/Message/Room/Presence)
                    ↓
              PostgreSQL + Redis
```

**Long Polling Sync:**
1. Client calls `GET /_api/v1/sync?since=token&timeout=30000`
2. API Gateway forwards to Message Service
3. Message Service subscribes client to long poll manager
4. When new messages arrive, notification sent immediately
5. Otherwise, returns empty response after 30s timeout

## Database

**PostgreSQL Schema:**
- `users` - User accounts
- `direct_messages` - P2P messages
- `rooms` - Group chat rooms
- `room_members` - Room membership
- `room_events` - Room messages and events
- `message_queue` - Pending messages
- `presence` - User online status

**Migrations:** Located in `migrations/` directory, automatically run on container startup.

## User ID Format

User IDs follow Matrix convention: `@username:domain`
- Example: `@alice:localhost`
- Domain is configured via `SERVER_DOMAIN` environment variable

Room IDs: `!roomid:domain`
- Example: `!abc123:localhost`

## Testing Notes

- Unit tests exist for auth, message, and room services
- Tests use in-memory or mock implementations
- Run tests before committing changes
- Test files: `*_test.go`

## Common Patterns

**Adding a new API endpoint:**
1. Define gRPC service in `pkg/proto/*.proto`
2. Run `make proto` to generate Go code
3. Implement handler in `internal/[service]/handler/grpc_handler.go`
4. Implement service logic in `internal/[service]/service/`
5. Add HTTP route in `internal/api-gateway/router/router.go`
6. Add handler method in `internal/api-gateway/handler/api_handler.go`

**Adding a new database table:**
1. Create migration files in `migrations/`: `NNN_name.up.sql` and `NNN_name.down.sql`
2. Add model struct in `pkg/models/`
3. Add repository methods in `internal/[service]/repository/`
4. Restart services to apply migrations

## Important Notes

- JWT secret is configured via `JWT_SECRET` environment variable (change in production)
- Long polling timeout is 30 seconds (hardcoded in multiple places)
- Redis is used for message pub/sub and presence caching
- All services use structured logging with zap
- CORS is enabled for all origins (restrict in production)
- Rate limiting: 100 requests/second per user (configurable in router)

## Module Path

Go module: `github.com/arcane/arcanelink`

When adding imports, use this module path prefix.

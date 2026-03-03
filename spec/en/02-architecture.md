# Architecture Design

## Overall Architecture

### System Components

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  Client A    │     │  Client B    │     │  Client C    │
│ @alice:srv-a │     │ @bob:srv-b   │     │ @carol:srv-c │
└──────┬───────┘     └──────┬───────┘     └──────┬───────┘
       │                    │                    │
       │ HTTP Long Polling  │                    │
       │                    │                    │
┌──────▼───────┐     ┌──────▼───────┐     ┌──────▼───────┐
│ Homeserver A │◄───►│ Homeserver B │◄───►│ Homeserver C │
│  srv-a.com   │     │  srv-b.com   │     │  srv-c.com   │
└──────────────┘     └──────────────┘     └──────────────┘
   Federation           Federation
```

### Homeserver Architecture

```
┌─────────────────────────────────────────┐
│           Homeserver                    │
├─────────────────────────────────────────┤
│  Client API Layer                       │
│  - Long polling management              │
│  - Message sending interface            │
│  - Room management interface            │
├─────────────────────────────────────────┤
│  Business Logic Layer                   │
│  - User management                      │
│  - Direct message routing               │
│  - Room event processing                │
│  - Presence management                  │
├─────────────────────────────────────────┤
│  Federation Protocol Layer              │
│  - Server discovery                     │
│  - Message forwarding                   │
│  - Room synchronization                 │
├─────────────────────────────────────────┤
│  Storage Layer                          │
│  - User data                            │
│  - Direct messages                      │
│  - Room data                            │
│  - Message queue                        │
└─────────────────────────────────────────┘
```

## Dual-Channel Message Model

### Direct Channel

Features:
- No Room object creation
- Direct message routing
- Lightweight processing
- Both servers store separately

Processing flow:
```
1. Alice sends message to Bob
2. Alice's client POSTs to Server-A
3. Server-A validates and stores message
4. Server-A forwards to Server-B via federation
5. Server-B stores to Bob's message queue
6. Bob's long polling request receives response
```

### Room Channel

Features:
- Uses Room object
- Maintains member list
- Preserves complete history
- Room's homeserver is the primary node

Processing flow:
```
1. Alice sends message in Room
2. Alice's client POSTs to Room's Server
3. Server validates permissions and stores event
4. Server forwards to all members' Homeservers in parallel
5. Each member's Homeserver stores to message queue
6. Each member's long polling request receives response
```

## HTTP Long Polling Mechanism

### How It Works

```
Client                          Server
  │                              │
  ├─ GET /sync?since=token ─────►│
  │                              │ Check for new messages
  │                              │ No messages, hold request
  │                              │ Waiting...
  │                              │ New message arrives!
  │◄──── Return messages ────────┤
  │                              │
  ├─ GET /sync?since=new_token ─►│
  │                              │ Immediately start next request
```

### Parameters

- `since`: Token from last sync, marks read position
- `timeout`: Timeout in milliseconds, recommended 30000 (30 seconds)

### Server-side Implementation

```
Pseudocode:
function handleSync(user_id, since_token, timeout):
    messages = getNewMessages(user_id, since_token)

    if messages.length > 0:
        return {
            next_token: generateToken(),
            messages: messages
        }

    # No new messages, wait
    wait_result = waitForMessages(user_id, timeout)

    if wait_result.has_messages:
        return {
            next_token: generateToken(),
            messages: wait_result.messages
        }
    else:
        # Timeout
        return {
            next_token: since_token,
            messages: []
        }
```

## Server Federation

### Server Discovery

Method 1: DNS SRV Record
```
_matrix-fed._tcp.example.com. 3600 IN SRV 10 0 8448 matrix.example.com.
```

Method 2: .well-known
```
GET https://example.com/.well-known/matrix/server
{
  "m.server": "matrix.example.com:8448"
}
```

### Federation Connection

- Protocol: HTTPS
- Port: 8448 (default)
- Path prefix: `/_fed/v1/`

### Message Forwarding Flow

Direct message forwarding:
```
Server-A                    Server-B
    │                          │
    ├─ POST /_fed/v1/send_direct ─►│
    │  {sender, recipient, ...}    │
    │                          │ Validate
    │                          │ Store
    │◄──── 200 OK ─────────────┤
```

Room message forwarding:
```
Room Server                Member Server
    │                          │
    ├─ POST /_fed/v1/send_room ──►│
    │  {room_id, event, ...}      │
    │                          │ Validate
    │                          │ Store
    │◄──── 200 OK ─────────────┤
```

## Data Storage Model

### User Data

```sql
CREATE TABLE users (
    user_id VARCHAR(255) PRIMARY KEY,
    username VARCHAR(100),
    password_hash VARCHAR(255),
    created_at TIMESTAMP,
    profile_data JSON
);
```

### Direct Messages

```sql
CREATE TABLE direct_messages (
    msg_id VARCHAR(100) PRIMARY KEY,
    sender VARCHAR(255),
    recipient VARCHAR(255),
    content JSON,
    timestamp BIGINT,
    INDEX idx_recipient_time (recipient, timestamp)
);
```

### Room Data

```sql
CREATE TABLE rooms (
    room_id VARCHAR(255) PRIMARY KEY,
    creator VARCHAR(255),
    name VARCHAR(200),
    topic TEXT,
    created_at TIMESTAMP
);

CREATE TABLE room_members (
    room_id VARCHAR(255),
    user_id VARCHAR(255),
    joined_at TIMESTAMP,
    PRIMARY KEY (room_id, user_id)
);

CREATE TABLE room_events (
    event_id VARCHAR(100) PRIMARY KEY,
    room_id VARCHAR(255),
    sender VARCHAR(255),
    event_type VARCHAR(50),
    content JSON,
    timestamp BIGINT,
    INDEX idx_room_time (room_id, timestamp)
);
```

### Message Queue

```sql
CREATE TABLE message_queue (
    queue_id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(255),
    message_type ENUM('direct', 'room'),
    message_id VARCHAR(100),
    created_at TIMESTAMP,
    INDEX idx_user_created (user_id, created_at)
);
```

## Presence Management

### Status Determination

- Online: Had sync request within last 60 seconds
- Offline: No sync request for over 60 seconds

### Status Storage

```sql
CREATE TABLE presence (
    user_id VARCHAR(255) PRIMARY KEY,
    status ENUM('online', 'offline'),
    last_active TIMESTAMP
);
```

### Status Updates

Update on each sync request:
```sql
UPDATE presence
SET status = 'online', last_active = NOW()
WHERE user_id = ?;
```

Periodic cleanup (run every minute):
```sql
UPDATE presence
SET status = 'offline'
WHERE last_active < NOW() - INTERVAL 60 SECOND;
```

## Performance Optimization

### Message Queue Optimization

Use in-memory queue + persistence:
- New messages enter memory queue first
- Asynchronously write to database
- Read from memory queue when user is online
- Read historical messages from database

### Long Polling Optimization

- Use event notification mechanism (epoll/kqueue)
- Avoid polling database
- Single server supports tens of thousands of concurrent connections

### Federation Caching

- Cache server discovery results
- Cache user's homeserver
- Reduce DNS queries

## Scalability Design

### Horizontal Scaling

- Multiple Homeserver instances
- Load balancer distributes requests
- Shared database or distributed storage

### Message Sharding

- Shard by user ID hash
- Each shard processes independently
- Improve concurrency capability

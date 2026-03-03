# Federation API Specification

## Basic Information

- Protocol: HTTPS
- Port: 8448 (default federation port)
- Path prefix: `/_fed/v1/`
- Authentication: Server signature (optional implementation)

## Server Discovery

### Method 1: DNS SRV Record

```
Query: _matrix-fed._tcp.example.com

Record format:
_matrix-fed._tcp.example.com. 3600 IN SRV 10 0 8448 matrix.example.com.

Parameters:
- 10: Priority
- 0: Weight
- 8448: Port
- matrix.example.com: Target host
```

### Method 2: .well-known

```
GET https://example.com/.well-known/matrix/server

Response:
{
  "m.server": "matrix.example.com:8448"
}
```

### Server Connection Flow

```
1. Parse target user ID: @bob:example.com
2. Extract domain: example.com
3. Query DNS SRV record or .well-known
4. Get actual server address and port
5. Establish HTTPS connection
```

## Direct Message Forwarding

### Send Direct Message

```
POST /_fed/v1/send_direct

Request body:
{
  "sender": "@alice:server-a.com",
  "recipient": "@bob:server-b.com",
  "msg_id": "msg_abc123",
  "content": {
    "msgtype": "m.text",
    "body": "Hello"
  },
  "timestamp": 1234567890000
}

Response:
{
  "success": true,
  "received_at": 1234567890100
}

Error response:
{
  "error": "USER_NOT_FOUND",
  "message": "Recipient does not exist"
}
```

### Validation Flow

Receiving server should validate:
1. Recipient belongs to this server
2. Sender format is correct
3. msg_id is unique (prevent duplicates)
4. Timestamp is reasonable (not too old or too new)

## Room Message Forwarding

### Send Room Event

```
POST /_fed/v1/send_room

Request body:
{
  "room_id": "!abc123:server-a.com",
  "event_id": "evt_xyz789",
  "sender": "@alice:server-a.com",
  "event_type": "m.room.message",
  "content": {
    "msgtype": "m.text",
    "body": "Hello everyone"
  },
  "timestamp": 1234567890000
}

Response:
{
  "success": true,
  "received_at": 1234567890100
}

Error response:
{
  "error": "NOT_IN_ROOM",
  "message": "No local users in this room"
}
```

### Validation Flow

Receiving server should validate:
1. This server has users in the room
2. Sender has permission to send messages
3. event_id is unique
4. room_id format is correct

## User Query

### Query User Existence

```
GET /_fed/v1/query_user/@user:domain.com

Response:
{
  "user_id": "@user:domain.com",
  "exists": true,
  "presence": "online",
  "last_active": 1234567890000
}

User not found:
{
  "user_id": "@user:domain.com",
  "exists": false
}
```

## Room Operations

### Invite User to Room

```
POST /_fed/v1/room/invite

Request body:
{
  "room_id": "!abc123:server-a.com",
  "inviter": "@alice:server-a.com",
  "invitee": "@bob:server-b.com",
  "room_name": "My Group",
  "room_topic": "Discussion"
}

Response:
{
  "success": true
}
```

### Join Room

```
POST /_fed/v1/room/join

Request body:
{
  "room_id": "!abc123:server-a.com",
  "user_id": "@bob:server-b.com"
}

Response:
{
  "success": true,
  "room_state": {
    "name": "My Group",
    "topic": "Discussion",
    "members": [
      "@alice:server-a.com",
      "@carol:server-c.com"
    ]
  }
}

Error response:
{
  "error": "ROOM_NOT_FOUND",
  "message": "Room does not exist"
}
```

### Leave Room

```
POST /_fed/v1/room/leave

Request body:
{
  "room_id": "!abc123:server-a.com",
  "user_id": "@bob:server-b.com"
}

Response:
{
  "success": true
}
```

### Get Room State

```
GET /_fed/v1/room/{room_id}/state

Response:
{
  "room_id": "!abc123:server-a.com",
  "name": "My Group",
  "topic": "Discussion",
  "creator": "@alice:server-a.com",
  "created_at": 1234567890000,
  "member_count": 5
}
```

### Get Room Members

```
GET /_fed/v1/room/{room_id}/members

Response:
{
  "members": [
    {
      "user_id": "@alice:server-a.com",
      "joined_at": 1234567890000
    },
    {
      "user_id": "@bob:server-b.com",
      "joined_at": 1234567891000
    }
  ]
}
```

## Server Health Check

### Health Check Endpoint

```
GET /_fed/v1/health

Response:
{
  "status": "ok",
  "version": "1.0",
  "server_name": "example.com",
  "timestamp": 1234567890000
}
```

## Error Handling

### Error Response Format

```json
{
  "error": "ERROR_CODE",
  "message": "Human readable error message"
}
```

### Common Error Codes

- `USER_NOT_FOUND`: User does not exist
- `ROOM_NOT_FOUND`: Room does not exist
- `NOT_IN_ROOM`: User not in room
- `FORBIDDEN`: No permission to perform operation
- `INVALID_REQUEST`: Invalid request format
- `SERVER_ERROR`: Internal server error
- `TIMEOUT`: Request timeout

### HTTP Status Codes

- 200: Success
- 400: Bad request
- 403: Forbidden
- 404: Not found
- 500: Server error
- 503: Service unavailable

## Retry Mechanism

### Message Forwarding Retry

When message forwarding fails:
1. Retry immediately once
2. If still fails, wait 5 seconds and retry
3. If still fails, wait 30 seconds and retry
4. If still fails, wait 5 minutes and retry
5. Maximum 5 retries, then give up

### Exponential Backoff

```
retry_interval = min(base_delay * 2^retry_count, max_delay)

base_delay = 5 seconds
max_delay = 5 minutes
```

## Security Considerations

### Prevent Message Forgery

Recommended implementation:
- Server signature verification
- TLS certificate verification
- Message ID deduplication

### Prevent Abuse

- Rate limiting: Maximum 100 requests per server per second
- Connection limiting: Maximum 10 concurrent connections per server
- Message size limit: Maximum 1MB per message

### Prevent Loops

- Check sender's domain when forwarding messages
- Don't forward messages from target server
- Maintain message ID history to prevent duplicate processing

## Performance Optimization

### Connection Reuse

- Use HTTP/2 or HTTP/1.1 Keep-Alive
- Maintain connection pool between servers
- Batch send messages

### Parallel Forwarding

When forwarding room messages:
- Send to all member servers in parallel
- Don't wait for all responses
- Handle failed retries asynchronously

### Caching

- Cache server discovery results (TTL: 1 hour)
- Cache room state (TTL: 5 minutes)
- Cache user's homeserver (TTL: 1 day)

## Implementation Recommendations

### Minimal Implementation

Must implement endpoints:
- `POST /_fed/v1/send_direct`
- `POST /_fed/v1/send_room`
- `GET /_fed/v1/query_user/{user_id}`
- `GET /_fed/v1/health`

### Complete Implementation

Recommend implementing all endpoints for full functionality.

### Testing

Provide testing tools to verify:
- Server discovery
- Message forwarding
- Room operations
- Error handling
- Retry mechanism

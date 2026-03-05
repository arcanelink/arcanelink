# Client API Specification

## Basic Information

- Protocol: HTTPS
- Port: 443 (standard HTTPS port)
- Path prefix: `/_api/v1/`
- Authentication: Bearer Token

## Authentication

### Login

```
POST /_api/v1/login

Request body:
{
  "user_id": "@alice:example.com",
  "password": "password123"
}

Response:
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user_id": "@alice:example.com",
  "expires_in": 86400
}
```

### Logout

```
POST /_api/v1/logout

Headers:
Authorization: Bearer <access_token>

Response:
{
  "success": true
}
```

## Message Sync (Long Polling)

### Sync Endpoint

```
GET /_api/v1/sync?since={token}&timeout={ms}

Headers:
Authorization: Bearer <access_token>

Parameters:
- since: Token from last sync (optional for first request)
- timeout: Timeout in milliseconds, recommended 30000

Response:
{
  "next_token": "t_12345_67890",
  "direct_messages": [
    {
      "msg_id": "msg_001",
      "sender": "@bob:example.com",
      "content": {
        "msgtype": "m.text",
        "body": "Hello"
      },
      "timestamp": 1234567890000
    }
  ],
  "room_events": [
    {
      "event_id": "evt_001",
      "room_id": "!abc:example.com",
      "sender": "@carol:example.com",
      "event_type": "m.room.message",
      "content": {
        "msgtype": "m.text",
        "body": "Hi everyone"
      },
      "timestamp": 1234567891000
    }
  ]
}
```

## Direct Messages

### Send Direct Message

```
POST /_api/v1/messages

Headers:
Authorization: Bearer <access_token>

Request body:
{
  "recipient": "@bob:example.com",
  "content": {
    "msgtype": "m.text",
    "body": "Hello Bob"
  }
}

Response:
{
  "msg_id": "msg_12345",
  "timestamp": 1234567890000
}
```

### Get Direct Message History

```
GET /_api/v1/messages?peer={user_id}&limit={n}&before={token}

Headers:
Authorization: Bearer <access_token>

Parameters:
- peer: Peer user ID
- limit: Number of messages to return (default 50, max 100)
- before: Get messages before this token (optional)

Response:
{
  "messages": [
    {
      "msg_id": "msg_100",
      "sender": "@alice:example.com",
      "recipient": "@bob:example.com",
      "content": {
        "msgtype": "m.text",
        "body": "Message text"
      },
      "timestamp": 1234567890000
    }
  ],
  "prev_token": "t_prev_token",
  "has_more": true
}
```

## Room Management

### Create Room

```
POST /_api/v1/rooms

Headers:
Authorization: Bearer <access_token>

Request body:
{
  "name": "My Group Chat",
  "topic": "Discussion about project",
  "invite": ["@bob:example.com", "@carol:example.com"]
}

Response:
{
  "room_id": "!abc123:example.com",
  "created_at": 1234567890000
}
```

### Join Room

```
POST /_api/v1/rooms/{room_id}/members

Headers:
Authorization: Bearer <access_token>

Response:
{
  "success": true,
  "joined_at": 1234567890000
}
```

### Leave Room

```
DELETE /_api/v1/rooms/{room_id}/members

Headers:
Authorization: Bearer <access_token>

Response:
{
  "success": true
}
```

**Note**: Room creators cannot leave rooms, they must delete the room instead.

### Invite User

```
POST /_api/v1/rooms/{room_id}/members

Headers:
Authorization: Bearer <access_token>

Request body:
{
  "user_id": "@dave:example.com"
}

Response:
{
  "success": true
}
```

**Note**: The system validates that the invited user exists and is not already a room member.

### Delete Room

```
DELETE /_api/v1/rooms/{room_id}

Headers:
Authorization: Bearer <access_token>

Response:
{
  "success": true
}
```

**Permission**: Only the room creator can delete the room.

**Behavior**: Deleting a room automatically removes all members.

### Get Room State

```
GET /_api/v1/rooms/{room_id}

Headers:
Authorization: Bearer <access_token>

Response:
{
  "room_id": "!abc123:example.com",
  "name": "My Group Chat",
  "topic": "Discussion about project",
  "creator": "@alice:example.com",
  "created_at": 1234567890000
}
```

### Get Room Members

```
GET /_api/v1/rooms/{room_id}/members

Headers:
Authorization: Bearer <access_token>

Response:
{
  "members": [
    {
      "user_id": "@alice:example.com",
      "joined_at": 1234567890000
    },
    {
      "user_id": "@bob:example.com",
      "joined_at": 1234567891000
    }
  ]
}
```

### Get Room List

```
GET /_api/v1/rooms

Headers:
Authorization: Bearer <access_token>

Response:
{
  "rooms": [
    {
      "room_id": "!abc:example.com",
      "name": "My Group",
      "topic": "Discussion",
      "member_count": 5,
      "last_activity": 1234567890000
    }
  ]
}
```

## Room Messages

### Send Room Message

```
POST /_api/v1/rooms/{room_id}/messages

Headers:
Authorization: Bearer <access_token>

Request body:
{
  "content": {
    "msgtype": "m.text",
    "body": "Hello everyone"
  }
}

Response:
{
  "event_id": "evt_12345",
  "timestamp": 1234567890000
}
```

### Get Room History

```
GET /_api/v1/rooms/{room_id}/messages?limit={n}&before={token}

Headers:
Authorization: Bearer <access_token>

Parameters:
- limit: Number of events to return (default 50, max 100)
- before: Get events before this token (optional)

Response:
{
  "events": [
    {
      "event_id": "evt_100",
      "sender": "@alice:example.com",
      "event_type": "m.room.message",
      "content": {
        "msgtype": "m.text",
        "body": "Message text"
      },
      "timestamp": 1234567890000
    }
  ],
  "prev_token": "t_prev_token",
  "has_more": true
}
```

## Link Preview

### Get Link Preview

```
GET /_api/v1/link_preview?url={encoded_url}

Headers:
Authorization: Bearer <access_token>

Parameters:
- url: URL-encoded webpage address

Response:
{
  "url": "https://example.com",
  "title": "Example Website",
  "description": "This is an example website",
  "image": "https://example.com/image.jpg",
  "site_name": "Example"
}
```

**Features**:
- Automatically fetches Open Graph and Twitter Card metadata from webpages
- Returns basic information (URL as title) if metadata cannot be retrieved
- Used to display rich link preview cards in messages

## User Information

### Get User Profile

```
GET /_api/v1/user/{user_id}/profile

Response:
{
  "user_id": "@alice:example.com",
  "display_name": "Alice",
  "avatar_url": "https://example.com/avatar.jpg"
}
```

### Update User Profile

```
PUT /_api/v1/user/profile

Headers:
Authorization: Bearer <access_token>

Request body:
{
  "display_name": "Alice Smith",
  "avatar_url": "https://example.com/new_avatar.jpg"
}

Response:
{
  "success": true
}
```

## Presence

### Get Presence

```
GET /_api/v1/presence/{user_id}

Response:
{
  "user_id": "@alice:example.com",
  "presence": "online",
  "last_active": 1234567890000
}
```

## Error Responses

All APIs return a unified format on error:

```json
{
  "error": "ERROR_CODE",
  "message": "Human readable error message"
}
```

Common error codes:
- `UNAUTHORIZED`: Unauthorized or invalid token
- `FORBIDDEN`: No permission to perform operation
- `NOT_FOUND`: Resource not found
- `BAD_REQUEST`: Invalid request parameters
- `RATE_LIMITED`: Too many requests
- `SERVER_ERROR`: Internal server error

## Rate Limiting

To prevent abuse, APIs implement rate limiting:

- Login: 5 times per IP per minute
- Send message: 10 messages per user per second
- Create room: 10 rooms per user per hour
- Other APIs: 100 requests per user per second

Exceeding limits returns 429 status code:
```json
{
  "error": "RATE_LIMITED",
  "message": "Too many requests",
  "retry_after": 30
}
```

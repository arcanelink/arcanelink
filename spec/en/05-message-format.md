# Message Format Specification

## Message Type Overview

The protocol defines two major message categories:
1. Direct Messages
2. Room Events

## Direct Message Format

### Basic Structure

```json
{
  "msg_id": "msg_abc123",
  "sender": "@alice:example.com",
  "recipient": "@bob:example.com",
  "content": {
    "msgtype": "m.text",
    ...
  },
  "timestamp": 1234567890000
}
```

### Field Descriptions

- `msg_id`: Unique message identifier, format: `msg_` + random string
- `sender`: Sender user ID
- `recipient`: Recipient user ID
- `content`: Message content object
- `timestamp`: Unix timestamp (milliseconds)

### Text Message

```json
{
  "msgtype": "m.text",
  "body": "Hello, how are you?"
}
```

### Image Message

```json
{
  "msgtype": "m.image",
  "body": "image.jpg",
  "url": "https://example.com/files/image.jpg",
  "info": {
    "mimetype": "image/jpeg",
    "size": 102400,
    "w": 1920,
    "h": 1080
  }
}
```

### File Message

```json
{
  "msgtype": "m.file",
  "body": "document.pdf",
  "url": "https://example.com/files/document.pdf",
  "info": {
    "mimetype": "application/pdf",
    "size": 524288
  }
}
```

### Audio Message

```json
{
  "msgtype": "m.audio",
  "body": "voice.mp3",
  "url": "https://example.com/files/voice.mp3",
  "info": {
    "mimetype": "audio/mpeg",
    "size": 204800,
    "duration": 30000
  }
}
```

### Video Message

```json
{
  "msgtype": "m.video",
  "body": "video.mp4",
  "url": "https://example.com/files/video.mp4",
  "info": {
    "mimetype": "video/mp4",
    "size": 10485760,
    "duration": 60000,
    "w": 1920,
    "h": 1080
  }
}
```

### Location Message

```json
{
  "msgtype": "m.location",
  "body": "My Location",
  "geo_uri": "geo:37.7749,-122.4194",
  "info": {
    "latitude": 37.7749,
    "longitude": -122.4194
  }
}
```

## Room Event Format

### Basic Structure

```json
{
  "event_id": "evt_xyz789",
  "room_id": "!abc123:example.com",
  "sender": "@alice:example.com",
  "event_type": "m.room.message",
  "content": {
    ...
  },
  "timestamp": 1234567890000
}
```

### Field Descriptions

- `event_id`: Unique event identifier, format: `evt_` + random string
- `room_id`: Room ID
- `sender`: Sender user ID
- `event_type`: Event type
- `content`: Event content object
- `timestamp`: Unix timestamp (milliseconds)

## Room Event Types

### m.room.message - Room Message

Same content format as direct messages:

```json
{
  "event_type": "m.room.message",
  "content": {
    "msgtype": "m.text",
    "body": "Hello everyone"
  }
}
```

### m.room.create - Create Room

```json
{
  "event_type": "m.room.create",
  "content": {
    "creator": "@alice:example.com",
    "room_version": "1"
  }
}
```

### m.room.name - Set Room Name

```json
{
  "event_type": "m.room.name",
  "content": {
    "name": "My Group Chat"
  }
}
```

### m.room.topic - Set Room Topic

```json
{
  "event_type": "m.room.topic",
  "content": {
    "topic": "Discussion about the project"
  }
}
```

### m.room.member - Member Changes

Join:
```json
{
  "event_type": "m.room.member",
  "content": {
    "membership": "join",
    "user_id": "@bob:example.com"
  }
}
```

Leave:
```json
{
  "event_type": "m.room.member",
  "content": {
    "membership": "leave",
    "user_id": "@bob:example.com"
  }
}
```

Invite:
```json
{
  "event_type": "m.room.member",
  "content": {
    "membership": "invite",
    "user_id": "@carol:example.com",
    "inviter": "@alice:example.com"
  }
}
```

Kick:
```json
{
  "event_type": "m.room.member",
  "content": {
    "membership": "kick",
    "user_id": "@dave:example.com",
    "reason": "Violation of rules"
  }
}
```

## Extended Fields

### Reply Message

```json
{
  "msgtype": "m.text",
  "body": "That's a great idea!",
  "m.relates_to": {
    "rel_type": "m.reply",
    "event_id": "evt_original_message"
  }
}
```

### Edit Message

```json
{
  "msgtype": "m.text",
  "body": "Corrected message",
  "m.relates_to": {
    "rel_type": "m.replace",
    "event_id": "evt_original_message"
  }
}
```

### Delete Message

```json
{
  "msgtype": "m.text",
  "body": "Message deleted",
  "m.relates_to": {
    "rel_type": "m.delete",
    "event_id": "evt_message_to_delete"
  }
}
```

### Emoji Reaction

```json
{
  "msgtype": "m.reaction",
  "m.relates_to": {
    "rel_type": "m.annotation",
    "event_id": "evt_target_message",
    "key": "👍"
  }
}
```

## User Profile Format

```json
{
  "user_id": "@alice:example.com",
  "display_name": "Alice Smith",
  "avatar_url": "https://example.com/avatars/alice.jpg",
  "status_msg": "Working on the project"
}
```

## Room State Format

```json
{
  "room_id": "!abc123:example.com",
  "name": "My Group Chat",
  "topic": "Discussion about the project",
  "creator": "@alice:example.com",
  "created_at": 1234567890000,
  "member_count": 5,
  "avatar_url": "https://example.com/room_avatars/abc123.jpg"
}
```

## Presence Format

```json
{
  "user_id": "@alice:example.com",
  "presence": "online",
  "last_active": 1234567890000,
  "status_msg": "Available"
}
```

Presence values:
- `online`: Online
- `offline`: Offline
- `away`: Away
- `busy`: Busy

## Sync Response Format

```json
{
  "next_token": "t_12345_67890",
  "direct_messages": [
    {
      "msg_id": "msg_001",
      "sender": "@bob:example.com",
      "content": {...},
      "timestamp": 1234567890000
    }
  ],
  "room_events": [
    {
      "event_id": "evt_001",
      "room_id": "!abc:example.com",
      "sender": "@carol:example.com",
      "event_type": "m.room.message",
      "content": {...},
      "timestamp": 1234567891000
    }
  ],
  "presence_updates": [
    {
      "user_id": "@dave:example.com",
      "presence": "online",
      "last_active": 1234567892000
    }
  ]
}
```

## Data Types

### User ID

Format: `@username:domain.com`

Rules:
- Starts with `@`
- username can only contain lowercase letters, numbers, `.`, `_`, `-`
- domain is a valid domain name

Examples:
- `@alice:example.com` ✓
- `@bob_123:chat.example.org` ✓
- `@Alice:example.com` ✗ (uppercase letters)

### Room ID

Format: `!roomid:domain.com`

Rules:
- Starts with `!`
- roomid is a random string
- domain is the server domain where room was created

Example:
- `!abc123xyz:example.com` ✓

### Message ID

Format: `msg_` + random string

Example:
- `msg_a1b2c3d4e5f6` ✓

### Event ID

Format: `evt_` + random string

Example:
- `evt_x1y2z3a4b5c6` ✓

### Timestamp

Unix timestamp (milliseconds)

Example:
- `1234567890000`

### URL

Must be HTTPS URL (production environment)

Examples:
- `https://example.com/files/image.jpg` ✓
- `http://example.com/files/image.jpg` ✗ (HTTP not secure)

## Size Limits

- Text message body: Maximum 64KB
- File URL: Maximum 2KB
- User display_name: Maximum 100 characters
- Room name: Maximum 200 characters
- Room topic: Maximum 1000 characters
- Single event total size: Maximum 1MB

## Character Encoding

All text must use UTF-8 encoding.

## Extensibility

### Custom Message Types

Custom msgtype can be defined, recommend using namespace:

```json
{
  "msgtype": "com.example.custom",
  "body": "Fallback text",
  "com.example.data": {
    "custom_field": "value"
  }
}
```

### Custom Event Types

Custom event_type can be defined, recommend using namespace:

```json
{
  "event_type": "com.example.custom_event",
  "content": {
    "custom_data": "value"
  }
}
```

## Backward Compatibility

- Clients should ignore unknown msgtype and event_type
- Clients should ignore unknown fields
- New fields should have reasonable default values

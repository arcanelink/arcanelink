#!/bin/bash

# Run database migrations
echo "Running database migrations..."

docker exec -i mbot-postgres psql -U mbot -d mbot << 'EOF'
-- 001_create_users.up.sql
CREATE TABLE IF NOT EXISTS users (
    user_id VARCHAR(255) PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    display_name VARCHAR(100),
    avatar_url VARCHAR(500),
    status_msg VARCHAR(200)
);

CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

-- 002_create_direct_messages.up.sql
CREATE TABLE IF NOT EXISTS direct_messages (
    msg_id VARCHAR(100) PRIMARY KEY,
    sender VARCHAR(255) NOT NULL,
    recipient VARCHAR(255) NOT NULL,
    content JSONB NOT NULL,
    timestamp BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sender) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (recipient) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_direct_messages_recipient_time ON direct_messages(recipient, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_direct_messages_sender_time ON direct_messages(sender, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_direct_messages_timestamp ON direct_messages(timestamp DESC);

-- 003_create_rooms.up.sql
CREATE TABLE IF NOT EXISTS rooms (
    room_id VARCHAR(255) PRIMARY KEY,
    creator VARCHAR(255) NOT NULL,
    name VARCHAR(200),
    topic TEXT,
    avatar_url VARCHAR(500),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (creator) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_rooms_creator ON rooms(creator);

-- 004_create_room_members.up.sql
CREATE TABLE IF NOT EXISTS room_members (
    room_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (room_id, user_id),
    FOREIGN KEY (room_id) REFERENCES rooms(room_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_room_members_user ON room_members(user_id);
CREATE INDEX IF NOT EXISTS idx_room_members_room ON room_members(room_id);

-- 005_create_room_events.up.sql
CREATE TABLE IF NOT EXISTS room_events (
    event_id VARCHAR(100) PRIMARY KEY,
    room_id VARCHAR(255) NOT NULL,
    sender VARCHAR(255) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    content JSONB NOT NULL,
    timestamp BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (room_id) REFERENCES rooms(room_id) ON DELETE CASCADE,
    FOREIGN KEY (sender) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_room_events_room_time ON room_events(room_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_room_events_timestamp ON room_events(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_room_events_type ON room_events(event_type);

-- 006_create_message_queue.up.sql
CREATE TABLE IF NOT EXISTS message_queue (
    queue_id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    message_type VARCHAR(20) NOT NULL CHECK (message_type IN ('direct', 'room')),
    message_id VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    delivered BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_message_queue_user_created ON message_queue(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_message_queue_delivered ON message_queue(delivered, created_at);

-- 007_create_presence.up.sql
CREATE TABLE IF NOT EXISTS presence (
    user_id VARCHAR(255) PRIMARY KEY,
    status VARCHAR(20) NOT NULL DEFAULT 'offline' CHECK (status IN ('online', 'offline', 'away', 'busy')),
    last_active TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status_msg VARCHAR(200),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_presence_status ON presence(status);
CREATE INDEX IF NOT EXISTS idx_presence_last_active ON presence(last_active);

EOF

echo "Migrations completed!"

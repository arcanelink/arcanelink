CREATE TABLE room_events (
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

CREATE INDEX idx_room_events_room_time ON room_events(room_id, timestamp DESC);
CREATE INDEX idx_room_events_timestamp ON room_events(timestamp DESC);
CREATE INDEX idx_room_events_type ON room_events(event_type);

CREATE TABLE rooms (
    room_id VARCHAR(255) PRIMARY KEY,
    creator VARCHAR(255) NOT NULL,
    name VARCHAR(200),
    topic TEXT,
    avatar_url VARCHAR(500),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (creator) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE INDEX idx_rooms_creator ON rooms(creator);

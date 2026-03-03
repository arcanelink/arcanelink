CREATE TABLE direct_messages (
    msg_id VARCHAR(100) PRIMARY KEY,
    sender VARCHAR(255) NOT NULL,
    recipient VARCHAR(255) NOT NULL,
    content JSONB NOT NULL,
    timestamp BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sender) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (recipient) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE INDEX idx_direct_messages_recipient_time ON direct_messages(recipient, timestamp DESC);
CREATE INDEX idx_direct_messages_sender_time ON direct_messages(sender, timestamp DESC);
CREATE INDEX idx_direct_messages_timestamp ON direct_messages(timestamp DESC);

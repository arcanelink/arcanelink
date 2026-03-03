CREATE TABLE message_queue (
    queue_id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    message_type VARCHAR(20) NOT NULL CHECK (message_type IN ('direct', 'room')),
    message_id VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    delivered BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE INDEX idx_message_queue_user_created ON message_queue(user_id, created_at DESC);
CREATE INDEX idx_message_queue_delivered ON message_queue(delivered, created_at);

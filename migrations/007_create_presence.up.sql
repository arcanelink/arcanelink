CREATE TABLE presence (
    user_id VARCHAR(255) PRIMARY KEY,
    status VARCHAR(20) NOT NULL DEFAULT 'offline' CHECK (status IN ('online', 'offline', 'away', 'busy')),
    last_active TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status_msg VARCHAR(200),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE INDEX idx_presence_status ON presence(status);
CREATE INDEX idx_presence_last_active ON presence(last_active);

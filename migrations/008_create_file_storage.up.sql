-- Create file_storage table for storing uploaded files metadata
CREATE TABLE IF NOT EXISTS file_storage (
    file_id VARCHAR(255) PRIMARY KEY,
    uploader VARCHAR(255) NOT NULL,
    filename VARCHAR(512) NOT NULL,
    content_type VARCHAR(128) NOT NULL,
    file_size BIGINT NOT NULL,
    storage_path TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (uploader) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Create index for faster lookups
CREATE INDEX idx_file_storage_uploader ON file_storage(uploader);
CREATE INDEX idx_file_storage_created_at ON file_storage(created_at);

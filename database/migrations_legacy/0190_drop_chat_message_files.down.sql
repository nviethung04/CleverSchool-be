-- +migrate Down

-- Tạo lại bảng chat_message_files khi rollback
CREATE TABLE chat_message_files (
    id BIGSERIAL PRIMARY KEY,
    message_id BIGINT NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_url TEXT NOT NULL,
    file_type VARCHAR(50),
    file_size BIGINT,
    mime_type VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (message_id) REFERENCES chat_messages(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX idx_chat_message_files_message_id ON chat_message_files(message_id);

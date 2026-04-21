-- +migrate Up

-- Bảng lưu trữ trạng thái đã đọc tin nhắn chat
CREATE TABLE chat_message_reads (
    id BIGSERIAL PRIMARY KEY,
    message_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    read_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (message_id) REFERENCES chat_messages(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(message_id, user_id)
);

-- Indexes for performance
CREATE INDEX idx_chat_message_reads_message_id ON chat_message_reads(message_id);
CREATE INDEX idx_chat_message_reads_user_id ON chat_message_reads(user_id);
CREATE INDEX idx_chat_message_reads_message_user ON chat_message_reads(message_id, user_id);

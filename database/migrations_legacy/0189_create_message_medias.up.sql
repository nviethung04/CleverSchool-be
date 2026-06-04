-- +migrate Up

-- Bảng nối giữa chat_messages và medias để hỗ trợ nhiều file cho 1 tin nhắn
CREATE TABLE message_medias (
    id BIGSERIAL PRIMARY KEY,
    message_id BIGINT NOT NULL,
    media_id BIGINT NOT NULL,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (message_id) REFERENCES chat_messages(id) ON DELETE CASCADE,
    FOREIGN KEY (media_id) REFERENCES medias(id) ON DELETE CASCADE,
    
    -- Đảm bảo không trùng lặp (1 media chỉ thuộc 1 message)
    UNIQUE(message_id, media_id)
);

-- Indexes for performance
CREATE INDEX idx_message_medias_message_id ON message_medias(message_id);
CREATE INDEX idx_message_medias_media_id ON message_medias(media_id);
CREATE INDEX idx_message_medias_sort_order ON message_medias(message_id, sort_order);

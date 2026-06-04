-- Drop indexes first (nếu cần)
DROP INDEX IF EXISTS idx_chat_messages_reply_to;
DROP INDEX IF EXISTS idx_chat_messages_course_created;
DROP INDEX IF EXISTS idx_chat_messages_created_at;
DROP INDEX IF EXISTS idx_chat_messages_user_id;
DROP INDEX IF EXISTS idx_chat_messages_course_id;

-- Drop dependent tables first
DROP TABLE IF EXISTS chat_message_reads;
DROP TABLE IF EXISTS chat_message_reactions;
DROP TABLE IF EXISTS message_medias;

-- Now drop the parent table
DROP TABLE IF EXISTS chat_messages;


-- Create chat_messages table
CREATE TABLE chat_messages (
    id BIGSERIAL PRIMARY KEY,
    course_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    content TEXT,
    message_type VARCHAR(20) DEFAULT 'text',
    is_pinned BOOLEAN DEFAULT FALSE,
    is_edited BOOLEAN DEFAULT FALSE,
    edited_at TIMESTAMP,
    reply_to_message_id BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,

    CONSTRAINT chat_messages_message_type_check
      CHECK (message_type IN ('text', 'file', 'image', 'system')),

    FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (reply_to_message_id) REFERENCES chat_messages(id) ON DELETE SET NULL
);

-- Indexes for performance
CREATE INDEX idx_chat_messages_course_id ON chat_messages(course_id);
CREATE INDEX idx_chat_messages_user_id ON chat_messages(user_id);
CREATE INDEX idx_chat_messages_created_at ON chat_messages(created_at DESC);
CREATE INDEX idx_chat_messages_course_created ON chat_messages(course_id, created_at DESC);
CREATE INDEX idx_chat_messages_reply_to ON chat_messages(reply_to_message_id);

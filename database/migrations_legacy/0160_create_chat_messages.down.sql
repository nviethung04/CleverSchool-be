-- +migrate Down

DROP INDEX IF EXISTS idx_chat_messages_reply_to;
DROP INDEX IF EXISTS idx_chat_messages_course_created;
DROP INDEX IF EXISTS idx_chat_messages_created_at;
DROP INDEX IF EXISTS idx_chat_messages_user_id;
DROP INDEX IF EXISTS idx_chat_messages_course_id;

DROP TABLE IF EXISTS chat_messages;

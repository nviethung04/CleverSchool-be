-- +migrate Down

DROP INDEX IF EXISTS idx_chat_message_reads_message_user;
DROP INDEX IF EXISTS idx_chat_message_reads_user_id;
DROP INDEX IF EXISTS idx_chat_message_reads_message_id;

DROP TABLE IF EXISTS chat_message_reads;

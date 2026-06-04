-- +migrate Down

DROP INDEX IF EXISTS idx_chat_message_files_message_id;

DROP TABLE IF EXISTS chat_message_files;

-- Drop chat_message_reads table and its indexes
DROP INDEX IF EXISTS idx_chat_message_reads_read_at;
DROP INDEX IF EXISTS idx_chat_message_reads_user_id;
DROP INDEX IF EXISTS idx_chat_message_reads_message_id;
DROP TABLE IF EXISTS chat_message_reads;
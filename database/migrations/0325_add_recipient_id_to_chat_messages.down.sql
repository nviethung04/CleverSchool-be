DROP INDEX IF EXISTS idx_chat_messages_recipient_created;
DROP INDEX IF EXISTS idx_chat_messages_user_recipient;
DROP INDEX IF EXISTS idx_chat_messages_recipient_id;

ALTER TABLE chat_messages
DROP CONSTRAINT IF EXISTS fk_chat_messages_recipient;

ALTER TABLE chat_messages
DROP COLUMN IF EXISTS recipient_id;



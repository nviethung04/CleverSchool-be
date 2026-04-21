ALTER TABLE chat_messages 
ADD COLUMN recipient_id BIGINT;

CREATE INDEX idx_chat_messages_recipient_id ON chat_messages(recipient_id);

CREATE INDEX idx_chat_messages_user_recipient ON chat_messages(user_id, recipient_id) 
WHERE recipient_id IS NOT NULL;

CREATE INDEX idx_chat_messages_recipient_created ON chat_messages(recipient_id, created_at DESC) 
WHERE recipient_id IS NOT NULL;



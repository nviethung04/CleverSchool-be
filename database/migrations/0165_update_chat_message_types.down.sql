-- +migrate Down

-- Revert message_type constraint to original types
ALTER TABLE chat_messages 
DROP CONSTRAINT chat_messages_message_type_check;

ALTER TABLE chat_messages 
ADD CONSTRAINT chat_messages_message_type_check 
CHECK (message_type IN ('text', 'file', 'image', 'system'));

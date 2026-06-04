-- +migrate Down

DROP INDEX IF EXISTS idx_chat_message_reactions_user_id;
DROP INDEX IF EXISTS idx_chat_message_reactions_message_id;

DROP TABLE IF EXISTS chat_message_reactions;

-- +migrate Down

-- Xóa constraint mới
ALTER TABLE chat_message_reactions 
DROP CONSTRAINT IF EXISTS unique_user_message_emoji;

-- Thêm lại constraint cũ
ALTER TABLE chat_message_reactions 
ADD CONSTRAINT chat_message_reactions_message_id_user_id_emoji_key 
UNIQUE(message_id, user_id, emoji);

-- Xóa cột quantity
ALTER TABLE chat_message_reactions 
DROP COLUMN IF EXISTS quantity;

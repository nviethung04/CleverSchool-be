-- +migrate Up

-- Thêm cột quantity vào bảng chat_message_reactions
ALTER TABLE chat_message_reactions 
ADD COLUMN quantity INTEGER NOT NULL DEFAULT 1;

-- Cập nhật constraint unique để loại bỏ emoji vì giờ chúng ta cho phép nhiều reaction cùng emoji
-- Xóa constraint cũ
ALTER TABLE chat_message_reactions 
DROP CONSTRAINT IF EXISTS chat_message_reactions_message_id_user_id_emoji_key;

-- Thêm constraint mới: một user chỉ có một record cho mỗi emoji trên một message
-- Nhưng có thể có quantity > 1
ALTER TABLE chat_message_reactions 
ADD CONSTRAINT unique_user_message_emoji 
UNIQUE(message_id, user_id, emoji);

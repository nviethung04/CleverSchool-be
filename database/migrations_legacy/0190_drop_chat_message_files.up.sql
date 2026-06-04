-- +migrate Up

-- Drop bảng chat_message_files vì không còn sử dụng
-- Thay thế bằng bảng nối message_medias
DROP TABLE IF EXISTS chat_message_files;

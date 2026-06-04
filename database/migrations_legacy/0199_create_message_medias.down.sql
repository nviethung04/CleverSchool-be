-- Drop message_medias table and its indexes
DROP INDEX IF EXISTS idx_message_medias_sort_order;
DROP INDEX IF EXISTS idx_message_medias_media_id;
DROP INDEX IF EXISTS idx_message_medias_message_id;
DROP TABLE IF EXISTS message_medias;
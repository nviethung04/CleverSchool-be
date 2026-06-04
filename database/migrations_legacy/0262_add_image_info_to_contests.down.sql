-- +migrate Down
-- Remove image_info column from contests table
ALTER TABLE contests DROP COLUMN IF EXISTS image_info;
